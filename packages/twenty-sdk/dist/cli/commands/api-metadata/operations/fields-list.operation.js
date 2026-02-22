"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.runFieldsList = runFieldsList;
async function runFieldsList(ctx) {
    const fields = await ctx.services.metadata.listFields();
    let filtered = fields;
    if (ctx.options.object) {
        const targetId = await resolveObjectId(ctx);
        if (targetId) {
            filtered = fields.filter((field) => field.objectMetadataId === targetId);
        }
    }
    await ctx.services.output.render(filtered, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
    });
}
async function resolveObjectId(ctx) {
    const value = ctx.options.object;
    if (!value)
        return undefined;
    if (looksLikeUuid(value)) {
        return value;
    }
    const objects = await ctx.services.metadata.listObjects();
    const match = objects.find((obj) => obj.nameSingular === value || obj.namePlural === value);
    return match?.id;
}
function looksLikeUuid(value) {
    return value.length === 36 && value[8] === '-' && value[13] === '-';
}
