import { Command } from "commander";
import { type GraphQLResponse } from "../../utilities/api/graphql-response";
import { CliError } from "../../utilities/errors/cli-error";
import { applyGlobalOptions, resolveGlobalOptions } from "../../utilities/shared/global-options";
import { createServices } from "../../utilities/shared/services";

interface EmailingDomainsOptions {
  domain?: string;
  driver?: string;
}

const EMAILING_DOMAIN_FIELDS = `
  id
  createdAt
  updatedAt
  domain
  driver
  status
  verificationRecords {
    type
    key
    value
    priority
  }
  verifiedAt
`;

const LIST_EMAILING_DOMAINS_QUERY = `query GetEmailingDomains {
  getEmailingDomains {
    ${EMAILING_DOMAIN_FIELDS}
  }
}`;

const CREATE_EMAILING_DOMAIN_MUTATION = `mutation CreateEmailingDomain($domain: String!, $driver: EmailingDomainDriver!) {
  createEmailingDomain(domain: $domain, driver: $driver) {
    ${EMAILING_DOMAIN_FIELDS}
  }
}`;

const VERIFY_EMAILING_DOMAIN_MUTATION = `mutation VerifyEmailingDomain($id: String!) {
  verifyEmailingDomain(id: $id) {
    ${EMAILING_DOMAIN_FIELDS}
  }
}`;

const DELETE_EMAILING_DOMAIN_MUTATION = `mutation DeleteEmailingDomain($id: String!) {
  deleteEmailingDomain(id: $id)
}`;

export function registerEmailingDomainsCommand(program: Command): void {
  const endpoint = "/graphql";
  const cmd = program
    .command("emailing-domains")
    .description("Manage emailing domains")
    .argument("<operation>", "list, create, verify, delete")
    .argument("[id]", "Emailing domain ID")
    .option("--domain <domain>", "Emailing domain name")
    .option("--driver <driver>", "Emailing domain driver", "AWS_SES");

  applyGlobalOptions(cmd);

  cmd.action(
    async (
      operation: string,
      id: string | undefined,
      options: EmailingDomainsOptions,
      command: Command,
    ) => {
      const globalOptions = resolveGlobalOptions(command);
      const services = createServices(globalOptions);
      const op = operation.toLowerCase();

      switch (op) {
        case "list": {
          const response = await services.api.post<
            GraphQLResponse<{ getEmailingDomains?: unknown[] }>
          >(endpoint, {
            query: LIST_EMAILING_DOMAINS_QUERY,
          });

          await services.output.render(response.data?.data?.getEmailingDomains ?? [], {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "create": {
          const domain = requireDomain(options.domain);
          const driver = normalizeDriver(options.driver);
          const response = await services.api.post<
            GraphQLResponse<{ createEmailingDomain?: unknown }>
          >(endpoint, {
            query: CREATE_EMAILING_DOMAIN_MUTATION,
            variables: { domain, driver },
          });

          await services.output.render(response.data?.data?.createEmailingDomain, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "verify": {
          if (!id) {
            throw new CliError("Missing emailing domain ID.", "INVALID_ARGUMENTS");
          }

          const response = await services.api.post<
            GraphQLResponse<{ verifyEmailingDomain?: unknown }>
          >(endpoint, {
            query: VERIFY_EMAILING_DOMAIN_MUTATION,
            variables: { id },
          });

          await services.output.render(response.data?.data?.verifyEmailingDomain, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "delete": {
          if (!id) {
            throw new CliError("Missing emailing domain ID.", "INVALID_ARGUMENTS");
          }

          const response = await services.api.post<
            GraphQLResponse<{ deleteEmailingDomain?: boolean }>
          >(endpoint, {
            query: DELETE_EMAILING_DOMAIN_MUTATION,
            variables: { id },
          });

          await services.output.render(
            {
              success: response.data?.data?.deleteEmailingDomain ?? false,
              id,
            },
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
    },
  );
}

function requireDomain(domain: string | undefined): string {
  if (!domain) {
    throw new CliError("Missing --domain option.", "INVALID_ARGUMENTS");
  }

  return domain;
}

function normalizeDriver(driver: string | undefined): string {
  const normalized = driver?.trim().toUpperCase() ?? "AWS_SES";

  if (normalized !== "AWS_SES") {
    throw new CliError(`Unsupported emailing domain driver "${driver}".`, "INVALID_ARGUMENTS");
  }

  return normalized;
}
