"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runExportOperation = runExportOperation;
const parse_1 = require("../../../utilities/shared/parse");
const cli_error_1 = require("../../../utilities/errors/cli-error");
const OUTPUT_FORMATS = new Set(['json', 'csv', 'text']);
async function runExportOperation(ctx) {
    const format = (ctx.options.format ?? 'json').toLowerCase();
    if (format !== 'json' && format !== 'csv') {
        throw new cli_error_1.CliError(`Unsupported export format ${JSON.stringify(format)}.`, 'INVALID_ARGUMENTS');
    }
    const params = (0, parse_1.parseKeyValuePairs)(ctx.options.param);
    const limit = ctx.options.limit ? Number(ctx.options.limit) : 200;
    const listOptions = {
        limit: Number.isNaN(limit) ? 200 : limit,
        cursor: ctx.options.cursor,
        filter: ctx.options.filter,
        include: ctx.options.include,
        sort: ctx.options.sort,
        order: ctx.options.order,
        fields: ctx.options.fields,
        params,
    };
    const shouldAll = ctx.options.all === true;
    const response = shouldAll
        ? await ctx.services.records.listAll(ctx.object, listOptions)
        : await ctx.services.records.list(ctx.object, listOptions);
    let outputFile = ctx.options.outputFile;
    if (!outputFile && ctx.options.output && !OUTPUT_FORMATS.has(ctx.options.output)) {
        outputFile = ctx.options.output;
    }
    await ctx.services.exporter.export(response.data, {
        format: format,
        output: outputFile,
    });
}
