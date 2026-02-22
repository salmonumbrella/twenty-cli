"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runObjectsList = runObjectsList;
async function runObjectsList(ctx) {
    const objects = await ctx.services.metadata.listObjects();
    await ctx.services.output.render(objects, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
