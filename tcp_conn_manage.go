package go_epoll

import "sync"

type ConnManage struct {
	conns     map[int]*Conn
	connsLock sync.RWMutex
}

func NewConnManage() *ConnManage {
	return &ConnManage{
		conns:     make(map[int]*Conn),
		connsLock: sync.RWMutex{},
	}
}

func (cm *ConnManage) AddConn(conn *Conn) {
	cm.connsLock.Lock()
	defer cm.connsLock.Unlock()
	cm.conns[conn.fd] = conn
}

func (cm *ConnManage) DelConn(conn *Conn) {
	cm.connsLock.Lock()
	defer cm.connsLock.Unlock()
	delete(cm.conns, conn.fd)
}

func (cm *ConnManage) GetConn(fd int) (*Conn, bool) {
	cm.connsLock.RLock()
	defer cm.connsLock.RUnlock()
	conn, ok := cm.conns[fd]
	return conn, ok
}
