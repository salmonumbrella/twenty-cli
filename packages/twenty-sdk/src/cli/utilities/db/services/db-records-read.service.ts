import type { Client } from "pg";
import type { MetadataService } from "../../metadata/services/metadata.service";
import { DbConnectionError, UnsupportedDbReadError } from "../../readbackend/types";
import type {
  GetOptions,
  GroupByParams,
  ListOptions,
  ListResponse,
} from "../../records/services/api-records-read.service";
import type { ResolvedDbConfig } from "./db-config-resolver.service";
import { DbConnectionService } from "./db-connection.service";
import type { DbFilterClause } from "./db-filter-compiler.service";
import { DbFilterCompilerService } from "./db-filter-compiler.service";
import type { DbRelationPlan } from "./db-metadata-planner.service";
import { DbMetadataPlannerService } from "./db-metadata-planner.service";

type MetadataClient = Pick<MetadataService, "listObjects" | "getObject">;
type MetadataPlanner = Pick<DbMetadataPlannerService, "planObject">;
type FilterCompiler = Pick<DbFilterCompilerService, "compile">;
type DbConnector = Pick<DbConnectionService, "connect">;

interface DbCountRow {
  totalCount: number | string;
}

interface DbRecordRow {
  rowData: unknown;
}

interface DbGroupByRow extends Record<string, unknown> {
  countNotEmptyId?: number | string;
}

interface SupportedDbGroupByRequest {
  fields: string[];
  aggregateField?: "countNotEmptyId";
  filter?: string;
  limit: number;
}

const DEFAULT_LIST_LIMIT = 20;
const DEFAULT_GROUP_BY_LIMIT = 50;
const DB_CURSOR_PATTERN = /^db:(\d+)$/;
const SIMPLE_IDENTIFIER_PATTERN = /^[A-Za-z_][A-Za-z0-9_]*$/;

export class DbRecordsReadService {
  private readonly metadataPlanner: MetadataPlanner;
  private readonly filterCompiler: FilterCompiler;
  private readonly dbConnectionService: DbConnector;

  constructor(
    metadataService: MetadataClient,
    metadataPlanner?: MetadataPlanner,
    filterCompiler?: FilterCompiler,
    dbConnectionService?: DbConnector,
  ) {
    this.metadataPlanner =
      metadataPlanner ?? new DbMetadataPlannerService(metadataService as MetadataService);
    this.filterCompiler = filterCompiler ?? new DbFilterCompilerService();
    this.dbConnectionService = dbConnectionService ?? new DbConnectionService();
  }

  async list(
    target: ResolvedDbConfig,
    object: string,
    options: ListOptions = {},
  ): Promise<ListResponse> {
    assertSupportedListOptions(options);

    const connectionOptions = resolveConnectionOptions(target);
    const plan = await this.metadataPlanner.planObject(object, { include: options.include });
    const clauses = this.filterCompiler.compile(options.filter);
    const limit = resolveLimit(options.limit);
    const offset = resolveOffset(options.cursor);
    const client = await this.connectClient(connectionOptions, "list");

    try {
      const totalCount = await this.queryTotalCount(client, plan.tableName, clauses);
      const data = await this.queryRows(client, plan.tableName, clauses, {
        limit,
        offset,
        sort: options.sort,
        order: options.order,
      });
      const lastOffset = offset + data.length - 1;

      return {
        data,
        totalCount,
        pageInfo: {
          hasNextPage: offset + data.length < totalCount,
          endCursor: data.length > 0 ? encodeCursor(lastOffset) : undefined,
        },
      };
    } finally {
      await client.end();
    }
  }

  async listAll(
    target: ResolvedDbConfig,
    object: string,
    options: ListOptions = {},
  ): Promise<ListResponse> {
    const all: unknown[] = [];
    let cursor = options.cursor ?? "";
    let totalCount: number | undefined;
    let pageInfo: ListResponse["pageInfo"];

    while (true) {
      const response = await this.list(target, object, { ...options, cursor });
      all.push(...response.data);
      totalCount = response.totalCount ?? totalCount;
      pageInfo = response.pageInfo;

      if (!pageInfo?.hasNextPage || !pageInfo.endCursor) {
        break;
      }

      cursor = pageInfo.endCursor;
    }

    return {
      data: all,
      totalCount,
      pageInfo,
    };
  }

  async get(
    target: ResolvedDbConfig,
    object: string,
    id: string,
    options?: GetOptions,
  ): Promise<unknown> {
    const connectionOptions = resolveConnectionOptions(target);
    const plan = await this.metadataPlanner.planObject(object, { include: options?.include });
    const client = await this.connectClient(connectionOptions, "get");

    try {
      const baseRecord = await this.queryRecordById(client, plan.tableName, id);

      if (!baseRecord || plan.includes.length === 0) {
        return baseRecord;
      }

      return this.hydrateIncludes(client, baseRecord, plan.includes);
    } finally {
      await client.end();
    }
  }

  async groupBy(
    target: ResolvedDbConfig,
    object: string,
    payload?: unknown,
    params?: GroupByParams,
  ): Promise<unknown> {
    const request = parseSupportedGroupByRequest(payload, params);
    const connectionOptions = resolveConnectionOptions(target);
    const plan = await this.metadataPlanner.planObject(object);
    const clauses = this.filterCompiler.compile(request.filter);
    const client = await this.connectClient(connectionOptions, "groupBy");

    try {
      return await this.queryGroupByRows(client, plan.tableName, request, clauses);
    } finally {
      await client.end();
    }
  }

  private async queryTotalCount(
    client: Pick<Client, "query">,
    tableName: string,
    clauses: DbFilterClause[],
  ): Promise<number> {
    const { sql, params } = buildWhereClause(clauses);
    const result = await client.query<DbCountRow>(
      `
        select count(*)::text as "totalCount"
        from ${quoteIdentifier(tableName)} as t
        ${sql}
      `,
      params,
    );

    return toNumber(result.rows[0]?.totalCount);
  }

  private async queryRows(
    client: Pick<Client, "query">,
    tableName: string,
    clauses: DbFilterClause[],
    options: {
      limit: number;
      offset: number;
      sort?: string;
      order?: string;
    },
  ): Promise<unknown[]> {
    const { sql: whereSql, params } = buildWhereClause(clauses);
    const orderBy = buildOrderByClause(options.sort, options.order);
    const result = await client.query<DbRecordRow>(
      `
        select to_jsonb(t) as "rowData"
        from ${quoteIdentifier(tableName)} as t
        ${whereSql}
        ${orderBy}
        limit $${params.length + 1}
        offset $${params.length + 2}
      `,
      [...params, options.limit, options.offset],
    );

    return result.rows.map((row) => row.rowData);
  }

  private async queryRecordById(
    client: Pick<Client, "query">,
    tableName: string,
    id: string,
  ): Promise<Record<string, unknown> | undefined> {
    const result = await client.query<DbRecordRow>(
      `
        select to_jsonb(t) as "rowData"
        from ${quoteIdentifier(tableName)} as t
        where t."id" = $1
        limit 1
      `,
      [id],
    );

    return asRecord(result.rows[0]?.rowData);
  }

  private async queryGroupByRows(
    client: Pick<Client, "query">,
    tableName: string,
    request: SupportedDbGroupByRequest,
    clauses: DbFilterClause[],
  ): Promise<Array<Record<string, unknown>>> {
    const { sql: whereSql, params } = buildWhereClause(clauses);
    const fields = request.fields.map((field) => quoteColumn(field));
    const selectSql = fields
      .map((field, index) => `${field} as ${quoteIdentifier(getGroupByAlias(index))}`)
      .join(", ");
    const aggregateSql = request.aggregateField
      ? `, count(*)::text as "${request.aggregateField}"`
      : "";
    const groupBySql = fields.join(", ");
    const orderBySql = fields.map((field) => `${field} asc nulls first`).join(", ");
    const result = await client.query<DbGroupByRow>(
      `
        select ${selectSql}${aggregateSql}
        from ${quoteIdentifier(tableName)} as t
        ${whereSql}
        group by ${groupBySql}
        order by ${orderBySql}
        limit $${params.length + 1}
      `,
      [...params, request.limit],
    );

    return result.rows.map((row) => {
      const group: Record<string, unknown> = {
        groupByDimensionValues: request.fields.map((_, index) => row[getGroupByAlias(index)]),
      };

      if (request.aggregateField) {
        group[request.aggregateField] = row[request.aggregateField] ?? null;
      }

      return group;
    });
  }

  private async hydrateIncludes(
    client: Pick<Client, "query">,
    baseRecord: Record<string, unknown>,
    includes: DbRelationPlan[],
  ): Promise<Record<string, unknown>> {
    const hydratedRecord = { ...baseRecord };

    for (const relationPlan of includes) {
      const relationIdFieldName = relationPlan.joinColumnName;

      if (!hasOwn(hydratedRecord, relationIdFieldName)) {
        throw new UnsupportedDbReadError(
          `DB get does not support include hydration for relation ${JSON.stringify(relationPlan.relationName)} because ${JSON.stringify(relationIdFieldName)} is not present on the base row.`,
        );
      }

      const relationId = hydratedRecord[relationIdFieldName];

      if (typeof relationId !== "string" || relationId.length === 0) {
        hydratedRecord[relationPlan.relationName] = null;
        continue;
      }

      const relatedRecord = await this.queryRecordById(client, relationPlan.tableName, relationId);

      hydratedRecord[relationPlan.relationName] = relatedRecord ?? null;
    }

    return hydratedRecord;
  }

  private async connectClient(
    connectionOptions: ReturnType<typeof resolveConnectionOptions>,
    operation: "list" | "get" | "groupBy",
  ) {
    try {
      return await this.dbConnectionService.connect(connectionOptions);
    } catch (error) {
      throw new DbConnectionError(resolveRecordsOperationUnavailableMessage(operation), {
        cause: error,
      });
    }
  }
}

function assertSupportedListOptions(options: ListOptions): void {
  if (options.include) {
    throw new UnsupportedDbReadError("DB list does not support include hydration yet.");
  }

  if (options.params && Object.keys(options.params).length > 0) {
    throw new UnsupportedDbReadError("DB list does not support custom query params.");
  }
}

function resolveConnectionOptions(target: ResolvedDbConfig) {
  if (!target.databaseUrl) {
    throw new UnsupportedDbReadError("DB records read requires a resolved database URL.");
  }

  return {
    databaseUrl: target.databaseUrl,
  };
}

function parseSupportedGroupByRequest(
  payload?: unknown,
  params?: GroupByParams,
): SupportedDbGroupByRequest {
  const payloadRecord = asRecord(payload);

  if (!payloadRecord) {
    throw new UnsupportedDbReadError("DB groupBy requires a payload with groupBy fields.");
  }

  assertSupportedGroupByPayloadKeys(payloadRecord);
  assertSupportedGroupByParamsKeys(params);

  const fields = parseGroupByFields(payloadRecord.groupBy);
  const payloadFilter = parseGroupByPayloadFilter(payloadRecord.filter);
  const paramsFilter = parseSingleStringParam(params, "filter");

  if (payloadFilter && paramsFilter) {
    throw new UnsupportedDbReadError("DB groupBy does not support both payload and param filters.");
  }

  return {
    fields,
    aggregateField: parseGroupByAggregate(params),
    filter: payloadFilter ?? paramsFilter,
    limit: parseGroupByLimit(params),
  };
}

function assertSupportedGroupByPayloadKeys(payload: Record<string, unknown>): void {
  for (const [key, value] of Object.entries(payload)) {
    if (value === undefined) {
      continue;
    }

    if (key === "groupBy" || key === "filter") {
      continue;
    }

    throw new UnsupportedDbReadError(
      `DB groupBy does not support payload key ${JSON.stringify(key)}.`,
    );
  }
}

function assertSupportedGroupByParamsKeys(params?: GroupByParams): void {
  if (!params) {
    return;
  }

  for (const [key, value] of Object.entries(params)) {
    if (!Array.isArray(value) || value.length === 0) {
      continue;
    }

    if (key === "filter" || key === "limit" || key === "aggregate") {
      continue;
    }

    throw new UnsupportedDbReadError(`DB groupBy does not support param ${JSON.stringify(key)}.`);
  }
}

function parseGroupByFields(groupBy: unknown): string[] {
  if (!Array.isArray(groupBy) || groupBy.length === 0) {
    throw new UnsupportedDbReadError("DB groupBy requires a non-empty groupBy array.");
  }

  return groupBy.map((entry) => parseGroupByFieldEntry(entry));
}

function parseGroupByFieldEntry(entry: unknown): string {
  if (!entry || typeof entry !== "object" || Array.isArray(entry)) {
    throw new UnsupportedDbReadError("DB groupBy only supports simple field entries.");
  }

  const fields = Object.entries(entry as Record<string, unknown>).filter(
    ([, value]) => value !== undefined,
  );

  if (fields.length !== 1) {
    throw new UnsupportedDbReadError("DB groupBy only supports one field per groupBy entry.");
  }

  const [fieldName, enabled] = fields[0];

  if (enabled !== true) {
    throw new UnsupportedDbReadError(
      `DB groupBy does not support advanced field definition for ${JSON.stringify(fieldName)}.`,
    );
  }

  return fieldName;
}

function parseGroupByPayloadFilter(filter: unknown): string | undefined {
  if (filter === undefined) {
    return undefined;
  }

  if (typeof filter !== "string") {
    throw new UnsupportedDbReadError("DB groupBy only supports string filters.");
  }

  return filter;
}

function parseGroupByAggregate(params?: GroupByParams): "countNotEmptyId" | undefined {
  const aggregate = parseSingleStringParam(params, "aggregate");

  if (aggregate === undefined) {
    return undefined;
  }

  const normalizedFields = normalizeAggregateFields(aggregate);

  if (normalizedFields.length !== 1 || normalizedFields[0] !== "countNotEmptyId") {
    throw new UnsupportedDbReadError(
      `DB groupBy only supports aggregate ${JSON.stringify("countNotEmptyId")}.`,
    );
  }

  return "countNotEmptyId";
}

function parseGroupByLimit(params?: GroupByParams): number {
  const rawLimit = parseSingleStringParam(params, "limit");

  if (rawLimit === undefined) {
    return DEFAULT_GROUP_BY_LIMIT;
  }

  if (!/^\d+$/.test(rawLimit)) {
    throw new UnsupportedDbReadError(
      `DB groupBy does not support limit ${JSON.stringify(rawLimit)}.`,
    );
  }

  const limit = Number(rawLimit);

  if (!Number.isFinite(limit) || limit < 1) {
    throw new UnsupportedDbReadError(
      `DB groupBy does not support limit ${JSON.stringify(rawLimit)}.`,
    );
  }

  return Math.floor(limit);
}

function parseSingleStringParam(
  params: GroupByParams | undefined,
  key: string,
): string | undefined {
  const values = params?.[key];

  if (!values || values.length === 0) {
    return undefined;
  }

  if (values.length !== 1) {
    throw new UnsupportedDbReadError(
      `DB groupBy does not support multiple values for ${JSON.stringify(key)}.`,
    );
  }

  return values[0];
}

function normalizeAggregateFields(raw: string): string[] {
  const trimmed = raw.trim();

  if (!trimmed) {
    return [];
  }

  if (trimmed === "totalCount") {
    return ["countNotEmptyId"];
  }

  if (!trimmed.startsWith("[")) {
    return [trimmed];
  }

  let parsed: unknown;

  try {
    parsed = JSON.parse(trimmed);
  } catch {
    throw new UnsupportedDbReadError(
      `DB groupBy could not parse aggregate ${JSON.stringify(raw)} as a JSON string array.`,
    );
  }

  if (!Array.isArray(parsed) || parsed.some((entry) => typeof entry !== "string")) {
    throw new UnsupportedDbReadError(
      `DB groupBy could not parse aggregate ${JSON.stringify(raw)} as a JSON string array.`,
    );
  }

  return parsed.map((entry) => (entry === "totalCount" ? "countNotEmptyId" : entry));
}

function resolveLimit(limit?: number): number {
  if (!Number.isFinite(limit) || !limit || limit < 1) {
    return DEFAULT_LIST_LIMIT;
  }

  return Math.floor(limit);
}

function resolveOffset(cursor?: string): number {
  if (!cursor) {
    return 0;
  }

  const match = DB_CURSOR_PATTERN.exec(cursor);

  if (!match) {
    throw new UnsupportedDbReadError(`DB list does not support cursor ${JSON.stringify(cursor)}.`);
  }

  return Number(match[1]) + 1;
}

function buildWhereClause(clauses: DbFilterClause[]): { sql: string; params: unknown[] } {
  if (clauses.length === 0) {
    return { sql: "", params: [] };
  }

  const params: unknown[] = [];
  const fragments = clauses.map((clause) => compileFilterClause(clause, params));

  return {
    sql: `where ${fragments.join(" and ")}`,
    params,
  };
}

function compileFilterClause(clause: DbFilterClause, params: unknown[]): string {
  const field = quoteColumn(clause.field);

  switch (clause.operator) {
    case "eq":
      if (clause.value === null) {
        return `${field} is null`;
      }
      params.push(clause.value);
      return `${field} = $${params.length}`;
    case "neq":
      if (clause.value === null) {
        return `${field} is not null`;
      }
      params.push(clause.value);
      return `${field} <> $${params.length}`;
    case "gt":
      params.push(clause.value);
      return `${field} > $${params.length}`;
    case "gte":
      params.push(clause.value);
      return `${field} >= $${params.length}`;
    case "lt":
      params.push(clause.value);
      return `${field} < $${params.length}`;
    case "lte":
      params.push(clause.value);
      return `${field} <= $${params.length}`;
    case "in":
      if (!Array.isArray(clause.value)) {
        throw new UnsupportedDbReadError(
          `DB list requires array values for [in] filters on ${JSON.stringify(clause.field)}.`,
        );
      }
      params.push(clause.value);
      return `${field} = any($${params.length})`;
    case "is":
      if (clause.value === null || clause.value === "null") {
        return `${field} is null`;
      }
      if (clause.value === "not_null") {
        return `${field} is not null`;
      }
      throw new UnsupportedDbReadError(
        `DB list does not support [is] value ${JSON.stringify(clause.value)}.`,
      );
    default:
      throw new UnsupportedDbReadError(
        `DB list does not support filter operator ${JSON.stringify(clause.operator)}.`,
      );
  }
}

function buildOrderByClause(sort?: string, order?: string): string {
  const direction = order?.toLowerCase() === "desc" ? "desc" : "asc";
  const nullsOrdering = direction === "desc" ? "nulls last" : "nulls first";

  if (!sort) {
    return `order by ${quoteColumn("id")} asc`;
  }

  const sortColumn = quoteColumn(sort);

  if (sort === "id") {
    return `order by ${sortColumn} ${direction} ${nullsOrdering}`;
  }

  return `order by ${sortColumn} ${direction} ${nullsOrdering}, ${quoteColumn("id")} asc`;
}

function quoteColumn(column: string): string {
  if (!SIMPLE_IDENTIFIER_PATTERN.test(column)) {
    throw new UnsupportedDbReadError(`DB list does not support field ${JSON.stringify(column)}.`);
  }

  return `t.${quoteIdentifier(column)}`;
}

function quoteIdentifier(value: string): string {
  return `"${value.replace(/"/g, '""')}"`;
}

function encodeCursor(offset: number): string {
  return `db:${offset}`;
}

function getGroupByAlias(index: number): string {
  return `group_${index}`;
}

function resolveRecordsOperationUnavailableMessage(operation: "list" | "get" | "groupBy"): string {
  switch (operation) {
    case "get":
      return "DB records get is unavailable.";
    case "groupBy":
      return "DB records groupBy is unavailable.";
    case "list":
    default:
      return "DB records list is unavailable.";
  }
}

function asRecord(value: unknown): Record<string, unknown> | undefined {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return undefined;
  }

  return value as Record<string, unknown>;
}

function hasOwn(record: Record<string, unknown>, key: string): boolean {
  return Object.prototype.hasOwnProperty.call(record, key);
}

function toNumber(value: number | string | undefined): number {
  if (typeof value === "number") {
    return value;
  }

  if (typeof value === "string" && value.trim().length > 0) {
    return Number(value);
  }

  return 0;
}
