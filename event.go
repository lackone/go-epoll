package go_epoll

type EventType uint32

const (
	EventRead EventType = 1 << iota
	EventWrite
	EventClose
	EventError
)

type Event struct {
	Fd        int       //表示文件描述符
	EventType EventType //表示事件类型，可读，可写
}
