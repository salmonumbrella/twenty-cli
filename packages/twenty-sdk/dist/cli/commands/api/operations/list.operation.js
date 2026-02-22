"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runListOperation = runListOperation;
const parse_1 = require("../../../utilities/shared/parse");
async function runListOperation(ctx) {
    const { services, globalOptions } = ctx;
    const limit = ctx.options.limit ? Number(ctx.options.limit) : undefined;
    const params = (0, parse_1.parseKeyValuePairs)(ctx.options.param);
    const listOptions = {
        limit,
        cursor: ctx.options.cursor,
        filter: ctx.options.filter,
        include: ctx.options.include,
        sort: ctx.options.sort,
        order: ctx.options.order,
        fields: ctx.options.fields,
        params,
    };
    const result = ctx.options.all
        ? await services.records.listAll(ctx.object, listOptions)
        : await services.records.list(ctx.object, listOptions);
    await services.output.render(result.data, {
        format: globalOptions.output,
        query: globalOptions.query,
    });
}
