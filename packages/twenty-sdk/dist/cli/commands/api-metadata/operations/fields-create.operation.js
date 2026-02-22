"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runFieldsCreate = runFieldsCreate;
const body_1 = require("../../../utilities/shared/body");
async function runFieldsCreate(ctx) {
    const payload = await (0, body_1.parseBody)(ctx.options.data, ctx.options.file);
    const response = await ctx.services.metadata.createField(payload);
    await ctx.services.output.render(response, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
