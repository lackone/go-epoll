package go_epoll

import (
	"errors"
	"sync"
)

type Reactor struct {
	demultiplexer     map[int]EventDemultiplexer   //多路复用器
	demultiplexerSize int                          //多路复用器数量
	handlers          map[int]map[int]EventHandler //事件handler
	handlersLock      sync.RWMutex                 //handler锁
	wg                sync.WaitGroup               //等待组
}

func NewReactor(demultiplexerType EventDemultiplexerType, demultiplexerSize int, eventSize int) (*Reactor, error) {
	if demultiplexerSize <= 0 {
		return nil, errors.New("demultiplexerSize >= 1")
	}

	demultiplexer := make(map[int]EventDemultiplexer)
	for i := 0; i < demultiplexerSize; i++ {
		d, err := NewEventDemultiplexer(demultiplexerType, eventSize)
		if err != nil {
			return nil, err
		}
		demultiplexer[i] = d
	}

	return &Reactor{
		demultiplexer:     demultiplexer,
		demultiplexerSize: demultiplexerSize,
		handlers:          make(map[int]map[int]EventHandler),
		handlersLock:      sync.RWMutex{},
		wg:                sync.WaitGroup{},
	}, nil
}

// 获取fd最终分配到哪个复用器下面
func (r *Reactor) getIndex(ev Event) int {
	return ev.Fd % r.demultiplexerSize
}

// 添加事件handler
func (r *Reactor) AddHandler(ev Event, fn EventHandler) error {
	r.handlersLock.Lock()
	defer r.handlersLock.Unlock()

	index := r.getIndex(ev)

	if r.handlers[index] == nil {
		r.handlers[index] = make(map[int]EventHandler)
	}
	r.handlers[index][ev.Fd] = fn

	return r.demultiplexer[index].AddEvent(ev)
}

// 删除事件handler
func (r *Reactor) DelHandler(ev Event) error {
	r.handlersLock.Lock()
	defer r.handlersLock.Unlock()

	index := r.getIndex(ev)

	if _, ok := r.handlers[index]; !ok {
		return errors.New("no handler")
	}

	delete(r.handlers[index], ev.Fd)

	return r.demultiplexer[index].DelEvent(ev)
}

// 修改事件handler
func (r *Reactor) ModHandler(ev Event, fn EventHandler) error {
	r.handlersLock.Lock()
	defer r.handlersLock.Unlock()

	index := r.getIndex(ev)

	if r.handlers[index] == nil {
		r.handlers[index] = make(map[int]EventHandler)
	}
	r.handlers[index][ev.Fd] = fn

	return r.demultiplexer[index].ModEvent(ev)
}

// 运行，等待事件发，并调用handler
func (r *Reactor) Run() {
	defer r.Close()

	r.wg.Add(r.demultiplexerSize)

	for _, d := range r.demultiplexer {
		go func(d EventDemultiplexer) {
			defer r.wg.Done()

			for {
				events, err := d.Wait()
				if err != nil {
					return
				}
				for _, ev := range events {
					go r.handlers[r.getIndex(*ev)][ev.Fd](ev)
				}
			}
		}(d)
	}

	r.wg.Wait()
}

// 关闭
func (r *Reactor) Close() {
	for _, d := range r.demultiplexer {
		d.Close()
	}
}
