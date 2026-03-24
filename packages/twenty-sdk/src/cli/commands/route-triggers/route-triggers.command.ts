import { Command } from "commander";
import { type GraphQLResponse } from "../../utilities/api/graphql-response";
import { CliError } from "../../utilities/errors/cli-error";
import { parseBody } from "../../utilities/shared/body";
import { applyGlobalOptions, resolveGlobalOptions } from "../../utilities/shared/global-options";
import { createServices } from "../../utilities/shared/services";

interface RouteTriggersOptions {
  data?: string;
  file?: string;
  set?: string[];
}

const ROUTE_TRIGGER_FIELDS = `
  id
  path
  isAuthRequired
  httpMethod
  createdAt
  updatedAt
`;

const FIND_MANY_ROUTE_TRIGGERS_QUERY = `query FindManyRouteTriggers {
  findManyRouteTriggers {
    ${ROUTE_TRIGGER_FIELDS}
  }
}`;

const FIND_ONE_ROUTE_TRIGGER_QUERY = `query FindOneRouteTrigger($id: String!) {
  findOneRouteTrigger(input: { id: $id }) {
    ${ROUTE_TRIGGER_FIELDS}
  }
}`;

const CREATE_ONE_ROUTE_TRIGGER_MUTATION = `mutation CreateOneRouteTrigger($input: CreateRouteTriggerInput!) {
  createOneRouteTrigger(input: $input) {
    ${ROUTE_TRIGGER_FIELDS}
  }
}`;

const UPDATE_ONE_ROUTE_TRIGGER_MUTATION = `mutation UpdateOneRouteTrigger($input: UpdateRouteTriggerInput!) {
  updateOneRouteTrigger(input: $input) {
    ${ROUTE_TRIGGER_FIELDS}
  }
}`;

const DELETE_ONE_ROUTE_TRIGGER_MUTATION = `mutation DeleteOneRouteTrigger($input: RouteTriggerIdInput!) {
  deleteOneRouteTrigger(input: $input) {
    ${ROUTE_TRIGGER_FIELDS}
  }
}`;

function collect(value: string, previous: string[] = []): string[] {
  return previous.concat([value]);
}

export function registerRouteTriggersCommand(program: Command): void {
  const endpoint = "/metadata";
  const cmd = program
    .command("route-triggers")
    .description("Manage route triggers")
    .argument("<operation>", "list, get, create, update, delete")
    .argument("[id]", "Route trigger ID")
    .option("-d, --data <json>", "JSON payload")
    .option("-f, --file <path>", "JSON file")
    .option("--set <key=value>", "Set a field value", collect);

  applyGlobalOptions(cmd);

  cmd.action(
    async (
      operation: string,
      id: string | undefined,
      options: RouteTriggersOptions,
      command: Command,
    ) => {
      const globalOptions = resolveGlobalOptions(command);
      const services = createServices(globalOptions);
      const op = operation.toLowerCase();

      switch (op) {
        case "list": {
          const response = await services.api.post<
            GraphQLResponse<{ findManyRouteTriggers?: unknown[] }>
          >(endpoint, {
            query: FIND_MANY_ROUTE_TRIGGERS_QUERY,
          });

          await services.output.render(response.data?.data?.findManyRouteTriggers ?? [], {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "get": {
          if (!id) {
            throw new CliError("Missing route trigger ID.", "INVALID_ARGUMENTS");
          }

          const response = await services.api.post<
            GraphQLResponse<{ findOneRouteTrigger?: unknown }>
          >(endpoint, {
            query: FIND_ONE_ROUTE_TRIGGER_QUERY,
            variables: { id },
          });

          await services.output.render(response.data?.data?.findOneRouteTrigger, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "create": {
          const payload = await parseBody(options.data, options.file, options.set);
          const response = await services.api.post<
            GraphQLResponse<{ createOneRouteTrigger?: unknown }>
          >(endpoint, {
            query: CREATE_ONE_ROUTE_TRIGGER_MUTATION,
            variables: { input: payload },
          });

          await services.output.render(response.data?.data?.createOneRouteTrigger, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "update": {
          if (!id) {
            throw new CliError("Missing route trigger ID.", "INVALID_ARGUMENTS");
          }

          const payload = await parseBody(options.data, options.file, options.set);
          const response = await services.api.post<
            GraphQLResponse<{ updateOneRouteTrigger?: unknown }>
          >(endpoint, {
            query: UPDATE_ONE_ROUTE_TRIGGER_MUTATION,
            variables: {
              input: {
                id,
                update: payload,
              },
            },
          });

          await services.output.render(response.data?.data?.updateOneRouteTrigger, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "delete": {
          if (!id) {
            throw new CliError("Missing route trigger ID.", "INVALID_ARGUMENTS");
          }

          const response = await services.api.post<
            GraphQLResponse<{ deleteOneRouteTrigger?: unknown }>
          >(endpoint, {
            query: DELETE_ONE_ROUTE_TRIGGER_MUTATION,
            variables: {
              input: {
                id,
              },
            },
          });

          await services.output.render(response.data?.data?.deleteOneRouteTrigger, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        default:
          throw new CliError(`Unknown operation: ${operation}`, "INVALID_ARGUMENTS");
      }
    },
  );
}
