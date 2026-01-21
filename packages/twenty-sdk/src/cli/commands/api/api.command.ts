import { Command } from 'commander';
import { applyGlobalOptions, resolveGlobalOptions } from '../../utilities/shared/global-options';
import { createServices } from '../../utilities/shared/services';
import { CliError } from '../../utilities/errors/cli-error';
import { ApiCommandOptions, ApiOperationContext } from './operations/types';
import { runListOperation } from './operations/list.operation';
import { runGetOperation } from './operations/get.operation';
import { runCreateOperation } from './operations/create.operation';
import { runUpdateOperation } from './operations/update.operation';
import { runDeleteOperation } from './operations/delete.operation';
import { runDestroyOperation } from './operations/destroy.operation';
import { runRestoreOperation } from './operations/restore.operation';
import { runBatchCreateOperation } from './operations/batch-create.operation';
import { runBatchUpdateOperation } from './operations/batch-update.operation';
import { runBatchDeleteOperation } from './operations/batch-delete.operation';
import { runImportOperation } from './operations/import.operation';
import { runExportOperation } from './operations/export.operation';
import { runGroupByOperation } from './operations/group-by.operation';
import { runFindDuplicatesOperation } from './operations/find-duplicates.operation';
import { runMergeOperation } from './operations/merge.operation';

const operationHandlers: Record<string, (ctx: ApiOperationContext) => Promise<void>> = {
  list: runListOperation,
  get: runGetOperation,
  create: runCreateOperation,
  update: runUpdateOperation,
  delete: runDeleteOperation,
  destroy: runDestroyOperation,
  restore: runRestoreOperation,
  'batch-create': runBatchCreateOperation,
  'batch-update': runBatchUpdateOperation,
  'batch-delete': runBatchDeleteOperation,
  import: runImportOperation,
  export: runExportOperation,
  'group-by': runGroupByOperation,
  'find-duplicates': runFindDuplicatesOperation,
  merge: runMergeOperation,
};

export function registerApiCommand(program: Command): void {
  const cmd = program
    .command('api')
    .description('Record operations')
    .argument('<object>', 'Object name (plural)')
    .argument('<operation>', 'Operation to perform')
    .argument('[arg]', 'Record ID or file path')
    .argument('[arg2]', 'Secondary argument')
    .option('--limit <number>', 'Limit number of records')
    .option('--all', 'Fetch all records')
    .option('--filter <expression>', 'Filter expression')
    .option('--include <relations>', 'Include related records')
    .option('--cursor <cursor>', 'Pagination cursor')
    .option('--sort <field>', 'Sort field')
    .option('--order <direction>', 'Sort order (asc or desc)')
    .option('--fields <fields>', 'Fields selection (comma-separated)')
    .option('--param <key=value>', 'Additional query params', collect)
    .option('-d, --data <json>', 'JSON payload')
    .option('-f, --file <path>', 'JSON/CSV file payload (use - for stdin)')
    .option('--set <key=value>', 'Set a field value', collect)
    .option('--force', 'Skip confirmation prompt')
    .option('--yes', 'Skip confirmation prompt (alias for --force)')
    .option('--ids <ids>', 'Comma-separated IDs')
    .option('--format <format>', 'Export format (json or csv)')
    .option('--output-file <path>', 'Output file path')
    .option('--batch-size <number>', 'Batch size (import)')
    .option('--dry-run', 'Preview without executing')
    .option('--continue-on-error', 'Continue on batch errors')
    .option('--field <field>', 'Group-by field')
    .option('--source <id>', 'Source record ID (merge)')
    .option('--target <id>', 'Target record ID (merge)')
    .option('--priority <index>', 'Conflict priority index (merge)');

  applyGlobalOptions(cmd);

  cmd.action(async (object: string, operation: string, arg?: string, arg2?: string, options?: ApiCommandOptions | Command, command?: Command) => {
    const op = operation.toLowerCase();
    const handler = operationHandlers[op];
    if (!handler) {
      throw new CliError(`Unknown operation ${JSON.stringify(operation)}.`, 'INVALID_ARGUMENTS');
    }

    const resolvedCommand = command ?? (options instanceof Command ? options : cmd);
    const globalOptions = resolveGlobalOptions(resolvedCommand);
    const services = createServices(globalOptions);

    const rawOptions = resolvedCommand.opts() as ApiCommandOptions;
    const mergedOptions: ApiCommandOptions = {
      ...rawOptions,
      force: rawOptions.force || rawOptions.yes,
    };

    await handler({
      object,
      arg,
      arg2,
      options: mergedOptions,
      services,
      globalOptions,
    });
  });
}

function collect(value: string, previous: string[] = []): string[] {
  return previous.concat([value]);
}
