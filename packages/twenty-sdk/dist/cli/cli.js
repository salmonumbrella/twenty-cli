#!/usr/bin/env node
"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const commander_1 = require("commander");
const api_command_1 = require("./commands/api/api.command");
const api_metadata_command_1 = require("./commands/api-metadata/api-metadata.command");
const rest_command_1 = require("./commands/raw/rest.command");
const graphql_command_1 = require("./commands/raw/graphql.command");
const error_handler_1 = require("./utilities/errors/error-handler");
async function main() {
    const program = new commander_1.Command();
    program.name('twenty');
    program.description('Twenty CLI (TypeScript port)');
    program.exitOverride();
    (0, api_command_1.registerApiCommand)(program);
    (0, api_metadata_command_1.registerApiMetadataCommand)(program);
    (0, rest_command_1.registerRestCommand)(program);
    (0, graphql_command_1.registerGraphqlCommand)(program);
    try {
        await program.parseAsync(process.argv);
    }
    catch (error) {
        const messages = (0, error_handler_1.formatError)(error);
        for (const line of messages) {
            // eslint-disable-next-line no-console
            console.error(line);
        }
        process.exitCode = (0, error_handler_1.toExitCode)(error);
    }
}
main();
