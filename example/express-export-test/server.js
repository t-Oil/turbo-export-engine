const express = require("express");
const path = require("path");
const fs = require("fs").promises;
const {
  exportXLSXBuffer,
  exportCSVBuffer,
  splitZipBuffer,
} = require("../../node-wrapper/dist/index.js");

const app = express();
const PORT = 3011;

// Generate sample data as array of arrays (required format for Go engine)
function generateRows(count) {
  const rows = [];
  for (let i = 1; i <= count; i++) {
    rows.push([
      i, // ID
      `User ${i}`, // Name
      `user${i}@example.com`, // Email
      20 + (i % 50), // Age
      ["New York", "London", "Tokyo", "Paris", "Berlin"][i % 5], // City
      30000 + (i % 100) * 1000, // Salary
      ["Sales", "Engineering", "Marketing", "HR", "Finance"][i % 5], // Department
      new Date(2020, i % 12, (i % 28) + 1).toISOString().split("T")[0], // Join Date
      i % 3 === 0 ? "Active" : "Inactive", // Status
    ]);
  }
  return rows;
}

// Headers for the export
const HEADERS = [
  "ID",
  "Name",
  "Email",
  "Age",
  "City",
  "Salary",
  "Department",
  "Join Date",
  "Status",
];

// Endpoint to export Excel using Go engine
app.get("/export/xlsx", async (req, res) => {
  const rowCount = parseInt(req.query.rows) || 500000;
  const workers = parseInt(req.query.workers) || 8;
  const mode = req.query.mode || "parallel";

  console.log(`Starting XLSX export of ${rowCount} rows using Go engine...`);
  console.log(`Mode: ${mode}, Workers: ${workers}`);
  const startTime = Date.now();

  try {
    // Generate data
    const rows = generateRows(rowCount);

    // Export using Go engine
    const buffer = await exportXLSXBuffer(HEADERS, rows, {
      mode,
      workers,
      chunkSize: 10000,
    });

    const duration = ((Date.now() - startTime) / 1000).toFixed(2);
    console.log(`XLSX export completed in ${duration} seconds`);

    // Set response headers
    res.setHeader(
      "Content-Type",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    );
    res.setHeader(
      "Content-Disposition",
      `attachment; filename=export_${rowCount}_rows.xlsx`,
    );
    res.setHeader("X-Export-Duration", duration);
    res.setHeader("X-Export-Rows", rowCount);
    res.setHeader("X-Export-Engine", "Go");

    res.send(buffer);
  } catch (error) {
    console.error("Export error:", error);
    res.status(500).json({ error: "Export failed", message: error.message });
  }
});

// Endpoint to export CSV using Go engine
app.get("/export/csv", async (req, res) => {
  const rowCount = parseInt(req.query.rows) || 500000;
  const workers = parseInt(req.query.workers) || 8;
  const mode = req.query.mode || "parallel";

  console.log(`Starting CSV export of ${rowCount} rows using Go engine...`);
  console.log(`Mode: ${mode}, Workers: ${workers}`);
  const startTime = Date.now();

  try {
    // Generate data
    const rows = generateRows(rowCount);

    // Export using Go engine
    const buffer = await exportCSVBuffer(HEADERS, rows, {
      mode,
      workers,
      chunkSize: 10000,
    });

    const duration = ((Date.now() - startTime) / 1000).toFixed(2);
    console.log(`CSV export completed in ${duration} seconds`);

    // Set response headers
    res.setHeader("Content-Type", "text/csv");
    res.setHeader(
      "Content-Disposition",
      `attachment; filename=export_${rowCount}_rows.csv`,
    );
    res.setHeader("X-Export-Duration", duration);
    res.setHeader("X-Export-Rows", rowCount);
    res.setHeader("X-Export-Engine", "Go");

    res.send(buffer);
  } catch (error) {
    console.error("Export error:", error);
    res.status(500).json({ error: "Export failed", message: error.message });
  }
});

// Endpoint to split and zip using Go engine
app.get("/export/split-zip", async (req, res) => {
  const rowCount = parseInt(req.query.rows) || 500000;
  const workers = parseInt(req.query.workers) || 8;
  const mode = req.query.mode || "parallel";
  const format = req.query.format || "csv"; // csv or xlsx
  const chunkSize = parseInt(req.query.chunkSize) || 100000; // rows per file

  console.log(
    `Starting Split-Zip export of ${rowCount} rows using Go engine...`,
  );
  console.log(
    `Mode: ${mode}, Workers: ${workers}, Format: ${format}, ChunkSize: ${chunkSize}`,
  );
  const startTime = Date.now();

  try {
    // Generate data
    const rows = generateRows(rowCount);

    // Export using Go engine with split+zip
    const buffer = await splitZipBuffer(HEADERS, rows, {
      mode,
      workers,
      chunkSize,
      format,
      includeHeaders: true,
    });

    const duration = ((Date.now() - startTime) / 1000).toFixed(2);
    const totalParts = Math.ceil(rowCount / chunkSize);
    console.log(
      `Split-Zip export completed in ${duration} seconds (${totalParts} parts)`,
    );

    // Set response headers
    res.setHeader("Content-Type", "application/zip");
    res.setHeader(
      "Content-Disposition",
      `attachment; filename=export_${rowCount}_rows_split.zip`,
    );
    res.setHeader("X-Export-Duration", duration);
    res.setHeader("X-Export-Rows", rowCount);
    res.setHeader("X-Export-Parts", totalParts);
    res.setHeader("X-Export-Format", format);
    res.setHeader("X-Export-Engine", "Go");

    res.send(buffer);
  } catch (error) {
    console.error("Export error:", error);
    res.status(500).json({ error: "Export failed", message: error.message });
  }
});

// Health check endpoint
app.get("/health", (req, res) => {
  res.json({ status: "ok", message: "Export server is running" });
});

// Root endpoint with instructions
app.get("/", (req, res) => {
  res.json({
    message: "Go Export Engine Test Server",
    engine: "High-performance Go-powered export engine",
    endpoints: {
      "/export/xlsx": "Export data to Excel (default: 500k rows)",
      "/export/csv": "Export data to CSV (default: 500k rows)",
      "/export/split-zip":
        "Split data into multiple files and zip (default: 500k rows)",
      "/health": "Health check",
    },
    parameters: {
      rows: "Number of rows to export (e.g., ?rows=100000)",
      workers: "Number of workers (e.g., ?workers=8)",
      mode: "Execution mode: sync, parallel, or global_pool (e.g., ?mode=parallel)",
      format:
        "[split-zip only] Output format: csv or xlsx (e.g., ?format=xlsx)",
      chunkSize: "[split-zip only] Rows per file (e.g., ?chunkSize=100000)",
    },
    examples: {
      xlsx_500k: "http://localhost:3011/export/xlsx",
      xlsx_custom:
        "http://localhost:3011/export/xlsx?rows=100000&workers=8&mode=parallel",
      csv_500k: "http://localhost:3011/export/csv",
      csv_custom:
        "http://localhost:3011/export/csv?rows=100000&workers=8&mode=parallel",
      split_zip_csv:
        "http://localhost:3011/export/split-zip?rows=500000&chunkSize=100000&format=csv",
      split_zip_xlsx:
        "http://localhost:3011/export/split-zip?rows=500000&chunkSize=100000&format=xlsx",
    },
    performance: {
      engine: "Go (native)",
      expected_xlsx_500k: "8-15 seconds",
      expected_csv_500k: "2-4 seconds",
      expected_split_zip: "Varies by format and chunk size",
      memory: "Low (streaming architecture)",
    },
  });
});

app.listen(PORT, () => {
  console.log("=".repeat(60));
  console.log(`Go Export Engine Test Server`);
  console.log("=".repeat(60));
  console.log(`Server running on http://localhost:${PORT}`);
  console.log("");
  console.log("XLSX Export:");
  console.log(`  Default:  http://localhost:${PORT}/export/xlsx`);
  console.log(
    `  Custom:   http://localhost:${PORT}/export/xlsx?rows=100000&workers=8`,
  );
  console.log("");
  console.log("CSV Export:");
  console.log(`  Default:  http://localhost:${PORT}/export/csv`);
  console.log(
    `  Custom:   http://localhost:${PORT}/export/csv?rows=100000&workers=8`,
  );
  console.log("");
  console.log("Split-Zip Export:");
  console.log(
    `  CSV:      http://localhost:${PORT}/export/split-zip?format=csv&chunkSize=100000`,
  );
  console.log(
    `  XLSX:     http://localhost:${PORT}/export/split-zip?format=xlsx&chunkSize=100000`,
  );
  console.log("");
  console.log("Modes: sync, parallel (default), global_pool");
  console.log("=".repeat(60));
});
