package splitzip

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"html"
	"strings"

	"github.com/turbo-export-engine/pkg/types"
)

func writeXLSXPartToZip(zw *zip.Writer, filename string, headers []string, rows []types.Row, includeHeaders bool) error {
	w, err := zw.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	xlsxWriter := zip.NewWriter(w)
	defer xlsxWriter.Close()

	return writeXLSXStructure(xlsxWriter, headers, rows, includeHeaders)
}

func generateXLSXPartData(headers []string, rows []types.Row, includeHeaders bool) ([]byte, error) {
	var buf bytes.Buffer
	xlsxWriter := zip.NewWriter(&buf)

	if err := writeXLSXStructure(xlsxWriter, headers, rows, includeHeaders); err != nil {
		xlsxWriter.Close()
		return nil, err
	}

	if err := xlsxWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close xlsx writer: %w", err)
	}

	return buf.Bytes(), nil
}

func writeXLSXStructure(zw *zip.Writer, headers []string, rows []types.Row, includeHeaders bool) error {
	if err := writeContentTypes(zw); err != nil {
		return err
	}
	if err := writeRels(zw); err != nil {
		return err
	}
	if err := writeWorkbookRels(zw); err != nil {
		return err
	}
	if err := writeWorkbook(zw); err != nil {
		return err
	}
	return writeSheet(zw, headers, rows, includeHeaders)
}

func writeContentTypes(zw *zip.Writer) error {
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

func writeRels(zw *zip.Writer) error {
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

func writeWorkbookRels(zw *zip.Writer) error {
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

func writeWorkbook(zw *zip.Writer) error {
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

func writeSheet(zw *zip.Writer, headers []string, rows []types.Row, includeHeaders bool) error {
	w, err := zw.Create("xl/worksheets/sheet1.xml")
	if err != nil {
		return err
	}

	buffered := bufio.NewWriterSize(w, 128*1024)

	sheetHeader := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
`
	if _, err := buffered.WriteString(sheetHeader); err != nil {
		return err
	}

	rowNum := 1

	if includeHeaders && len(headers) > 0 {
		if _, err := buffered.WriteString(buildRowXML(rowNum, headers)); err != nil {
			return err
		}
		rowNum++
	}

	for _, row := range rows {
		cells := make([]string, len(row))
		for i, cell := range row {
			cells[i] = fmt.Sprintf("%v", cell)
		}
		if _, err := buffered.WriteString(buildRowXML(rowNum, cells)); err != nil {
			return err
		}
		rowNum++
	}

	if _, err := buffered.WriteString("  </sheetData>\n</worksheet>"); err != nil {
		return err
	}

	return buffered.Flush()
}

func buildRowXML(rowNum int, cells []string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("    <row r=\"%d\">", rowNum))
	for colIdx, cellValue := range cells {
		colName := columnName(colIdx)
		cellRef := fmt.Sprintf("%s%d", colName, rowNum)
		sb.WriteString(fmt.Sprintf("<c r=\"%s\" t=\"inlineStr\"><is><t>%s</t></is></c>",
			cellRef, html.EscapeString(cellValue)))
	}
	sb.WriteString("</row>\n")
	return sb.String()
}

func columnName(col int) string {
	name := ""
	col++
	for col > 0 {
		col--
		name = string(rune('A'+(col%26))) + name
		col /= 26
	}
	return name
}
