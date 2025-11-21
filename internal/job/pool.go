package job

import (
	"fmt"
	"sync"

	"github.com/turbo-export-engine/internal/csv"
	"github.com/turbo-export-engine/internal/worker"
	"github.com/turbo-export-engine/internal/xlsx"
	"github.com/turbo-export-engine/pkg/types"
)

// PoolExecutor handles job execution using a global worker pool
type PoolExecutor struct {
	queue *worker.Queue
	once  sync.Once
}

var globalPoolExecutor *PoolExecutor

// NewPoolExecutor creates or returns the singleton pool executor
func NewPoolExecutor(workers int) *PoolExecutor {
	if globalPoolExecutor == nil {
		globalPoolExecutor = &PoolExecutor{}
		processor := &poolProcessor{}
		globalPoolExecutor.queue = worker.GetGlobalQueue(workers, processor)
	}
	return globalPoolExecutor
}

// Execute submits a job to the global worker pool
func (e *PoolExecutor) Execute(job *types.ExportJob) error {
	if job.Result == nil {
		job.Result = make(chan error, 1)
	}

	e.queue.Submit(job)

	// Wait for result
	return <-job.Result
}

// Shutdown gracefully shuts down the global pool
func (e *PoolExecutor) Shutdown() {
	if e.queue != nil {
		e.queue.Shutdown()
	}
}

// poolProcessor implements JobProcessor for global pool
type poolProcessor struct{}

func (p *poolProcessor) Process(job *types.ExportJob) error {
	switch job.Config.Format {
	case types.FormatCSV:
		writer := csv.NewWriter(job.Config)
		if job.Config.Mode == types.ModeSync {
			return writer.WriteSync(job.Headers, job.Rows)
		}
		return writer.WriteParallel(job.Headers, job.Rows)
	case types.FormatXLSX:
		builder := xlsx.NewBuilder(job.Config)
		return builder.Build(job.Headers, job.Rows)
	default:
		return fmt.Errorf("unsupported format: %s", job.Config.Format)
	}
}
