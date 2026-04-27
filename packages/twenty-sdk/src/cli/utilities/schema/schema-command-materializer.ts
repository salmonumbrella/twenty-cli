import { Command } from "commander";
import { createCommandContext } from "../shared/context";
import { applyGlobalOptions } from "../shared/global-options";
import { parseArrayPayload, parseBody } from "../shared/body";
import { CliError } from "../errors/cli-error";
import { SchemaCacheEntry } from "./schema-cache.service";
import { readCachedSchemaEntries } from "./schema-cache-reader";
import { readJsonInput } from "../shared/io";
import { mergeSets } from "../shared/parse";
import { requireYes } from "../shared/confirmation";

export type DynamicRecordOperation =
  | "batch-create"
  | "batch-delete"
  | "batch-update"
  | "create"
  | "delete"
  | "destroy"
  | "find-duplicates"
  | "get"
  | "group-by"
  | "list"
  | "merge"
  | "restore"
  | "update";

export type DynamicMetadataOperation = "create" | "delete" | "get" | "list" | "update";

export interface DynamicResource<TOperation extends string> {
  apiName: string;
  operations: TOperation[];
}

export interface CachedSchemaCommandEntries {
  coreOpenApi?: SchemaCacheEntry;
  metadataOpenApi?: SchemaCacheEntry;
}

interface DynamicCommandOptions {
  limit?: string;
  cursor?: string;
  filter?: string;
  sort?: string;
  order?: string;
  include?: string;
  data?: string;
  file?: string;
  set?: string[];
  ids?: string;
  field?: string;
  object?: string;
  view?: string;
  yes?: boolean;
}

interface DynamicRecordsService {
  batchUpdate(object: string, records: Record<string, unknown>[]): Promise<unknown>;
  updateMany(
    object: string,
    data: Record<string, unknown>,
    options: { filter: string },
  ): Promise<unknown>;
}

const RECORD_OPERATION_ORDER: DynamicRecordOperation[] = [
  "batch-create",
  "batch-delete",
  "batch-update",
  "create",
  "delete",
  "destroy",
  "find-duplicates",
  "get",
  "group-by",
  "list",
  "merge",
  "restore",
  "update",
];

const METADATA_OPERATION_ORDER: DynamicMetadataOperation[] = [
  "create",
  "delete",
  "get",
  "list",
  "update",
];

export function registerCachedSchemaCommands(
  program: Command,
  entries?: CachedSchemaCommandEntries,
): void {
  const cachedEntries = entries ?? safeReadCachedSchemaEntries();
  const records = program.command("records").description("Cache-backed record commands");
  applyGlobalOptions(records);

  for (const resource of extractCoreRecordResources(cachedEntries.coreOpenApi?.schema)) {
    registerRecordResource(records, resource);
  }

  const metadata = program.command("metadata").description("Cache-backed metadata commands");
  applyGlobalOptions(metadata);

  for (const resource of extractMetadataResources(cachedEntries.metadataOpenApi?.schema)) {
    registerMetadataResource(metadata, resource);
  }
}

function safeReadCachedSchemaEntries(): CachedSchemaCommandEntries {
  try {
    return readCachedSchemaEntries();
  } catch {
    return {};
  }
}

export function extractCoreRecordResources(
  schema: unknown,
): DynamicResource<DynamicRecordOperation>[] {
  const resources = new Map<string, Set<DynamicRecordOperation>>();

  for (const [rawPath, methods] of Object.entries(extractOpenApiPaths(schema))) {
    if (!isRecord(methods)) continue;
    const normalizedPath = normalizeOpenApiPath(rawPath);
    const methodSet = new Set(Object.keys(methods).map((method) => method.toLowerCase()));
    const match = corePathToResourceOperation(normalizedPath, methodSet);
    if (!match || match.operations.length === 0) continue;

    const operations = resources.get(match.resource) ?? new Set<DynamicRecordOperation>();
    for (const operation of match.operations) operations.add(operation);
    resources.set(match.resource, operations);
  }

  return Array.from(resources.entries())
    .map(([apiName, operations]) => ({
      apiName,
      operations: sortOperations(Array.from(operations), RECORD_OPERATION_ORDER),
    }))
    .sort((left, right) => left.apiName.localeCompare(right.apiName));
}

export function extractMetadataResources(
  schema: unknown,
): DynamicResource<DynamicMetadataOperation>[] {
  const resources = new Map<string, Set<DynamicMetadataOperation>>();

  for (const [rawPath, methods] of Object.entries(extractOpenApiPaths(schema))) {
    if (!isRecord(methods)) continue;
    const normalizedPath = normalizeOpenApiPath(rawPath);
    const match = metadataPathToResourceOperation(
      normalizedPath,
      new Set(Object.keys(methods).map((method) => method.toLowerCase())),
    );
    if (!match || match.operations.length === 0) continue;

    const operations = resources.get(match.resource) ?? new Set<DynamicMetadataOperation>();
    for (const operation of match.operations) operations.add(operation);
    resources.set(match.resource, operations);
  }

  return Array.from(resources.entries())
    .map(([apiName, operations]) => ({
      apiName,
      operations: sortOperations(Array.from(operations), METADATA_OPERATION_ORDER),
    }))
    .sort((left, right) => left.apiName.localeCompare(right.apiName));
}

function registerRecordResource(
  parent: Command,
  resource: DynamicResource<DynamicRecordOperation>,
): void {
  const commandName = toKebabCase(resource.apiName);
  const command = parent
    .command(commandName)
    .description(`Cached record commands for ${commandName}`);
  applyGlobalOptions(command);

  for (const operation of resource.operations) {
    const operationCommand = command
      .command(operation)
      .description(recordOperationSummary(operation, commandName));
    applyRecordOptions(operationCommand);
    if (isDestructiveRecordOperation(operation)) {
      applyRecordDestructiveOptions(operationCommand);
    }
    if (operationNeedsId(operation)) {
      operationCommand.argument("[id]", "Record ID");
    }
    applyGlobalOptions(operationCommand);
    operationCommand.action(
      async (
        idOrOptions: string | DynamicCommandOptions | undefined,
        maybeOptions?: DynamicCommandOptions | Command,
        maybeCommand?: Command,
      ) => {
        const id = typeof idOrOptions === "string" ? idOrOptions : undefined;
        const actionCommand =
          maybeCommand ?? (maybeOptions instanceof Command ? maybeOptions : operationCommand);
        await runRecordOperation(actionCommand, resource.apiName, operation, id);
      },
    );
  }
}

function registerMetadataResource(
  parent: Command,
  resource: DynamicResource<DynamicMetadataOperation>,
): void {
  const commandName = toKebabCase(resource.apiName);
  const command = parent
    .command(commandName)
    .description(`Cached metadata commands for ${commandName}`);
  applyGlobalOptions(command);

  for (const operation of resource.operations) {
    const operationCommand = command
      .command(operation)
      .description(metadataOperationSummary(operation, commandName));
    applyMetadataOptions(operationCommand);
    if (operation !== "list" && operation !== "create") {
      operationCommand.argument("[id]", "Identifier");
    }
    applyGlobalOptions(operationCommand);
    operationCommand.action(
      async (
        idOrOptions: string | DynamicCommandOptions | undefined,
        maybeOptions?: DynamicCommandOptions | Command,
        maybeCommand?: Command,
      ) => {
        const id = typeof idOrOptions === "string" ? idOrOptions : undefined;
        const actionCommand =
          maybeCommand ?? (maybeOptions instanceof Command ? maybeOptions : operationCommand);
        await runMetadataOperation(actionCommand, resource.apiName, operation, id);
      },
    );
  }
}

async function runRecordOperation(
  command: Command,
  object: string,
  operation: DynamicRecordOperation,
  id: string | undefined,
): Promise<void> {
  const { globalOptions, services } = createCommandContext(command);
  const options = command.opts() as DynamicCommandOptions;
  let result: unknown;

  switch (operation) {
    case "list":
      result = await services.records.list(object, buildListOptions(options));
      break;
    case "get":
      assertId(id, "record");
      result = await services.records.get(object, id, { include: options.include });
      break;
    case "create":
      result = await services.records.create(
        object,
        await parseBody(options.data, options.file, options.set),
      );
      break;
    case "update":
      assertId(id, "record");
      result = await services.records.update(
        object,
        id,
        await parseBody(options.data, options.file, options.set),
      );
      break;
    case "delete":
      assertId(id, "record");
      requireYes(options, "Delete");
      result = await services.records.delete(object, id);
      break;
    case "destroy":
      if (id) {
        requireYes(options, "Destroy");
        result = await services.records.destroy(object, id);
      } else {
        const filter = resolveBulkFilter(options);
        requireYes(options, "Destroy");
        result = await services.records.destroyMany(object, { filter });
      }
      break;
    case "restore":
      if (id) {
        result = await services.records.restore(object, id);
      } else {
        result = await services.records.restoreMany(object, { filter: resolveBulkFilter(options) });
      }
      break;
    case "batch-create":
      result = await services.records.batchCreate(
        object,
        (await parseArrayPayload(options.data, options.file)) as Record<string, unknown>[],
      );
      break;
    case "batch-update":
      result = await runBatchUpdate(object, options, services.records);
      break;
    case "batch-delete":
      requireYes(options, "Batch delete");
      result = await services.records.batchDelete(object, await parseBatchDeleteIds(options));
      break;
    case "group-by":
      result = await services.records.groupBy(object, await buildGroupByPayload(options));
      break;
    case "find-duplicates":
      result = await services.records.findDuplicates(
        object,
        await parseBody(options.data, options.file, options.set),
      );
      break;
    case "merge":
      result = await services.records.merge(
        object,
        await parseBody(options.data, options.file, options.set),
      );
      break;
  }

  await services.output.render(result, {
    format: globalOptions.output,
    query: globalOptions.query,
  });
}

async function runMetadataOperation(
  command: Command,
  resource: string,
  operation: DynamicMetadataOperation,
  id: string | undefined,
): Promise<void> {
  const { globalOptions, services } = createCommandContext(command);
  const options = command.opts() as DynamicCommandOptions;
  const basePath = `/rest/metadata/${resource}`;
  let result: unknown;

  switch (operation) {
    case "list":
      result = (await services.api.get(basePath, { params: metadataQueryParams(options) })).data;
      break;
    case "get":
      assertId(id, "metadata resource");
      result = (await services.api.get(`${basePath}/${id}`)).data;
      break;
    case "create":
      result = (
        await services.api.post(basePath, await parseBody(options.data, options.file, options.set))
      ).data;
      break;
    case "update":
      assertId(id, "metadata resource");
      result = (
        await services.api.patch(
          `${basePath}/${id}`,
          await parseBody(options.data, options.file, options.set),
        )
      ).data;
      break;
    case "delete":
      assertId(id, "metadata resource");
      result = (await services.api.delete(`${basePath}/${id}`)).data;
      break;
  }

  await services.output.render(result ?? null, {
    format: globalOptions.output,
    query: globalOptions.query,
  });
}

function extractOpenApiPaths(schema: unknown): Record<string, Record<string, unknown>> {
  if (!isRecord(schema) || !isRecord(schema.paths)) return {};
  return schema.paths as Record<string, Record<string, unknown>>;
}

function corePathToResourceOperation(
  openApiPath: string,
  methods: Set<string>,
): { resource: string; operations: DynamicRecordOperation[] } | undefined {
  let match = openApiPath.match(/^\/batch\/([^/]+)$/);
  if (match) {
    return {
      resource: match[1],
      operations: [
        ...(methods.has("post") ? ["batch-create" as const] : []),
        ...(methods.has("patch") ? ["batch-update" as const] : []),
        ...(methods.has("delete") ? ["batch-delete" as const] : []),
      ],
    };
  }

  match = openApiPath.match(/^\/restore\/([^/]+)\/\{[^/]+\}$/);
  if (match && methods.has("patch")) {
    return { resource: match[1], operations: ["restore"] };
  }

  match = openApiPath.match(/^\/restore\/([^/]+)$/);
  if (match && methods.has("patch")) {
    return { resource: match[1], operations: ["restore"] };
  }

  match = openApiPath.match(/^\/([^/]+)\/\{[^/]+\}$/);
  if (match) {
    return {
      resource: match[1],
      operations: [
        ...(methods.has("get") ? ["get" as const] : []),
        ...(methods.has("patch") ? ["update" as const] : []),
        ...(methods.has("delete") ? (["delete", "destroy"] as const) : []),
      ],
    };
  }

  match = openApiPath.match(/^\/([^/]+)\/duplicates$/);
  if (match && methods.has("post")) {
    return { resource: match[1], operations: ["find-duplicates"] };
  }

  match = openApiPath.match(/^\/([^/]+)\/groupBy$/);
  if (match && methods.has("get")) {
    return { resource: match[1], operations: ["group-by"] };
  }

  match = openApiPath.match(/^\/([^/]+)\/merge$/);
  if (match && methods.has("patch")) {
    return { resource: match[1], operations: ["merge"] };
  }

  match = openApiPath.match(/^\/([^/]+)$/);
  if (match) {
    return {
      resource: match[1],
      operations: [
        ...(methods.has("get") ? ["list" as const] : []),
        ...(methods.has("post") ? ["create" as const] : []),
        ...(methods.has("patch") ? ["batch-update" as const] : []),
        ...(methods.has("delete") ? ["batch-delete" as const] : []),
      ],
    };
  }

  return undefined;
}

function metadataPathToResourceOperation(
  openApiPath: string,
  methods: Set<string>,
): { resource: string; operations: DynamicMetadataOperation[] } | undefined {
  let match = openApiPath.match(/^(?:\/metadata)?\/([^/]+)\/\{[^/]+\}$/);
  if (match) {
    return {
      resource: match[1],
      operations: [
        ...(methods.has("get") ? ["get" as const] : []),
        ...(methods.has("patch") ? ["update" as const] : []),
        ...(methods.has("delete") ? ["delete" as const] : []),
      ],
    };
  }

  match = openApiPath.match(/^(?:\/metadata)?\/([^/]+)$/);
  if (match) {
    return {
      resource: match[1],
      operations: [
        ...(methods.has("get") ? ["list" as const] : []),
        ...(methods.has("post") ? ["create" as const] : []),
      ],
    };
  }

  return undefined;
}

function normalizeOpenApiPath(openApiPath: string): string {
  return openApiPath.replace(/^\/rest(?=\/)/, "");
}

function sortOperations<TOperation extends string>(
  operations: TOperation[],
  order: readonly TOperation[],
): TOperation[] {
  return operations.sort((left, right) => order.indexOf(left) - order.indexOf(right));
}

function applyRecordOptions(command: Command): void {
  command
    .option("--limit <number>", "Limit number of records")
    .option("--cursor <cursor>", "Pagination cursor")
    .option("--filter <expression>", "Filter expression")
    .option("--sort <field>", "Sort field")
    .option("--order <direction>", "Sort order (asc or desc)")
    .option("--include <relations>", "Include related records")
    .option("-d, --data <json>", "JSON payload")
    .option("-f, --file <path>", "JSON file")
    .option("--set <key=value>", "Set a field value", collect)
    .option("--ids <ids>", "Comma-separated IDs")
    .option("--field <field>", "Group-by field");
}

function applyRecordDestructiveOptions(command: Command): void {
  command.option("--yes", "Confirm destructive operations");
}

function applyMetadataOptions(command: Command): void {
  command
    .option("-d, --data <json>", "JSON payload")
    .option("-f, --file <path>", "JSON file")
    .option("--set <key=value>", "Set a field value", collect)
    .option("--object <nameOrId>", "Filter by object name or metadata ID")
    .option("--view <id>", "Filter by view ID");
}

function buildListOptions(
  options: DynamicCommandOptions,
): Record<string, string | number | undefined> {
  return {
    limit: options.limit ? Number.parseInt(options.limit, 10) : undefined,
    cursor: options.cursor,
    filter: options.filter,
    sort: options.sort,
    order: options.order,
  };
}

async function runBatchUpdate(
  object: string,
  options: DynamicCommandOptions,
  records: DynamicRecordsService,
): Promise<unknown> {
  const rawPayload = await readJsonInput(options.data, options.file);

  if (Array.isArray(rawPayload) && !options.set?.length) {
    return records.batchUpdate(object, rawPayload as Record<string, unknown>[]);
  }

  return records.updateMany(object, requireObjectPayload(rawPayload, options.set), {
    filter: resolveBulkFilter(options),
  });
}

function requireObjectPayload(
  payload: unknown | undefined,
  sets: string[] | undefined,
): Record<string, unknown> {
  const merged = mergeOptionalSets(payload, sets);
  if (merged == null) {
    throw new CliError("Missing JSON payload; use --data, --file, or --set.", "INVALID_ARGUMENTS");
  }
  if (!isRecord(merged)) {
    throw new CliError("Payload must be a JSON object.", "INVALID_ARGUMENTS");
  }
  return merged;
}

async function parseBatchDeleteIds(options: DynamicCommandOptions): Promise<string[]> {
  const optionIds = parseOptionalIds(options.ids);
  if (optionIds.length > 0) return optionIds;

  const payload = await readJsonInput(options.data, options.file);
  if (payload == null) {
    throw new CliError("Missing JSON payload; use --data, --file, or --ids.", "INVALID_ARGUMENTS");
  }
  if (!Array.isArray(payload)) {
    throw new CliError("Batch payload must be a JSON array.", "INVALID_ARGUMENTS");
  }

  const ids = payload.map((value) => String(value).trim()).filter(Boolean);
  if (ids.length === 0) {
    throw new CliError("No valid IDs provided.", "INVALID_ARGUMENTS");
  }
  return ids;
}

function resolveBulkFilter(options: DynamicCommandOptions): string {
  if (options.filter?.trim()) {
    return options.filter.trim();
  }

  const ids = parseOptionalIds(options.ids);
  if (ids.length > 0) {
    return `id[in]:[${ids.join(",")}]`;
  }

  throw new CliError("Missing record ID.", "INVALID_ARGUMENTS");
}

async function buildGroupByPayload(options: DynamicCommandOptions): Promise<unknown> {
  const hasPayload = Boolean(options.data || options.file || options.set?.length);
  let payload: unknown | undefined;

  if (hasPayload) {
    const rawPayload = await readJsonInput(options.data, options.file);
    payload = normalizeGroupByPayload(mergeOptionalSets(rawPayload, options.set));
  } else if (options.field) {
    payload = { groupBy: [{ [options.field]: true }] };
  }

  if (options.filter) {
    payload = mergeGroupByFilter(payload, options.filter);
  }

  if (payload == null) {
    throw new CliError(
      "Missing group-by payload; use --field, --data, --file, or --set.",
      "INVALID_ARGUMENTS",
    );
  }

  return payload;
}

function mergeOptionalSets(payload: unknown | undefined, sets: string[] | undefined): unknown {
  if (!sets?.length) return payload;
  if (payload != null && (typeof payload !== "object" || Array.isArray(payload))) {
    throw new CliError("Payload must be a JSON object when using --set.", "INVALID_ARGUMENTS");
  }
  return mergeSets((payload ?? {}) as Record<string, unknown>, sets);
}

function normalizeGroupByPayload(payload: unknown): unknown {
  if (Array.isArray(payload)) {
    return { groupBy: payload };
  }

  if (typeof payload !== "object" || payload === null) {
    return payload;
  }

  const record = payload as Record<string, unknown>;
  if (typeof record.groupBy === "string") {
    return {
      ...record,
      groupBy: [{ [record.groupBy]: true }],
    };
  }

  return payload;
}

function mergeGroupByFilter(payload: unknown, filter: string): unknown {
  if (payload == null) {
    return { filter };
  }

  if (typeof payload !== "object" || Array.isArray(payload)) {
    return payload;
  }

  const record = payload as Record<string, unknown>;
  if (record.filter !== undefined) {
    return payload;
  }

  return {
    ...record,
    filter,
  };
}

function metadataQueryParams(options: DynamicCommandOptions): Record<string, string> {
  return Object.fromEntries(
    Object.entries({
      object: options.object,
      view: options.view,
    }).filter((entry): entry is [string, string] => typeof entry[1] === "string"),
  );
}

function operationNeedsId(operation: DynamicRecordOperation): boolean {
  return (
    operation === "get" ||
    operation === "update" ||
    operation === "delete" ||
    operation === "destroy" ||
    operation === "restore"
  );
}

function isDestructiveRecordOperation(operation: DynamicRecordOperation): boolean {
  return operation === "delete" || operation === "destroy" || operation === "batch-delete";
}

function assertId(id: string | undefined, label: string): asserts id is string {
  if (!id) throw new CliError(`Missing ${label} ID.`, "INVALID_ARGUMENTS");
}

function parseOptionalIds(value: string | undefined): string[] {
  return (
    value
      ?.split(",")
      .map((id) => id.trim())
      .filter(Boolean) ?? []
  );
}

function recordOperationSummary(operation: DynamicRecordOperation, resource: string): string {
  return `${capitalize(operation.replace(/-/g, " "))} ${resource}`;
}

function metadataOperationSummary(operation: DynamicMetadataOperation, resource: string): string {
  return `${capitalize(operation)} ${resource}`;
}

function collect(value: string, previous: string[] = []): string[] {
  return previous.concat([value]);
}

function toKebabCase(value: string): string {
  return value
    .replace(/([A-Z]+)([A-Z][a-z])/g, "$1-$2")
    .replace(/([a-z0-9])([A-Z])/g, "$1-$2")
    .replace(/[_\s]+/g, "-")
    .toLowerCase();
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function capitalize(value: string): string {
  return value.length === 0 ? value : `${value[0].toUpperCase()}${value.slice(1)}`;
}
