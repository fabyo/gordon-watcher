package watcher

import (
	"context"
	"sync"

	"github.com/fabyo/gordon-watcher/internal/metrics"
)

// WorkerPool manages a pool of workers for processing files
type WorkerPool struct {
	maxWorkers int
	queue      chan string
	wg         sync.WaitGroup
	stop       chan struct{}
	processor  func(context.Context, string) error
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(maxWorkers, queueSize int, processor func(context.Context, string) error) *WorkerPool {
	return &WorkerPool{
		maxWorkers: maxWorkers,
		queue:      make(chan string, queueSize),
		stop:       make(chan struct{}),
		processor:  processor,
	}
}

// Start starts the worker pool
func (p *WorkerPool) Start() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// Stop stops the worker pool
func (p *WorkerPool) Stop() {
	close(p.stop)
	close(p.queue)
	p.wg.Wait()
}

// Submit submits a file path to the worker pool
func (p *WorkerPool) Submit(path string) {
	select {
	case p.queue <- path:
		metrics.WorkerPoolQueueSize.Set(float64(len(p.queue)))
	case <-p.stop:
		// Pool is stopped, ignore
	default:
		// Queue is full, log and drop
		metrics.RateLimitDropped.Inc()
	}
}

// SubmitBlocking submits a file path to the worker pool, blocking if the queue is full
func (p *WorkerPool) SubmitBlocking(path string) {
	select {
	case p.queue <- path:
		metrics.WorkerPoolQueueSize.Set(float64(len(p.queue)))
	case <-p.stop:
		// Pool is stopped, ignore
	}
}

// worker processes files from the queue
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.stop:
			return

		case path, ok := <-p.queue:
			if !ok {
				return
			}

			metrics.WorkerPoolActiveWorkers.Inc()
			metrics.WorkerPoolQueueSize.Set(float64(len(p.queue)))

			ctx := context.Background()
			if err := p.processor(ctx, path); err != nil {
				// Error already logged in processor
			}

			metrics.WorkerPoolActiveWorkers.Dec()
		}
	}
}
