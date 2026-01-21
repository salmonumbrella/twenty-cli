"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runBatchDeleteOperation = runBatchDeleteOperation;
const io_1 = require("../../../utilities/shared/io");
const cli_error_1 = require("../../../utilities/errors/cli-error");
async function runBatchDeleteOperation(ctx) {
    if (!ctx.options.force) {
        // eslint-disable-next-line no-console
        console.log(`About to batch delete ${ctx.object}. Use --force to confirm.`);
        return;
    }
    let ids = [];
    if (ctx.options.ids) {
        ids = ctx.options.ids.split(',').map((id) => id.trim()).filter(Boolean);
    }
    else {
        const payload = await (0, io_1.readJsonInput)(ctx.options.data, ctx.options.file);
        if (!payload) {
            throw new cli_error_1.CliError('Missing JSON payload; use --data, --file, or --ids.', 'INVALID_ARGUMENTS');
        }
        if (!Array.isArray(payload)) {
            throw new cli_error_1.CliError('Batch payload must be a JSON array.', 'INVALID_ARGUMENTS');
        }
        ids = payload.map((value) => String(value));
    }
    if (ids.length === 0) {
        throw new cli_error_1.CliError('No valid IDs provided.', 'INVALID_ARGUMENTS');
    }
    const response = await ctx.services.records.batchDelete(ctx.object, ids);
    await ctx.services.output.render(response, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
