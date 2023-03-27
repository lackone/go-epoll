package go_epoll

import "strings"

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

func (et EventType) String() string {
	str := make([]string, 0)
	if et&EventRead != 0 {
		str = append(str, "read")
	}
	if et&EventWrite != 0 {
		str = append(str, "write")
	}
	if et&EventClose != 0 {
		str = append(str, "close")
	}
	if et&EventError != 0 {
		str = append(str, "error")
	}
	return strings.Join(str, ",")
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
