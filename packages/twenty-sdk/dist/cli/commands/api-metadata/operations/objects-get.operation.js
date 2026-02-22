"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runObjectsGet = runObjectsGet;
const cli_error_1 = require("../../../utilities/errors/cli-error");
async function runObjectsGet(ctx) {
    const id = ctx.arg;
    if (!id) {
        throw new cli_error_1.CliError('Missing object identifier.', 'INVALID_ARGUMENTS');
    }
    const object = await ctx.services.metadata.getObject(id);
    await ctx.services.output.render(object, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
