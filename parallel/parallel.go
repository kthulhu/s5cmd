package parallel

import (
	"runtime"
	"sync"
)

const minNumWorkers = 2

var global *Manager

type Task func() error

func Init(workercount int) {
	global = New(workercount)
}

func Close() { global.Close() }

func Run(task Task, waiter *Waiter) { global.Run(task, waiter) }

type Manager struct {
	wg        *sync.WaitGroup
	semaphore chan bool
}

func New(workercount int) *Manager {
	if workercount < 0 {
		workercount = runtime.NumCPU() * -workercount
	}

	if workercount < minNumWorkers {
		workercount = minNumWorkers
	}

	return &Manager{
		wg:        &sync.WaitGroup{},
		semaphore: make(chan bool, workercount),
	}
}

// acquire limits concurrency by trying to acquire the semaphore.
func (p *Manager) acquire() {
	p.semaphore <- true
	p.wg.Add(1)
}

// release releases the acquired semaphore to signal that a task is finished.
func (p *Manager) release() {
	p.wg.Done()
	<-p.semaphore
}

// Run runs the given task while limiting the concurrency.
func (p *Manager) Run(fn Task, waiter *Waiter) {
	waiter.wg.Add(1)
	p.acquire()
	go func() {
		defer waiter.wg.Done()
		defer p.release()

		if err := fn(); err != nil {
			waiter.errch <- err
		}
	}()
}

// Close waits all tasks to finish.
func (p *Manager) Close() {
	p.wg.Wait()
	close(p.semaphore)
}

type Waiter struct {
	wg    sync.WaitGroup
	errch chan error
}

func NewWaiter() *Waiter {
	return &Waiter{
		errch: make(chan error),
	}
}

func (w *Waiter) Wait() {
	w.wg.Wait()
	close(w.errch)
}

func (w *Waiter) Err() <-chan error {
	return w.errch
}
