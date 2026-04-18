import type { ResolvedDbConfig } from "../db/services/db-config-resolver.service";
import type {
  GetOptions,
  GroupByParams,
  ListOptions,
  ListResponse,
} from "../records/services/api-records-read.service";
import type { SearchOptions, SearchResponse } from "../search/services/api-search.service";

export class UnsupportedDbReadError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "UnsupportedDbReadError";
  }
}

export class DbConnectionError extends Error {
  constructor(message: string, options?: { cause?: unknown }) {
    super(message);
    this.name = "DbConnectionError";

    if (options?.cause !== undefined) {
      Object.defineProperty(this, "cause", {
        value: options.cause,
        enumerable: false,
        configurable: true,
      });
    }
  }
}

export interface SearchReadBackend {
  runSearch(options: SearchOptions): Promise<SearchResponse>;
}

export interface DbSearchReadService {
  search(target: ResolvedDbConfig, options: SearchOptions): Promise<SearchResponse>;
}

export interface RecordsReadBackend {
  list(object: string, options?: ListOptions): Promise<ListResponse>;
  listAll(object: string, options?: ListOptions): Promise<ListResponse>;
  get(object: string, id: string, options?: GetOptions): Promise<unknown>;
  groupBy(object: string, payload?: unknown, params?: GroupByParams): Promise<unknown>;
}

export interface DbRecordsReadService {
  list(target: ResolvedDbConfig, object: string, options?: ListOptions): Promise<ListResponse>;
  listAll(target: ResolvedDbConfig, object: string, options?: ListOptions): Promise<ListResponse>;
  get(target: ResolvedDbConfig, object: string, id: string, options?: GetOptions): Promise<unknown>;
  groupBy(
    target: ResolvedDbConfig,
    object: string,
    payload?: unknown,
    params?: GroupByParams,
  ): Promise<unknown>;
}

export interface ReadBackendDbReads {
  search?: DbSearchReadService;
  records?: Partial<DbRecordsReadService>;
}
