import { Command } from "commander";
import { type GraphQLResponse } from "../../utilities/api/graphql-response";
import { CliError } from "../../utilities/errors/cli-error";
import { applyGlobalOptions } from "../../utilities/shared/global-options";
import { requireYes } from "../../utilities/shared/confirmation";
import { createCommandContext } from "../../utilities/shared/context";

interface ApprovedAccessDomainsOptions {
  validationToken?: string;
  yes?: boolean;
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
  const endpoint = "/graphql";
  const cmd = program.command("approved-access-domains").description("Manage approved access domains");
  applyGlobalOptions(cmd);

  const listCmd = cmd.command("list").description("List approved access domains");
  applyGlobalOptions(listCmd);
  listCmd.action(async (_options: unknown, command: Command) => {
    const { globalOptions, services } = createCommandContext(command);
    const response = await services.api.post<
      GraphQLResponse<{ getApprovedAccessDomains?: unknown[] }>
    >(endpoint, {
      query: LIST_APPROVED_ACCESS_DOMAINS_QUERY,
    });

    await services.output.render(response.data?.data?.getApprovedAccessDomains ?? [], {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });

  const deleteCmd = cmd
    .command("delete")
    .description("Delete an approved access domain")
    .argument("[id]", "Approved access domain ID")
    .option("--yes", "Confirm destructive operations");
  applyGlobalOptions(deleteCmd);
  deleteCmd.action(async (id: string | undefined, options: ApprovedAccessDomainsOptions, command: Command) => {
    const { globalOptions, services } = createCommandContext(command);
    if (!id) {
      throw new CliError("Missing approved access domain ID.", "INVALID_ARGUMENTS");
    }
    requireYes(options, "Delete");

    const response = await services.api.post<GraphQLResponse<{ deleteApprovedAccessDomain?: boolean }>>(
      endpoint,
      {
        query: DELETE_APPROVED_ACCESS_DOMAIN_MUTATION,
        variables: {
          input: { id },
        },
      },
    );

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
  });

  const validateCmd = cmd
    .command("validate")
    .description("Validate an approved access domain")
    .argument("[id]", "Approved access domain ID")
    .option("--validation-token <token>", "Domain validation token");
  applyGlobalOptions(validateCmd);
  validateCmd.action(
    async (id: string | undefined, options: ApprovedAccessDomainsOptions, command: Command) => {
      const { globalOptions, services } = createCommandContext(command);
      if (!id || !options.validationToken) {
        throw new CliError(
          "Missing required validate inputs: <id> and --validation-token.",
          "INVALID_ARGUMENTS",
        );
      }

      const response = await services.api.post<
        GraphQLResponse<{ validateApprovedAccessDomain?: unknown }>
      >(endpoint, {
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
    },
  );
}
