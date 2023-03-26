package go_epoll

type EventType uint32

const (
	EventRead EventType = 1 << iota
	EventWrite
	EventClose
	EventError
	EventET
	EventOneShot
)

type Event struct {
	Fd        int       //表示文件描述符
	EventType EventType //表示事件类型，可读，可写
}

func (e *Event) IsRead() bool {
	if e.EventType&EventRead != 0 {
		return true
	}
	return false
}

func (e *Event) IsWrite() bool {
	if e.EventType&EventWrite != 0 {
		return true
	}
	return false
}

func (e *Event) IsClose() bool {
	if e.EventType&EventClose != 0 {
		return true
	}
	return false
}

func (e *Event) IsError() bool {
	if e.EventType&EventError != 0 {
		return true
	}
	return false
}
