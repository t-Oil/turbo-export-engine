import { Injectable } from '@nestjs/common';
import { ExportEngine, Row, ExportOptions } from '@export-engine/core';

@Injectable()
export class ExportService {
  private engine: ExportEngine;

  constructor() {
    this.engine = new ExportEngine();
  }

  async generateCSV(
    headers: string[],
    rows: Row[],
    options: ExportOptions = { mode: 'parallel', workers: 8 }
  ): Promise<Buffer> {
    return this.engine.exportCSVBuffer(headers, rows, options);
  }

  async generateXLSX(
    headers: string[],
    rows: Row[],
    options: ExportOptions = { mode: 'parallel', workers: 8 }
  ): Promise<Buffer> {
    return this.engine.exportXLSXBuffer(headers, rows, options);
  }

  async generateLargeCSV(
    headers: string[],
    rows: Row[]
  ): Promise<Buffer> {
    // For very large datasets, use global_pool mode
    return this.engine.exportCSVBuffer(headers, rows, {
      mode: 'global_pool',
      workers: 4,
      chunkSize: 50000,
    });
  }

  async generateLargeXLSX(
    headers: string[],
    rows: Row[]
  ): Promise<Buffer> {
    // For very large datasets, use global_pool mode
    return this.engine.exportXLSXBuffer(headers, rows, {
      mode: 'global_pool',
      workers: 4,
      chunkSize: 50000,
    });
  }

  async exportToFile(
    format: 'csv' | 'xlsx',
    headers: string[],
    rows: Row[],
    outputPath: string,
    options: ExportOptions = { mode: 'parallel', workers: 8 }
  ): Promise<{ filePath: string; rowCount: number; duration: number }> {
    if (format === 'csv') {
      return this.engine.exportCSV(headers, rows, outputPath, options);
    } else {
      return this.engine.exportXLSX(headers, rows, outputPath, options);
    }
  }
}
