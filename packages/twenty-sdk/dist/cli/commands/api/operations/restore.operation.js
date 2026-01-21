"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runRestoreOperation = runRestoreOperation;
const cli_error_1 = require("../../../utilities/errors/cli-error");
async function runRestoreOperation(ctx) {
    const id = ctx.arg;
    if (!id) {
        throw new cli_error_1.CliError('Missing record ID.', 'INVALID_ARGUMENTS');
    }
    const response = await ctx.services.records.restore(ctx.object, id);
    await ctx.services.output.render(response, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
