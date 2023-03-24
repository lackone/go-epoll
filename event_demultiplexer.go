package go_epoll

type EventDemultiplexerType uint32

const (
	EpollType EventDemultiplexerType = iota + 1
)

// 事件多路复用器
type EventDemultiplexer interface {
	//添加事件
	AddEvent(ev Event) error
	//删除事件
	DelEvent(ev Event) error
	//修改事件
	ModEvent(ev Event) error
	//等待事件，并返回已经触发的事件
	Wait() ([]*Event, error)
	//关闭
	Close() error
}

// 创建多路复用器
func NewEventDemultiplexer(t EventDemultiplexerType, eventSize int) (EventDemultiplexer, error) {
	switch t {
	case EpollType:
		return NewEpoll(eventSize)
	default:
		return nil, DemultiplexerTypeUnknown
	}
}
