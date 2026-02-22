import { Command } from 'commander';
import { parseBooleanEnv } from './parse';

export interface GlobalOptions {
  output?: string;
  query?: string;
  workspace?: string;
  debug?: boolean;
  noRetry?: boolean;
}

export interface GlobalOptionSettings {
  includeQuery?: boolean;
}

export function applyGlobalOptions(command: Command, settings: GlobalOptionSettings = {}): void {
  const includeQuery = settings.includeQuery !== false;
  command.option('-o, --output <format>', 'Output format: text, json, csv');
  if (includeQuery) {
    command.option('--query <expression>', 'JMESPath query filter');
  }
  command.option('--workspace <name>', 'Workspace profile to use');
  command.option('--debug', 'Show request/response details');
  command.option('--no-retry', 'Disable automatic retry');
}

export function resolveGlobalOptions(command: Command, overrides?: { outputQuery?: string }): GlobalOptions {
  const opts = getCommandOptions(command);

  const rawOutput = typeof opts.output === 'string' ? opts.output : (process.env.TWENTY_OUTPUT ?? 'text');
  const output = isValidOutputFormat(rawOutput) ? rawOutput : 'text';
  const query = overrides?.outputQuery
    ?? (typeof opts.query === 'string' ? opts.query : undefined)
    ?? process.env.TWENTY_QUERY
    ?? undefined;
  const workspace = typeof opts.workspace === 'string' ? opts.workspace : process.env.TWENTY_PROFILE;
  const debug = typeof opts.debug === 'boolean' ? opts.debug : parseBooleanEnv(process.env.TWENTY_DEBUG) ?? false;
  const envNoRetry = parseBooleanEnv(process.env.TWENTY_NO_RETRY) ?? false;
  const retry = typeof opts.retry === 'boolean' ? opts.retry : undefined;
  const noRetry = retry === false ? true : envNoRetry;

  return {
    output,
    query,
    workspace,
    debug,
    noRetry,
  };
}

function getCommandOptions(command: Command): Record<string, unknown> {
  const optsFn = (command as any).optsWithGlobals as undefined | (() => Record<string, unknown>);
  if (typeof optsFn === 'function') {
    return optsFn.call(command);
  }
  return command.opts();
}

function isValidOutputFormat(value: unknown): value is string {
  return value === 'text' || value === 'json' || value === 'csv';
}
