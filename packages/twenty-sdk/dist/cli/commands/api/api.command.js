"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerApiCommand = registerApiCommand;
const commander_1 = require("commander");
const global_options_1 = require("../../utilities/shared/global-options");
const services_1 = require("../../utilities/shared/services");
const cli_error_1 = require("../../utilities/errors/cli-error");
const list_operation_1 = require("./operations/list.operation");
const get_operation_1 = require("./operations/get.operation");
const create_operation_1 = require("./operations/create.operation");
const update_operation_1 = require("./operations/update.operation");
const delete_operation_1 = require("./operations/delete.operation");
const destroy_operation_1 = require("./operations/destroy.operation");
const restore_operation_1 = require("./operations/restore.operation");
const batch_create_operation_1 = require("./operations/batch-create.operation");
const batch_update_operation_1 = require("./operations/batch-update.operation");
const batch_delete_operation_1 = require("./operations/batch-delete.operation");
const import_operation_1 = require("./operations/import.operation");
const export_operation_1 = require("./operations/export.operation");
const group_by_operation_1 = require("./operations/group-by.operation");
const find_duplicates_operation_1 = require("./operations/find-duplicates.operation");
const merge_operation_1 = require("./operations/merge.operation");
const operationHandlers = {
    list: list_operation_1.runListOperation,
    get: get_operation_1.runGetOperation,
    create: create_operation_1.runCreateOperation,
    update: update_operation_1.runUpdateOperation,
    delete: delete_operation_1.runDeleteOperation,
    destroy: destroy_operation_1.runDestroyOperation,
    restore: restore_operation_1.runRestoreOperation,
    'batch-create': batch_create_operation_1.runBatchCreateOperation,
    'batch-update': batch_update_operation_1.runBatchUpdateOperation,
    'batch-delete': batch_delete_operation_1.runBatchDeleteOperation,
    import: import_operation_1.runImportOperation,
    export: export_operation_1.runExportOperation,
    'group-by': group_by_operation_1.runGroupByOperation,
    'find-duplicates': find_duplicates_operation_1.runFindDuplicatesOperation,
    merge: merge_operation_1.runMergeOperation,
};
function registerApiCommand(program) {
    const cmd = program
        .command('api')
        .description('Record operations')
        .argument('<object>', 'Object name (plural)')
        .argument('<operation>', 'Operation to perform')
        .argument('[arg]', 'Record ID or file path')
        .argument('[arg2]', 'Secondary argument')
        .option('--limit <number>', 'Limit number of records')
        .option('--all', 'Fetch all records')
        .option('--filter <expression>', 'Filter expression')
        .option('--include <relations>', 'Include related records')
        .option('--cursor <cursor>', 'Pagination cursor')
        .option('--sort <field>', 'Sort field')
        .option('--order <direction>', 'Sort order (asc or desc)')
        .option('--fields <fields>', 'Fields selection (comma-separated)')
        .option('--param <key=value>', 'Additional query params', collect)
        .option('-d, --data <json>', 'JSON payload')
        .option('-f, --file <path>', 'JSON/CSV file payload (use - for stdin)')
        .option('--set <key=value>', 'Set a field value', collect)
        .option('--force', 'Skip confirmation prompt')
        .option('--yes', 'Skip confirmation prompt (alias for --force)')
        .option('--ids <ids>', 'Comma-separated IDs')
        .option('--format <format>', 'Export format (json or csv)')
        .option('--output-file <path>', 'Output file path')
        .option('--batch-size <number>', 'Batch size (import)')
        .option('--dry-run', 'Preview without executing')
        .option('--continue-on-error', 'Continue on batch errors')
        .option('--field <field>', 'Group-by field')
        .option('--source <id>', 'Source record ID (merge)')
        .option('--target <id>', 'Target record ID (merge)')
        .option('--priority <index>', 'Conflict priority index (merge)');
    (0, global_options_1.applyGlobalOptions)(cmd);
    cmd.action(async (object, operation, arg, arg2, options, command) => {
        const op = operation.toLowerCase();
        const handler = operationHandlers[op];
        if (!handler) {
            throw new cli_error_1.CliError(`Unknown operation ${JSON.stringify(operation)}.`, 'INVALID_ARGUMENTS');
        }
        const resolvedCommand = command ?? (options instanceof commander_1.Command ? options : cmd);
        const globalOptions = (0, global_options_1.resolveGlobalOptions)(resolvedCommand);
        const services = (0, services_1.createServices)(globalOptions);
        const rawOptions = resolvedCommand.opts();
        const mergedOptions = {
            ...rawOptions,
            force: rawOptions.force || rawOptions.yes,
        };
        await handler({
            object,
            arg,
            arg2,
            options: mergedOptions,
            services,
            globalOptions,
        });
    });
}
function collect(value, previous = []) {
    return previous.concat([value]);
}
