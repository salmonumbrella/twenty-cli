import { Command } from 'commander';
import { applyGlobalOptions, resolveGlobalOptions } from '../../utilities/shared/global-options';
import { createServices } from '../../utilities/shared/services';
import { CliError } from '../../utilities/errors/cli-error';

interface GraphQLResponse<T = unknown> {
  data?: T;
  errors?: Array<{ message: string }>;
}

interface ApiKeyOptions {
  name?: string;
  expiresAt?: string;
}

export function registerApiKeysCommand(program: Command): void {
  const cmd = program
    .command('api-keys')
    .description('Manage API keys')
    .argument('<operation>', 'list, get, create, revoke')
    .argument('[id]', 'API key ID')
    .option('--name <name>', 'API key name')
    .option('--expires-at <date>', 'Expiration date (ISO format)');

  applyGlobalOptions(cmd);

  cmd.action(async (operation: string, id: string | undefined, options: ApiKeyOptions, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const op = operation.toLowerCase();

    switch (op) {
      case 'list': {
        const response = await services.api.post<GraphQLResponse<{ apiKeys: unknown[] }>>('/graphql', {
          query: `query { apiKeys { id name expiresAt revokedAt createdAt } }`,
        });
        await services.output.render(response.data?.data?.apiKeys ?? [], { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'get': {
        if (!id) throw new CliError('Missing API key ID.', 'INVALID_ARGUMENTS');
        const response = await services.api.post<GraphQLResponse<{ apiKey: unknown }>>('/graphql', {
          query: `query($id: UUID!) { apiKey(id: $id) { id name expiresAt revokedAt createdAt updatedAt } }`,
          variables: { id },
        });
        await services.output.render(response.data?.data?.apiKey, { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'create': {
        if (!options.name) throw new CliError('Missing --name option.', 'INVALID_ARGUMENTS');
        const response = await services.api.post<GraphQLResponse<{ createApiKey: { id: string; name: string; expiresAt?: string } }>>('/graphql', {
          query: `mutation($name: String!, $expiresAt: DateTime) { createApiKey(data: { name: $name, expiresAt: $expiresAt }) { id name expiresAt } }`,
          variables: { name: options.name, expiresAt: options.expiresAt },
        });
        const result = response.data?.data?.createApiKey;
        if (!result) throw new CliError('Failed to create API key.', 'API_ERROR');
        // Also get the token
        const tokenResponse = await services.api.post<GraphQLResponse<{ generateApiKeyToken: { token: string } }>>('/graphql', {
          query: `mutation($id: UUID!) { generateApiKeyToken(apiKeyId: $id) { token } }`,
          variables: { id: result.id },
        });
        const output = {
          ...result,
          token: tokenResponse.data?.data?.generateApiKeyToken?.token,
        };
        await services.output.render(output, { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'revoke': {
        if (!id) throw new CliError('Missing API key ID.', 'INVALID_ARGUMENTS');
        await services.api.post<GraphQLResponse<{ revokeApiKey: { id: string } }>>('/graphql', {
          query: `mutation($id: UUID!) { revokeApiKey(id: $id) { id } }`,
          variables: { id },
        });
        // eslint-disable-next-line no-console
        console.log(`API key ${id} revoked.`);
        break;
      }
      default:
        throw new CliError(`Unknown operation: ${operation}`, 'INVALID_ARGUMENTS');
    }
  });
}
