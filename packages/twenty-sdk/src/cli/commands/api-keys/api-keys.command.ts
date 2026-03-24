import { Command } from "commander";
import { type GraphQLResponse } from "../../utilities/api/graphql-response";
import { applyGlobalOptions, resolveGlobalOptions } from "../../utilities/shared/global-options";
import { createServices } from "../../utilities/shared/services";
import { CliError } from "../../utilities/errors/cli-error";

interface ApiKeyOptions {
  name?: string;
  expiresAt?: string;
  roleId?: string;
  revokedAt?: string;
}

export function registerApiKeysCommand(program: Command): void {
  const endpoint = "/metadata";
  const cmd = program
    .command("api-keys")
    .description("Manage API keys")
    .argument("<operation>", "list, get, create, update, revoke, assign-role")
    .argument("[id]", "API key ID")
    .option("--name <name>", "API key name")
    .option("--expires-at <date>", "Expiration date (ISO format)")
    .option("--role-id <id>", "Role ID for the API key")
    .option("--revoked-at <date>", "Revoked date (ISO format)");

  applyGlobalOptions(cmd);

  cmd.action(
    async (operation: string, id: string | undefined, options: ApiKeyOptions, command: Command) => {
      const globalOptions = resolveGlobalOptions(command);
      const services = createServices(globalOptions);
      const op = operation.toLowerCase();

      switch (op) {
        case "list": {
          const response = await services.api.post<GraphQLResponse<{ apiKeys: unknown[] }>>(
            endpoint,
            {
              query: `query { apiKeys { id name expiresAt revokedAt createdAt role { id name } } }`,
            },
          );
          await services.output.render(response.data?.data?.apiKeys ?? [], {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "get": {
          if (!id) throw new CliError("Missing API key ID.", "INVALID_ARGUMENTS");
          const response = await services.api.post<GraphQLResponse<{ apiKey: unknown }>>(endpoint, {
            query: `query($id: UUID!) { apiKey(input: { id: $id }) { id name expiresAt revokedAt createdAt updatedAt role { id name } } }`,
            variables: { id },
          });
          await services.output.render(response.data?.data?.apiKey, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "create": {
          if (!options.name) throw new CliError("Missing --name option.", "INVALID_ARGUMENTS");
          if (!options.expiresAt)
            throw new CliError("Missing --expires-at option.", "INVALID_ARGUMENTS");
          if (!options.roleId) throw new CliError("Missing --role-id option.", "INVALID_ARGUMENTS");
          const response = await services.api.post<
            GraphQLResponse<{ createApiKey: { id: string; name: string; expiresAt?: string } }>
          >(endpoint, {
            query: `mutation($input: CreateApiKeyInput!) { createApiKey(input: $input) { id name expiresAt role { id name } } }`,
            variables: {
              input: {
                name: options.name,
                expiresAt: options.expiresAt,
                roleId: options.roleId,
              },
            },
          });
          const result = response.data?.data?.createApiKey;
          if (!result) throw new CliError("Failed to create API key.", "API_ERROR");
          const tokenResponse = await services.api.post<
            GraphQLResponse<{ generateApiKeyToken: { token: string } }>
          >(endpoint, {
            query: `mutation($id: UUID!, $expiresAt: String!) { generateApiKeyToken(apiKeyId: $id, expiresAt: $expiresAt) { token } }`,
            variables: { id: result.id, expiresAt: options.expiresAt },
          });
          const output = {
            ...result,
            token: tokenResponse.data?.data?.generateApiKeyToken?.token,
          };
          await services.output.render(output, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "update": {
          if (!id) throw new CliError("Missing API key ID.", "INVALID_ARGUMENTS");
          if (!options.name && !options.expiresAt && options.revokedAt === undefined) {
            throw new CliError(
              "Provide at least one of --name, --expires-at, or --revoked-at.",
              "INVALID_ARGUMENTS",
            );
          }

          const response = await services.api.post<GraphQLResponse<{ updateApiKey: unknown }>>(
            endpoint,
            {
              query: `mutation($input: UpdateApiKeyInput!) { updateApiKey(input: $input) { id name expiresAt revokedAt role { id name } } }`,
              variables: {
                input: {
                  id,
                  name: options.name,
                  expiresAt: options.expiresAt,
                  revokedAt: options.revokedAt,
                },
              },
            },
          );
          await services.output.render(response.data?.data?.updateApiKey, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "revoke": {
          if (!id) throw new CliError("Missing API key ID.", "INVALID_ARGUMENTS");
          await services.api.post<GraphQLResponse<{ revokeApiKey: { id: string } }>>(endpoint, {
            query: `mutation($id: UUID!) { revokeApiKey(input: { id: $id }) { id } }`,
            variables: { id },
          });
          // eslint-disable-next-line no-console
          console.log(`API key ${id} revoked.`);
          break;
        }
        case "assign-role": {
          if (!id) throw new CliError("Missing API key ID.", "INVALID_ARGUMENTS");
          if (!options.roleId) throw new CliError("Missing --role-id option.", "INVALID_ARGUMENTS");

          const response = await services.api.post<
            GraphQLResponse<{ assignRoleToApiKey: boolean }>
          >(endpoint, {
            query: `mutation($id: UUID!, $roleId: UUID!) { assignRoleToApiKey(apiKeyId: $id, roleId: $roleId) }`,
            variables: { id, roleId: options.roleId },
          });

          await services.output.render(
            {
              apiKeyId: id,
              roleId: options.roleId,
              assigned: response.data?.data?.assignRoleToApiKey ?? false,
            },
            { format: globalOptions.output, query: globalOptions.query },
          );
          break;
        }
        default:
          throw new CliError(`Unknown operation: ${operation}`, "INVALID_ARGUMENTS");
      }
    },
  );
}
