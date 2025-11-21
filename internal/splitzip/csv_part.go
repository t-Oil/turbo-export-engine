package splitzip

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/turbo-export-engine/pkg/types"
)

func writeCSVPartToZip(zw *zip.Writer, filename string, headers []string, rows []types.Row, includeHeaders bool) error {
	w, err := zw.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	buffered := bufio.NewWriterSize(w, 64*1024)
	csvWriter := csv.NewWriter(buffered)

	if includeHeaders && len(headers) > 0 {
		if err := csvWriter.Write(headers); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}

	for _, row := range rows {
		record := make([]string, len(row))
		for i, cell := range row {
			record[i] = fmt.Sprintf("%v", cell)
		}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("csv writer error: %w", err)
	}

	return buffered.Flush()
}

func generateCSVPartData(headers []string, rows []types.Row, includeHeaders bool) ([]byte, error) {
	var buf bytes.Buffer
	buffered := bufio.NewWriterSize(&buf, 64*1024)
	csvWriter := csv.NewWriter(buffered)

	if includeHeaders && len(headers) > 0 {
		if err := csvWriter.Write(headers); err != nil {
			return nil, fmt.Errorf("failed to write headers: %w", err)
		}
	}

	for _, row := range rows {
		record := make([]string, len(row))
		for i, cell := range row {
			record[i] = fmt.Sprintf("%v", cell)
		}
		if err := csvWriter.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write row: %w", err)
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("csv writer error: %w", err)
	}

	if err := buffered.Flush(); err != nil {
		return nil, fmt.Errorf("buffer flush error: %w", err)
	}

	return buf.Bytes(), nil
}
