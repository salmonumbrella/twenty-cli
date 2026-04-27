import Papa from "papaparse";
import type { OutputFormat } from "../../shared/global-options";
import { toLightPayload } from "./compact-aliases";
import { QueryService } from "./query.service";
import { TableService } from "./table.service";

export interface OutputOptions {
  format?: OutputFormat;
  query?: string;
  light?: boolean;
  full?: boolean;
  agentMode?: boolean;
}

interface OutputServiceDefaults extends OutputOptions {}

export class OutputService {
  constructor(
    private table: TableService,
    private queryService: QueryService,
    private defaults: OutputServiceDefaults = {},
  ) {}

  async render(data: unknown, options: OutputOptions = {}): Promise<void> {
    const query = options.query ?? this.defaults.query;
    const full = options.full ?? this.defaults.full ?? false;
    const light = !full && (options.light ?? this.defaults.light ?? false);
    let result: unknown = data;
    if (query) {
      result = this.queryService.apply(result, query);
    }
    if (light) {
      result = toLightPayload(result);
    }

    const format = options.format ?? this.defaults.format ?? "json";
    switch (format) {
      case "json":
        // eslint-disable-next-line no-console
        console.log(JSON.stringify(result));
        break;
      case "jsonl":
        // eslint-disable-next-line no-console
        console.log(this.formatJsonLines(result));
        break;
      case "csv":
        // eslint-disable-next-line no-console
        console.log(this.formatCsv(result));
        break;
      case "text":
        {
          const { data: textData, cliMessage } = this.extractTextCliDiagnostic(result);
          if (cliMessage) {
            // eslint-disable-next-line no-console
            console.log(`Note: ${cliMessage}`);
          }
          this.table.render(textData);
        }
        break;
      default:
        throw new Error(`Unsupported output format: ${format}`);
    }
  }

  private extractTextCliDiagnostic(data: unknown): { data: unknown; cliMessage?: string } {
    if (!isRecord(data)) {
      return { data };
    }

    const cli = data._cli;
    if (!isRecord(cli) || typeof cli.message !== "string" || cli.message.trim() === "") {
      return { data };
    }

    const { _cli, ...rest } = data;
    return {
      data: rest,
      cliMessage: cli.message,
    };
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
