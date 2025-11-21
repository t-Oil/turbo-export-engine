# Turbo Export Engine

High-performance data export engine for CSV and XLSX generation, powered by Go.

## Features

- CSV and XLSX export (no-style XLSX for speed)
- Split + ZIP: Split large datasets into multiple files
- Three modes: `sync`, `parallel`, `global_pool`
- Streaming architecture (low memory usage)
- Node.js wrapper included

## Installation

```bash
# Build Go binary
go build -o bin/export-engine ./cmd/export-engine

# Copy to node-wrapper
cp bin/export-engine node-wrapper/bin/export-engine-macos  # or -linux, -win.exe

# Build Node wrapper
cd node-wrapper && npm install && npm run build
```

## CLI Usage

### Export CSV
```bash
./export-engine csv --input data.json --output out.csv --mode parallel --workers 8
```

### Export XLSX
```bash
./export-engine xlsx --input data.json --output out.xlsx --mode parallel --workers 8
```

### Split + ZIP
```bash
# Split into multiple CSV files, zipped
./export-engine split-zip --input data.json --output out.zip --format csv --chunk-size 100000

# Split into multiple XLSX files, zipped
./export-engine split-zip --input data.json --output out.zip --format xlsx --chunk-size 100000
```

### Input Format
```json
{
  "headers": ["Name", "Email", "Age"],
  "rows": [
    ["John", "john@example.com", 30],
    ["Jane", "jane@example.com", 25]
  ]
}
```

### Modes
| Mode | Description |
|------|-------------|
| `sync` | Single-threaded, low memory |
| `parallel` | Per-job worker pool (default) |
| `global_pool` | Shared pool for concurrent jobs |

## Node.js Usage

```typescript
import {
  exportCSVBuffer,
  exportXLSXBuffer,
  splitZipBuffer
} from '@turbo-export-engine/core';

// CSV buffer
const csv = await exportCSVBuffer(headers, rows, {
  mode: 'parallel',
  workers: 8
});

// XLSX buffer
const xlsx = await exportXLSXBuffer(headers, rows, {
  mode: 'parallel',
  workers: 8
});

// Split + ZIP buffer
const zip = await splitZipBuffer(headers, rows, {
  mode: 'parallel',
  workers: 8,
  chunkSize: 100000,
  format: 'csv',        // or 'xlsx'
  includeHeaders: true
});
```

## Express Example

```javascript
const { splitZipBuffer } = require('@turbo-export-engine/core');

app.get('/export/split-zip', async (req, res) => {
  const { rows, chunkSize, format } = req.query;

  const buffer = await splitZipBuffer(headers, data, {
    mode: 'parallel',
    workers: 8,
    chunkSize: parseInt(chunkSize) || 100000,
    format: format || 'csv'
  });

  res.setHeader('Content-Type', 'application/zip');
  res.setHeader('Content-Disposition', 'attachment; filename=export.zip');
  res.send(buffer);
});
```

## API Reference

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--input` | required | Input JSON file |
| `--output` | required | Output file path |
| `--mode` | `sync` | Execution mode |
| `--workers` | `4` | Number of workers |
| `--chunk-size` | `10000` | Rows per chunk |
| `--format` | `csv` | Output format (split-zip only) |
| `--include-headers` | `true` | Headers in each part (split-zip only) |

### Node.js Options

```typescript
interface ExportOptions {
  mode?: 'sync' | 'parallel' | 'global_pool';
  workers?: number;
  chunkSize?: number;
}

interface SplitZipOptions extends ExportOptions {
  format?: 'csv' | 'xlsx';
  includeHeaders?: boolean;
}
```

## Project Structure

```
turbo-export-engine/
├── cmd/export-engine/main.go    # CLI entry point
├── internal/
│   ├── csv/                     # CSV writer
│   ├── xlsx/                    # XLSX builder
│   ├── job/                     # Job executors
│   └── splitzip/                # Split + ZIP logic
├── pkg/types/                   # Type definitions
├── node-wrapper/                # Node.js wrapper
└── example/
    ├── express-export-test/     # Express example
    └── nestjs-example/          # NestJS example
```

## License

MIT
