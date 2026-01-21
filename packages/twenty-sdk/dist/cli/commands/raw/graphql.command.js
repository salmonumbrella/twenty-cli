"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerGraphqlCommand = registerGraphqlCommand;
const commander_1 = require("commander");
const global_options_1 = require("../../utilities/shared/global-options");
const services_1 = require("../../utilities/shared/services");
const fs_extra_1 = __importDefault(require("fs-extra"));
const io_1 = require("../../utilities/shared/io");
const cli_error_1 = require("../../utilities/errors/cli-error");
function registerGraphqlCommand(program) {
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
    (0, global_options_1.applyGlobalOptions)(cmd, { includeQuery: false });
    cmd.action(async (operation, options, command) => {
        const resolvedCommand = command ?? (options instanceof commander_1.Command ? options : cmd);
        const rawOptions = resolvedCommand.opts();
        const outputQuery = rawOptions.outputQuery;
        const globalOptions = (0, global_options_1.resolveGlobalOptions)(resolvedCommand, { outputQuery });
        const services = (0, services_1.createServices)(globalOptions);
        const op = operation.toLowerCase();
        if (op === 'schema') {
            const payload = { query: introspectionQuery };
            const response = await services.api.post(normalizeEndpoint(rawOptions.endpoint), payload);
            await outputGraphqlResult(response.data, globalOptions, services, rawOptions.outputFile);
            return;
        }
        if (op !== 'query' && op !== 'mutate') {
            throw new cli_error_1.CliError(`Unknown GraphQL operation ${JSON.stringify(operation)}.`, 'INVALID_ARGUMENTS');
        }
        const query = await readGraphqlQuery(rawOptions.query, rawOptions.file);
        const variables = await readVariables(rawOptions.variables, rawOptions.variablesFile);
        const payload = { query };
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
async function readGraphqlQuery(query, filePath) {
    if (filePath) {
        const content = await (0, io_1.readFileOrStdin)(filePath);
        return content.trim();
    }
    if (!query) {
        throw new cli_error_1.CliError('Missing GraphQL query; use --query or --file.', 'INVALID_ARGUMENTS');
    }
    return query;
}
async function readVariables(raw, filePath) {
    const payload = await (0, io_1.readJsonInput)(raw, filePath);
    if (!payload)
        return {};
    if (typeof payload !== 'object' || Array.isArray(payload)) {
        throw new cli_error_1.CliError('GraphQL variables must be a JSON object.', 'INVALID_ARGUMENTS');
    }
    return payload;
}
function normalizeEndpoint(endpoint) {
    if (endpoint.startsWith('/'))
        return endpoint;
    return `/${endpoint}`;
}
async function outputGraphqlResult(data, globalOptions, services, outputFile) {
    if (outputFile) {
        const content = typeof data === 'string' ? data : JSON.stringify(data, null, 2);
        await fs_extra_1.default.writeFile(outputFile, content);
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
