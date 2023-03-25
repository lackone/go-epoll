package go_epoll

import (
	"bytes"
	"golang.org/x/sys/unix"
	"sync"
)

type TcpServerHandler interface {
	OnConnect(conn *Conn)
	OnData(conn *Conn)
	OnClose(conn *Conn)
}

type TcpServer struct {
	addr       string
	fd         int
	reactor    *Reactor
	handler    TcpServerHandler
	connManage *ConnManage
	bufPool    *sync.Pool
}

func NewServer(addr string, dType EventDemultiplexerType, dSize int, eventSize int, workCount int) (*TcpServer, error) {
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
	}

	s.reactor, err = NewReactor(dType, dSize, eventSize, workCount)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// 监听
func (s *TcpServer) Listen() error {
	var err error
	// 创建监听socket
	s.fd, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.O_NONBLOCK, 0)
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

	//新来的连接，往反应堆里添加读事件，注意这里使用ET模式
	err = s.reactor.AddHandler(Event{
		Fd:        nfd,
		EventType: EventRead | EventError | EventET,
	}, s.connEventHandler)

	if err != nil {
		return
	}

	//转换成IP字符串
	addr = GetIPBySockAddr(sa)

	conn := NewConn(nfd, addr, s)
	s.connManage.AddConn(conn)

	//处理连接回调
	s.handler.OnConnect(conn)

	return
}

// 连接事件处理
func (s *TcpServer) connEventHandler(ev *Event) {
	conn, ok := s.connManage.GetConn(ev.Fd)
	if !ok {
		return
	}
	if ev.IsError() {
		s.handler.OnClose(conn)
		conn.Close()
		return
	}
	if ev.IsRead() {
		conn.readFromFD()
	}
	if ev.IsWrite() {
		conn.writeFromFD()
	}
}

// 使Socket重用地址
func (s *TcpServer) ReuseAddr(fd int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
}

// 使Socket重用端口
func (s *TcpServer) ReusePort(fd int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
}
