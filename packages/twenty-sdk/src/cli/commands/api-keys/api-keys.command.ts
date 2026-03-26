import { Command } from "commander";
import { type GraphQLResponse } from "../../utilities/api/graphql-response";
import { CliError } from "../../utilities/errors/cli-error";
import { applyGlobalOptions } from "../../utilities/shared/global-options";
import { createCommandContext } from "../../utilities/shared/context";

interface ApiKeyOptions {
  name?: string;
  expiresAt?: string;
  roleId?: string;
  revokedAt?: string;
}

export function registerApiKeysCommand(program: Command): void {
  const endpoint = "/metadata";
  const cmd = program.command("api-keys").description("Manage API keys");
  applyGlobalOptions(cmd);

  const listCmd = cmd.command("list").description("List API keys");
  applyGlobalOptions(listCmd);
  listCmd.action(async (_options: unknown, command: Command) => {
    const { globalOptions, services } = createCommandContext(command);
    const response = await services.api.post<GraphQLResponse<{ apiKeys: unknown[] }>>(endpoint, {
      query: `query { apiKeys { id name expiresAt revokedAt createdAt role { id name } } }`,
    });
    await services.output.render(response.data?.data?.apiKeys ?? [], {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });

  const getCmd = cmd.command("get").description("Get an API key").argument("[id]", "API key ID");
  applyGlobalOptions(getCmd);
  getCmd.action(async (id: string | undefined, _options: unknown, command: Command) => {
    const { globalOptions, services } = createCommandContext(command);
    if (!id) throw new CliError("Missing API key ID.", "INVALID_ARGUMENTS");
    const response = await services.api.post<GraphQLResponse<{ apiKey: unknown }>>(endpoint, {
      query: `query($id: UUID!) { apiKey(input: { id: $id }) { id name expiresAt revokedAt createdAt updatedAt role { id name } } }`,
      variables: { id },
    });
    await services.output.render(response.data?.data?.apiKey, {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });

  const createCmd = cmd.command("create").description("Create an API key");
  createCmd
    .option("--name <name>", "API key name")
    .option("--expires-at <date>", "Expiration date (ISO format)")
    .option("--role-id <id>", "Role ID for the API key");
  applyGlobalOptions(createCmd);
  createCmd.action(async (options: ApiKeyOptions, command: Command) => {
    const { globalOptions, services } = createCommandContext(command);
    if (!options.name) throw new CliError("Missing --name option.", "INVALID_ARGUMENTS");
    if (!options.expiresAt) throw new CliError("Missing --expires-at option.", "INVALID_ARGUMENTS");
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
  });

  const updateCmd = cmd
    .command("update")
    .description("Update an API key")
    .argument("[id]", "API key ID");
  updateCmd
    .option("--name <name>", "API key name")
    .option("--expires-at <date>", "Expiration date (ISO format)")
    .option("--revoked-at <date>", "Revoked date (ISO format)");
  applyGlobalOptions(updateCmd);
  updateCmd.action(async (id: string | undefined, options: ApiKeyOptions, command: Command) => {
    const { globalOptions, services } = createCommandContext(command);
    if (!id) throw new CliError("Missing API key ID.", "INVALID_ARGUMENTS");
    if (!options.name && !options.expiresAt && options.revokedAt === undefined) {
      throw new CliError(
        "Provide at least one of --name, --expires-at, or --revoked-at.",
        "INVALID_ARGUMENTS",
      );
    }

    const response = await services.api.post<GraphQLResponse<{ updateApiKey: unknown }>>(endpoint, {
      query: `mutation($input: UpdateApiKeyInput!) { updateApiKey(input: $input) { id name expiresAt revokedAt role { id name } } }`,
      variables: {
        input: {
          id,
          name: options.name,
          expiresAt: options.expiresAt,
          revokedAt: options.revokedAt,
        },
      },
    });
    await services.output.render(response.data?.data?.updateApiKey, {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });

  const revokeCmd = cmd
    .command("revoke")
    .description("Revoke an API key")
    .argument("[id]", "API key ID");
  applyGlobalOptions(revokeCmd);
  revokeCmd.action(async (id: string | undefined, _options: unknown, command: Command) => {
    const { services } = createCommandContext(command);
    if (!id) throw new CliError("Missing API key ID.", "INVALID_ARGUMENTS");
    await services.api.post<GraphQLResponse<{ revokeApiKey: { id: string } }>>(endpoint, {
      query: `mutation($id: UUID!) { revokeApiKey(input: { id: $id }) { id } }`,
      variables: { id },
    });
    // eslint-disable-next-line no-console
    console.log(`API key ${id} revoked.`);
  });

  const assignRoleCmd = cmd
    .command("assign-role")
    .description("Assign a role to an API key")
    .argument("[id]", "API key ID")
    .option("--role-id <id>", "Role ID for the API key");
  applyGlobalOptions(assignRoleCmd);
  assignRoleCmd.action(async (id: string | undefined, options: ApiKeyOptions, command: Command) => {
    const { globalOptions, services } = createCommandContext(command);
    if (!id) throw new CliError("Missing API key ID.", "INVALID_ARGUMENTS");
    if (!options.roleId) throw new CliError("Missing --role-id option.", "INVALID_ARGUMENTS");

    const response = await services.api.post<GraphQLResponse<{ assignRoleToApiKey: boolean }>>(
      endpoint,
      {
        query: `mutation($id: UUID!, $roleId: UUID!) { assignRoleToApiKey(apiKeyId: $id, roleId: $roleId) }`,
        variables: { id, roleId: options.roleId },
      },
    );

    await services.output.render(
      {
        apiKeyId: id,
        roleId: options.roleId,
        assigned: response.data?.data?.assignRoleToApiKey ?? false,
      },
      { format: globalOptions.output, query: globalOptions.query },
    );
  });
}
