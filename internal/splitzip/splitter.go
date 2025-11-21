package splitzip

import (
	"archive/zip"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/turbo-export-engine/pkg/types"
)

type Splitter struct {
	config *types.SplitZipConfig
}

func NewSplitter(config *types.SplitZipConfig) *Splitter {
	return &Splitter{config: config}
}

func (s *Splitter) Execute(headers []string, rows []types.Row) (*types.SplitZipResult, error) {
	if !s.config.Split || !s.config.Zip {
		return nil, fmt.Errorf("split and zip must both be enabled")
	}

	chunkSize := s.config.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 10000
	}

	totalRows := len(rows)
	numParts := (totalRows + chunkSize - 1) / chunkSize
	if numParts == 0 {
		numParts = 1
	}

	file, err := os.Create(s.config.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	var partFiles []string

	switch s.config.Mode {
	case types.ModeSync:
		partFiles, err = s.executeSync(zipWriter, headers, rows, chunkSize, numParts)
	case types.ModeParallel, types.ModeGlobalPool:
		partFiles, err = s.executeParallel(zipWriter, headers, rows, chunkSize, numParts)
	default:
		partFiles, err = s.executeSync(zipWriter, headers, rows, chunkSize, numParts)
	}

	if err != nil {
		return nil, err
	}

	return &types.SplitZipResult{
		OutputPath: s.config.OutputPath,
		TotalParts: numParts,
		TotalRows:  totalRows,
		PartFiles:  partFiles,
	}, nil
}

func (s *Splitter) executeSync(zw *zip.Writer, headers []string, rows []types.Row, chunkSize, numParts int) ([]string, error) {
	partFiles := make([]string, 0, numParts)

	for partIdx := 0; partIdx < numParts; partIdx++ {
		startIdx := partIdx * chunkSize
		endIdx := startIdx + chunkSize
		if endIdx > len(rows) {
			endIdx = len(rows)
		}

		partRows := rows[startIdx:endIdx]
		filename := s.getPartFilename(partIdx)

		if err := s.writePartToZip(zw, filename, headers, partRows); err != nil {
			return nil, fmt.Errorf("failed to write part %d: %w", partIdx+1, err)
		}

		partFiles = append(partFiles, filename)
	}

	return partFiles, nil
}

func (s *Splitter) executeParallel(zw *zip.Writer, headers []string, rows []types.Row, chunkSize, numParts int) ([]string, error) {
	workers := s.config.Workers
	if workers <= 0 {
		workers = 4
	}

	resultChan := make(chan types.PartResult, numParts)
	errChan := make(chan error, workers)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, workers)

	for partIdx := 0; partIdx < numParts; partIdx++ {
		startIdx := partIdx * chunkSize
		endIdx := startIdx + chunkSize
		if endIdx > len(rows) {
			endIdx = len(rows)
		}

		partRows := rows[startIdx:endIdx]

		wg.Add(1)
		go func(idx int, data []types.Row) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			partData, err := s.generatePartData(headers, data)
			if err != nil {
				errChan <- fmt.Errorf("part %d: %w", idx+1, err)
				return
			}

			resultChan <- types.PartResult{
				PartIndex: idx,
				Data:      partData,
				RowCount:  len(data),
			}
		}(partIdx, partRows)
	}

	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	results := make([]types.PartResult, 0, numParts)
	for result := range resultChan {
		results = append(results, result)
	}

	select {
	case err := <-errChan:
		if err != nil {
			return nil, err
		}
	default:
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].PartIndex < results[j].PartIndex
	})

	partFiles := make([]string, 0, numParts)
	for _, result := range results {
		filename := s.getPartFilename(result.PartIndex)

		w, err := zw.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to create zip entry %s: %w", filename, err)
		}

		if _, err := w.Write(result.Data); err != nil {
			return nil, fmt.Errorf("failed to write zip entry %s: %w", filename, err)
		}

		partFiles = append(partFiles, filename)
	}

	return partFiles, nil
}

func (s *Splitter) getPartFilename(partIdx int) string {
	ext := "csv"
	if s.config.Format == types.FormatXLSX {
		ext = "xlsx"
	}
	return fmt.Sprintf("part_%d.%s", partIdx+1, ext)
}

func (s *Splitter) writePartToZip(zw *zip.Writer, filename string, headers []string, rows []types.Row) error {
	switch s.config.Format {
	case types.FormatCSV:
		return writeCSVPartToZip(zw, filename, headers, rows, s.config.IncludeHeaders)
	case types.FormatXLSX:
		return writeXLSXPartToZip(zw, filename, headers, rows, s.config.IncludeHeaders)
	default:
		return fmt.Errorf("unsupported format: %s", s.config.Format)
	}
}

func (s *Splitter) generatePartData(headers []string, rows []types.Row) ([]byte, error) {
	switch s.config.Format {
	case types.FormatCSV:
		return generateCSVPartData(headers, rows, s.config.IncludeHeaders)
	case types.FormatXLSX:
		return generateXLSXPartData(headers, rows, s.config.IncludeHeaders)
	default:
		return nil, fmt.Errorf("unsupported format: %s", s.config.Format)
	}
}
