"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runFieldsGet = runFieldsGet;
const cli_error_1 = require("../../../utilities/errors/cli-error");
async function runFieldsGet(ctx) {
    const id = ctx.arg;
    if (!id) {
        throw new cli_error_1.CliError('Missing field ID.', 'INVALID_ARGUMENTS');
    }
    const field = await ctx.services.metadata.getField(id);
    await ctx.services.output.render(field, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
