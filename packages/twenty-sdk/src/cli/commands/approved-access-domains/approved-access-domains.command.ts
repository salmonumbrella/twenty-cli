import { Command } from "commander";
import { type GraphQLResponse } from "../../utilities/api/graphql-response";
import { CliError } from "../../utilities/errors/cli-error";
import { applyGlobalOptions, resolveGlobalOptions } from "../../utilities/shared/global-options";
import { createServices } from "../../utilities/shared/services";

interface ApprovedAccessDomainsOptions {
  validationToken?: string;
}

const LIST_APPROVED_ACCESS_DOMAINS_QUERY = `query GetApprovedAccessDomains {
  getApprovedAccessDomains {
    id
    domain
    isValidated
    createdAt
  }
}`;

const DELETE_APPROVED_ACCESS_DOMAIN_MUTATION = `mutation DeleteApprovedAccessDomain($input: DeleteApprovedAccessDomainInput!) {
  deleteApprovedAccessDomain(input: $input)
}`;

const VALIDATE_APPROVED_ACCESS_DOMAIN_MUTATION = `mutation ValidateApprovedAccessDomain($input: ValidateApprovedAccessDomainInput!) {
  validateApprovedAccessDomain(input: $input) {
    id
    domain
    isValidated
    createdAt
  }
}`;

export function registerApprovedAccessDomainsCommand(program: Command): void {
  const cmd = program
    .command("approved-access-domains")
    .description("Manage approved access domains")
    .argument("<operation>", "list, delete, validate")
    .argument("[id]", "Approved access domain ID")
    .option("--validation-token <token>", "Domain validation token");

  applyGlobalOptions(cmd);

  cmd.action(
    async (
      operation: string,
      id: string | undefined,
      options: ApprovedAccessDomainsOptions,
      command: Command,
    ) => {
      const globalOptions = resolveGlobalOptions(command);
      const services = createServices(globalOptions);
      const op = operation.toLowerCase();

      switch (op) {
        case "list": {
          const response = await services.api.post<
            GraphQLResponse<{ getApprovedAccessDomains?: unknown[] }>
          >("/graphql", {
            query: LIST_APPROVED_ACCESS_DOMAINS_QUERY,
          });

          await services.output.render(response.data?.data?.getApprovedAccessDomains ?? [], {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "delete": {
          if (!id) {
            throw new CliError("Missing approved access domain ID.", "INVALID_ARGUMENTS");
          }

          const response = await services.api.post<
            GraphQLResponse<{ deleteApprovedAccessDomain?: boolean }>
          >("/graphql", {
            query: DELETE_APPROVED_ACCESS_DOMAIN_MUTATION,
            variables: {
              input: { id },
            },
          });

          await services.output.render(
            {
              success: response.data?.data?.deleteApprovedAccessDomain ?? false,
              id,
            },
            {
              format: globalOptions.output,
              query: globalOptions.query,
            },
          );
          break;
        }
        case "validate": {
          if (!id || !options.validationToken) {
            throw new CliError(
              "Missing required validate inputs: <id> and --validation-token.",
              "INVALID_ARGUMENTS",
            );
          }

          const response = await services.api.post<
            GraphQLResponse<{ validateApprovedAccessDomain?: unknown }>
          >("/graphql", {
            query: VALIDATE_APPROVED_ACCESS_DOMAIN_MUTATION,
            variables: {
              input: {
                approvedAccessDomainId: id,
                validationToken: options.validationToken,
              },
            },
          });

          await services.output.render(response.data?.data?.validateApprovedAccessDomain, {
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
