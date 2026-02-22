"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runCreateOperation = runCreateOperation;
const body_1 = require("../../../utilities/shared/body");
async function runCreateOperation(ctx) {
    const payload = await (0, body_1.parseBody)(ctx.options.data, ctx.options.file, ctx.options.set);
    const record = await ctx.services.records.create(ctx.object, payload);
    await ctx.services.output.render(record, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
