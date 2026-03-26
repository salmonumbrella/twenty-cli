import { Command } from "commander";
import { loadCliEnvironment } from "../config/services/environment.service";
import { parseBooleanEnv } from "./parse";

export interface GlobalOptions {
  output?: string;
  query?: string;
  workspace?: string;
  debug?: boolean;
  noRetry?: boolean;
  envFile?: string;
  outputKind?: string;
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
    description: "Output format: text, json, jsonl, agent, csv",
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

  const rawOutput =
    typeof opts.output === "string" ? opts.output : (process.env.TWENTY_OUTPUT ?? "text");
  const output = isValidOutputFormat(rawOutput) ? rawOutput : "text";
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
  };
}

function getCommandOptions(command: Command): Record<string, unknown> {
  const optsFn = (command as any).optsWithGlobals as undefined | (() => Record<string, unknown>);
  if (typeof optsFn === "function") {
    return optsFn.call(command);
  }
  return command.opts();
}

function isValidOutputFormat(value: unknown): value is string {
  return (
    value === "text" ||
    value === "json" ||
    value === "jsonl" ||
    value === "agent" ||
    value === "csv"
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
