package go_epoll

import (
	"context"
	"sync"
	"sync/atomic"
)

type Reactor struct {
	demultiplexer     map[int]EventDemultiplexer   //多路复用器
	demultiplexerSize int                          //多路复用器数量
	handlers          map[int]map[int]EventHandler //事件handler
	handlersLock      sync.RWMutex                 //handler锁
	wg                sync.WaitGroup               //等待组
	totalEventNums    int32                        //监控的Event数量
	eventWorkPool     *EventWorkPool               //事件工作池
	stop              chan struct{}
}

func NewReactor(dType EventDemultiplexerType, dSize int, eventSize int, workCount int) (*Reactor, error) {
	if dSize <= 0 {
		return nil, DemultiplexerSizeError
	}

	demultiplexer := make(map[int]EventDemultiplexer)
	for i := 0; i < dSize; i++ {
		d, err := NewEventDemultiplexer(dType, eventSize)
		if err != nil {
			logger.Error(context.Background(), "NewEventDemultiplexer error : ", err.Error())
			return nil, err
		}
		demultiplexer[i] = d
	}

	return &Reactor{
		demultiplexer:     demultiplexer,
		demultiplexerSize: dSize,
		handlers:          make(map[int]map[int]EventHandler),
		handlersLock:      sync.RWMutex{},
		wg:                sync.WaitGroup{},
		totalEventNums:    0,
		eventWorkPool:     NewEventWorkPool(workCount),
		stop:              make(chan struct{}),
	}, nil
}

// 获取event最终分配到哪个复用器下面
func (r *Reactor) GetIndex(ev Event) int {
	return ev.Fd % r.demultiplexerSize
}

// 添加事件handler
func (r *Reactor) AddHandler(ev Event, handler EventHandler) error {
	r.handlersLock.Lock()
	defer r.handlersLock.Unlock()

	index := r.GetIndex(ev)

	if r.handlers[index] == nil {
		r.handlers[index] = make(map[int]EventHandler)
	}
	r.handlers[index][ev.Fd] = handler

	err := r.demultiplexer[index].AddEvent(ev)
	if err == nil {
		atomic.AddInt32(&r.totalEventNums, 1)
	}

	return err
}

// 删除事件handler
func (r *Reactor) DelHandler(ev Event) error {
	r.handlersLock.Lock()
	defer r.handlersLock.Unlock()

	index := r.GetIndex(ev)

	if _, ok := r.handlers[index]; !ok {
		return EventHandlerNotFound
	}

	delete(r.handlers[index], ev.Fd)

	err := r.demultiplexer[index].DelEvent(ev)
	if err == nil {
		atomic.AddInt32(&r.totalEventNums, -1)
	}

	return err
}

// 修改事件handler
func (r *Reactor) ModHandler(ev Event, handler EventHandler) error {
	r.handlersLock.Lock()
	defer r.handlersLock.Unlock()

	index := r.GetIndex(ev)

	if r.handlers[index] == nil {
		r.handlers[index] = make(map[int]EventHandler)
	}
	r.handlers[index][ev.Fd] = handler

	return r.demultiplexer[index].ModEvent(ev)
}

// 运行，等待事件发，并调用handler
func (r *Reactor) Run() {
	go r.eventWorkPool.Run()

	r.wg.Add(r.demultiplexerSize)

	for _, d := range r.demultiplexer {
		go func(d EventDemultiplexer) {
			defer r.wg.Done()

			for {
				select {
				case <-r.stop:
					return
				default:
					//等待事件触发
					events, err := d.Wait()
					if err != nil {
						logger.Error(context.Background(), "wait error : ", err.Error())
						return
					}
					for _, ev := range events {
						//把事件压入工作池中执行
						r.eventWorkPool.PushTaskFunc(r.handlers[r.GetIndex(*ev)][ev.Fd], ev)
					}
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

	r.stop <- struct{}{}
	close(r.stop)

	r.eventWorkPool.Close()
}
