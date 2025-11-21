package job

import (
	"fmt"

	"github.com/turbo-export-engine/internal/csv"
	"github.com/turbo-export-engine/internal/xlsx"
	"github.com/turbo-export-engine/pkg/types"
)

// ParallelExecutor handles parallel job execution with per-job worker pool
type ParallelExecutor struct{}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor() *ParallelExecutor {
	return &ParallelExecutor{}
}

// Execute runs the export job with parallel workers
func (e *ParallelExecutor) Execute(job *types.ExportJob) error {
	// Ensure parallel mode is set
	job.Config.Mode = types.ModeParallel

	switch job.Config.Format {
	case types.FormatCSV:
		writer := csv.NewWriter(job.Config)
		return writer.WriteParallel(job.Headers, job.Rows)
	case types.FormatXLSX:
		builder := xlsx.NewBuilder(job.Config)
		return builder.Build(job.Headers, job.Rows)
	default:
		return fmt.Errorf("unsupported format: %s", job.Config.Format)
	}
}

// Process implements the JobProcessor interface
func (e *ParallelExecutor) Process(job *types.ExportJob) error {
	return e.Execute(job)
}
