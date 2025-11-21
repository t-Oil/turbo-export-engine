package xlsx

import (
	"archive/zip"
	"bufio"
	"fmt"
	"html"
	"os"
	"strings"
	"sync"

	"github.com/turbo-export-engine/pkg/types"
)

// Builder handles streaming XLSX file generation
type Builder struct {
	config *types.ExportConfig
	mu     sync.Mutex
}

// NewBuilder creates a new XLSX builder
func NewBuilder(config *types.ExportConfig) *Builder {
	return &Builder{
		config: config,
	}
}

// Build creates an XLSX file with the given data
func (b *Builder) Build(headers []string, rows []types.Row) error {
	// Create output file
	file, err := os.Create(b.config.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Create zip writer
	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// Write [Content_Types].xml
	if err := b.writeContentTypes(zipWriter); err != nil {
		return err
	}

	// Write _rels/.rels
	if err := b.writeRels(zipWriter); err != nil {
		return err
	}

	// Write xl/_rels/workbook.xml.rels
	if err := b.writeWorkbookRels(zipWriter); err != nil {
		return err
	}

	// Write xl/workbook.xml
	if err := b.writeWorkbook(zipWriter); err != nil {
		return err
	}

	// Write xl/worksheets/sheet1.xml (streaming)
	if err := b.writeSheet(zipWriter, headers, rows); err != nil {
		return err
	}

	return nil
}

func (b *Builder) writeContentTypes(zw *zip.Writer) error {
	w, err := zw.Create("[Content_Types].xml")
	if err != nil {
		return err
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
</Types>`

	_, err = w.Write([]byte(content))
	return err
}

func (b *Builder) writeRels(zw *zip.Writer) error {
	w, err := zw.Create("_rels/.rels")
	if err != nil {
		return err
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`

	_, err = w.Write([]byte(content))
	return err
}

func (b *Builder) writeWorkbookRels(zw *zip.Writer) error {
	w, err := zw.Create("xl/_rels/workbook.xml.rels")
	if err != nil {
		return err
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
</Relationships>`

	_, err = w.Write([]byte(content))
	return err
}

func (b *Builder) writeWorkbook(zw *zip.Writer) error {
	w, err := zw.Create("xl/workbook.xml")
	if err != nil {
		return err
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Sheet1" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`

	_, err = w.Write([]byte(content))
	return err
}

func (b *Builder) writeSheet(zw *zip.Writer, headers []string, rows []types.Row) error {
	w, err := zw.Create("xl/worksheets/sheet1.xml")
	if err != nil {
		return err
	}

	buffered := bufio.NewWriterSize(w, 128*1024)
	defer buffered.Flush()

	// Write header
	header := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
`
	if _, err := buffered.WriteString(header); err != nil {
		return err
	}

	rowNum := 1

	// Write header row
	if len(headers) > 0 {
		rowXML := b.buildRowXML(rowNum, headers)
		if _, err := buffered.WriteString(rowXML); err != nil {
			return err
		}
		rowNum++
	}

	// Process rows based on mode
	if b.config.Mode == types.ModeSync {
		// Write rows synchronously
		for _, row := range rows {
			cells := make([]string, len(row))
			for i, cell := range row {
				cells[i] = fmt.Sprintf("%v", cell)
			}
			rowXML := b.buildRowXML(rowNum, cells)
			if _, err := buffered.WriteString(rowXML); err != nil {
				return err
			}
			rowNum++
		}
	} else {
		// Write rows with parallel processing
		chunkSize := b.config.ChunkSize
		if chunkSize <= 0 {
			chunkSize = 10000
		}

		workers := b.config.Workers
		if workers <= 0 {
			workers = 4
		}

		chunks := splitIntoChunks(rows, chunkSize)
		resultChan := make(chan processedChunk, len(chunks))
		errChan := make(chan error, workers)

		var wg sync.WaitGroup
		semaphore := make(chan struct{}, workers)

		for idx, chunk := range chunks {
			wg.Add(1)
			go func(chunkIdx int, chunkData []types.Row, startRow int) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				processed, err := b.processChunkXML(chunkIdx, chunkData, startRow)
				if err != nil {
					errChan <- err
					return
				}
				resultChan <- processed
			}(idx, chunk, rowNum+idx*chunkSize)
		}

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
			if _, err := buffered.WriteString(result.XML); err != nil {
				return err
			}
		}
	}

	// Write footer
	footer := `  </sheetData>
</worksheet>`
	if _, err := buffered.WriteString(footer); err != nil {
		return err
	}

	return nil
}

func (b *Builder) buildRowXML(rowNum int, cells []string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("    <row r=\"%d\">", rowNum))

	for colIdx, cellValue := range cells {
		colName := columnName(colIdx)
		cellRef := fmt.Sprintf("%s%d", colName, rowNum)
		escapedValue := html.EscapeString(cellValue)

		sb.WriteString(fmt.Sprintf("<c r=\"%s\" t=\"inlineStr\"><is><t>%s</t></is></c>",
			cellRef, escapedValue))
	}

	sb.WriteString("</row>\n")
	return sb.String()
}

type processedChunk struct {
	Index int
	XML   string
}

func (b *Builder) processChunkXML(index int, rows []types.Row, startRowNum int) (processedChunk, error) {
	var sb strings.Builder

	for i, row := range rows {
		cells := make([]string, len(row))
		for j, cell := range row {
			cells[j] = fmt.Sprintf("%v", cell)
		}
		rowXML := b.buildRowXML(startRowNum+i, cells)
		sb.WriteString(rowXML)
	}

	return processedChunk{Index: index, XML: sb.String()}, nil
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

// columnName converts a column index to Excel column name (A, B, ..., Z, AA, AB, ...)
func columnName(col int) string {
	name := ""
	col++ // Excel columns are 1-based
	for col > 0 {
		col--
		name = string(rune('A'+(col%26))) + name
		col /= 26
	}
	return name
}
