package csv

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"sync"

	"github.com/turbo-export-engine/pkg/types"
)

// Writer handles streaming CSV writing
type Writer struct {
	config *types.ExportConfig
	mu     sync.Mutex
}

// NewWriter creates a new CSV writer
func NewWriter(config *types.ExportConfig) *Writer {
	return &Writer{
		config: config,
	}
}

// WriteSync writes rows synchronously without workers
func (w *Writer) WriteSync(headers []string, rows []types.Row) error {
	file, err := os.Create(w.config.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	buffered := bufio.NewWriterSize(file, 64*1024)
	defer buffered.Flush()

	csvWriter := csv.NewWriter(buffered)
	defer csvWriter.Flush()

	// Write headers
	if len(headers) > 0 {
		if err := csvWriter.Write(headers); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}

	// Write rows
	for _, row := range rows {
		record := make([]string, len(row))
		for i, cell := range row {
			record[i] = fmt.Sprintf("%v", cell)
		}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// WriteParallel writes rows using parallel worker pool
func (w *Writer) WriteParallel(headers []string, rows []types.Row) error {
	chunkSize := w.config.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 10000
	}

	workers := w.config.Workers
	if workers <= 0 {
		workers = 4
	}

	// Create output file
	file, err := os.Create(w.config.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	buffered := bufio.NewWriterSize(file, 128*1024)
	defer buffered.Flush()

	csvWriter := csv.NewWriter(buffered)
	defer csvWriter.Flush()

	// Write headers
	if len(headers) > 0 {
		if err := csvWriter.Write(headers); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}

	// Split rows into chunks
	chunks := splitIntoChunks(rows, chunkSize)
	resultChan := make(chan processedChunk, len(chunks))
	errChan := make(chan error, workers)

	// Worker pool
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, workers)

	for idx, chunk := range chunks {
		wg.Add(1)
		go func(chunkIdx int, chunkData []types.Row) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			processed, err := processChunk(chunkIdx, chunkData)
			if err != nil {
				errChan <- err
				return
			}
			resultChan <- processed
		}(idx, chunk)
	}

	// Wait for all workers
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	// Collect results in order
	results := make([]processedChunk, len(chunks))
	for result := range resultChan {
		results[result.Index] = result
	}

	// Check for errors
	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
	default:
	}

	// Write results in order
	for _, result := range results {
		for _, record := range result.Records {
			if err := csvWriter.Write(record); err != nil {
				return fmt.Errorf("failed to write record: %w", err)
			}
		}
	}

	return nil
}

type processedChunk struct {
	Index   int
	Records [][]string
}

func processChunk(index int, rows []types.Row) (processedChunk, error) {
	records := make([][]string, len(rows))
	for i, row := range rows {
		record := make([]string, len(row))
		for j, cell := range row {
			record[j] = fmt.Sprintf("%v", cell)
		}
		records[i] = record
	}
	return processedChunk{Index: index, Records: records}, nil
}

func splitIntoChunks(rows []types.Row, chunkSize int) [][]types.Row {
	var chunks [][]types.Row
	for i := 0; i < len(rows); i += chunkSize {
		end := i + chunkSize
		if end > len(rows) {
			end = len(rows)
		}
		chunks = append(chunks, rows[i:end])
	}
	return chunks
}

// Write is the main entry point for writing CSV
func (w *Writer) Write(headers []string, rows []types.Row) error {
	if w.config.Mode == types.ModeSync {
		return w.WriteSync(headers, rows)
	}
	return w.WriteParallel(headers, rows)
}
