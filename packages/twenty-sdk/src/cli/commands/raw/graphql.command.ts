import { Command } from 'commander';
import { applyGlobalOptions, resolveGlobalOptions } from '../../utilities/shared/global-options';
import { createServices } from '../../utilities/shared/services';
import fs from 'fs-extra';
import { readFileOrStdin, readJsonInput } from '../../utilities/shared/io';
import { CliError } from '../../utilities/errors/cli-error';

export function registerGraphqlCommand(program: Command): void {
  const cmd = program
    .command('graphql')
    .description('Raw GraphQL API access')
    .argument('<operation>', 'query, mutate, or schema')
    .option('-q, --query <query>', 'GraphQL query string')
    .option('-f, --file <path>', 'GraphQL query file (use - for stdin)')
    .option('--variables <json>', 'JSON variables')
    .option('--variables-file <path>', 'JSON variables file (use - for stdin)')
    .option('--operation-name <name>', 'GraphQL operation name')
    .option('--endpoint <path>', 'GraphQL endpoint path', 'graphql')
    .option('--output-query <expression>', 'JMESPath query filter')
    .option('--output-file <path>', 'Output file (schema command)');

  applyGlobalOptions(cmd, { includeQuery: false });

  cmd.action(async (operation: string, options: GraphqlOptions | Command, command?: Command) => {
    const resolvedCommand = command ?? (options instanceof Command ? options : cmd);
    const rawOptions = resolvedCommand.opts() as GraphqlOptions;
    const outputQuery = rawOptions.outputQuery;
    const globalOptions = resolveGlobalOptions(resolvedCommand, { outputQuery });
    const services = createServices(globalOptions);

    const op = operation.toLowerCase();
    if (op === 'schema') {
      const payload = { query: introspectionQuery };
      const response = await services.api.post(normalizeEndpoint(rawOptions.endpoint), payload);
      await outputGraphqlResult(response.data, globalOptions, services, rawOptions.outputFile);
      return;
    }

    if (op !== 'query' && op !== 'mutate') {
      throw new CliError(`Unknown GraphQL operation ${JSON.stringify(operation)}.`, 'INVALID_ARGUMENTS');
    }

    const query = await readGraphqlQuery(rawOptions.query, rawOptions.file);
    const variables = await readVariables(rawOptions.variables, rawOptions.variablesFile);

    const payload: Record<string, unknown> = { query };
    if (Object.keys(variables).length > 0) {
      payload.variables = variables;
    }
    if (rawOptions.operationName) {
      payload.operationName = rawOptions.operationName;
    }

    const response = await services.api.post(normalizeEndpoint(rawOptions.endpoint), payload);
    await outputGraphqlResult(response.data, globalOptions, services, undefined);
  });
}

interface GraphqlOptions {
  query?: string;
  file?: string;
  variables?: string;
  variablesFile?: string;
  operationName?: string;
  endpoint: string;
  outputQuery?: string;
  outputFile?: string;
}

async function readGraphqlQuery(query?: string, filePath?: string): Promise<string> {
  if (filePath) {
    const content = await readFileOrStdin(filePath);
    return content.trim();
  }
  if (!query) {
    throw new CliError('Missing GraphQL query; use --query or --file.', 'INVALID_ARGUMENTS');
  }
  return query;
}

async function readVariables(raw?: string, filePath?: string): Promise<Record<string, unknown>> {
  const payload = await readJsonInput(raw, filePath);
  if (!payload) return {};
  if (typeof payload !== 'object' || Array.isArray(payload)) {
    throw new CliError('GraphQL variables must be a JSON object.', 'INVALID_ARGUMENTS');
  }
  return payload as Record<string, unknown>;
}

function normalizeEndpoint(endpoint: string): string {
  if (endpoint.startsWith('/')) return endpoint;
  return `/${endpoint}`;
}

async function outputGraphqlResult(
  data: unknown,
  globalOptions: { output?: string; query?: string },
  services: ReturnType<typeof createServices>,
  outputFile?: string,
): Promise<void> {
  if (outputFile) {
    const content = typeof data === 'string' ? data : JSON.stringify(data, null, 2);
    await fs.writeFile(outputFile, content);
    // eslint-disable-next-line no-console
    console.error(`Wrote schema output to ${outputFile}`);
    return;
  }

  await services.output.render(data, {
    format: globalOptions.output,
    query: globalOptions.query,
  });
}

const introspectionQuery = `
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types {
      ...FullType
    }
    directives {
      name
      description
      locations
      args {
        ...InputValue
      }
    }
  }
}

fragment FullType on __Type {
  kind
  name
  description
  fields(includeDeprecated: true) {
    name
    description
    args {
      ...InputValue
    }
    type {
      ...TypeRef
    }
    isDeprecated
    deprecationReason
  }
  inputFields {
    ...InputValue
  }
  interfaces {
    ...TypeRef
  }
  enumValues(includeDeprecated: true) {
    name
    description
    isDeprecated
    deprecationReason
  }
  possibleTypes {
    ...TypeRef
  }
}

fragment InputValue on __InputValue {
  name
  description
  type { ...TypeRef }
  defaultValue
}

fragment TypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
        }
      }
    }
  }
}
`;
