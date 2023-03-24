package go_epoll

import (
	"golang.org/x/sys/unix"
)

type TcpServer struct {
	addr    string
	fd      int
	reactor *Reactor
}

func NewServer(addr string, dType EventDemultiplexerType, dSize int, eventSize int, workCount int) (*TcpServer, error) {
	var err error

	s := &TcpServer{
		addr: addr,
	}

	s.reactor, err = NewReactor(dType, dSize, eventSize, workCount)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// 监听
func (s *TcpServer) listen() error {
	var err error
	// 创建监听socket
	s.fd, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.O_NONBLOCK, unix.IPPROTO_TCP)
	if err != nil {
		return err
	}
	// 重用socket地址
	if err = s.reuseAddr(s.fd); err != nil {
		return err
	}
	// 重用socket端口
	if err = s.reusePort(s.fd); err != nil {
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

func (s *TcpServer) connEventHandler(ev *Event) {

}

// 接收连接
func (s *TcpServer) accept() (nfd int, addr string, err error) {
	nfd, sa, err := unix.Accept(s.fd)
	if err != nil {
		return
	}
	err = unix.SetNonblock(nfd, true)
	if err != nil {
		return
	}

	//新来的连接，往反应堆里添加读事件
	err = s.reactor.AddHandler(Event{
		Fd:        nfd,
		EventType: EventRead | EventError,
	}, s.connEventHandler)
	if err != nil {
		return
	}

	//转换成IP字符串
	addr = GetIPBySockAddr(sa)

	return
}

// 使Socket重用地址
func (s *TcpServer) reuseAddr(fd int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
}

// 使Socket重用端口
func (s *TcpServer) reusePort(fd int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
}
