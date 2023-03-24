package go_epoll

import "sync"

type TaskHandler func(args ...interface{})

type Task struct {
	fn   TaskHandler
	args []interface{}
}

func NewTask(fn TaskHandler, args ...interface{}) *Task {
	return &Task{
		fn:   fn,
		args: args,
	}
}

func (t *Task) Exec() {
	t.fn(t.args...)
}

type WorkPool struct {
	workCount int
	taskQueue chan *Task
	wg        sync.WaitGroup
	stop      chan struct{}
	onPanic   func(msg interface{})
}

func NewWorkPool(workCount int) *WorkPool {
	return &WorkPool{
		workCount: workCount,
		taskQueue: make(chan *Task),
		wg:        sync.WaitGroup{},
		stop:      make(chan struct{}),
	}
}

func (wp *WorkPool) Run() {
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

func (wp *WorkPool) Close() {
	wp.stop <- struct{}{}
	close(wp.stop)
	close(wp.taskQueue)
}

func (wp *WorkPool) OnPanic(fn func(msg interface{})) {
	wp.onPanic = fn
}

func (wp *WorkPool) PushTask(t *Task) {
	wp.taskQueue <- t
}

func (wp *WorkPool) PushTaskFunc(fn TaskHandler, args ...interface{}) {
	wp.taskQueue <- NewTask(fn, args...)
}
