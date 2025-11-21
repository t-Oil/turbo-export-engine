export type ExportMode = 'sync' | 'parallel' | 'global_pool';
export type ExportFormat = 'csv' | 'xlsx';
export type Row = (string | number | boolean | null)[];

export interface ExportOptions {
  mode?: ExportMode;
  workers?: number;
  chunkSize?: number;
}

export interface ExportResult {
  filePath: string;
  rowCount: number;
  duration: number;
}

export interface ExportData {
  headers: string[];
  rows: Row[];
}

export interface SplitZipOptions {
  mode?: ExportMode;
  workers?: number;
  chunkSize?: number;
  format?: ExportFormat;
  includeHeaders?: boolean;
}

export interface SplitZipResult {
  filePath: string;
  totalParts: number;
  totalRows: number;
  partFiles: string[];
  duration: number;
}
