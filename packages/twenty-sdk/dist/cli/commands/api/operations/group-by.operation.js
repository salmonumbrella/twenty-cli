"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runGroupByOperation = runGroupByOperation;
const io_1 = require("../../../utilities/shared/io");
const parse_1 = require("../../../utilities/shared/parse");
async function runGroupByOperation(ctx) {
    let payload;
    if (ctx.options.data || ctx.options.file) {
        payload = await (0, io_1.readJsonInput)(ctx.options.data, ctx.options.file);
    }
    else if (ctx.options.field) {
        payload = { groupBy: ctx.options.field };
    }
    const params = (0, parse_1.parseKeyValuePairs)(ctx.options.param);
    if (ctx.options.filter) {
        params.filter = [ctx.options.filter];
    }
    const response = await ctx.services.records.groupBy(ctx.object, payload, Object.keys(params).length ? params : undefined);
    await ctx.services.output.render(response, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
