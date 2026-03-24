import { Command } from "commander";
import { type GraphQLResponse } from "../../utilities/api/graphql-response";
import { CliError } from "../../utilities/errors/cli-error";
import { applyGlobalOptions, resolveGlobalOptions } from "../../utilities/shared/global-options";
import { createServices } from "../../utilities/shared/services";

interface PostgresProxyOptions {
  showPassword?: boolean;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

const POSTGRES_CREDENTIAL_FIELDS = `
  id
  user
  password
  workspaceId
`;

const GET_POSTGRES_CREDENTIALS_QUERY = `query GetPostgresCredentials {
  getPostgresCredentials {
    ${POSTGRES_CREDENTIAL_FIELDS}
  }
}`;

const ENABLE_POSTGRES_PROXY_MUTATION = `mutation EnablePostgresProxy {
  enablePostgresProxy {
    ${POSTGRES_CREDENTIAL_FIELDS}
  }
}`;

const DISABLE_POSTGRES_PROXY_MUTATION = `mutation DisablePostgresProxy {
  disablePostgresProxy {
    ${POSTGRES_CREDENTIAL_FIELDS}
  }
}`;

export function registerPostgresProxyCommand(program: Command): void {
  const endpoint = "/graphql";
  const cmd = program
    .command("postgres-proxy")
    .description("Manage Postgres proxy credentials")
    .argument("<operation>", "get, enable, disable")
    .option("--show-password", "Show the Postgres proxy password");

  applyGlobalOptions(cmd);

  cmd.action(async (operation: string, options: PostgresProxyOptions, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const op = operation.toLowerCase();

    switch (op) {
      case "get": {
        const response = await services.api.post<
          GraphQLResponse<{ getPostgresCredentials?: unknown }>
        >(endpoint, {
          query: GET_POSTGRES_CREDENTIALS_QUERY,
        });

        await services.output.render(
          sanitizeCredentials(
            resolvePostgresResult(response.data, "getPostgresCredentials"),
            options,
          ),
          {
            format: globalOptions.output,
            query: globalOptions.query,
          },
        );
        break;
      }
      case "enable": {
        const response = await services.api.post<
          GraphQLResponse<{ enablePostgresProxy?: unknown }>
        >(endpoint, {
          query: ENABLE_POSTGRES_PROXY_MUTATION,
        });

        await services.output.render(
          sanitizeCredentials(resolvePostgresResult(response.data, "enablePostgresProxy"), options),
          {
            format: globalOptions.output,
            query: globalOptions.query,
          },
        );
        break;
      }
      case "disable": {
        const response = await services.api.post<
          GraphQLResponse<{ disablePostgresProxy?: unknown }>
        >(endpoint, {
          query: DISABLE_POSTGRES_PROXY_MUTATION,
        });

        await services.output.render(
          sanitizeCredentials(
            resolvePostgresResult(response.data, "disablePostgresProxy"),
            options,
          ),
          {
            format: globalOptions.output,
            query: globalOptions.query,
          },
        );
        break;
      }
      default:
        throw new CliError(`Unknown operation: ${operation}`, "INVALID_ARGUMENTS");
    }
  });
}

function sanitizeCredentials(value: unknown, options: PostgresProxyOptions): unknown {
  if (options.showPassword || value == null || typeof value !== "object" || Array.isArray(value)) {
    return value;
  }

  const record = { ...(value as Record<string, unknown>) };
  if ("password" in record) {
    record.password = maskSecret(record.password);
  }

  return record;
}

function resolvePostgresResult<T>(
  response: GraphQLResponse<Record<string, T | null>> | undefined,
  key: string,
): T | null {
  if (Array.isArray(response?.errors) && response.errors.length > 0) {
    if (response.errors.some((error) => error.message?.includes(key))) {
      throw new CliError(
        `Postgres proxy is not available on this workspace because it does not expose ${key}.`,
        "API_ERROR",
      );
    }

    throw new CliError(
      response.errors
        .map((error) => error.message?.trim())
        .filter((message): message is string => Boolean(message))
        .join("\n") || "Postgres proxy request failed.",
      "API_ERROR",
    );
  }

  const data = response?.data;
  if (!isRecord(data) || !Object.prototype.hasOwnProperty.call(data, key)) {
    throw new CliError("Postgres proxy request returned an unexpected response.", "API_ERROR");
  }

  return (data as Record<string, T | null>)[key] ?? null;
}

function maskSecret(value: unknown): unknown {
  if (value == null || value === "") {
    return value;
  }

  return "[hidden]";
}
