import { ApiService } from "../../api/services/api.service";
import { CliError } from "../../errors/cli-error";
import { capitalize, singularize } from "../../shared/parse";

export interface ListOptions {
  limit?: number;
  cursor?: string;
  filter?: string;
  sort?: string;
  order?: string;
  include?: string;
  params?: Record<string, string[]>;
}

export interface PageInfo {
  hasNextPage?: boolean;
  endCursor?: string;
}

export interface ListResponse {
  data: unknown[];
  totalCount?: number;
  pageInfo?: PageInfo;
}

interface BulkMutationOptions {
  filter: string;
  include?: string;
}

export class RecordsService {
  constructor(private api: ApiService) {}

  async list(object: string, options: ListOptions = {}): Promise<ListResponse> {
    const params: Record<string, string | string[]> = {};
    if (options.limit) params.limit = String(options.limit);
    if (options.cursor) params.starting_after = options.cursor;
    if (options.sort) params.order_by = formatOrderBy(options.sort, options.order);
    if (options.include) params.depth = "1";
    if (options.filter) params.filter = options.filter;
    if (options.params) {
      for (const [key, values] of Object.entries(options.params)) {
        params[key] = values.length === 1 ? values[0] : values;
      }
    }

    const response = await this.api.get(`/rest/${object}`, { params });
    const payload = response.data as any;
    const dataSection = payload?.data ?? {};
    const records = extractArray(dataSection, object);
    return {
      data: records,
      totalCount: payload?.totalCount,
      pageInfo: payload?.pageInfo,
    };
  }

  async listAll(object: string, options: ListOptions = {}): Promise<ListResponse> {
    const all: unknown[] = [];
    let cursor = options.cursor ?? "";
    let pageInfo: PageInfo | undefined;
    let totalCount: number | undefined;

    while (true) {
      const response = await this.list(object, { ...options, cursor });
      all.push(...response.data);
      pageInfo = response.pageInfo;
      totalCount = response.totalCount ?? totalCount;
      if (!pageInfo?.hasNextPage || !pageInfo?.endCursor) {
        break;
      }
      cursor = pageInfo.endCursor;
    }

    return { data: all, totalCount, pageInfo };
  }

  async get(object: string, id: string, options?: { include?: string }): Promise<unknown> {
    const params: Record<string, string> = {};
    if (options?.include) {
      params.depth = "1";
    }
    const response = await this.api.get(`/rest/${object}/${id}`, { params });
    const payload = response.data as any;
    const dataSection = payload?.data ?? {};
    const singular = singularize(object);
    return dataSection[singular] ?? dataSection[object] ?? firstValue(dataSection);
  }

  async create(object: string, data: Record<string, unknown>): Promise<unknown> {
    const response = await this.api.post(`/rest/${object}`, data);
    const payload = response.data as any;
    const dataSection = payload?.data ?? {};
    const key = `create${capitalize(singularize(object))}`;
    return dataSection[key] ?? firstValue(dataSection);
  }

  async update(object: string, id: string, data: Record<string, unknown>): Promise<unknown> {
    const response = await this.api.patch(`/rest/${object}/${id}`, data);
    const payload = response.data as any;
    const dataSection = payload?.data ?? {};
    const key = `update${capitalize(singularize(object))}`;
    return dataSection[key] ?? firstValue(dataSection);
  }

  async delete(object: string, id: string): Promise<unknown> {
    const response = await this.api.delete(`/rest/${object}/${id}`, {
      params: { soft_delete: "true" },
    });
    return response.data ?? null;
  }

  async destroy(object: string, id: string): Promise<unknown> {
    const response = await this.api.delete(`/rest/${object}/${id}`);
    return response.data ?? null;
  }

  async restore(object: string, id: string): Promise<unknown> {
    const response = await this.api.patch(`/rest/restore/${object}/${id}`);
    return response.data ?? null;
  }

  async batchCreate(object: string, records: Record<string, unknown>[]): Promise<unknown> {
    const response = await this.api.post(`/rest/batch/${object}`, records);
    return response.data ?? null;
  }

  async batchUpdate(object: string, records: Record<string, unknown>[]): Promise<unknown> {
    const updates = await Promise.all(
      records.map((record) => {
        const id = extractRecordId(record);
        const { id: _ignoredId, ...data } = record;
        return this.update(object, id, data);
      }),
    );

    return updates;
  }

  async updateMany(
    object: string,
    data: Record<string, unknown>,
    options: BulkMutationOptions,
  ): Promise<unknown> {
    const response = await this.api.patch(`/rest/${object}`, data, {
      params: buildBulkParams(options),
    });
    return response.data ?? null;
  }

  async batchDelete(object: string, ids: string[]): Promise<unknown> {
    const filter = `id[in]:[${ids.join(",")}]`;
    const response = await this.api.delete(`/rest/${object}`, {
      params: {
        filter,
        soft_delete: "true",
      },
    });
    return response.data ?? null;
  }

  async restoreMany(object: string, options: BulkMutationOptions): Promise<unknown> {
    const response = await this.api.patch(`/rest/restore/${object}`, undefined, {
      params: buildBulkParams(options),
    });
    return response.data ?? null;
  }

  async destroyMany(object: string, options: BulkMutationOptions): Promise<unknown> {
    const response = await this.api.delete(`/rest/${object}`, {
      params: buildBulkParams(options),
    });
    return response.data ?? null;
  }

  async groupBy(
    object: string,
    payload?: unknown,
    params?: Record<string, string[]>,
  ): Promise<unknown> {
    const path = `/rest/${object}/groupBy`;
    const response = await this.api.get(path, {
      params: {
        ...flattenParams(params),
        ...serializeGroupByPayload(payload),
      },
    });
    return response.data ?? null;
  }

  async findDuplicates(object: string, payload: unknown): Promise<unknown> {
    const response = await this.api.post(`/rest/${object}/duplicates`, payload);
    return response.data ?? null;
  }

  async merge(object: string, payload: unknown): Promise<unknown> {
    const response = await this.api.patch(`/rest/${object}/merge`, payload);
    return response.data ?? null;
  }
}

function extractArray(dataSection: Record<string, unknown>, object: string): unknown[] {
  const raw = dataSection?.[object];
  if (Array.isArray(raw)) return raw;
  for (const value of Object.values(dataSection)) {
    if (Array.isArray(value)) return value as unknown[];
  }
  return [];
}

function firstValue(dataSection: Record<string, unknown>): unknown {
  const values = Object.values(dataSection);
  if (values.length === 0) return undefined;
  return values[0];
}

function flattenParams(
  params?: Record<string, string[]>,
): Record<string, string | string[]> | undefined {
  if (!params) return undefined;
  const out: Record<string, string | string[]> = {};
  for (const [key, values] of Object.entries(params)) {
    out[key] = values.length === 1 ? values[0] : values;
  }
  return out;
}

function formatOrderBy(sort: string, order?: string): string {
  const direction = order?.toLowerCase() === "desc" ? "DescNullsLast" : "AscNullsFirst";

  return `${sort}[${direction}]`;
}

function extractRecordId(record: Record<string, unknown>): string {
  const id = record.id;

  if (typeof id !== "string" || id.length === 0) {
    throw new CliError(
      "Batch update requires every record to include a string id.",
      "INVALID_ARGUMENTS",
    );
  }

  return id;
}

function serializeGroupByPayload(payload?: unknown): Record<string, string | string[]> {
  if (Array.isArray(payload)) {
    return {
      group_by: JSON.stringify(payload),
    };
  }

  if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
    return {};
  }

  const serialized: Record<string, string | string[]> = {};

  for (const [key, value] of Object.entries(payload)) {
    if (value === undefined) {
      continue;
    }

    if (key === "groupBy" || key === "group_by") {
      serialized.group_by = JSON.stringify(value);
      continue;
    }

    if (key === "orderBy" || key === "order_by") {
      serialized.order_by = JSON.stringify(value);
      continue;
    }

    if (key === "viewId" || key === "view_id") {
      serialized.view_id = String(value);
      continue;
    }

    if (key === "includeRecordsSample" || key === "include_records_sample") {
      serialized.include_records_sample = String(value);
      continue;
    }

    if (key === "filter" && isRecord(value)) {
      serialized.filter = serializeFilter(value);
      continue;
    }

    if (Array.isArray(value)) {
      serialized[key] = value.map(String);
      continue;
    }

    serialized[key] = typeof value === "object" ? JSON.stringify(value) : String(value);
  }

  return serialized;
}

function serializeFilter(filter: Record<string, unknown>): string {
  return Object.entries(filter)
    .map(([field, condition]) => {
      if (!isRecord(condition)) {
        return `${field}[eq]:${String(condition)}`;
      }

      const clauses = Object.entries(condition).map(([operator, value]) => {
        if (Array.isArray(value)) {
          return `${field}[${operator}]:[${value.map(String).join(",")}]`;
        }

        return `${field}[${operator}]:${String(value)}`;
      });

      return clauses.join(";");
    })
    .join(";");
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function buildBulkParams(options: BulkMutationOptions): Record<string, string> {
  const params: Record<string, string> = {
    filter: options.filter,
  };

  if (options.include) {
    params.depth = "1";
  }

  return params;
}
