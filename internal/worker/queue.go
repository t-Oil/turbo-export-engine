package worker

import (
	"sync"

	"github.com/turbo-export-engine/pkg/types"
)

// GlobalQueue is a singleton global job queue
var (
	globalQueue     *Queue
	globalQueueOnce sync.Once
)

// Queue represents a global job queue with a fixed worker pool
type Queue struct {
	pool      *Pool
	processor JobProcessor
	mu        sync.Mutex
}

// GetGlobalQueue returns the singleton global queue instance
func GetGlobalQueue(workers int, processor JobProcessor) *Queue {
	globalQueueOnce.Do(func() {
		globalQueue = NewQueue(workers, processor)
		globalQueue.Start()
	})
	return globalQueue
}

// NewQueue creates a new queue with a worker pool
func NewQueue(workers int, processor JobProcessor) *Queue {
	return &Queue{
		pool:      NewPool(workers, 1000, processor),
		processor: processor,
	}
}

// Start starts the queue's worker pool
func (q *Queue) Start() {
	q.pool.Start()
}

// Submit adds a job to the global queue
func (q *Queue) Submit(job *types.ExportJob) {
	q.pool.Submit(job)
}

// Shutdown gracefully shuts down the queue
func (q *Queue) Shutdown() {
	q.pool.Shutdown()
}
