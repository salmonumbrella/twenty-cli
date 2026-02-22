"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runUpdateOperation = runUpdateOperation;
const body_1 = require("../../../utilities/shared/body");
const cli_error_1 = require("../../../utilities/errors/cli-error");
async function runUpdateOperation(ctx) {
    const id = ctx.arg;
    if (!id) {
        throw new cli_error_1.CliError('Missing record ID.', 'INVALID_ARGUMENTS');
    }
    const payload = await (0, body_1.parseBody)(ctx.options.data, ctx.options.file, ctx.options.set);
    const record = await ctx.services.records.update(ctx.object, id, payload);
    await ctx.services.output.render(record, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
