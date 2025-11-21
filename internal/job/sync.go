package job

import (
	"fmt"

	"github.com/turbo-export-engine/internal/csv"
	"github.com/turbo-export-engine/internal/xlsx"
	"github.com/turbo-export-engine/pkg/types"
)

// SyncExecutor handles synchronous job execution
type SyncExecutor struct{}

// NewSyncExecutor creates a new sync executor
func NewSyncExecutor() *SyncExecutor {
	return &SyncExecutor{}
}

// Execute runs the export job synchronously
func (e *SyncExecutor) Execute(job *types.ExportJob) error {
	switch job.Config.Format {
	case types.FormatCSV:
		writer := csv.NewWriter(job.Config)
		return writer.WriteSync(job.Headers, job.Rows)
	case types.FormatXLSX:
		builder := xlsx.NewBuilder(job.Config)
		return builder.Build(job.Headers, job.Rows)
	default:
		return fmt.Errorf("unsupported format: %s", job.Config.Format)
	}
}

// Process implements the JobProcessor interface
func (e *SyncExecutor) Process(job *types.ExportJob) error {
	return e.Execute(job)
}
