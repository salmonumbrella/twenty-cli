"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runMergeOperation = runMergeOperation;
const io_1 = require("../../../utilities/shared/io");
const cli_error_1 = require("../../../utilities/errors/cli-error");
async function runMergeOperation(ctx) {
    let payload;
    const parsedPriority = ctx.options.priority ? Number(ctx.options.priority) : 0;
    const priority = Number.isNaN(parsedPriority) ? 0 : parsedPriority;
    if (ctx.options.source || ctx.options.target) {
        if (!ctx.options.source || !ctx.options.target) {
            throw new cli_error_1.CliError('Both --source and --target are required for merge.', 'INVALID_ARGUMENTS');
        }
        payload = {
            ids: [ctx.options.source, ctx.options.target],
            conflictPriorityIndex: priority,
        };
    }
    else if (ctx.options.ids) {
        const ids = ctx.options.ids.split(',').map((id) => id.trim()).filter(Boolean);
        if (ids.length === 0) {
            throw new cli_error_1.CliError('No valid IDs provided for merge.', 'INVALID_ARGUMENTS');
        }
        payload = {
            ids,
            conflictPriorityIndex: priority,
        };
    }
    else if (ctx.options.data || ctx.options.file) {
        const raw = await (0, io_1.readJsonInput)(ctx.options.data, ctx.options.file);
        if (!raw || typeof raw !== 'object') {
            throw new cli_error_1.CliError('Invalid merge payload.', 'INVALID_ARGUMENTS');
        }
        payload = raw;
    }
    if (!payload) {
        throw new cli_error_1.CliError('Missing payload; use --ids, --source/--target, --data, or --file.', 'INVALID_ARGUMENTS');
    }
    if (ctx.options.dryRun) {
        payload.dryRun = true;
    }
    const response = await ctx.services.records.merge(ctx.object, payload);
    await ctx.services.output.render(response, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
