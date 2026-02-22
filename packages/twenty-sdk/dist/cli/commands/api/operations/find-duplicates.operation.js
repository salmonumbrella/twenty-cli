"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runFindDuplicatesOperation = runFindDuplicatesOperation;
const io_1 = require("../../../utilities/shared/io");
const cli_error_1 = require("../../../utilities/errors/cli-error");
async function runFindDuplicatesOperation(ctx) {
    let payload;
    if (ctx.options.data || ctx.options.file) {
        payload = await (0, io_1.readJsonInput)(ctx.options.data, ctx.options.file);
    }
    else if (ctx.options.fields) {
        const fields = ctx.options.fields.split(',').map((field) => field.trim()).filter(Boolean);
        if (fields.length === 0) {
            throw new cli_error_1.CliError('No fields provided for duplicate detection.', 'INVALID_ARGUMENTS');
        }
        payload = { fields };
    }
    if (!payload) {
        throw new cli_error_1.CliError('Missing payload; use --fields, --data, or --file.', 'INVALID_ARGUMENTS');
    }
    const response = await ctx.services.records.findDuplicates(ctx.object, payload);
    await ctx.services.output.render(response, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
