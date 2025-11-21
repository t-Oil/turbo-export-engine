# turbo-export-engine

High-performance CSV and XLSX export engine powered by Go. Export millions of rows in seconds.

## Installation

```bash
npm install turbo-export-engine
```

## Usage

```javascript
const { ExportEngine } = require('turbo-export-engine');

const engine = new ExportEngine();

// Sample data
const data = [
  { id: 1, name: 'John', email: 'john@example.com' },
  { id: 2, name: 'Jane', email: 'jane@example.com' }
];

// Export to CSV file
const csvResult = await engine.exportCsv(data, '/tmp/output.csv', {
  mode: 'parallel',
  workers: 4
});

// Export to XLSX file
const xlsxResult = await engine.exportXlsx(data, '/tmp/output.xlsx', {
  mode: 'parallel',
  workers: 4
});

// Export to buffer (in-memory)
const csvBuffer = await engine.exportCsvBuffer(data, { mode: 'parallel' });
const xlsxBuffer = await engine.exportXlsxBuffer(data, { mode: 'parallel' });

// Split large data into multiple files and zip
const zipResult = await engine.splitZip(data, '/tmp/output.zip', {
  format: 'csv',
  chunkSize: 1000,
  mode: 'parallel',
  workers: 4
});
```

## API

### ExportEngine

#### `exportCsv(data, outputPath, options?)`
Export data to CSV file.

#### `exportXlsx(data, outputPath, options?)`
Export data to XLSX file.

#### `exportCsvBuffer(data, options?)`
Export data to CSV buffer.

#### `exportXlsxBuffer(data, options?)`
Export data to XLSX buffer.

#### `splitZip(data, outputPath, options?)`
Split data into multiple files and package as ZIP.

#### `splitZipBuffer(data, options?)`
Split data into multiple files and return ZIP as buffer.

### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| mode | `'sync'` \| `'parallel'` \| `'global_pool'` | `'sync'` | Execution mode |
| workers | number | 4 | Number of parallel workers |
| chunkSize | number | 10000 | Rows per chunk (for splitZip) |
| format | `'csv'` \| `'xlsx'` | `'csv'` | Output format (for splitZip) |
| includeHeaders | boolean | true | Include headers in output |

## Performance

- **500K rows CSV**: ~1.5 seconds
- **500K rows XLSX**: ~3 seconds
- **Parallel mode**: Up to 4x faster than sync mode

## License

MIT
