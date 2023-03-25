package go_epoll

import "sync"

type EventTask struct {
	fn EventHandler
	ev *Event
}

func NewTask(fn EventHandler, ev *Event) *EventTask {
	return &EventTask{
		fn: fn,
		ev: ev,
	}
}

func (t *EventTask) Exec() {
	t.fn(t.ev)
}

type EventWorkPool struct {
	workCount int
	taskQueue chan *EventTask
	wg        sync.WaitGroup
	stop      chan struct{}
	onPanic   func(msg interface{})
}

func NewEventWorkPool(workCount int) *EventWorkPool {
	return &EventWorkPool{
		workCount: workCount,
		taskQueue: make(chan *EventTask),
		wg:        sync.WaitGroup{},
		stop:      make(chan struct{}),
	}
}

func (wp *EventWorkPool) Run() {
	wp.wg.Add(wp.workCount)

	for i := 0; i < wp.workCount; i++ {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					if wp.onPanic != nil {
						wp.onPanic(err)
					}
				}
			}()

			defer wp.wg.Done()

			for {
				select {
				case <-wp.stop:
					return
				case task, ok := <-wp.taskQueue:
					if !ok {
						return
					}
					task.Exec()
				}
			}
		}()
	}

	wp.wg.Wait()
}

func (wp *EventWorkPool) Close() {
	wp.stop <- struct{}{}
	close(wp.stop)
	close(wp.taskQueue)
}

func (wp *EventWorkPool) OnPanic(fn func(msg interface{})) {
	wp.onPanic = fn
}

func (wp *EventWorkPool) PushTask(t *EventTask) {
	wp.taskQueue <- t
}

func (wp *EventWorkPool) PushTaskFunc(fn EventHandler, ev *Event) {
	wp.taskQueue <- NewTask(fn, ev)
}
