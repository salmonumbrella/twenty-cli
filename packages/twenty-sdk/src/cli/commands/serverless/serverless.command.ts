import { Command } from 'commander';
import { applyGlobalOptions, resolveGlobalOptions } from '../../utilities/shared/global-options';
import { createServices } from '../../utilities/shared/services';
import { CliError } from '../../utilities/errors/cli-error';
import { parseBody } from '../../utilities/shared/body';

interface GraphQLResponse<T = unknown> {
  data?: T;
  errors?: Array<{ message: string }>;
}

export function registerServerlessCommand(program: Command): void {
  const cmd = program
    .command('serverless')
    .description('Manage serverless functions')
    .argument('<operation>', 'list, get, create, update, delete, execute, publish, source')
    .argument('[id]', 'Function ID')
    .option('-d, --data <json>', 'JSON payload')
    .option('-f, --file <path>', 'JSON file')
    .option('--name <name>', 'Function name')
    .option('--description <text>', 'Function description');

  applyGlobalOptions(cmd);

  cmd.action(async (operation: string, id: string | undefined, options: ServerlessOptions, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const op = operation.toLowerCase();

    switch (op) {
      case 'list': {
        const response = await services.api.post<GraphQLResponse<{ findManyServerlessFunctions: unknown[] }>>('/graphql', {
          query: `query { findManyServerlessFunctions { id name description syncStatus createdAt updatedAt } }`,
        });
        await services.output.render(response.data?.data?.findManyServerlessFunctions ?? [], { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'get': {
        if (!id) throw new CliError('Missing function ID.', 'INVALID_ARGUMENTS');
        const response = await services.api.post<GraphQLResponse<{ findOneServerlessFunction: unknown }>>('/graphql', {
          query: `query($id: UUID!) { findOneServerlessFunction(id: $id) { id name description syncStatus sourceCodeFullPath createdAt updatedAt } }`,
          variables: { id },
        });
        await services.output.render(response.data?.data?.findOneServerlessFunction, { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'create': {
        if (!options.name) throw new CliError('Missing --name option.', 'INVALID_ARGUMENTS');
        const response = await services.api.post<GraphQLResponse<{ createOneServerlessFunction: unknown }>>('/graphql', {
          query: `mutation($name: String!, $description: String) { createOneServerlessFunction(input: { name: $name, description: $description }) { id name description } }`,
          variables: { name: options.name, description: options.description },
        });
        await services.output.render(response.data?.data?.createOneServerlessFunction, { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'update': {
        if (!id) throw new CliError('Missing function ID.', 'INVALID_ARGUMENTS');
        const payload = (options.data || options.file) ? await parseBody(options.data, options.file) : {};
        const response = await services.api.post<GraphQLResponse<{ updateOneServerlessFunction: unknown }>>('/graphql', {
          query: `mutation($id: UUID!, $name: String, $description: String) { updateOneServerlessFunction(input: { id: $id, name: $name, description: $description }) { id name description } }`,
          variables: { id, ...payload },
        });
        await services.output.render(response.data?.data?.updateOneServerlessFunction, { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'delete': {
        if (!id) throw new CliError('Missing function ID.', 'INVALID_ARGUMENTS');
        await services.api.post<GraphQLResponse<{ deleteOneServerlessFunction: { id: string } }>>('/graphql', {
          query: `mutation($id: UUID!) { deleteOneServerlessFunction(input: { id: $id }) { id } }`,
          variables: { id },
        });
        // eslint-disable-next-line no-console
        console.log(`Serverless function ${id} deleted.`);
        break;
      }
      case 'execute': {
        if (!id) throw new CliError('Missing function ID.', 'INVALID_ARGUMENTS');
        const payload = (options.data || options.file) ? await parseBody(options.data, options.file) : {};
        const response = await services.api.post<GraphQLResponse<{ executeOneServerlessFunction: unknown }>>('/graphql', {
          query: `mutation($id: UUID!, $payload: JSON!) { executeOneServerlessFunction(input: { id: $id, payload: $payload }) { data status duration } }`,
          variables: { id, payload },
        });
        await services.output.render(response.data?.data?.executeOneServerlessFunction, { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'publish': {
        if (!id) throw new CliError('Missing function ID.', 'INVALID_ARGUMENTS');
        const response = await services.api.post<GraphQLResponse<{ publishServerlessFunction: unknown }>>('/graphql', {
          query: `mutation($id: UUID!) { publishServerlessFunction(input: { id: $id }) { id syncStatus } }`,
          variables: { id },
        });
        await services.output.render(response.data?.data?.publishServerlessFunction, { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'source': {
        if (!id) throw new CliError('Missing function ID.', 'INVALID_ARGUMENTS');
        const response = await services.api.post<GraphQLResponse<{ getServerlessFunctionSourceCode: string }>>('/graphql', {
          query: `query($id: UUID!) { getServerlessFunctionSourceCode(input: { id: $id }) }`,
          variables: { id },
        });
        // eslint-disable-next-line no-console
        console.log(response.data?.data?.getServerlessFunctionSourceCode ?? '');
        break;
      }
      default:
        throw new CliError(`Unknown operation: ${operation}`, 'INVALID_ARGUMENTS');
    }
  });
}

interface ServerlessOptions {
  data?: string;
  file?: string;
  name?: string;
  description?: string;
}
