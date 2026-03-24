import fs from "fs-extra";
import path from "path";
import { Command, Option } from "commander";

interface HelpOperationMetadata {
  name: string;
  summary?: string;
  mutates?: boolean;
}

interface HelpMetadata {
  examples?: string[];
  operations?: HelpOperationMetadata[];
}

export interface HelpArgument {
  name: string;
  required: boolean;
  variadic: boolean;
  description?: string;
}

export interface HelpOption {
  name: string;
  flags: string;
  type: string;
  default?: string;
  required: boolean;
  global: boolean;
  description: string;
}

export interface HelpOperation {
  name: string;
  summary: string;
  mutates: boolean;
}

export interface HelpExitCode {
  code: number;
  summary: string;
}

export interface HelpOutputContract {
  query_language: "JMESPath";
  query_applies_before_format: boolean;
  formats: Array<{
    name: "agent" | "csv" | "json" | "jsonl" | "text";
    summary: string;
  }>;
}

export interface HelpSubcommand {
  name: string;
  summary: string;
}

export interface HelpCapabilities {
  mutates: boolean;
  supports_query: boolean;
  supports_workspace: boolean;
  supports_output: boolean;
}

export interface HelpDocument {
  schema_version: 1;
  kind: "root" | "command";
  name: string;
  aliases: string[];
  path: string[];
  summary: string;
  description?: string;
  usage: string;
  examples: string[];
  args: HelpArgument[];
  options: HelpOption[];
  operations: HelpOperation[];
  subcommands: HelpSubcommand[];
  capabilities: HelpCapabilities;
  exit_codes: HelpExitCode[];
  output_contract?: HelpOutputContract;
}

type HelpWriter = (text: string) => void;

const ROOT_HELP_FALLBACK = `twenty - CLI for Twenty

Discovery:
  twenty --help-json
  twenty --hj
  twenty CMD --help-json
  twenty CMD --hj
  twenty CMD --help
  Command names are canonical; only --help-json also has the short --hj alias.

Agent Use:
  Prefer twenty CMD --help-json before executing mutations.
  Stable JSON fields: path, args, options, operations, capabilities, exit_codes, output_contract.

Env Precedence:
  .env then .env.local then the explicit env file; existing environment variables win.
`;

const HELP_JSON_FLAG_ALIASES = ["--help-json", "--hj"];

const EXIT_CODES: HelpExitCode[] = [
  {
    code: 0,
    summary: "Success, help output, or version output",
  },
  {
    code: 1,
    summary: "General or unexpected error",
  },
  {
    code: 2,
    summary: "Invalid arguments or command usage error",
  },
  {
    code: 3,
    summary: "Authentication or permission error",
  },
  {
    code: 4,
    summary: "Network error or request failed before a response was received",
  },
  {
    code: 5,
    summary: "Rate limited (HTTP 429)",
  },
];

const OUTPUT_CONTRACT: HelpOutputContract = {
  query_language: "JMESPath",
  query_applies_before_format: true,
  formats: [
    {
      name: "text",
      summary: "Best-effort table rendering for objects and arrays.",
    },
    {
      name: "json",
      summary: "Pretty-printed JSON with 2-space indentation.",
    },
    {
      name: "jsonl",
      summary: "Newline-delimited JSON, one record per line for arrays.",
    },
    {
      name: "agent",
      summary: "Stable agent envelopes: arrays as items, objects as item, primitives as data.",
    },
    {
      name: "csv",
      summary: "Wraps singleton values as one record and JSON-encodes nested values.",
    },
  ],
};

const GLOBAL_OPTION_NAMES = new Set([
  "output",
  "query",
  "workspace",
  "env-file",
  "debug",
  "no-retry",
]);

const METADATA: Record<string, HelpMetadata> = {
  twenty: {
    examples: [
      "twenty auth status",
      "twenty api person list -o json",
      "twenty dashboards duplicate <dashboard-id>",
      "twenty public-domains list",
      "twenty emailing-domains list",
      "twenty event-logs list --table workspace-event",
      "twenty postgres-proxy get",
      "twenty roles list --include-targets",
      "twenty skills list",
      "twenty openapi core",
      "twenty serverless list",
      "twenty routes invoke public/ping",
      "twenty workflows invoke-webhook <workflow-id>",
    ],
  },
  "twenty api": {
    operations: [
      { name: "list" },
      { name: "get" },
      { name: "create" },
      { name: "update" },
      { name: "delete" },
      {
        name: "destroy",
        summary: "Permanently destroy a record",
      },
      {
        name: "restore",
        summary: "Restore a soft-deleted record",
      },
      {
        name: "batch-create",
        summary: "Create multiple records",
      },
      {
        name: "batch-update",
        summary: "Update multiple records",
      },
      {
        name: "batch-delete",
        summary: "Delete multiple records",
      },
      {
        name: "import",
        summary: "Import records",
      },
      {
        name: "export",
        summary: "Export records",
        mutates: false,
      },
      {
        name: "group-by",
        summary: "Group records by a field",
        mutates: false,
      },
      {
        name: "find-duplicates",
        summary: "Find duplicate records",
        mutates: false,
      },
      {
        name: "merge",
        summary: "Merge duplicate records",
      },
    ],
    examples: [
      "twenty api person list --limit 10 -o json",
      'twenty api note create --data \'{"title":"Hello"}\'',
    ],
  },
  "twenty graphql": {
    operations: [
      { name: "query", summary: "Run a GraphQL query", mutates: false },
      { name: "mutate", summary: "Run a GraphQL mutation", mutates: true },
      { name: "schema", summary: "Inspect the GraphQL schema", mutates: false },
    ],
  },
  "twenty routes": {
    examples: [
      "twenty routes invoke public/ping",
      'twenty routes invoke hooks/import --method post --data \'{"source":"cli"}\'',
    ],
  },
  "twenty roles": {
    operations: [
      {
        name: "upsert-permission-flags",
        summary: "Upsert permission flags for a role",
      },
      {
        name: "upsert-object-permissions",
        summary: "Upsert object permissions for a role",
      },
      {
        name: "upsert-field-permissions",
        summary: "Upsert field permissions for a role",
      },
      {
        name: "assign-agent",
        summary: "Assign a role to an agent",
      },
      {
        name: "remove-agent",
        summary: "Remove a role from an agent",
      },
    ],
    examples: [
      "twenty roles list --include-targets",
      "twenty roles get <role-id> --include-permissions",
      'twenty roles create --data \'{"label":"Support"}\'',
    ],
  },
  "twenty skills": {
    examples: [
      "twenty skills list",
      "twenty skills get <skill-id> -o json",
      'twenty skills create --data \'{"name":"workflow-building"}\'',
    ],
  },
  "twenty workflows": {
    operations: [
      { name: "invoke-webhook", summary: "Invoke a public workflow webhook", mutates: true },
      { name: "activate", summary: "Activate a workflow version", mutates: true },
      { name: "deactivate", summary: "Deactivate a workflow version", mutates: true },
      { name: "run", summary: "Run a workflow version", mutates: true },
      { name: "stop-run", summary: "Stop a workflow run", mutates: true },
    ],
    examples: [
      "twenty workflows invoke-webhook <workflow-id> --workspace-id <workspace-id>",
      "twenty workflows invoke-webhook <workflow-id> --method get --param source=cli",
      "twenty workflows activate <workflow-version-id>",
      'twenty workflows run <workflow-version-id> --data \'{"source":"cli"}\'',
      "twenty workflows stop-run <workflow-run-id>",
    ],
  },
  "twenty api-keys": {
    operations: [
      { name: "revoke", summary: "Revoke an API key" },
      { name: "assign-role", summary: "Assign a role to an API key" },
    ],
    examples: [
      "twenty api-keys list",
      "twenty api-keys get <api-key-id>",
      'twenty api-keys create --data \'{"name":"CI key"}\'',
    ],
  },
  "twenty applications": {
    operations: [
      { name: "sync", summary: "Sync an application manifest" },
      { name: "uninstall", summary: "Uninstall an application" },
      {
        name: "update-variable",
        summary: "Update an application variable",
      },
      {
        name: "create-development",
        summary: "Create a development application",
      },
      {
        name: "generate-token",
        summary: "Generate tokens for an application",
      },
    ],
    examples: [
      "twenty applications list",
      "twenty applications get <application-id>",
      "twenty applications sync --manifest-file ./manifest.json",
      "twenty applications create-development com.example.app --name 'Example App'",
      "twenty applications uninstall com.example.app",
    ],
  },
  "twenty application-registrations": {
    operations: [
      {
        name: "stats",
        summary: "Get stats for an application registration",
        mutates: false,
      },
      {
        name: "tarball-url",
        summary: "Get the tarball URL for an application registration",
        mutates: false,
      },
      {
        name: "list-variables",
        summary: "List variables for an application registration",
        mutates: false,
      },
      {
        name: "create-variable",
        summary: "Create a variable for an application registration",
        mutates: true,
      },
      {
        name: "update-variable",
        summary: "Update a variable for an application registration",
        mutates: true,
      },
      {
        name: "delete-variable",
        summary: "Delete a variable for an application registration",
        mutates: true,
      },
      {
        name: "rotate-secret",
        summary: "Rotate the client secret for an application registration",
        mutates: true,
      },
      {
        name: "transfer-ownership",
        summary: "Transfer ownership of an application registration",
        mutates: true,
      },
    ],
    examples: [
      "twenty application-registrations list",
      "twenty application-registrations get <application-registration-id>",
      "twenty application-registrations rotate-secret <application-registration-id>",
    ],
  },
  "twenty marketplace-apps": {
    operations: [{ name: "install", summary: "Install a marketplace app" }],
    examples: [
      "twenty marketplace-apps list",
      "twenty marketplace-apps get com.example.app",
      "twenty marketplace-apps install com.example.app --version 1.2.0",
    ],
  },
  "twenty dashboards duplicate": {
    examples: ["twenty dashboards duplicate <dashboard-id>"],
  },
  "twenty public-domains": {
    operations: [
      {
        name: "check-records",
        summary: "Check DNS records for a public domain",
        mutates: false,
      },
    ],
    examples: [
      "twenty public-domains list",
      "twenty public-domains create --domain app.example.com",
      "twenty public-domains check-records --domain app.example.com",
    ],
  },
  "twenty approved-access-domains": {
    operations: [
      {
        name: "validate",
        summary: "Validate an approved access domain",
        mutates: false,
      },
    ],
  },
  "twenty emailing-domains": {
    operations: [
      {
        name: "verify",
        summary: "Verify an emailing domain",
      },
    ],
    examples: [
      "twenty emailing-domains list",
      "twenty emailing-domains create --domain mail.example.com",
      "twenty emailing-domains verify <emailing-domain-id>",
    ],
  },
  "twenty postgres-proxy": {
    operations: [
      { name: "enable", summary: "Enable the Postgres proxy" },
      { name: "disable", summary: "Disable the Postgres proxy" },
    ],
    examples: [
      "twenty postgres-proxy get",
      "twenty postgres-proxy enable --show-password",
      "twenty postgres-proxy disable",
    ],
  },
  "twenty connected-accounts": {
    operations: [
      {
        name: "sync",
        summary: "Start a sync for a connected account",
      },
      {
        name: "get-imap-smtp-caldav",
        summary: "Get IMAP, SMTP, and CalDAV settings for a connected account",
        mutates: false,
      },
      {
        name: "save-imap-smtp-caldav",
        summary: "Save IMAP, SMTP, and CalDAV settings for a connected account",
      },
    ],
    examples: [
      "twenty connected-accounts list",
      "twenty connected-accounts get <connected-account-id>",
      "twenty connected-accounts get-imap-smtp-caldav <connected-account-id>",
      "twenty connected-accounts save-imap-smtp-caldav --account-owner-id <workspace-member-id> --handle mailbox@example.com --file account.json",
    ],
  },
  "twenty serverless": {
    operations: [
      { name: "list", summary: "List serverless functions", mutates: false },
      { name: "get", summary: "Get one serverless function", mutates: false },
      { name: "create", summary: "Create a serverless function", mutates: true },
      { name: "update", summary: "Update a serverless function", mutates: true },
      { name: "delete", summary: "Delete a serverless function", mutates: true },
      { name: "publish", summary: "Publish a serverless function", mutates: true },
      { name: "execute", summary: "Execute a serverless function", mutates: true },
      {
        name: "packages",
        summary: "List available packages for a serverless function",
        mutates: false,
      },
      {
        name: "source",
        summary: "Get the source code for a serverless function",
        mutates: false,
      },
      {
        name: "logs",
        summary: "Stream logs for a serverless function",
        mutates: false,
      },
      { name: "create-layer", summary: "Create a serverless layer", mutates: true },
      {
        name: "available-packages",
        summary: "List available packages for a serverless function",
        mutates: false,
      },
    ],
    examples: [
      "twenty serverless list",
      "twenty serverless source <serverless-function-id> -o json",
      "twenty serverless logs <serverless-function-id> --max-events 1 -o jsonl",
      "twenty serverless create-layer --package-json '{\"dependencies\":{}}' --yarn-lock 'lockfile'",
      "twenty serverless publish <serverless-function-id>",
    ],
  },
  "twenty event-logs": {
    operations: [{ name: "list", summary: "Query event logs", mutates: false }],
    examples: [
      "twenty event-logs list --table workspace-event",
      "twenty event-logs list --table pageview --include-page-info -o json",
      "twenty event-logs list --table object-event --object-metadata-id <object-metadata-id>",
    ],
  },
  "twenty openapi": {
    examples: [
      "twenty openapi core",
      "twenty openapi metadata --output-file metadata-openapi.json",
    ],
  },
  "twenty files": {
    operations: [
      { name: "upload", summary: "Upload a file", mutates: true },
      { name: "download", summary: "Download a file", mutates: false },
      { name: "public-asset", summary: "Download a public asset", mutates: false },
    ],
  },
};

export async function maybeHandleInlineHelp(
  program: Command,
  args: string[],
  write: HelpWriter = defaultWrite,
): Promise<boolean> {
  if (hasHelpJsonFlag(args)) {
    write(JSON.stringify(buildHelpJson(program, args), null, 2));
    return true;
  }

  if (shouldRenderRootHelp(args)) {
    write(loadRootHelpText());
    return true;
  }

  return false;
}

export function buildHelpJson(program: Command, args: string[]): HelpDocument {
  const { command, path: commandPath } = resolveTargetCommand(program, args);
  const commandKey = commandPath.join(" ");
  const metadata = METADATA[commandKey] ?? {};
  const options = getHelpOptions(command);
  const operations = getHelpOperations(command, commandKey, metadata);
  const capabilities = {
    mutates: operations.some((operation) => operation.mutates),
    supports_query: options.some((option) => option.name === "query"),
    supports_workspace: options.some((option) => option.name === "workspace"),
    supports_output: options.some((option) => option.name === "output"),
  };

  return {
    schema_version: 1,
    kind: command === program ? "root" : "command",
    name: command.name(),
    aliases: command.aliases(),
    path: commandPath,
    summary: command.description() || "",
    description: command.description() || undefined,
    usage: buildUsage(command, commandPath),
    examples: metadata.examples ?? [],
    args: getHelpArguments(command),
    options,
    operations,
    subcommands: getVisibleSubcommands(command),
    capabilities,
    exit_codes: EXIT_CODES,
    output_contract:
      capabilities.supports_output || capabilities.supports_query ? OUTPUT_CONTRACT : undefined,
  };
}

function loadRootHelpText(): string {
  const helpPath = path.join(__dirname, "help.txt");
  if (fs.pathExistsSync(helpPath)) {
    return fs.readFileSync(helpPath, "utf-8");
  }

  return ROOT_HELP_FALLBACK;
}

function shouldRenderRootHelp(args: string[]): boolean {
  if (args.length === 0) {
    return true;
  }

  const hasHelpFlag = args.includes("--help") || args.includes("-h");
  if (!hasHelpFlag) {
    return false;
  }

  const firstCommandToken = args.find((token) => !token.startsWith("-"));
  return firstCommandToken === undefined;
}

function resolveTargetCommand(
  program: Command,
  args: string[],
): { command: Command; path: string[] } {
  const sanitizedArgs = args.filter((token) => !isTruthyHelpJsonFlag(token));
  const pathParts = [program.name()];
  let current = program;

  for (let index = 0; index < sanitizedArgs.length; index += 1) {
    const token = sanitizedArgs[index];
    if (token === "--") {
      break;
    }

    const option = findMatchingOption(current, token);
    if (option) {
      if ((option.required || option.optional) && !token.includes("=")) {
        index += 1;
      }
      continue;
    }

    if (token.startsWith("-")) {
      continue;
    }

    const nextCommand = current.commands.find(
      (candidate) => !candidate.name().startsWith("help") && candidate.name() === token,
    );

    if (!nextCommand) {
      break;
    }

    current = nextCommand;
    pathParts.push(nextCommand.name());
  }

  return { command: current, path: pathParts };
}

function hasHelpJsonFlag(args: string[]): boolean {
  return args.some((token) => isTruthyHelpJsonFlag(token));
}

function isTruthyHelpJsonFlag(token: string): boolean {
  if (HELP_JSON_FLAG_ALIASES.includes(token)) {
    return true;
  }

  for (const flag of HELP_JSON_FLAG_ALIASES) {
    if (!token.startsWith(`${flag}=`)) {
      continue;
    }

    const rawValue = token
      .slice(flag.length + 1)
      .trim()
      .toLowerCase();
    return rawValue === "1" || rawValue === "true";
  }

  return false;
}

function findMatchingOption(command: Command, token: string): Option | undefined {
  return command.options.find((option) => {
    if (option.long === token || option.short === token) {
      return true;
    }

    return Boolean(option.long && token.startsWith(`${option.long}=`));
  });
}

function getHelpArguments(command: Command): HelpArgument[] {
  return (command.registeredArguments ?? []).map((argument) => ({
    name: argument.name(),
    required: argument.required,
    variadic: argument.variadic,
    description: argument.description || undefined,
  }));
}

function getHelpOptions(command: Command): HelpOption[] {
  return command.options
    .filter((option) => option.long !== "--help")
    .map((option) => {
      const longName = (option.long ?? `--${option.attributeName()}`).replace(/^--/, "");

      return {
        name: longName,
        flags: option.flags,
        type: inferOptionType(option),
        default:
          option.defaultValue === undefined ? undefined : JSON.stringify(option.defaultValue),
        required: option.mandatory ?? false,
        global: GLOBAL_OPTION_NAMES.has(longName),
        description: option.description || "",
      };
    });
}

function inferOptionType(option: Option): string {
  if (option.required || option.optional) {
    return "string";
  }

  return "boolean";
}

function getHelpOperations(
  command: Command,
  commandKey: string,
  metadata: HelpMetadata,
): HelpOperation[] {
  const parsedOperations = parseOperationsFromArguments(command.registeredArguments ?? []) ?? [];
  const sourceOperations = mergeOperationMetadata(parsedOperations, metadata.operations ?? []);

  return sourceOperations.map((operation) => ({
    name: operation.name,
    summary: operation.summary ?? summarizeOperation(commandKey, operation.name),
    mutates: operation.mutates ?? inferMutation(operation.name),
  }));
}

function mergeOperationMetadata(
  parsedOperations: HelpOperationMetadata[],
  metadataOperations: HelpOperationMetadata[],
): HelpOperationMetadata[] {
  if (parsedOperations.length === 0) {
    return metadataOperations;
  }

  if (metadataOperations.length === 0) {
    return parsedOperations;
  }

  const metadataByName = new Map(
    metadataOperations.map((operation) => [operation.name, operation]),
  );
  const merged = parsedOperations.map((operation) => ({
    ...operation,
    ...metadataByName.get(operation.name),
  }));

  for (const operation of metadataOperations) {
    if (!merged.some((candidate) => candidate.name === operation.name)) {
      merged.push(operation);
    }
  }

  return merged;
}

function parseOperationsFromArguments(
  argumentsList: ReadonlyArray<{ name(): string; description?: string }>,
): HelpOperationMetadata[] | undefined {
  const operationArgument = argumentsList.find((argument) => argument.name() === "operation");
  if (!operationArgument?.description) {
    return undefined;
  }

  const normalized = operationArgument.description.replace(/\bor\b/gi, ",");
  const operations = normalized
    .split(",")
    .map((value) => value.trim())
    .filter((value) => /^[a-z][a-z0-9-]*$/i.test(value));

  if (operations.length === 0) {
    return undefined;
  }

  return operations.map((name) => ({ name }));
}

function summarizeOperation(commandKey: string, operation: string): string {
  const parts = commandKey.split(" ");
  const resource = humanizeResource(parts[parts.length - 1] ?? "resource");
  const singularResource = singularizeResource(resource);

  switch (operation) {
    case "list":
      return `List ${resource}`;
    case "get":
      return `Get one ${singularResource}`;
    case "create":
      return `Create ${indefiniteArticle(singularResource)} ${singularResource}`;
    case "update":
      return `Update ${indefiniteArticle(singularResource)} ${singularResource}`;
    case "delete":
      return `Delete ${indefiniteArticle(singularResource)} ${singularResource}`;
    case "activate":
      return `Activate ${indefiniteArticle(singularResource)} ${singularResource}`;
    case "deactivate":
      return `Deactivate ${indefiniteArticle(singularResource)} ${singularResource}`;
    default:
      return operation.replace(/-/g, " ");
  }
}

function humanizeResource(resource: string): string {
  return resource.replace(/-/g, " ");
}

function singularizeResource(resource: string): string {
  if (resource.endsWith("ies")) {
    return `${resource.slice(0, -3)}y`;
  }

  if (resource.endsWith("ss")) {
    return resource;
  }

  if (resource.endsWith("s")) {
    return resource.slice(0, -1);
  }

  return resource;
}

function indefiniteArticle(resource: string): "a" | "an" {
  return /^[aeiou]/i.test(resource) ? "an" : "a";
}

function inferMutation(operation: string): boolean {
  const readOnlyOperations = new Set([
    "discover",
    "find-duplicates",
    "get",
    "group-by",
    "list",
    "packages",
    "query",
    "schema",
    "search",
    "source",
    "status",
    "validate",
    "workspace",
  ]);

  return !readOnlyOperations.has(operation);
}

function getVisibleSubcommands(command: Command): HelpSubcommand[] {
  return command.commands
    .filter(
      (candidate) => candidate.name() !== "help" && !candidate.name().startsWith("completion"),
    )
    .map((candidate) => ({
      name: candidate.name(),
      summary: candidate.description() || "",
    }));
}

function buildUsage(command: Command, commandPath: string[]): string {
  if (commandPath.length === 1) {
    return `${commandPath[0]} [command] [options]`;
  }

  const usage = command.usage();
  if (!usage) {
    return commandPath.join(" ");
  }

  return `${commandPath.join(" ")} ${usage}`;
}

function defaultWrite(text: string): void {
  // eslint-disable-next-line no-console
  console.log(text);
}
