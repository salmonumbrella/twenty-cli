import { Command } from "commander";
import { loadCliEnvironment } from "../config/services/environment.service";
import { CliError } from "../errors/cli-error";
import { parseBooleanEnv } from "./parse";

export type OutputFormat = "json" | "jsonl" | "csv" | "text";

export interface GlobalOptions {
  output?: OutputFormat;
  query?: string;
  workspace?: string;
  debug?: boolean;
  noRetry?: boolean;
  envFile?: string;
  outputKind?: string;
  light?: boolean;
  full?: boolean;
  agentMode?: boolean;
}

export interface GlobalOptionSettings {
  includeQuery?: boolean;
}

interface GlobalOptionDefinition {
  name: string;
  flags: string;
  description: string;
  takesValue: boolean;
}

const GLOBAL_OPTION_DEFINITIONS: GlobalOptionDefinition[] = [
  {
    name: "output",
    flags: "-o, --output <format>",
    description: "Output format: json, jsonl, csv, text",
    takesValue: true,
  },
  {
    name: "query",
    flags: "--query <expression>",
    description: "JMESPath query filter",
    takesValue: true,
  },
  {
    name: "workspace",
    flags: "--workspace <name>",
    description: "Workspace profile to use",
    takesValue: true,
  },
  {
    name: "env-file",
    flags: "--env-file <path>",
    description: "Load environment variables from file",
    takesValue: true,
  },
  {
    name: "debug",
    flags: "--debug",
    description: "Show request/response details",
    takesValue: false,
  },
  {
    name: "no-retry",
    flags: "--no-retry",
    description: "Disable automatic retry",
    takesValue: false,
  },
  {
    name: "light",
    flags: "--light",
    description: "Render compact short-key JSON",
    takesValue: false,
  },
  {
    name: "li",
    flags: "--li",
    description: "Alias for --light",
    takesValue: false,
  },
  {
    name: "full",
    flags: "--full",
    description: "Render canonical full JSON",
    takesValue: false,
  },
  {
    name: "agent-mode",
    flags: "--agent-mode",
    description: "Agent mode: JSON output with light payloads by default",
    takesValue: false,
  },
  {
    name: "ai",
    flags: "--ai",
    description: "Alias for --agent-mode",
    takesValue: false,
  },
];

export const GLOBAL_OPTION_NAMES = new Set(
  GLOBAL_OPTION_DEFINITIONS.map((definition) => definition.name),
);

export const GLOBAL_OPTION_VALUE_TOKENS = new Set(
  GLOBAL_OPTION_DEFINITIONS.flatMap((definition) => {
    if (!definition.takesValue) {
      return [];
    }

    return definition.flags
      .split(",")
      .map((flag) => flag.trim().split(" ")[0]!)
      .filter(Boolean);
  }),
);

export function isGlobalOptionValueToken(token: string): boolean {
  return [...GLOBAL_OPTION_VALUE_TOKENS].some(
    (option) => token === option || token.startsWith(`${option}=`),
  );
}

export function applyGlobalOptions(command: Command, settings: GlobalOptionSettings = {}): void {
  for (const definition of GLOBAL_OPTION_DEFINITIONS) {
    if (definition.name === "query" && settings.includeQuery === false) {
      continue;
    }

    command.option(definition.flags, definition.description);
  }
}

export function resolveGlobalOptions(
  command: Command,
  overrides?: { outputQuery?: string },
): GlobalOptions {
  const opts = getCommandOptions(command);
  const envFile = typeof opts.envFile === "string" ? opts.envFile : undefined;

  loadCliEnvironment({
    argv: process.argv,
    cwd: process.cwd(),
    explicitEnvFile: envFile,
  });

  const agentMode = Boolean(opts.agentMode || opts.ai || parseBooleanEnv(process.env.TWENTY_AGENT));
  const rawOutput =
    typeof opts.output === "string" ? opts.output : (process.env.TWENTY_OUTPUT ?? "json");
  let output = parseOutputFormat(rawOutput);
  if (agentMode) {
    output = "json";
  }
  const full = Boolean(opts.full);
  const explicitLight = Boolean(opts.light || opts.li);
  if (explicitLight && full) {
    throw new CliError("--light and --full cannot be used together.", "INVALID_ARGUMENTS");
  }
  const defaultsToLight = output === "json" || output === "jsonl";
  const light = full ? false : explicitLight || defaultsToLight;
  const query =
    overrides?.outputQuery ??
    (typeof opts.query === "string" ? opts.query : undefined) ??
    process.env.TWENTY_QUERY ??
    undefined;
  const workspace =
    typeof opts.workspace === "string" ? opts.workspace : process.env.TWENTY_PROFILE;
  const debug =
    typeof opts.debug === "boolean"
      ? opts.debug
      : (parseBooleanEnv(process.env.TWENTY_DEBUG) ?? false);
  const envNoRetry = parseBooleanEnv(process.env.TWENTY_NO_RETRY) ?? false;
  const retry = typeof opts.retry === "boolean" ? opts.retry : undefined;
  const noRetry = retry === false ? true : envNoRetry;

  return {
    output,
    query,
    workspace,
    debug,
    noRetry,
    envFile,
    outputKind: deriveCommandKind(command),
    light,
    full,
    agentMode,
  };
}

function getCommandOptions(command: Command): Record<string, unknown> {
  const optsFn = (command as any).optsWithGlobals as undefined | (() => Record<string, unknown>);
  if (typeof optsFn === "function") {
    return optsFn.call(command);
  }
  return command.opts();
}

function parseOutputFormat(value: unknown): OutputFormat {
  if (value === "agent") {
    throw new CliError(
      'Output format "agent" has been removed; use --agent-mode and optionally --li or --full.',
      "INVALID_ARGUMENTS",
    );
  }
  if (value === "json" || value === "jsonl" || value === "csv" || value === "text") {
    return value;
  }

  throw new CliError(
    `Unsupported output format ${JSON.stringify(value)}. Valid formats: json, jsonl, csv, text.`,
    "INVALID_ARGUMENTS",
  );
}

function deriveCommandKind(command: Command): string {
  const path: string[] = [];
  let current: Command | null = command;

  while (current) {
    const name = current.name();
    if (name && name !== "help") {
      path.unshift(name);
    }
    current = current.parent ?? null;
  }

  return path.join(".");
}
