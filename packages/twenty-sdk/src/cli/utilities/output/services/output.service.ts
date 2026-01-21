import Papa from 'papaparse';
import { QueryService } from './query.service';
import { TableService } from './table.service';

export interface OutputOptions {
  format?: string;
  query?: string;
}

export class OutputService {
  constructor(
    private table: TableService,
    private queryService: QueryService,
  ) {}

  async render(data: unknown, options: OutputOptions): Promise<void> {
    let result: unknown = data;
    if (options.query) {
      result = this.queryService.apply(result, options.query);
    }

    const format = options.format ?? 'text';
    switch (format) {
      case 'json':
        // eslint-disable-next-line no-console
        console.log(JSON.stringify(result, null, 2));
        break;
      case 'csv':
        // eslint-disable-next-line no-console
        console.log(this.formatCsv(result));
        break;
      case 'text':
      default:
        this.table.render(result);
        break;
    }
  }

  private formatCsv(data: unknown): string {
    const records = Array.isArray(data) ? data : [data];
    return Papa.unparse(records as any[]);
  }
}
