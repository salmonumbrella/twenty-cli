import { Command } from "commander";
import {
  formatGraphqlErrors,
  getGraphqlField,
  type GraphQLResponse,
  hasGraphqlField,
  hasSchemaErrorSymbol,
} from "../../utilities/api/graphql-response";
import { applyGlobalOptions } from "../../utilities/shared/global-options";
import { createCommandContext } from "../../utilities/shared/context";
import { createServices } from "../../utilities/shared/services";
import { CliError } from "../../utilities/errors/cli-error";
import { parseBody } from "../../utilities/shared/body";
import { readFileOrStdin, readJsonInput } from "../../utilities/shared/io";
import { requireYes } from "../../utilities/shared/confirmation";
import { MetadataSubscriptionService } from "../../utilities/api/services/metadata-subscription.service";

interface ServerlessOptions {
  data?: string;
  file?: string;
  set?: string[];
  name?: string;
  description?: string;
  timeoutSeconds?: number;
  universalIdentifier?: string;
  applicationId?: string;
  applicationUniversalIdentifier?: string;
  maxEvents?: number;
  waitSeconds?: number;
  packageJson?: string;
  packageJsonFile?: string;
  yarnLock?: string;
  yarnLockFile?: string;
  yes?: boolean;
}

interface OperationRequest {
  query: string;
  resultKey: string;
  variables?: Record<string, unknown>;
  schemaSymbols?: string[];
}

interface CompatibleOperation {
  current: OperationRequest;
  legacy?: OperationRequest;
  unavailableOnLegacyMessage?: string;
}

const endpoint = "/metadata";

const SERVERLESS_FUNCTION_FIELDS = `
  id
  name
  description
  runtime
  timeoutSeconds
  latestVersion
  publishedVersions
  handlerPath
  handlerName
  toolInputSchema
  isTool
  applicationId
  createdAt
  updatedAt
  cronTriggers {
    id
    settings
    createdAt
    updatedAt
  }
  databaseEventTriggers {
    id
    settings
    createdAt
    updatedAt
  }
  routeTriggers {
    id
    path
    isAuthRequired
    httpMethod
    createdAt
    updatedAt
  }
`;

const LEGACY_LOGIC_FUNCTION_FIELDS = `
  id
  name
  description
  runtime
  timeoutSeconds
  sourceHandlerPath
  handlerName
  toolInputSchema
  isTool
  cronTriggerSettings
  databaseEventTriggerSettings
  httpRouteTriggerSettings
  applicationId
  universalIdentifier
  createdAt
  updatedAt
`;

function collect(value: string, previous: string[] = []): string[] {
  return previous.concat([value]);
}

function hasPayloadInput(options: ServerlessOptions): boolean {
  return Boolean(options.data || options.file || options.set?.length);
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function hasOwnKey(value: unknown, key: string): boolean {
  return isRecord(value) && Object.prototype.hasOwnProperty.call(value, key);
}

function shouldFallbackToLegacy(
  response: GraphQLResponse<unknown>,
  current: OperationRequest,
): boolean {
  const symbols = [current.resultKey, ...(current.schemaSymbols ?? [])];
  return hasSchemaErrorSymbol(response, symbols);
}

function throwGraphqlError(
  response: GraphQLResponse<unknown>,
  fallbackMessage: string,
  code = "API_ERROR",
): never {
  throw new CliError(formatGraphqlErrors(response) ?? fallbackMessage, code);
}

function buildOptionInput(options: ServerlessOptions): Record<string, unknown> {
  const input: Record<string, unknown> = {};

  if (options.name !== undefined) {
    input.name = options.name;
  }

  if (options.description !== undefined) {
    input.description = options.description;
  }

  if (options.timeoutSeconds !== undefined) {
    input.timeoutSeconds = options.timeoutSeconds;
  }

  return input;
}

function normalizeCurrentServerlessInput(input: Record<string, unknown>): Record<string, unknown> {
  const normalized: Record<string, unknown> = { ...input };

  if (
    typeof normalized.sourceHandlerPath === "string" &&
    typeof normalized.handlerPath !== "string"
  ) {
    normalized.handlerPath = normalized.sourceHandlerPath;
  }

  const source = normalized.source;

  if (isRecord(source)) {
    const handlerPath =
      (typeof normalized.handlerPath === "string" && normalized.handlerPath) ||
      (typeof source.handlerPath === "string" && source.handlerPath) ||
      (typeof source.sourceHandlerPath === "string" && source.sourceHandlerPath) ||
      "src/index.ts";

    if (!hasOwnKey(normalized, "code")) {
      if (isRecord(source.code)) {
        normalized.code = source.code;
      } else if (typeof source.sourceHandlerCode === "string") {
        normalized.code = {
          [handlerPath]: source.sourceHandlerCode,
        };
      }
    }

    if (typeof normalized.handlerPath !== "string") {
      normalized.handlerPath = handlerPath;
    }

    if (typeof normalized.handlerName !== "string" && typeof source.handlerName === "string") {
      normalized.handlerName = source.handlerName;
    }

    if (!hasOwnKey(normalized, "toolInputSchema") && hasOwnKey(source, "toolInputSchema")) {
      normalized.toolInputSchema = source.toolInputSchema;
    }

    delete normalized.source;
  }

  delete normalized.sourceHandlerPath;

  return normalized;
}

async function buildCreateInput(
  options: ServerlessOptions,
  mode: "current" | "legacy",
): Promise<Record<string, unknown>> {
  const payload = hasPayloadInput(options)
    ? await parseBody(options.data, options.file, options.set)
    : {};
  const input = { ...payload, ...buildOptionInput(options) };

  if (typeof input.name !== "string" || input.name.trim() === "") {
    throw new CliError(
      "Missing serverless function name. Provide --name or include `name` in --data/--file.",
      "INVALID_ARGUMENTS",
    );
  }

  return mode === "current" ? normalizeCurrentServerlessInput(input) : input;
}

async function buildUpdateInput(
  id: string,
  options: ServerlessOptions,
  mode: "current" | "legacy",
): Promise<Record<string, unknown>> {
  const payload = hasPayloadInput(options)
    ? await parseBody(options.data, options.file, options.set)
    : {};
  const update = { ...payload, ...buildOptionInput(options) };

  if (Object.keys(update).length === 0) {
    throw new CliError(
      "Provide at least one field to update via --name, --description, --timeout-seconds, --data, --file, or --set.",
      "INVALID_ARGUMENTS",
    );
  }

  return { id, update: mode === "current" ? normalizeCurrentServerlessInput(update) : update };
}

async function buildExecutePayload(options: ServerlessOptions): Promise<Record<string, unknown>> {
  if (!hasPayloadInput(options)) {
    return {};
  }

  return parseBody(options.data, options.file, options.set);
}

async function buildCreateLayerInput(
  options: ServerlessOptions,
): Promise<{ packageJson: Record<string, unknown>; yarnLock: string }> {
  const packageJson = await readJsonInput(options.packageJson, options.packageJsonFile);

  if (packageJson == null) {
    throw new CliError(
      "Missing package.json input; provide --package-json or --package-json-file.",
      "INVALID_ARGUMENTS",
    );
  }

  if (typeof packageJson !== "object" || Array.isArray(packageJson)) {
    throw new CliError("package.json input must be a JSON object.", "INVALID_ARGUMENTS");
  }

  const yarnLock = await readLayerTextInput(options.yarnLock, options.yarnLockFile);
  if (!yarnLock) {
    throw new CliError(
      "Missing yarn.lock input; provide --yarn-lock or --yarn-lock-file.",
      "INVALID_ARGUMENTS",
    );
  }

  return {
    packageJson: packageJson as Record<string, unknown>,
    yarnLock,
  };
}

async function readLayerTextInput(
  inlineValue: string | undefined,
  filePath: string | undefined,
): Promise<string | undefined> {
  if (inlineValue && inlineValue.trim() !== "") {
    return inlineValue;
  }

  if (filePath && filePath.trim() !== "") {
    const content = await readFileOrStdin(filePath.trim());
    return content.trim() === "" ? undefined : content;
  }

  return undefined;
}

function buildLogsInput(
  id: string | undefined,
  options: ServerlessOptions,
): Record<string, unknown> {
  const input: Record<string, unknown> = {};

  if (id) {
    input.id = id;
  }
  if (options.name) {
    input.name = options.name;
  }
  if (options.universalIdentifier) {
    input.universalIdentifier = options.universalIdentifier;
  }
  if (options.applicationId) {
    input.applicationId = options.applicationId;
  }
  if (options.applicationUniversalIdentifier) {
    input.applicationUniversalIdentifier = options.applicationUniversalIdentifier;
  }

  if (Object.keys(input).length === 0) {
    throw new CliError(
      "Serverless logs require at least one filter: function ID, --name, --universal-identifier, --application-id, or --application-universal-identifier.",
      "INVALID_ARGUMENTS",
    );
  }

  return input;
}

function shouldCollectStreamOutput(format: string | undefined): boolean {
  return format === "json";
}

function isAbortError(error: unknown): boolean {
  return error instanceof Error && error.name === "AbortError";
}

async function streamLogicFunctionLogs(
  id: string | undefined,
  options: ServerlessOptions,
  globalOptions: { output?: string; query?: string; workspace?: string; debug?: boolean },
  services: ReturnType<typeof createServices>,
): Promise<void> {
  const maxEvents =
    typeof options.maxEvents === "number" && Number.isFinite(options.maxEvents)
      ? options.maxEvents
      : undefined;
  const waitSeconds =
    typeof options.waitSeconds === "number" && Number.isFinite(options.waitSeconds)
      ? options.waitSeconds
      : undefined;
  const format = globalOptions.output ?? "text";

  if (shouldCollectStreamOutput(format) && maxEvents === undefined && waitSeconds === undefined) {
    throw new CliError(
      "Streaming JSON output requires --max-events or --wait-seconds so the command can terminate with a complete array.",
      "INVALID_ARGUMENTS",
    );
  }

  const subscription = new MetadataSubscriptionService(undefined, {
    workspace: globalOptions.workspace,
    debug: globalOptions.debug,
  });
  const controller = new AbortController();
  const collected: unknown[] = [];
  let stopRequested = false;
  let seen = 0;
  let timeoutHandle: ReturnType<typeof setTimeout> | undefined;

  if (waitSeconds !== undefined && waitSeconds > 0) {
    timeoutHandle = setTimeout(() => {
      stopRequested = true;
      controller.abort();
    }, waitSeconds * 1000);
  }

  try {
    for await (const payload of subscription.subscribe<{ logicFunctionLogs?: unknown }>({
      query: `subscription LogicFunctionLogs($input: LogicFunctionLogsInput!) {
        logicFunctionLogs(input: $input) {
          logs
        }
      }`,
      variables: {
        input: buildLogsInput(id, options),
      },
      signal: controller.signal,
    })) {
      const event =
        typeof payload === "object" && payload !== null && "logicFunctionLogs" in payload
          ? (payload as { logicFunctionLogs?: unknown }).logicFunctionLogs
          : payload;

      if (shouldCollectStreamOutput(format)) {
        collected.push(event);
      } else {
        await renderServerlessFunction(event, services, globalOptions);
      }

      seen += 1;
      if (maxEvents !== undefined && seen >= maxEvents) {
        stopRequested = true;
        controller.abort();
        break;
      }
    }
  } catch (error) {
    if (!stopRequested || !isAbortError(error)) {
      throw error;
    }
  } finally {
    if (timeoutHandle) {
      clearTimeout(timeoutHandle);
    }
  }

  if (shouldCollectStreamOutput(format)) {
    await renderServerlessFunction(collected, services, globalOptions);
  }
}

async function postMetadataOperation<T>(
  services: ReturnType<typeof createServices>,
  request: OperationRequest,
): Promise<GraphQLResponse<Record<string, T>>> {
  const payload: Record<string, unknown> = {
    query: request.query,
  };

  if (request.variables) {
    payload.variables = request.variables;
  }

  const response = await services.api.post<GraphQLResponse<Record<string, T>>>(endpoint, payload);

  return response.data ?? {};
}

async function executeCompatibleOperation<T>(
  services: ReturnType<typeof createServices>,
  operation: CompatibleOperation,
): Promise<T> {
  const currentResponse = await postMetadataOperation<T>(services, operation.current);

  if (hasGraphqlField(currentResponse, operation.current.resultKey)) {
    return getGraphqlField(currentResponse, operation.current.resultKey) as T;
  }

  if (operation.legacy) {
    if (
      shouldFallbackToLegacy(currentResponse, operation.current) ||
      !formatGraphqlErrors(currentResponse)
    ) {
      const legacyResponse = await postMetadataOperation<T>(services, operation.legacy);

      if (hasGraphqlField(legacyResponse, operation.legacy.resultKey)) {
        return getGraphqlField(legacyResponse, operation.legacy.resultKey) as T;
      }

      throwGraphqlError(
        legacyResponse,
        `Unexpected GraphQL response missing ${operation.legacy.resultKey}.`,
      );
    }
  }

  if (
    operation.unavailableOnLegacyMessage &&
    shouldFallbackToLegacy(currentResponse, operation.current)
  ) {
    throw new CliError(operation.unavailableOnLegacyMessage, "INVALID_ARGUMENTS");
  }

  throwGraphqlError(
    currentResponse,
    `Unexpected GraphQL response missing ${operation.current.resultKey}.`,
  );
}

function renderServerlessFunction(
  value: unknown,
  services: ReturnType<typeof createServices>,
  globalOptions: { output?: string; query?: string },
): Promise<void> {
  return services.output.render(value, {
    format: globalOptions.output,
    query: globalOptions.query,
  });
}

function getServerlessOptions(command: Command): ServerlessOptions {
  return command.opts() as ServerlessOptions;
}

function applyMutationOptions(command: Command): void {
  command
    .option("-d, --data <json>", "JSON payload")
    .option("-f, --file <path>", "JSON file payload (use - for stdin)")
    .option("--set <key=value>", "Set a field value", collect)
    .option("--name <name>", "Function name")
    .option("--description <text>", "Function description")
    .option("--timeout-seconds <seconds>", "Function timeout in seconds", Number);
}

function applyExecutionOptions(command: Command): void {
  command
    .option("-d, --data <json>", "JSON payload")
    .option("-f, --file <path>", "JSON file payload (use - for stdin)")
    .option("--set <key=value>", "Set a field value", collect);
}

function applyLogsOptions(command: Command): void {
  command
    .option("--name <name>", "Function name")
    .option("--universal-identifier <id>", "Function universal identifier filter for logs")
    .option("--application-id <id>", "Application ID filter for logs")
    .option(
      "--application-universal-identifier <id>",
      "Application universal identifier filter for logs",
    )
    .option("--max-events <count>", "Stop streaming after N log payloads", Number)
    .option("--wait-seconds <seconds>", "Stop streaming after N seconds", Number);
}

function applyLayerOptions(command: Command): void {
  command
    .option("--package-json <json>", "Layer package.json JSON")
    .option("--package-json-file <path>", "Layer package.json file")
    .option("--yarn-lock <text>", "Layer yarn.lock content")
    .option("--yarn-lock-file <path>", "Layer yarn.lock file");
}

async function runServerlessListCommand(command: Command): Promise<void> {
  const { globalOptions, services } = createCommandContext(command);

  const result = await executeCompatibleOperation<unknown[]>(services, {
    current: {
      query: `query { findManyServerlessFunctions { ${SERVERLESS_FUNCTION_FIELDS} } }`,
      resultKey: "findManyServerlessFunctions",
      schemaSymbols: ["findManyServerlessFunctions"],
    },
    legacy: {
      query: `query { findManyLogicFunctions { ${LEGACY_LOGIC_FUNCTION_FIELDS} } }`,
      resultKey: "findManyLogicFunctions",
    },
  });

  await renderServerlessFunction(Array.isArray(result) ? result : [], services, globalOptions);
}

async function runServerlessGetCommand(id: string | undefined, command: Command): Promise<void> {
  const { globalOptions, services } = createCommandContext(command);

  if (!id) throw new CliError("Missing function ID.", "INVALID_ARGUMENTS");

  const result = await executeCompatibleOperation<unknown>(services, {
    current: {
      query: `query($input: ServerlessFunctionIdInput!) { findOneServerlessFunction(input: $input) { ${SERVERLESS_FUNCTION_FIELDS} } }`,
      variables: { input: { id } },
      resultKey: "findOneServerlessFunction",
      schemaSymbols: ["findOneServerlessFunction", "ServerlessFunctionIdInput"],
    },
    legacy: {
      query: `query($input: LogicFunctionIdInput!) { findOneLogicFunction(input: $input) { ${LEGACY_LOGIC_FUNCTION_FIELDS} } }`,
      variables: { input: { id } },
      resultKey: "findOneLogicFunction",
    },
  });

  await renderServerlessFunction(result, services, globalOptions);
}

async function runServerlessCreateCommand(command: Command): Promise<void> {
  const { globalOptions, services } = createCommandContext(command);
  const options = getServerlessOptions(command);
  const currentInput = await buildCreateInput(options, "current");

  const result = await executeCompatibleOperation<unknown>(services, {
    current: {
      query: `mutation($input: CreateServerlessFunctionInput!) { createOneServerlessFunction(input: $input) { ${SERVERLESS_FUNCTION_FIELDS} } }`,
      variables: { input: currentInput },
      resultKey: "createOneServerlessFunction",
      schemaSymbols: ["createOneServerlessFunction", "CreateServerlessFunctionInput"],
    },
    legacy: {
      query: `mutation($input: CreateLogicFunctionFromSourceInput!) { createOneLogicFunction(input: $input) { ${LEGACY_LOGIC_FUNCTION_FIELDS} } }`,
      variables: { input: await buildCreateInput(options, "legacy") },
      resultKey: "createOneLogicFunction",
    },
  });

  await renderServerlessFunction(result, services, globalOptions);
}

async function runServerlessUpdateCommand(id: string | undefined, command: Command): Promise<void> {
  const { globalOptions, services } = createCommandContext(command);
  const options = getServerlessOptions(command);

  if (!id) throw new CliError("Missing function ID.", "INVALID_ARGUMENTS");

  const currentInput = await buildUpdateInput(id, options, "current");

  const result = await executeCompatibleOperation<unknown>(services, {
    current: {
      query: `mutation($input: UpdateServerlessFunctionInput!) { updateOneServerlessFunction(input: $input) { ${SERVERLESS_FUNCTION_FIELDS} } }`,
      variables: { input: currentInput },
      resultKey: "updateOneServerlessFunction",
      schemaSymbols: ["updateOneServerlessFunction", "UpdateServerlessFunctionInput"],
    },
    legacy: {
      query: `mutation($input: UpdateLogicFunctionFromSourceInput!) { updateOneLogicFunction(input: $input) }`,
      variables: { input: await buildUpdateInput(id, options, "legacy") },
      resultKey: "updateOneLogicFunction",
    },
  });

  await renderServerlessFunction(
    {
      id,
      updated: Boolean(result),
    },
    services,
    globalOptions,
  );
}

async function runServerlessPackagesCommand(
  id: string | undefined,
  command: Command,
): Promise<void> {
  const { globalOptions, services } = createCommandContext(command);

  if (!id) throw new CliError("Missing function ID.", "INVALID_ARGUMENTS");

  const result = await executeCompatibleOperation<unknown>(services, {
    current: {
      query: `query($input: ServerlessFunctionIdInput!) { getAvailablePackages(input: $input) }`,
      variables: { input: { id } },
      resultKey: "getAvailablePackages",
      schemaSymbols: ["ServerlessFunctionIdInput"],
    },
    legacy: {
      query: `query($input: LogicFunctionIdInput!) { getAvailablePackages(input: $input) }`,
      variables: { input: { id } },
      resultKey: "getAvailablePackages",
    },
  });

  await renderServerlessFunction(result ?? {}, services, globalOptions);
}

async function runServerlessDeleteCommand(id: string | undefined, command: Command): Promise<void> {
  const { services } = createCommandContext(command);

  if (!id) throw new CliError("Missing function ID.", "INVALID_ARGUMENTS");
  requireYes(getServerlessOptions(command), "Delete");

  await executeCompatibleOperation<unknown>(services, {
    current: {
      query: `mutation($input: ServerlessFunctionIdInput!) { deleteOneServerlessFunction(input: $input) { id } }`,
      variables: { input: { id } },
      resultKey: "deleteOneServerlessFunction",
      schemaSymbols: ["deleteOneServerlessFunction", "ServerlessFunctionIdInput"],
    },
    legacy: {
      query: `mutation($input: LogicFunctionIdInput!) { deleteOneLogicFunction(input: $input) { id } }`,
      variables: { input: { id } },
      resultKey: "deleteOneLogicFunction",
    },
  });

  // eslint-disable-next-line no-console
  console.log(`Serverless function ${id} deleted.`);
}

async function runServerlessExecuteCommand(
  id: string | undefined,
  command: Command,
): Promise<void> {
  const { globalOptions, services } = createCommandContext(command);
  const options = getServerlessOptions(command);

  if (!id) throw new CliError("Missing function ID.", "INVALID_ARGUMENTS");

  const payload = await buildExecutePayload(options);
  const result = await executeCompatibleOperation<unknown>(services, {
    current: {
      query: `mutation($input: ExecuteServerlessFunctionInput!) { executeOneServerlessFunction(input: $input) { data logs duration status error } }`,
      variables: { input: { id, payload } },
      resultKey: "executeOneServerlessFunction",
      schemaSymbols: ["executeOneServerlessFunction", "ExecuteServerlessFunctionInput"],
    },
    legacy: {
      query: `mutation($input: ExecuteOneLogicFunctionInput!) { executeOneLogicFunction(input: $input) { data logs duration status error } }`,
      variables: { input: { id, payload } },
      resultKey: "executeOneLogicFunction",
    },
  });

  await renderServerlessFunction(result, services, globalOptions);
}

async function runServerlessSourceCommand(id: string | undefined, command: Command): Promise<void> {
  const { globalOptions, services } = createCommandContext(command);

  if (!id) throw new CliError("Missing function ID.", "INVALID_ARGUMENTS");

  const sourceCode = await executeCompatibleOperation<string | null>(services, {
    current: {
      query: `query($input: GetServerlessFunctionSourceCodeInput!) { getServerlessFunctionSourceCode(input: $input) }`,
      variables: { input: { id } },
      resultKey: "getServerlessFunctionSourceCode",
      schemaSymbols: ["getServerlessFunctionSourceCode", "GetServerlessFunctionSourceCodeInput"],
    },
    legacy: {
      query: `query($input: LogicFunctionIdInput!) { getLogicFunctionSourceCode(input: $input) }`,
      variables: { input: { id } },
      resultKey: "getLogicFunctionSourceCode",
    },
  });

  if (globalOptions.output === "json" || globalOptions.output === "csv") {
    await renderServerlessFunction({ sourceCode: sourceCode ?? "" }, services, globalOptions);
    return;
  }

  // eslint-disable-next-line no-console
  console.log(sourceCode ?? "");
}

async function runServerlessLogsCommand(id: string | undefined, command: Command): Promise<void> {
  const { globalOptions, services } = createCommandContext(command);

  await streamLogicFunctionLogs(id, getServerlessOptions(command), globalOptions, services);
}

async function runServerlessPublishCommand(
  id: string | undefined,
  command: Command,
): Promise<void> {
  const { globalOptions, services } = createCommandContext(command);

  if (!id) throw new CliError("Missing function ID.", "INVALID_ARGUMENTS");

  const result = await executeCompatibleOperation<unknown>(services, {
    current: {
      query: `mutation($input: PublishServerlessFunctionInput!) { publishServerlessFunction(input: $input) { ${SERVERLESS_FUNCTION_FIELDS} } }`,
      variables: { input: { id } },
      resultKey: "publishServerlessFunction",
      schemaSymbols: ["publishServerlessFunction", "PublishServerlessFunctionInput"],
    },
    unavailableOnLegacyMessage:
      "Publish is not available on this workspace because it still exposes the legacy LogicFunction schema.",
  });

  await renderServerlessFunction(result, services, globalOptions);
}

async function runServerlessCreateLayerCommand(command: Command): Promise<void> {
  const { globalOptions, services } = createCommandContext(command);
  const input = await buildCreateLayerInput(getServerlessOptions(command));

  const result = await executeCompatibleOperation<unknown>(services, {
    current: {
      query: `mutation($packageJson: JSON!, $yarnLock: String!) {
        createOneServerlessFunctionLayer(packageJson: $packageJson, yarnLock: $yarnLock) {
          id
          applicationId
          createdAt
          updatedAt
        }
      }`,
      variables: input,
      resultKey: "createOneServerlessFunctionLayer",
      schemaSymbols: ["createOneServerlessFunctionLayer", "CreateServerlessFunctionLayerInput"],
    },
    unavailableOnLegacyMessage:
      "Serverless layers are not available on this workspace because it does not expose createOneServerlessFunctionLayer.",
  });

  await renderServerlessFunction(result, services, globalOptions);
}

export function registerServerlessCommand(program: Command): void {
  const serverless = program.command("serverless").description("Manage serverless functions");
  applyGlobalOptions(serverless);

  const listCmd = serverless.command("list").description("List serverless functions");
  applyGlobalOptions(listCmd);
  listCmd.action(async (_options: unknown, command: Command) => {
    await runServerlessListCommand(command);
  });

  const getCmd = serverless
    .command("get")
    .description("Get a serverless function")
    .argument("[id]", "Function ID");
  applyGlobalOptions(getCmd);
  getCmd.action(async (id: string | undefined, _options: unknown, command: Command) => {
    await runServerlessGetCommand(id, command);
  });

  const createCmd = serverless.command("create").description("Create a serverless function");
  applyMutationOptions(createCmd);
  applyGlobalOptions(createCmd);
  createCmd.action(async (_options: unknown, command: Command) => {
    await runServerlessCreateCommand(command);
  });

  const updateCmd = serverless
    .command("update")
    .description("Update a serverless function")
    .argument("[id]", "Function ID");
  applyMutationOptions(updateCmd);
  applyGlobalOptions(updateCmd);
  updateCmd.action(async (id: string | undefined, _options: unknown, command: Command) => {
    await runServerlessUpdateCommand(id, command);
  });

  const deleteCmd = serverless
    .command("delete")
    .description("Delete a serverless function")
    .argument("[id]", "Function ID")
    .option("--yes", "Confirm destructive operations");
  applyGlobalOptions(deleteCmd);
  deleteCmd.action(async (id: string | undefined, _options: unknown, command: Command) => {
    await runServerlessDeleteCommand(id, command);
  });

  const publishCmd = serverless
    .command("publish")
    .description("Publish a serverless function")
    .argument("[id]", "Function ID");
  applyGlobalOptions(publishCmd);
  publishCmd.action(async (id: string | undefined, _options: unknown, command: Command) => {
    await runServerlessPublishCommand(id, command);
  });

  const executeCmd = serverless
    .command("execute")
    .description("Execute a serverless function")
    .argument("[id]", "Function ID");
  applyExecutionOptions(executeCmd);
  applyGlobalOptions(executeCmd);
  executeCmd.action(async (id: string | undefined, _options: unknown, command: Command) => {
    await runServerlessExecuteCommand(id, command);
  });

  const packagesCmd = serverless
    .command("packages")
    .alias("available-packages")
    .description("List available packages for a serverless function")
    .argument("[id]", "Function ID");
  applyGlobalOptions(packagesCmd);
  packagesCmd.action(async (id: string | undefined, _options: unknown, command: Command) => {
    await runServerlessPackagesCommand(id, command);
  });

  const sourceCmd = serverless
    .command("source")
    .description("Get serverless function source code")
    .argument("[id]", "Function ID");
  applyGlobalOptions(sourceCmd);
  sourceCmd.action(async (id: string | undefined, _options: unknown, command: Command) => {
    await runServerlessSourceCommand(id, command);
  });

  const logsCmd = serverless
    .command("logs")
    .description("Stream serverless function logs")
    .argument("[id]", "Function ID");
  applyLogsOptions(logsCmd);
  applyGlobalOptions(logsCmd);
  logsCmd.action(async (id: string | undefined, _options: unknown, command: Command) => {
    await runServerlessLogsCommand(id, command);
  });

  const createLayerCmd = serverless
    .command("create-layer")
    .description("Create a serverless function layer");
  applyLayerOptions(createLayerCmd);
  applyGlobalOptions(createLayerCmd);
  createLayerCmd.action(async (_options: unknown, command: Command) => {
    await runServerlessCreateLayerCommand(command);
  });
}
