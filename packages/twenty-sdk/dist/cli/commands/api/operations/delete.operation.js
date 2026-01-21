"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runDeleteOperation = runDeleteOperation;
const cli_error_1 = require("../../../utilities/errors/cli-error");
async function runDeleteOperation(ctx) {
    const id = ctx.arg;
    if (!id) {
        throw new cli_error_1.CliError('Missing record ID.', 'INVALID_ARGUMENTS');
    }
    if (!ctx.options.force) {
        // eslint-disable-next-line no-console
        console.log(`About to delete ${ctx.object} ${id}. Use --force to confirm.`);
        return;
    }
    const response = await ctx.services.records.delete(ctx.object, id);
    if (response == null || (typeof response === 'string' && response === '')) {
        // eslint-disable-next-line no-console
        console.log(`Deleted ${ctx.object} ${id}`);
        return;
    }
    await ctx.services.output.render(response, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
