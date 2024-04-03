package pool

import "time"

type worker struct {
	Pool     *Pool
	task     chan func()
	lastTime time.Time //执行任务最后的时间
}

func (w *worker) run() {
	go w.running()
}

func (w *worker) running() {
	for f := range w.task {
		if f == nil {
			return
		}
		f()
		//执行完任务后，将worker放回pool
		w.Pool.putWorker(w)
		w.Pool.decRunning()
	}

}
