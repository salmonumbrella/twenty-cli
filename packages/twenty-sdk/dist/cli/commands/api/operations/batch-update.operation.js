"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.runBatchUpdateOperation = runBatchUpdateOperation;
const path_1 = __importDefault(require("path"));
const body_1 = require("../../../utilities/shared/body");
async function runBatchUpdateOperation(ctx) {
    let records = [];
    if (ctx.options.file) {
        const ext = path_1.default.extname(ctx.options.file).toLowerCase();
        if (ext === '.csv') {
            records = await ctx.services.importer.import(ctx.options.file);
        }
        else {
            const payload = await (0, body_1.parseArrayPayload)(ctx.options.data, ctx.options.file);
            records = payload;
        }
    }
    else {
        const payload = await (0, body_1.parseArrayPayload)(ctx.options.data, ctx.options.file);
        records = payload;
    }
    const response = await ctx.services.records.batchUpdate(ctx.object, records);
    await ctx.services.output.render(response, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
