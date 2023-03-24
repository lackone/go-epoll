package go_epoll

import (
	"golang.org/x/sys/unix"
)

type Epoll struct {
	epollFD int
	events  []unix.EpollEvent
}

// 创建epoll
func NewEpoll(eventSize int) (*Epoll, error) {
	fd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}
	return &Epoll{
		epollFD: fd,
		events:  make([]unix.EpollEvent, eventSize),
	}, nil
}

// 添加事件
func (e *Epoll) AddEvent(ev Event) error {
	return unix.EpollCtl(e.epollFD, unix.EPOLL_CTL_ADD, ev.Fd, eventToEpollEvent(ev))
}

// 删除事件
func (e *Epoll) DelEvent(ev Event) error {
	return unix.EpollCtl(e.epollFD, unix.EPOLL_CTL_DEL, ev.Fd, nil)
}

// 修改事件
func (e *Epoll) ModEvent(ev Event) error {
	return unix.EpollCtl(e.epollFD, unix.EPOLL_CTL_MOD, ev.Fd, eventToEpollEvent(ev))
}

// 等待事件触发，并返回发生的事件
func (e *Epoll) Wait() ([]*Event, error) {
retry:
	n, err := unix.EpollWait(e.epollFD, e.events, -1)
	if err != nil {
		if err == unix.EINTR {
			goto retry
		}
		return nil, err
	}
	evs := make([]*Event, 0)
	for i := 0; i < n; i++ {
		evs = append(evs, epollEventToEvent(e.events[i]))
	}
	return evs, nil
}

// 关闭
func (e *Epoll) Close() error {
	return unix.Close(e.epollFD)
}

// 将自已的事件转换成epoll事件
func eventToEpollEvent(ev Event) *unix.EpollEvent {
	epEv := unix.EpollEvent{}
	epEv.Fd = int32(ev.Fd)

	if ev.EventType&EventRead != 0 {
		epEv.Events |= unix.EPOLLIN | unix.EPOLLPRI
	}
	if ev.EventType&EventWrite != 0 {
		epEv.Events |= unix.EPOLLOUT
	}
	if ev.EventType&EventError != 0 {
		epEv.Events |= unix.EPOLLERR | unix.EPOLLHUP | unix.EPOLLRDHUP
	}
	if ev.EventType&EventET != 0 {
		epEv.Events |= unix.EPOLLET
	}

	return &epEv
}

//EPOLLIN ：表示对应的文件描述符可以读（包括对端SOCKET正常关闭）；
//EPOLLOUT：表示对应的文件描述符可以写；
//EPOLLPRI：表示对应的文件描述符有紧急的数据可读（这里应该表示有带外数据到来）；
//EPOLLERR：表示对应的文件描述符发生错误；
//EPOLLHUP：表示对应的文件描述符被挂断；
//EPOLLRDHUP：表示对端Socket关闭连接（半关闭）或者半关闭的状态；
//EPOLLET： 将EPOLL设为边缘触发(Edge Triggered)模式，这是相对于水平触发(Level Triggered)来说的。
//EPOLLONESHOT：只监听一次事件，当监听完这次事件之后，如果还需要继续监听这个socket的话，需要再次把这个socket加入到EPOLL队列里

// 将epoll事件转换成自已的事件
func epollEventToEvent(epEv unix.EpollEvent) *Event {
	ev := Event{}
	ev.Fd = int(epEv.Fd)

	// 可读事件
	if epEv.Events&(unix.EPOLLIN|unix.EPOLLPRI) != 0 {
		ev.EventType |= EventRead
	}
	// 可写事件
	if epEv.Events&(unix.EPOLLOUT) != 0 {
		ev.EventType |= EventWrite
	}
	// 出错
	if epEv.Events&(unix.EPOLLERR|unix.EPOLLHUP|unix.EPOLLRDHUP) != 0 {
		ev.EventType |= EventError
	}

	return &ev
}
