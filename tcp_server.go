package go_epoll

import (
	"bytes"
	"context"
	"golang.org/x/sys/unix"
	"sync"
)

type TcpServerHandler interface {
	OnConnect(conn *Conn)
	OnData(conn *Conn)
	OnError(conn *Conn)
	OnClose(conn *Conn)
}

type TcpServer struct {
	addr       string           //地址
	fd         int              //文件描述符
	reactor    *Reactor         //多路复用反应堆
	handler    TcpServerHandler //回调函数
	connManage *ConnManage      //连接管理
	bufPool    *sync.Pool       //缓冲池，用于连接的读与写
	stop       chan struct{}    //关闭通道
}

func NewTcpServer(addr string, dType EventDemultiplexerType, dSize int, eventSize int, workCount int) (*TcpServer, error) {
	var err error

	s := &TcpServer{
		addr:       addr,
		connManage: NewConnManage(),
		bufPool: &sync.Pool{
			New: func() any {
				b := make([]byte, 0, 1024)
				return bytes.NewBuffer(b)
			},
		},
		stop: make(chan struct{}),
	}

	s.reactor, err = NewReactor(dType, dSize, eventSize, workCount)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// 设置回调函数
func (s *TcpServer) SetHandler(handler TcpServerHandler) {
	s.handler = handler
}

// 监听
func (s *TcpServer) Listen() error {
	var err error
	// 创建监听socket
	s.fd, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	// 重用socket地址
	if err = s.ReuseAddr(s.fd); err != nil {
		return err
	}
	// 重用socket端口
	if err = s.ReusePort(s.fd); err != nil {
		return err
	}
	inet4, err := GetSockAddrInet4(s.addr)
	if err != nil {
		return err
	}
	// 绑定地址
	if err = unix.Bind(s.fd, inet4); err != nil {
		return err
	}
	// 监听
	if err = unix.Listen(s.fd, 1024); err != nil {
		return err
	}
	return nil
}

// 接收连接
func (s *TcpServer) Accept() (nfd int, addr string, err error) {
	nfd, sa, err := unix.Accept(s.fd)
	if err != nil {
		return
	}
	err = unix.SetNonblock(nfd, true)
	if err != nil {
		return
	}

	//转换成IP字符串
	addr = GetIPBySockAddr(sa)

	//创建连接
	conn, err := NewConn(nfd, addr, s)
	if err != nil {
		return
	}

	//添加连接
	s.connManage.AddConn(conn)

	return
}

// 运行
func (s *TcpServer) Run() error {
	err := s.Listen()
	if err != nil {
		return err
	}

	logger.Infof(context.Background(), "server[%s] run ...", s.addr)

	go s.reactor.Run()

	for {
		select {
		case <-s.stop:
			return nil
		default:
			_, _, err = s.Accept()
			if err != nil {
				if err == unix.EINTR {
					continue
				}
				logger.Error(context.Background(), "Accept error : ", err.Error())
			}
		}
	}
}

// 关闭
func (s *TcpServer) Close() {
	s.stop <- struct{}{}
	close(s.stop)

	unix.Close(s.fd)

	s.connManage.Close()

	s.reactor.Close()
}

// 使Socket重用地址
func (s *TcpServer) ReuseAddr(fd int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
}

// 使Socket重用端口
func (s *TcpServer) ReusePort(fd int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
}
