import type { ResolvedDbConfig } from "../db/services/db-config-resolver.service";
import type { DbConfigResolverService } from "../db/services/db-config-resolver.service";
import type { ApiRecordsReadService } from "../records/services/api-records-read.service";
import type {
  GetOptions,
  GroupByParams,
  ListOptions,
  ListResponse,
} from "../records/services/api-records-read.service";
import type { ApiSearchService } from "../search/services/api-search.service";
import type { SearchOptions, SearchResponse } from "../search/services/api-search.service";
import {
  DbConnectionError,
  type ReadBackendDbReads,
  type RecordsReadBackend,
  type SearchReadBackend,
  UnsupportedDbReadError,
} from "./types";

type DbConfigResolver = Pick<DbConfigResolverService, "resolve">;
type ApiSearchReader = Pick<ApiSearchService, "search">;
type ApiRecordsReader = Pick<ApiRecordsReadService, "list" | "listAll" | "get" | "groupBy">;

interface ReadBackendServiceOptions {
  workspace?: string;
}

export class ReadBackendService implements SearchReadBackend, RecordsReadBackend {
  constructor(
    private readonly dbConfigResolver: DbConfigResolver,
    private readonly apiSearch: ApiSearchReader,
    private readonly apiRecords?: ApiRecordsReader,
    private readonly dbReads: ReadBackendDbReads = {},
    private readonly options: ReadBackendServiceOptions = {},
  ) {}

  async runSearch(options: SearchOptions): Promise<SearchResponse> {
    return this.runRead(
      () => this.apiSearch.search(options),
      this.dbReads.search ? (target) => this.dbReads.search!.search(target, options) : undefined,
    );
  }

  async list(object: string, options: ListOptions = {}): Promise<ListResponse> {
    return this.runRead(
      () => this.requireApiRecords().list(object, options),
      this.dbReads.records?.list
        ? (target) => this.dbReads.records!.list!(target, object, options)
        : undefined,
    );
  }

  async listAll(object: string, options: ListOptions = {}): Promise<ListResponse> {
    return this.runRead(
      () => this.requireApiRecords().listAll(object, options),
      this.dbReads.records?.listAll
        ? (target) => this.dbReads.records!.listAll!(target, object, options)
        : undefined,
    );
  }

  async get(object: string, id: string, options?: GetOptions): Promise<unknown> {
    return this.runRead(
      () => this.requireApiRecords().get(object, id, options),
      this.dbReads.records?.get
        ? (target) => this.dbReads.records!.get!(target, object, id, options)
        : undefined,
    );
  }

  async groupBy(object: string, payload?: unknown, params?: GroupByParams): Promise<unknown> {
    return this.runRead(
      () => this.requireApiRecords().groupBy(object, payload, params),
      this.dbReads.records?.groupBy
        ? (target) => this.dbReads.records!.groupBy!(target, object, payload, params)
        : undefined,
    );
  }

  private async runRead<T>(
    apiRead: () => Promise<T>,
    dbRead?: (target: ResolvedDbConfig) => Promise<T>,
  ): Promise<T> {
    const target = await this.dbConfigResolver.resolve({
      workspace: this.options.workspace,
    });

    if (target.mode !== "db" || !dbRead) {
      return apiRead();
    }

    try {
      return await dbRead(target);
    } catch (error) {
      if (error instanceof UnsupportedDbReadError || error instanceof DbConnectionError) {
        return apiRead();
      }

      throw error;
    }
  }

  private requireApiRecords(): ApiRecordsReader {
    if (!this.apiRecords) {
      throw new Error("ReadBackendService requires API records support for records reads.");
    }

    return this.apiRecords;
  }
}
