import { Command } from 'commander';
import { applyGlobalOptions, resolveGlobalOptions } from '../../utilities/shared/global-options';
import { createServices } from '../../utilities/shared/services';
import { CliError } from '../../utilities/errors/cli-error';
import { ApiMetadataContext, ApiMetadataOptions } from './operations/types';
import { runObjectsList } from './operations/objects-list.operation';
import { runObjectsGet } from './operations/objects-get.operation';
import { runObjectsCreate } from './operations/objects-create.operation';
import { runFieldsList } from './operations/fields-list.operation';
import { runFieldsGet } from './operations/fields-get.operation';
import { runFieldsCreate } from './operations/fields-create.operation';

const handlers: Record<string, (ctx: ApiMetadataContext) => Promise<void>> = {
  'objects:list': runObjectsList,
  'objects:get': runObjectsGet,
  'objects:create': runObjectsCreate,
  'fields:list': runFieldsList,
  'fields:get': runFieldsGet,
  'fields:create': runFieldsCreate,
};

export function registerApiMetadataCommand(program: Command): void {
  const cmd = program
    .command('api-metadata')
    .description('Schema operations')
    .argument('<type>', 'Metadata type (objects or fields)')
    .argument('<operation>', 'Operation to perform')
    .argument('[arg]', 'Identifier')
    .option('-d, --data <json>', 'JSON payload')
    .option('-f, --file <path>', 'JSON payload file (use - for stdin)')
    .option('--object <name>', 'Filter fields by object');

  applyGlobalOptions(cmd);

  cmd.action(async (type: string, operation: string, arg?: string, options?: ApiMetadataOptions | Command, command?: Command) => {
    const key = `${type.toLowerCase()}:${operation.toLowerCase()}`;
    const handler = handlers[key];
    if (!handler) {
      throw new CliError(`Unknown api-metadata operation ${JSON.stringify(type)} ${JSON.stringify(operation)}.`, 'INVALID_ARGUMENTS');
    }

    const resolvedCommand = command ?? (options instanceof Command ? options : cmd);
    const globalOptions = resolveGlobalOptions(resolvedCommand);
    const services = createServices(globalOptions);
    const rawOptions = resolvedCommand.opts() as ApiMetadataOptions;

    await handler({
      type,
      operation,
      arg,
      options: rawOptions ?? {},
      services,
      globalOptions,
    });
  });
}
