import {
  Controller,
  Post,
  Body,
  Res,
  StreamableFile,
  HttpException,
  HttpStatus,
} from '@nestjs/common';
import { Response } from 'express';
import { ExportService } from './export.service';
import { Row } from '@export-engine/core';

interface ExportRequest {
  headers: string[];
  rows: Row[];
  format?: 'csv' | 'xlsx';
  mode?: 'sync' | 'parallel' | 'global_pool';
  workers?: number;
}

@Controller('export')
export class ExportController {
  constructor(private readonly exportService: ExportService) {}

  @Post('csv')
  async exportCSV(
    @Body() body: ExportRequest,
    @Res({ passthrough: true }) res: Response
  ): Promise<StreamableFile> {
    try {
      const { headers, rows, mode = 'parallel', workers = 8 } = body;

      if (!headers || !rows) {
        throw new HttpException(
          'Headers and rows are required',
          HttpStatus.BAD_REQUEST
        );
      }

      const buffer = await this.exportService.generateCSV(headers, rows, {
        mode,
        workers,
      });

      res.set({
        'Content-Type': 'text/csv',
        'Content-Disposition': 'attachment; filename="export.csv"',
      });

      return new StreamableFile(buffer);
    } catch (error) {
      throw new HttpException(
        `Export failed: ${error.message}`,
        HttpStatus.INTERNAL_SERVER_ERROR
      );
    }
  }

  @Post('xlsx')
  async exportXLSX(
    @Body() body: ExportRequest,
    @Res({ passthrough: true }) res: Response
  ): Promise<StreamableFile> {
    try {
      const { headers, rows, mode = 'parallel', workers = 8 } = body;

      if (!headers || !rows) {
        throw new HttpException(
          'Headers and rows are required',
          HttpStatus.BAD_REQUEST
        );
      }

      const buffer = await this.exportService.generateXLSX(headers, rows, {
        mode,
        workers,
      });

      res.set({
        'Content-Type':
          'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'Content-Disposition': 'attachment; filename="export.xlsx"',
      });

      return new StreamableFile(buffer);
    } catch (error) {
      throw new HttpException(
        `Export failed: ${error.message}`,
        HttpStatus.INTERNAL_SERVER_ERROR
      );
    }
  }

  @Post('large-csv')
  async exportLargeCSV(
    @Body() body: ExportRequest,
    @Res({ passthrough: true }) res: Response
  ): Promise<StreamableFile> {
    try {
      const { headers, rows } = body;

      if (!headers || !rows) {
        throw new HttpException(
          'Headers and rows are required',
          HttpStatus.BAD_REQUEST
        );
      }

      // Use global_pool mode for large datasets
      const buffer = await this.exportService.generateLargeCSV(headers, rows);

      res.set({
        'Content-Type': 'text/csv',
        'Content-Disposition': 'attachment; filename="large-export.csv"',
      });

      return new StreamableFile(buffer);
    } catch (error) {
      throw new HttpException(
        `Export failed: ${error.message}`,
        HttpStatus.INTERNAL_SERVER_ERROR
      );
    }
  }

  @Post('large-xlsx')
  async exportLargeXLSX(
    @Body() body: ExportRequest,
    @Res({ passthrough: true }) res: Response
  ): Promise<StreamableFile> {
    try {
      const { headers, rows } = body;

      if (!headers || !rows) {
        throw new HttpException(
          'Headers and rows are required',
          HttpStatus.BAD_REQUEST
        );
      }

      // Use global_pool mode for large datasets
      const buffer = await this.exportService.generateLargeXLSX(headers, rows);

      res.set({
        'Content-Type':
          'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'Content-Disposition': 'attachment; filename="large-export.xlsx"',
      });

      return new StreamableFile(buffer);
    } catch (error) {
      throw new HttpException(
        `Export failed: ${error.message}`,
        HttpStatus.INTERNAL_SERVER_ERROR
      );
    }
  }

  @Post('dynamic')
  async exportDynamic(
    @Body() body: ExportRequest,
    @Res({ passthrough: true }) res: Response
  ): Promise<StreamableFile> {
    try {
      const {
        headers,
        rows,
        format = 'csv',
        mode = 'parallel',
        workers = 8,
      } = body;

      if (!headers || !rows) {
        throw new HttpException(
          'Headers and rows are required',
          HttpStatus.BAD_REQUEST
        );
      }

      let buffer: Buffer;
      let contentType: string;
      let filename: string;

      if (format === 'xlsx') {
        buffer = await this.exportService.generateXLSX(headers, rows, {
          mode,
          workers,
        });
        contentType =
          'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet';
        filename = 'export.xlsx';
      } else {
        buffer = await this.exportService.generateCSV(headers, rows, {
          mode,
          workers,
        });
        contentType = 'text/csv';
        filename = 'export.csv';
      }

      res.set({
        'Content-Type': contentType,
        'Content-Disposition': `attachment; filename="${filename}"`,
      });

      return new StreamableFile(buffer);
    } catch (error) {
      throw new HttpException(
        `Export failed: ${error.message}`,
        HttpStatus.INTERNAL_SERVER_ERROR
      );
    }
  }
}
