import type { Client } from "pg";
import type { ObjectMetadata, MetadataService } from "../../metadata/services/metadata.service";
import { DbConnectionError, UnsupportedDbReadError } from "../../readbackend/types";
import type {
  SearchOptions,
  SearchResponse,
  SearchResult,
} from "../../search/services/api-search.service";
import type { ResolvedDbConfig } from "./db-config-resolver.service";
import { DbConnectionService } from "./db-connection.service";
import { DbMetadataPlannerService } from "./db-metadata-planner.service";

type MetadataClient = Pick<MetadataService, "listObjects">;
type MetadataPlanner = Pick<DbMetadataPlannerService, "planObject">;
type DbConnector = Pick<DbConnectionService, "connect">;

interface DbSearchRow {
  recordId: string;
  rowData: unknown;
  tsRankCD: number | string;
  tsRank: number | string;
}

export class DbSearchService {
  private readonly metadataPlanner: MetadataPlanner;
  private readonly dbConnectionService: DbConnector;

  constructor(
    private readonly metadataService: MetadataClient,
    metadataPlanner?: MetadataPlanner,
    dbConnectionService?: DbConnector,
  ) {
    this.metadataPlanner =
      metadataPlanner ?? new DbMetadataPlannerService(metadataService as MetadataService);
    this.dbConnectionService = dbConnectionService ?? new DbConnectionService();
  }

  async search(target: ResolvedDbConfig, options: SearchOptions): Promise<SearchResponse> {
    if (options.filter) {
      throw new UnsupportedDbReadError("DB search does not support filters.");
    }

    const query = options.query.trim();

    if (!query) {
      return emptySearchResponse();
    }

    const connectionOptions = resolveConnectionOptions(target);
    const startOffset = resolveStartOffset(options.after);
    const objectNames = await this.resolveSearchObjects(options);
    const limit = options.limit ?? 20;
    const fetchLimit = startOffset + limit + 1;

    if (objectNames.length === 0) {
      return emptySearchResponse();
    }

    const client = await this.connectClient(connectionOptions);

    try {
      const resultEntries: Array<{ result: SearchResult; index: number }> = [];
      let index = 0;

      for (const objectName of objectNames) {
        try {
          const plan = await this.metadataPlanner.planObject(objectName);
          const rows = await this.queryObject(client, plan.tableName, query, fetchLimit);
          const objectNameSingular = plan.objectMetadata.nameSingular ?? objectName;
          const objectLabelSingular = getObjectLabel(plan.objectMetadata, objectNameSingular);

          for (const row of rows) {
            resultEntries.push({
              index,
              result: {
                recordId: row.recordId,
                objectNameSingular,
                objectLabelSingular,
                label: deriveLabel(row.rowData, row.recordId),
                imageUrl: deriveImageUrl(row.rowData),
                tsRankCD: toNumber(row.tsRankCD),
                tsRank: toNumber(row.tsRank),
              },
            });
            index += 1;
          }
        } catch (error) {
          if (error instanceof UnsupportedDbReadError) {
            continue;
          }

          throw error;
        }
      }

      const sortedEntries = resultEntries.sort((left, right) => {
        if (right.result.tsRankCD !== left.result.tsRankCD) {
          return right.result.tsRankCD - left.result.tsRankCD;
        }

        if (right.result.tsRank !== left.result.tsRank) {
          return right.result.tsRank - left.result.tsRank;
        }

        return left.index - right.index;
      });
      const pagedEntries = sortedEntries.slice(startOffset, startOffset + limit);
      const data = pagedEntries.map((entry, pageIndex) => ({
        ...entry.result,
        cursor: encodeCursor(startOffset + pageIndex),
      }));
      const endCursor = data[data.length - 1]?.cursor;

      return {
        data,
        pageInfo: {
          hasNextPage: sortedEntries.length > startOffset + limit,
          endCursor,
        },
      };
    } finally {
      await client.end();
    }
  }

  private async resolveSearchObjects(options: SearchOptions): Promise<string[]> {
    if (options.objects?.length) {
      return applyExcludedObjects(options.objects, options.excludeObjects);
    }

    const objects = await this.metadataService.listObjects();
    const names = objects
      .map((objectMetadata) => objectMetadata.nameSingular)
      .filter((value): value is string => typeof value === "string" && value.length > 0);

    return applyExcludedObjects(names, options.excludeObjects);
  }

  private async queryObject(
    client: Pick<Client, "query">,
    tableName: string,
    query: string,
    limit: number,
  ): Promise<DbSearchRow[]> {
    const sql = `
      with searchable_rows as (
        select
          t.id::text as "recordId",
          jsonb_strip_nulls(to_jsonb(t)) as "rowData",
          to_tsvector('simple', coalesce(jsonb_strip_nulls(to_jsonb(t))::text, '')) as document,
          plainto_tsquery('simple', $1) as query
        from ${quoteIdentifier(tableName)} as t
      )
      select
        "recordId",
        "rowData",
        ts_rank_cd(document, query) as "tsRankCD",
        ts_rank(document, query) as "tsRank"
      from searchable_rows
      where document @@ query
      order by "tsRankCD" desc, "tsRank" desc, "recordId" asc
      limit $2
    `;
    const result = await client.query(sql, [query, limit]);

    return result.rows as DbSearchRow[];
  }

  private async connectClient(connectionOptions: ReturnType<typeof resolveConnectionOptions>) {
    try {
      return await this.dbConnectionService.connect(connectionOptions);
    } catch (error) {
      throw new DbConnectionError("DB search is unavailable.", { cause: error });
    }
  }
}

function resolveConnectionOptions(target: ResolvedDbConfig) {
  const databaseUrl = target.databaseUrl;

  if (!databaseUrl) {
    throw new UnsupportedDbReadError("DB search requires a resolved database URL.");
  }

  return {
    databaseUrl,
  };
}

function applyExcludedObjects(objects: string[], excludeObjects?: string[]): string[] {
  const excluded = new Set((excludeObjects ?? []).map((value) => value.toLowerCase()));
  const seen = new Set<string>();
  const results: string[] = [];

  for (const objectName of objects) {
    const normalized = objectName.toLowerCase();

    if (excluded.has(normalized) || seen.has(normalized)) {
      continue;
    }

    seen.add(normalized);
    results.push(objectName);
  }

  return results;
}

function resolveStartOffset(after?: string): number {
  if (!after) {
    return 0;
  }

  const match = /^db:(\d+)$/.exec(after);

  if (!match) {
    throw new UnsupportedDbReadError(`DB search does not support cursor ${JSON.stringify(after)}.`);
  }

  return Number(match[1]) + 1;
}

function encodeCursor(offset: number): string {
  return `db:${offset}`;
}

function emptySearchResponse(): SearchResponse {
  return {
    data: [],
    pageInfo: {
      hasNextPage: false,
      endCursor: undefined,
    },
  };
}

function quoteIdentifier(value: string): string {
  return `"${value.replace(/"/g, '""')}"`;
}

function getObjectLabel(objectMetadata: ObjectMetadata, fallback: string): string {
  const labelSingular = objectMetadata.labelSingular;

  if (typeof labelSingular === "string" && labelSingular.trim().length > 0) {
    return labelSingular;
  }

  return fallback.charAt(0).toUpperCase() + fallback.slice(1);
}

function deriveLabel(rowData: unknown, fallback: string): string {
  const record = asRecord(rowData);

  if (!record) {
    return fallback;
  }

  const directString = firstString(record, [
    "label",
    "displayName",
    "fullName",
    "title",
    "subject",
    "name",
  ]);

  if (directString) {
    return directString;
  }

  const nestedName = deriveNameValue(record.name);

  if (nestedName) {
    return nestedName;
  }

  const rootName = joinNameParts(record);

  if (rootName) {
    return rootName;
  }

  const fallbackString = firstString(record, [
    "email",
    "primaryEmail",
    "phone",
    "linkedinLinkPrimary",
    "domainName",
  ]);

  return fallbackString ?? fallback;
}

function deriveImageUrl(rowData: unknown): string | null {
  const record = asRecord(rowData);

  if (!record) {
    return null;
  }

  return (
    firstString(record, ["imageUrl", "avatarUrl", "pictureUrl", "photoUrl", "logoUrl"]) ?? null
  );
}

function firstString(record: Record<string, unknown>, keys: string[]): string | undefined {
  for (const key of keys) {
    const value = record[key];

    if (typeof value === "string" && value.trim().length > 0) {
      return value;
    }
  }

  return undefined;
}

function deriveNameValue(value: unknown): string | undefined {
  const record = asRecord(value);

  if (!record) {
    return undefined;
  }

  const directString = firstString(record, ["fullName", "displayName", "name"]);

  if (directString) {
    return directString;
  }

  return joinNameParts(record);
}

function joinNameParts(record: Record<string, unknown>): string | undefined {
  const parts = ["firstName", "middleName", "lastName"]
    .map((key) => record[key])
    .filter((value): value is string => typeof value === "string" && value.trim().length > 0);

  if (parts.length === 0) {
    return undefined;
  }

  return parts.join(" ");
}

function asRecord(value: unknown): Record<string, unknown> | undefined {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return undefined;
  }

  return value as Record<string, unknown>;
}

function toNumber(value: number | string | undefined): number {
  if (typeof value === "number") {
    return value;
  }

  if (typeof value === "string" && value.length > 0) {
    return Number(value);
  }

  return 0;
}
