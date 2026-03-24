import { Command } from "commander";
import { type GraphQLResponse } from "../../utilities/api/graphql-response";
import { CliError } from "../../utilities/errors/cli-error";
import { applyGlobalOptions, resolveGlobalOptions } from "../../utilities/shared/global-options";
import { createServices } from "../../utilities/shared/services";

interface PublicDomainsOptions {
  domain?: string;
}

const PUBLIC_DOMAIN_FIELDS = `
  id
  domain
  isValidated
  createdAt
`;

const DOMAIN_VALID_RECORD_FIELDS = `
  id
  domain
  records {
    validationType
    type
    status
    key
    value
  }
`;

const LIST_PUBLIC_DOMAINS_QUERY = `query FindManyPublicDomains {
  findManyPublicDomains {
    ${PUBLIC_DOMAIN_FIELDS}
  }
}`;

const CREATE_PUBLIC_DOMAIN_MUTATION = `mutation CreatePublicDomain($domain: String!) {
  createPublicDomain(domain: $domain) {
    ${PUBLIC_DOMAIN_FIELDS}
  }
}`;

const DELETE_PUBLIC_DOMAIN_MUTATION = `mutation DeletePublicDomain($domain: String!) {
  deletePublicDomain(domain: $domain)
}`;

const CHECK_PUBLIC_DOMAIN_RECORDS_MUTATION = `mutation CheckPublicDomainValidRecords($domain: String!) {
  checkPublicDomainValidRecords(domain: $domain) {
    ${DOMAIN_VALID_RECORD_FIELDS}
  }
}`;

export function registerPublicDomainsCommand(program: Command): void {
  const endpoint = "/graphql";
  const cmd = program
    .command("public-domains")
    .description("Manage public domains")
    .argument("<operation>", "list, create, delete, check-records")
    .option("--domain <domain>", "Public domain name");

  applyGlobalOptions(cmd);

  cmd.action(async (operation: string, options: PublicDomainsOptions, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const op = operation.toLowerCase();

    switch (op) {
      case "list": {
        const response = await services.api.post<
          GraphQLResponse<{ findManyPublicDomains?: unknown[] }>
        >(endpoint, {
          query: LIST_PUBLIC_DOMAINS_QUERY,
        });

        await services.output.render(response.data?.data?.findManyPublicDomains ?? [], {
          format: globalOptions.output,
          query: globalOptions.query,
        });
        break;
      }
      case "create": {
        const domain = requireDomain(options.domain);
        const response = await services.api.post<GraphQLResponse<{ createPublicDomain?: unknown }>>(
          endpoint,
          {
            query: CREATE_PUBLIC_DOMAIN_MUTATION,
            variables: { domain },
          },
        );

        await services.output.render(response.data?.data?.createPublicDomain, {
          format: globalOptions.output,
          query: globalOptions.query,
        });
        break;
      }
      case "delete": {
        const domain = requireDomain(options.domain);
        const response = await services.api.post<GraphQLResponse<{ deletePublicDomain?: boolean }>>(
          endpoint,
          {
            query: DELETE_PUBLIC_DOMAIN_MUTATION,
            variables: { domain },
          },
        );

        await services.output.render(
          {
            success: response.data?.data?.deletePublicDomain ?? false,
            domain,
          },
          {
            format: globalOptions.output,
            query: globalOptions.query,
          },
        );
        break;
      }
      case "check-records": {
        const domain = requireDomain(options.domain);
        const response = await services.api.post<
          GraphQLResponse<{ checkPublicDomainValidRecords?: unknown }>
        >(endpoint, {
          query: CHECK_PUBLIC_DOMAIN_RECORDS_MUTATION,
          variables: { domain },
        });

        await services.output.render(response.data?.data?.checkPublicDomainValidRecords, {
          format: globalOptions.output,
          query: globalOptions.query,
        });
        break;
      }
      default:
        throw new CliError(`Unknown operation: ${operation}`, "INVALID_ARGUMENTS");
    }
  });
}

function requireDomain(domain: string | undefined): string {
  if (!domain) {
    throw new CliError("Missing --domain option.", "INVALID_ARGUMENTS");
  }

  return domain;
}
