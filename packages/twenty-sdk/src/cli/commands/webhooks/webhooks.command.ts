import { Command } from 'commander';
import { applyGlobalOptions, resolveGlobalOptions } from '../../utilities/shared/global-options';
import { createServices } from '../../utilities/shared/services';
import { CliError } from '../../utilities/errors/cli-error';
import { parseBody } from '../../utilities/shared/body';

interface GraphQLResponse<T = unknown> {
  data?: T;
  errors?: Array<{ message: string }>;
}

interface WebhooksOptions {
  data?: string;
  file?: string;
  set?: string[];
}

function collect(value: string, previous: string[] = []): string[] {
  return previous.concat([value]);
}

export function registerWebhooksCommand(program: Command): void {
  const cmd = program
    .command('webhooks')
    .description('Manage webhooks')
    .argument('<operation>', 'list, get, create, update, delete')
    .argument('[id]', 'Webhook ID')
    .option('-d, --data <json>', 'JSON payload')
    .option('-f, --file <path>', 'JSON file')
    .option('--set <key=value>', 'Set a field value', collect);

  applyGlobalOptions(cmd);

  cmd.action(async (operation: string, id: string | undefined, options: WebhooksOptions, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const op = operation.toLowerCase();

    switch (op) {
      case 'list': {
        const response = await services.api.post<GraphQLResponse<{ webhooks: unknown[] }>>('/graphql', {
          query: `query { webhooks { id targetUrl operations description createdAt } }`,
        });
        await services.output.render(response.data?.data?.webhooks ?? [], { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'get': {
        if (!id) throw new CliError('Missing webhook ID.', 'INVALID_ARGUMENTS');
        const response = await services.api.post<GraphQLResponse<{ webhook: unknown }>>('/graphql', {
          query: `query($id: UUID!) { webhook(id: $id) { id targetUrl operations description secret createdAt updatedAt } }`,
          variables: { id },
        });
        await services.output.render(response.data?.data?.webhook, { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'create': {
        const payload = await parseBody(options.data, options.file, options.set);
        const response = await services.api.post<GraphQLResponse<{ createWebhook: unknown }>>('/graphql', {
          query: `mutation($data: WebhookCreateInput!) { createWebhook(data: $data) { id targetUrl operations description } }`,
          variables: { data: payload },
        });
        await services.output.render(response.data?.data?.createWebhook, { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'update': {
        if (!id) throw new CliError('Missing webhook ID.', 'INVALID_ARGUMENTS');
        const payload = await parseBody(options.data, options.file, options.set);
        const response = await services.api.post<GraphQLResponse<{ updateWebhook: unknown }>>('/graphql', {
          query: `mutation($id: UUID!, $data: WebhookUpdateInput!) { updateWebhook(id: $id, data: $data) { id targetUrl operations description } }`,
          variables: { id, data: payload },
        });
        await services.output.render(response.data?.data?.updateWebhook, { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'delete': {
        if (!id) throw new CliError('Missing webhook ID.', 'INVALID_ARGUMENTS');
        await services.api.post<GraphQLResponse<{ deleteWebhook: { id: string } }>>('/graphql', {
          query: `mutation($id: UUID!) { deleteWebhook(id: $id) { id } }`,
          variables: { id },
        });
        // eslint-disable-next-line no-console
        console.log(`Webhook ${id} deleted.`);
        break;
      }
      default:
        throw new CliError(`Unknown operation: ${operation}`, 'INVALID_ARGUMENTS');
    }
  });
}
