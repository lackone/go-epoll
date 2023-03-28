package go_epoll

import (
	"bytes"
	"context"
	"golang.org/x/sys/unix"
	"io"
	"sync"
	"sync/atomic"
)

type Conn struct {
	fd       int           //文件描述符
	addr     string        //地址
	isClose  int32         //0正常，1关闭
	server   *TcpServer    //服务器指针
	rbuf     []byte        //读缓冲
	wbuf     []byte        //写缓冲
	readBuf  *bytes.Buffer //从fd中读取的数据
	writeBuf *bytes.Buffer //从fd中写入的数据
	rLock    *sync.Mutex   //读锁
	wLock    *sync.Mutex   //写锁
	ext      interface{}   //扩展数据
}

func NewConn(fd int, addr string, s *TcpServer) (*Conn, error) {
	conn := &Conn{
		fd:       fd,
		addr:     addr,
		isClose:  0,
		server:   s,
		rbuf:     make([]byte, 1024),
		wbuf:     make([]byte, 1024),
		readBuf:  s.bufPool.Get().(*bytes.Buffer),
		writeBuf: s.bufPool.Get().(*bytes.Buffer),
		rLock:    &sync.Mutex{},
		wLock:    &sync.Mutex{},
	}

	//新来的连接，往反应堆里添加读事件，注意这里使用ET模式
	err := s.reactor.AddHandler(Event{
		Fd:        fd,
		EventType: EventRead | EventError | EventET | EventOneShot,
	}, conn.eventHandle)

	if err != nil {
		logger.Error(context.Background(), "reactor AddHandler error : ", err.Error())
		return nil, err
	}

	//处理连接回调
	s.handler.OnConnect(conn)

	return conn, nil
}

// 获取文件描述符
func (c *Conn) GetFD() int {
	return c.fd
}

// 获取地址
func (c *Conn) GetAddr() string {
	return c.addr
}

// 设置扩展数据
func (c *Conn) SetExt(ext interface{}) {
	c.ext = ext
}

// 获取扩展数据
func (c *Conn) GetExt() interface{} {
	return c.ext
}

// 读数据
func (c *Conn) Read(p []byte) (int, error) {
	c.rLock.Lock()
	defer c.rLock.Unlock()

	return c.readBuf.Read(p)
}

// 写数据
func (c *Conn) Write(p []byte) (int, error) {
	c.wLock.Lock()
	defer c.wLock.Unlock()

	if c.server.endecoder != nil {
		encode, err := c.server.endecoder.Encode(p)
		if err != nil {
			return 0, err
		}
		p = encode
	}

	n, err := c.writeBuf.Write(p)

	c.eventHandleWrite()

	return n, err
}

// 关闭
func (c *Conn) Close() error {
	if atomic.CompareAndSwapInt32(&c.isClose, 0, 1) {
		//调用关闭回调函数
		c.server.handler.OnClose(c)

		//称除事件
		c.server.reactor.DelHandler(Event{Fd: c.fd})

		//关闭文件描述符
		unix.Close(c.fd)

		//删除连接
		c.server.connManage.DelConn(c)

		//归还buf到池中
		c.readBuf.Reset()
		c.server.bufPool.Put(c.readBuf)

		c.writeBuf.Reset()
		c.server.bufPool.Put(c.writeBuf)
	}
	return nil
}

// 事件处理
func (c *Conn) eventHandle(ev *Event) {
	//关闭
	if ev.IsClose() {
		c.Close()
		return
	}
	//出错
	if ev.IsError() {
		c.eventHandleError()
		c.Close()
		return
	}
	//可读
	if ev.IsRead() {
		c.eventHandleRead()
	}
	//可写
	if ev.IsWrite() {
		c.eventHandleWrite()
	}
}

// 出错事件处理
func (c *Conn) eventHandleError() {
	c.server.handler.OnError(c)
}

// epoll在ET模式下时，对于读操作，如果read一次没有读尽内核缓冲中的数据，那么下次将得不到读就绪的通知，造成内核缓冲中已有的数据无机会读出，除非有新的数据再次到达。
// 对于读操作，如果读缓冲区空了，对于阻塞socket，读操作将阻塞住。对于非阻塞socket，读操作将立即返回-1，同时errno设置为EAGAIN
// 所以在ET模式下，只要可读，就一直读，直到返回0，或者errno=EAGAIN
func (c *Conn) eventHandleRead() {
	for {
		//阻塞与非阻塞read返回值没有区分，都是 <0表示出错，=0表示连接关闭，>0表示接收到数据大小
		//非阻塞模式下返回值如果 <0时并且(errno == EINTR || errno == EWOULDBLOCK || errno == EAGAIN)的情况下认为连接是正常的，可以继续接收。
		n, err := unix.Read(c.fd, c.rbuf)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			// 内核中没有数据可读，为了防止数据丢失，重新注册事件，尝试再次读
			if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
				if e := c.server.reactor.ModHandler(Event{
					Fd:        c.fd,
					EventType: EventRead | EventError | EventET | EventOneShot,
				}, c.eventHandle); e != nil {
					logger.Error(context.Background(), "ModHandler read error : ", err.Error())
				}
			}
			if err != unix.EAGAIN && err != unix.EWOULDBLOCK {
				logger.Error(context.Background(), "eventHandleRead error : ", err.Error())
			}
			break
		}
		if n == 0 {
			//说明客户端已关闭
			c.Close()
			return
		}
		if n > 0 {
			//把从fd中读到的数据，写入我们自已的读buf中
			c.readBuf.Write(c.rbuf[:n])

			if c.server.endecoder == nil {
				//如果没有设置编解码，则直接把buf中的数据全部取出，然后reset
				c.server.handler.OnData(c, c.readBuf.Bytes())
				c.readBuf.Reset()
			} else {
				//如果设置了编解码，for循环解码，直到IO.EOF
				for {
					decode, err := c.server.endecoder.Decode(c.readBuf)
					if err != nil {
						if err != io.EOF {
							logger.Error(context.Background(), "Decode error : ", err.Error())
						}
						break
					}
					c.server.handler.OnData(c, decode)
				}
			}
		}
	}
}

// 对于写操作，如果写缓冲区满了，对于阻塞socket，写操作将阻塞住。对于非阻塞socket，写操作将立即返回-1，同时errno设置为EAGAIN
// 所以这个时候，在ET模式下，就需要你重新注册事件，尽量把数据写尽。
// 所以在ET模式下，只要可写，就一直写，直到数据发完，或者errno=EAGAIN
func (c *Conn) eventHandleWrite() {
	for {
		ix, err := c.writeBuf.Read(c.wbuf)
		if err != nil {
			if err == io.EOF {
				//我们自已的数据已经写完了，退出循环
				break
			}
		}
		//阻塞与非阻塞write返回值没有区分，都是 <0表示出错，=0表示连接关闭，>0表示发送数据大小
		//非阻塞模式下返回值 <0时并且 (errno == EINTR || errno == EWOULDBLOCK || errno == EAGAIN)的情况下认为连接是正常的，可以继续发送。
		n, err := unix.Write(c.fd, c.wbuf[:ix])
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			// 内核写缓冲区已满，为了防止数据丢失，重新注册事件，尝试再次写
			if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
				if e := c.server.reactor.ModHandler(Event{
					Fd:        c.fd,
					EventType: EventWrite | EventET | EventOneShot,
				}, c.eventHandle); e != nil {
					logger.Error(context.Background(), "ModHandler write error : ", e.Error())
				}
			}
			if err != unix.EAGAIN && err != unix.EWOULDBLOCK {
				logger.Error(context.Background(), "eventHandleWrite error : ", err.Error())
			}
			break
		}
		if n == 0 {
			//说明客户端已关闭
			c.Close()
			return
		}
	}
}
