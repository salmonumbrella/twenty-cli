"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runGetOperation = runGetOperation;
const cli_error_1 = require("../../../utilities/errors/cli-error");
async function runGetOperation(ctx) {
    const id = ctx.arg;
    if (!id) {
        throw new cli_error_1.CliError('Missing record ID.', 'INVALID_ARGUMENTS');
    }
    const record = await ctx.services.records.get(ctx.object, id, { include: ctx.options.include });
    await ctx.services.output.render(record, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
