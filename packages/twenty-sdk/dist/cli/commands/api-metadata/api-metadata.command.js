"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerApiMetadataCommand = registerApiMetadataCommand;
const commander_1 = require("commander");
const global_options_1 = require("../../utilities/shared/global-options");
const services_1 = require("../../utilities/shared/services");
const cli_error_1 = require("../../utilities/errors/cli-error");
const objects_list_operation_1 = require("./operations/objects-list.operation");
const objects_get_operation_1 = require("./operations/objects-get.operation");
const objects_create_operation_1 = require("./operations/objects-create.operation");
const fields_list_operation_1 = require("./operations/fields-list.operation");
const fields_get_operation_1 = require("./operations/fields-get.operation");
const fields_create_operation_1 = require("./operations/fields-create.operation");
const handlers = {
    'objects:list': objects_list_operation_1.runObjectsList,
    'objects:get': objects_get_operation_1.runObjectsGet,
    'objects:create': objects_create_operation_1.runObjectsCreate,
    'fields:list': fields_list_operation_1.runFieldsList,
    'fields:get': fields_get_operation_1.runFieldsGet,
    'fields:create': fields_create_operation_1.runFieldsCreate,
};
function registerApiMetadataCommand(program) {
    const cmd = program
        .command('api-metadata')
        .description('Schema operations')
        .argument('<type>', 'Metadata type (objects or fields)')
        .argument('<operation>', 'Operation to perform')
        .argument('[arg]', 'Identifier')
        .option('-d, --data <json>', 'JSON payload')
        .option('-f, --file <path>', 'JSON payload file (use - for stdin)')
        .option('--object <name>', 'Filter fields by object');
    (0, global_options_1.applyGlobalOptions)(cmd);
    cmd.action(async (type, operation, arg, options, command) => {
        const key = `${type.toLowerCase()}:${operation.toLowerCase()}`;
        const handler = handlers[key];
        if (!handler) {
            throw new cli_error_1.CliError(`Unknown api-metadata operation ${JSON.stringify(type)} ${JSON.stringify(operation)}.`, 'INVALID_ARGUMENTS');
        }
        const resolvedCommand = command ?? (options instanceof commander_1.Command ? options : cmd);
        const globalOptions = (0, global_options_1.resolveGlobalOptions)(resolvedCommand);
        const services = (0, services_1.createServices)(globalOptions);
        const rawOptions = resolvedCommand.opts();
        await handler({
            type,
            operation,
            arg,
            options: rawOptions ?? {},
            services,
            globalOptions,
        });
    });
}
