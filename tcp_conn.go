package go_epoll

import (
	"bytes"
	"golang.org/x/sys/unix"
	"io"
)

type Conn struct {
	fd       int           //文件描述符
	addr     string        //地址
	server   *TcpServer    //服务器指针
	readBuf  *bytes.Buffer //从fd中读取的数据
	writeBuf *bytes.Buffer //从fd中写入的数据
}

func NewConn(fd int, addr string, server *TcpServer) *Conn {
	return &Conn{
		fd:       fd,
		addr:     addr,
		server:   server,
		readBuf:  server.bufPool.Get().(*bytes.Buffer),
		writeBuf: server.bufPool.Get().(*bytes.Buffer),
	}
}

// 读数据
func (c *Conn) Read(p []byte) (int, error) {
	return c.readBuf.Read(p)
}

// 写数据
func (c *Conn) Write(p []byte) (int, error) {
	n, err := c.writeBuf.Write(p)
	c.writeFromFD()
	return n, err
}

// 关闭
func (c *Conn) Close() error {
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

	return nil
}

func (c *Conn) readFromFD() {
	buf := make([]byte, 1024)
	for {
		n, err := unix.Read(c.fd, buf)
		if err != nil {
			// 内核中没有数据可读，ET模式，为了防止数据丢失，重新修改事件，让它可以再次触发
			if err == unix.EAGAIN {
				c.server.reactor.ModHandler(Event{
					Fd:        c.fd,
					EventType: EventRead | EventError | EventET,
				}, c.server.connEventHandler)
				break
			}
		}
		if n == 0 {
			//说明客户端已关闭
			break
		}
		if n > 0 {
			c.readBuf.Write(buf[:n])
		}
	}
	c.server.handler.OnData(c)
}

func (c *Conn) writeFromFD() {
	buf := make([]byte, 1024)
	for {
		ix, err := c.writeBuf.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		n, err := unix.Write(c.fd, buf[:ix])
		if err != nil {
			// 内核写缓冲区已满，下一次继续，为了防止数据丢失，重新修改事件，让它可以再次触发
			if err == unix.EAGAIN {
				c.server.reactor.ModHandler(Event{
					Fd:        c.fd,
					EventType: EventWrite | EventET,
				}, c.server.connEventHandler)
				break
			}
		}
		if n == 0 {
			break
		}
	}
}
