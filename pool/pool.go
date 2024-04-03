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

var ()

type Pool struct {
	workers []*worker     //空闲worker
	cap     int32         //容量
	running int32         //正在运行的worker数量
	expire  time.Duration //空闲worker过期时间
	release chan sig      //释放资源
	lock    sync.Mutex
	once    sync.Once //保证只执行一次
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
	return p, nil
}

func (p *Pool) Submit(task func()) error {
	//获取池子里的worker 然后执行
	w := p.GetWorker()
	w.task <- task
	w.Pool.inRunning()
	return nil
}

func (p *Pool) GetWorker() *worker {
	//1、获取pool里的worker
	//2、如果有空闲的worker 直接获取
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
	//3、如果没有空闲的worker,则创建一个新的worker
	if p.running < p.cap {
		//创建一个新的worker
		w := &worker{
			Pool: p,
			task: make(chan func(), 1),
		}
		w.run()
		return w
	}
	//5、如果worker数量 + 运行的worker数量大于容量，则阻塞等待worker释放
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
