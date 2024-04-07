package pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type sig struct {
}

const DefaultExpire = 3

type Pool struct {
	workers     []*worker     //空闲worker
	cap         int32         //容量
	running     int32         //正在运行的worker数量
	expire      time.Duration //空闲worker过期时间
	release     chan sig      //释放资源
	lock        sync.Mutex
	once        sync.Once //保证只执行一次
	workerCache sync.Pool
}

func NewPool(cap int) (*Pool, error) {
	return NewTimePool(cap, DefaultExpire)
}

func NewTimePool(cap int, expire int) (*Pool, error) {
	if cap <= 0 {
		return nil, errors.New("cat cat not be less than 0")
	}
	if expire <= 0 {
		return nil, errors.New("expire time not be less than 0")
	}
	p := &Pool{
		cap:     int32(cap),
		expire:  time.Duration(expire) * time.Second,
		release: make(chan sig, 1),
	}
	// 启动一个协程定时清理过期的worker
	go p.expireWorker()
	return p, nil
}

// 清理过期的协程
func (p *Pool) expireWorker() {
	ticker := time.NewTicker(p.expire)
	for range ticker.C {
		currentTime := time.Now()
		if len(p.release) > 0 {
			break
		}
		p.lock.Lock()
		idleWorkers := p.workers
		n := -1
		for i, w := range idleWorkers {
			if currentTime.Sub(w.lastTime) <= p.expire {
				break
			}
			//需要清除的
			n = i
			w.task <- nil
			idleWorkers[i] = nil
		}
		if n > -1 {
			if n >= len(idleWorkers)-1 {
				p.workers = idleWorkers[:0]
			} else {
				p.workers = idleWorkers[n+1:]
			}
		}
		p.lock.Unlock()

	}
}

func (p *Pool) Submit(task func()) error {
	if len(p.release) > 0 {
		return errors.New("pool has bean released")
	}
	w := p.GetWorker()
	w.task <- task
	w.Pool.inRunning()
	return nil
}

func (p *Pool) GetWorker() *worker {

	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n >= 0 {
		p.lock.Lock()
		w := idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
		return w
	}

	if p.running < p.cap {
		//创建一个新的worker
		w := &worker{
			Pool: p,
			task: make(chan func(), 1),
		}
		w.run()
		return w
	}

	for {
		p.lock.Lock()
		idleWorkers := p.workers
		n := len(idleWorkers) - 1
		if n < 0 {
			p.lock.Unlock()
			continue
		}
		w := idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
		return w
	}
}

func (p *Pool) inRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) putWorker(w *worker) {
	w.lastTime = time.Now()
	p.lock.Lock()
	p.workers = append(p.workers, w)
	p.lock.Unlock()
}

func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *Pool) Release() {
	p.once.Do(func() {
		//只执行一次
		p.lock.Lock()
		workers := p.workers
		for i, w := range workers {
			w.task = nil
			w.Pool = nil
			workers[i] = nil
		}
		p.workers = nil
		p.lock.Unlock()
		p.release <- sig{}
	})
}

func (p *Pool) IsClosed() bool {

	return len(p.release) > 0
}

func (p *Pool) Restart() bool {
	if len(p.release) <= 0 {
		return true
	}
	_ = <-p.release
	return true
}
