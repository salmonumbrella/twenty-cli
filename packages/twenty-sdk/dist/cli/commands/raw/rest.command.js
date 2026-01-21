"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerRestCommand = registerRestCommand;
const commander_1 = require("commander");
const global_options_1 = require("../../utilities/shared/global-options");
const services_1 = require("../../utilities/shared/services");
const io_1 = require("../../utilities/shared/io");
const parse_1 = require("../../utilities/shared/parse");
function registerRestCommand(program) {
    const cmd = program
        .command('rest')
        .description('Raw REST API access')
        .argument('<method>', 'HTTP method')
        .argument('<path>', 'REST path')
        .option('-d, --data <json>', 'JSON payload')
        .option('-f, --file <path>', 'JSON file payload (use - for stdin)')
        .option('--param <key=value>', 'Query param', collect);
    (0, global_options_1.applyGlobalOptions)(cmd);
    cmd.action(async (method, path, options, command) => {
        const resolvedCommand = command ?? (options instanceof commander_1.Command ? options : cmd);
        const globalOptions = (0, global_options_1.resolveGlobalOptions)(resolvedCommand);
        const services = (0, services_1.createServices)(globalOptions);
        const rawOptions = resolvedCommand.opts();
        const payload = await (0, io_1.readJsonInput)(rawOptions.data, rawOptions.file);
        const params = (0, parse_1.parseKeyValuePairs)(rawOptions.param);
        const url = path.startsWith('/') ? path : `/${path}`;
        const response = await services.api.request({
            method: method.toLowerCase(),
            url,
            params: Object.keys(params).length ? params : undefined,
            data: payload,
        });
        await services.output.render(response.data, {
            format: globalOptions.output,
            query: globalOptions.query,
        });
    });
}
function collect(value, previous = []) {
    return previous.concat([value]);
}
