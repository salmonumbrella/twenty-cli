"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runImportOperation = runImportOperation;
const parse_1 = require("../../../utilities/shared/parse");
const cli_error_1 = require("../../../utilities/errors/cli-error");
async function runImportOperation(ctx) {
    const filePath = ctx.arg;
    if (!filePath) {
        throw new cli_error_1.CliError('Missing import file path.', 'INVALID_ARGUMENTS');
    }
    const batchSizeRaw = ctx.options.batchSize ? Number(ctx.options.batchSize) : 60;
    let batchSize = Number.isNaN(batchSizeRaw) || batchSizeRaw <= 0 ? 60 : batchSizeRaw;
    if (batchSize > 60)
        batchSize = 60;
    const records = await ctx.services.importer.import(filePath, { dryRun: ctx.options.dryRun });
    if (ctx.options.dryRun) {
        return;
    }
    if (records.length === 0) {
        // eslint-disable-next-line no-console
        console.log('No records to import.');
        return;
    }
    const batches = (0, parse_1.chunkArray)(records, batchSize);
    let imported = 0;
    let errors = 0;
    for (const batch of batches) {
        try {
            await ctx.services.records.batchCreate(ctx.object, batch);
            imported += batch.length;
        }
        catch (error) {
            errors += batch.length;
            if (!ctx.options.continueOnError) {
                throw error;
            }
        }
    }
    // eslint-disable-next-line no-console
    console.log(`Import complete: ${imported} imported${errors ? `, ${errors} failed` : ''}.`);
}
