package worker

import (
	"sync"

	"github.com/turbo-export-engine/pkg/types"
)

// Pool represents a worker pool for processing export jobs
type Pool struct {
	workers   int
	jobQueue  chan *types.ExportJob
	wg        sync.WaitGroup
	processor JobProcessor
	once      sync.Once
	stopped   bool
	mu        sync.Mutex
}

// JobProcessor defines the interface for processing jobs
type JobProcessor interface {
	Process(job *types.ExportJob) error
}

// NewPool creates a new worker pool
func NewPool(workers int, queueSize int, processor JobProcessor) *Pool {
	if workers <= 0 {
		workers = 1
	}
	if queueSize <= 0 {
		queueSize = 100
	}

	return &Pool{
		workers:   workers,
		jobQueue:  make(chan *types.ExportJob, queueSize),
		processor: processor,
		stopped:   false,
	}
}

// Start initializes and starts the worker pool
func (p *Pool) Start() {
	p.once.Do(func() {
		for i := 0; i < p.workers; i++ {
			p.wg.Add(1)
			go p.worker(i)
		}
	})
}

// worker is the main worker goroutine
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for job := range p.jobQueue {
		err := p.processor.Process(job)
		if job.Result != nil {
			job.Result <- err
			close(job.Result)
		}
	}
}

// Submit adds a job to the queue
func (p *Pool) Submit(job *types.ExportJob) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.stopped {
		p.jobQueue <- job
	}
}

// Shutdown gracefully shuts down the worker pool
func (p *Pool) Shutdown() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.stopped {
		p.stopped = true
		close(p.jobQueue)
		p.wg.Wait()
	}
}

// Wait waits for all jobs to complete
func (p *Pool) Wait() {
	p.wg.Wait()
}
