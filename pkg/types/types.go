package types

type ExportMode string

const (
	ModeSync       ExportMode = "sync"
	ModeParallel   ExportMode = "parallel"
	ModeGlobalPool ExportMode = "global_pool"
)

type ExportFormat string

const (
	FormatCSV  ExportFormat = "csv"
	FormatXLSX ExportFormat = "xlsx"
)

type Row []interface{}

type ExportConfig struct {
	Mode       ExportMode   `json:"mode"`
	Format     ExportFormat `json:"format"`
	Workers    int          `json:"workers"`
	ChunkSize  int          `json:"chunk_size"`
	InputPath  string       `json:"input_path"`
	OutputPath string       `json:"output_path"`
}

type ExportJob struct {
	ID      string
	Config  *ExportConfig
	Rows    []Row
	Headers []string
	Result  chan error
}

type SplitZipConfig struct {
	Split          bool         `json:"split"`
	Zip            bool         `json:"zip"`
	ChunkSize      int          `json:"chunk_size"`
	Format         ExportFormat `json:"format"`
	Mode           ExportMode   `json:"mode"`
	Workers        int          `json:"workers"`
	IncludeHeaders bool         `json:"include_headers"`
	OutputPath     string       `json:"output_path"`
}

type PartResult struct {
	PartIndex int
	Data      []byte
	RowCount  int
	Error     error
}

type SplitZipResult struct {
	OutputPath string   `json:"output_path"`
	TotalParts int      `json:"total_parts"`
	TotalRows  int      `json:"total_rows"`
	PartFiles  []string `json:"part_files"`
}
