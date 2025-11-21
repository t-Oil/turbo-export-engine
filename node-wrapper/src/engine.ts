import { spawn } from 'child_process';
import * as fs from 'fs/promises';
import * as path from 'path';
import * as os from 'os';
import { detectPlatform } from './detect';
import {
  ExportMode,
  ExportFormat,
  ExportOptions,
  ExportResult,
  ExportData,
  Row,
  SplitZipOptions,
  SplitZipResult,
} from './types';

export class ExportEngine {
  private binaryPath: string;

  constructor() {
    const { binaryPath } = detectPlatform();
    this.binaryPath = binaryPath;
  }

  async exportCSV(
    headers: string[],
    rows: Row[],
    outputPath: string,
    options: ExportOptions = {}
  ): Promise<ExportResult> {
    return this.export('csv', headers, rows, outputPath, options);
  }

  async exportXLSX(
    headers: string[],
    rows: Row[],
    outputPath: string,
    options: ExportOptions = {}
  ): Promise<ExportResult> {
    return this.export('xlsx', headers, rows, outputPath, options);
  }

  async exportCSVBuffer(
    headers: string[],
    rows: Row[],
    options: ExportOptions = {}
  ): Promise<Buffer> {
    const tmpOutput = path.join(os.tmpdir(), `export-${Date.now()}.csv`);
    await this.exportCSV(headers, rows, tmpOutput, options);
    const buffer = await fs.readFile(tmpOutput);
    await fs.unlink(tmpOutput);
    return buffer;
  }

  async exportXLSXBuffer(
    headers: string[],
    rows: Row[],
    options: ExportOptions = {}
  ): Promise<Buffer> {
    const tmpOutput = path.join(os.tmpdir(), `export-${Date.now()}.xlsx`);
    await this.exportXLSX(headers, rows, tmpOutput, options);
    const buffer = await fs.readFile(tmpOutput);
    await fs.unlink(tmpOutput);
    return buffer;
  }

  async splitZip(
    headers: string[],
    rows: Row[],
    outputPath: string,
    options: SplitZipOptions = {}
  ): Promise<SplitZipResult> {
    const startTime = Date.now();

    // Create temporary input file
    const tmpInput = path.join(os.tmpdir(), `input-${Date.now()}.json`);
    const data: ExportData = { headers, rows };
    await fs.writeFile(tmpInput, JSON.stringify(data));

    try {
      // Build command arguments
      const args = [
        'split-zip',
        '--mode', options.mode || 'parallel',
        '--workers', String(options.workers || 4),
        '--chunk-size', String(options.chunkSize || 10000),
        '--format', options.format || 'csv',
        '--include-headers', String(options.includeHeaders !== false),
        '--input', tmpInput,
        '--output', outputPath,
      ];

      // Execute binary and capture output
      const output = await this.executeBinaryWithOutput(args);

      const duration = Date.now() - startTime;

      // Parse output to get part count
      const partsMatch = output.match(/Total Parts:\s*(\d+)/);
      const totalParts = partsMatch ? parseInt(partsMatch[1]) : 0;

      // Extract part files from output
      const partFiles: string[] = [];
      const partRegex = /- (part_\d+\.(csv|xlsx))/g;
      let match;
      while ((match = partRegex.exec(output)) !== null) {
        partFiles.push(match[1]);
      }

      return {
        filePath: outputPath,
        totalParts,
        totalRows: rows.length,
        partFiles,
        duration,
      };
    } finally {
      // Clean up temp input file
      await fs.unlink(tmpInput).catch(() => {});
    }
  }

  async splitZipBuffer(
    headers: string[],
    rows: Row[],
    options: SplitZipOptions = {}
  ): Promise<Buffer> {
    const tmpOutput = path.join(os.tmpdir(), `split-export-${Date.now()}.zip`);
    await this.splitZip(headers, rows, tmpOutput, options);
    const buffer = await fs.readFile(tmpOutput);
    await fs.unlink(tmpOutput);
    return buffer;
  }

  private async export(
    format: ExportFormat,
    headers: string[],
    rows: Row[],
    outputPath: string,
    options: ExportOptions
  ): Promise<ExportResult> {
    const startTime = Date.now();

    // Create temporary input file
    const tmpInput = path.join(os.tmpdir(), `input-${Date.now()}.json`);
    const data: ExportData = { headers, rows };
    await fs.writeFile(tmpInput, JSON.stringify(data));

    try {
      // Build command arguments
      const args = [
        format,
        '--mode', options.mode || 'parallel',
        '--workers', String(options.workers || 4),
        '--chunk-size', String(options.chunkSize || 10000),
        '--input', tmpInput,
        '--output', outputPath,
      ];

      // Execute binary
      await this.executeBinary(args);

      const duration = Date.now() - startTime;

      return {
        filePath: outputPath,
        rowCount: rows.length,
        duration,
      };
    } finally {
      // Clean up temp input file
      await fs.unlink(tmpInput).catch(() => {});
    }
  }

  private executeBinary(args: string[]): Promise<void> {
    return new Promise((resolve, reject) => {
      const child = spawn(this.binaryPath, args, {
        stdio: ['ignore', 'pipe', 'pipe'],
      });

      let stdout = '';
      let stderr = '';

      child.stdout?.on('data', (data) => {
        stdout += data.toString();
      });

      child.stderr?.on('data', (data) => {
        stderr += data.toString();
      });

      child.on('error', (error) => {
        reject(new Error(`Failed to spawn binary: ${error.message}`));
      });

      child.on('close', (code) => {
        if (code !== 0) {
          reject(
            new Error(
              `Export process exited with code ${code}\nStderr: ${stderr}`
            )
          );
        } else {
          resolve();
        }
      });
    });
  }

  private executeBinaryWithOutput(args: string[]): Promise<string> {
    return new Promise((resolve, reject) => {
      const child = spawn(this.binaryPath, args, {
        stdio: ['ignore', 'pipe', 'pipe'],
      });

      let stdout = '';
      let stderr = '';

      child.stdout?.on('data', (data) => {
        stdout += data.toString();
      });

      child.stderr?.on('data', (data) => {
        stderr += data.toString();
      });

      child.on('error', (error) => {
        reject(new Error(`Failed to spawn binary: ${error.message}`));
      });

      child.on('close', (code) => {
        if (code !== 0) {
          reject(
            new Error(
              `Export process exited with code ${code}\nStderr: ${stderr}`
            )
          );
        } else {
          resolve(stdout);
        }
      });
    });
  }
}

// Singleton instance
let engineInstance: ExportEngine | null = null;

export function getEngine(): ExportEngine {
  if (!engineInstance) {
    engineInstance = new ExportEngine();
  }
  return engineInstance;
}

// Convenience exports
export async function exportCSV(
  headers: string[],
  rows: Row[],
  outputPath: string,
  options: ExportOptions = {}
): Promise<ExportResult> {
  return getEngine().exportCSV(headers, rows, outputPath, options);
}

export async function exportXLSX(
  headers: string[],
  rows: Row[],
  outputPath: string,
  options: ExportOptions = {}
): Promise<ExportResult> {
  return getEngine().exportXLSX(headers, rows, outputPath, options);
}

export async function exportCSVBuffer(
  headers: string[],
  rows: Row[],
  options: ExportOptions = {}
): Promise<Buffer> {
  return getEngine().exportCSVBuffer(headers, rows, options);
}

export async function exportXLSXBuffer(
  headers: string[],
  rows: Row[],
  options: ExportOptions = {}
): Promise<Buffer> {
  return getEngine().exportXLSXBuffer(headers, rows, options);
}

export async function splitZip(
  headers: string[],
  rows: Row[],
  outputPath: string,
  options: SplitZipOptions = {}
): Promise<SplitZipResult> {
  return getEngine().splitZip(headers, rows, outputPath, options);
}

export async function splitZipBuffer(
  headers: string[],
  rows: Row[],
  options: SplitZipOptions = {}
): Promise<Buffer> {
  return getEngine().splitZipBuffer(headers, rows, options);
}
