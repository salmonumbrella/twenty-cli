"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.applyGlobalOptions = applyGlobalOptions;
exports.resolveGlobalOptions = resolveGlobalOptions;
const parse_1 = require("./parse");
function applyGlobalOptions(command, settings = {}) {
    const includeQuery = settings.includeQuery !== false;
    command.option('-o, --output <format>', 'Output format: text, json, csv');
    if (includeQuery) {
        command.option('--query <expression>', 'JMESPath query filter');
    }
    command.option('--workspace <name>', 'Workspace profile to use');
    command.option('--debug', 'Show request/response details');
    command.option('--no-retry', 'Disable automatic retry');
}
function resolveGlobalOptions(command, overrides) {
    const opts = getCommandOptions(command);
    const rawOutput = typeof opts.output === 'string' ? opts.output : (process.env.TWENTY_OUTPUT ?? 'text');
    const output = isValidOutputFormat(rawOutput) ? rawOutput : 'text';
    const query = overrides?.outputQuery
        ?? (typeof opts.query === 'string' ? opts.query : undefined)
        ?? process.env.TWENTY_QUERY
        ?? undefined;
    const workspace = typeof opts.workspace === 'string' ? opts.workspace : process.env.TWENTY_PROFILE;
    const debug = typeof opts.debug === 'boolean' ? opts.debug : (0, parse_1.parseBooleanEnv)(process.env.TWENTY_DEBUG) ?? false;
    const envNoRetry = (0, parse_1.parseBooleanEnv)(process.env.TWENTY_NO_RETRY) ?? false;
    const retry = typeof opts.retry === 'boolean' ? opts.retry : undefined;
    const noRetry = retry === false ? true : envNoRetry;
    return {
        output,
        query,
        workspace,
        debug,
        noRetry,
    };
}
function getCommandOptions(command) {
    const optsFn = command.optsWithGlobals;
    if (typeof optsFn === 'function') {
        return optsFn.call(command);
    }
    return command.opts();
}
function isValidOutputFormat(value) {
    return value === 'text' || value === 'json' || value === 'csv';
}
