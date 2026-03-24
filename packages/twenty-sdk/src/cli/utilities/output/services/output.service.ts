import Papa from "papaparse";
import { QueryService } from "./query.service";
import { TableService } from "./table.service";

export interface OutputOptions {
  format?: string;
  query?: string;
  kind?: string;
}

interface OutputServiceDefaults {
  kind?: string;
}

export class OutputService {
  constructor(
    private table: TableService,
    private queryService: QueryService,
    private defaults: OutputServiceDefaults = {},
  ) {}

  async render(data: unknown, options: OutputOptions): Promise<void> {
    let result: unknown = data;
    if (options.query) {
      result = this.queryService.apply(result, options.query);
    }

    const format = options.format ?? "text";
    switch (format) {
      case "json":
        // eslint-disable-next-line no-console
        console.log(JSON.stringify(result, null, 2));
        break;
      case "jsonl":
        // eslint-disable-next-line no-console
        console.log(this.formatJsonLines(result));
        break;
      case "agent":
        // eslint-disable-next-line no-console
        console.log(JSON.stringify(this.wrapAgentPayload(result, options), null, 2));
        break;
      case "csv":
        // eslint-disable-next-line no-console
        console.log(this.formatCsv(result));
        break;
      case "text":
      default:
        this.table.render(result);
        break;
    }
  }

  private formatCsv(data: unknown): string {
    const records = Array.isArray(data) ? data : [data];
    const preprocessed = records.map((record) => this.preprocessForCsv(record));
    return Papa.unparse(preprocessed as any[]);
  }

  private formatJsonLines(data: unknown): string {
    const records = Array.isArray(data) ? data : [data];
    return records.map((record) => JSON.stringify(record)).join("\n");
  }

  private wrapAgentPayload(data: unknown, options: OutputOptions): Record<string, unknown> {
    const kind = options.kind ?? this.defaults.kind ?? "result";

    if (Array.isArray(data)) {
      return {
        kind,
        items: data,
      };
    }

    if (isRecord(data)) {
      return {
        kind,
        item: data,
      };
    }

    return {
      kind,
      data,
    };
  }

  private preprocessForCsv(record: unknown): unknown {
    if (record === null || record === undefined) {
      return record;
    }
    if (typeof record !== "object") {
      return record;
    }
    if (Array.isArray(record)) {
      return JSON.stringify(record);
    }
    const result: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(record as Record<string, unknown>)) {
      if (value === null || value === undefined) {
        result[key] = "";
      } else if (typeof value === "object") {
        result[key] = JSON.stringify(value);
      } else {
        result[key] = value;
      }
    }
    return result;
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
