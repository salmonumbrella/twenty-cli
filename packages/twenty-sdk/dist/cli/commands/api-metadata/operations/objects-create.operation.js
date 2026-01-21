"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runObjectsCreate = runObjectsCreate;
const body_1 = require("../../../utilities/shared/body");
async function runObjectsCreate(ctx) {
    const payload = await (0, body_1.parseBody)(ctx.options.data, ctx.options.file);
    const response = await ctx.services.metadata.createObject(payload);
    await ctx.services.output.render(response, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
