import { ApiMetadataContext } from './types';

export async function runFieldsList(ctx: ApiMetadataContext): Promise<void> {
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

async function resolveObjectId(ctx: ApiMetadataContext): Promise<string | undefined> {
  const value = ctx.options.object;
  if (!value) return undefined;
  if (looksLikeUuid(value)) {
    return value;
  }
  const objects = await ctx.services.metadata.listObjects();
  const match = objects.find((obj) => obj.nameSingular === value || obj.namePlural === value);
  return match?.id;
}

function looksLikeUuid(value: string): boolean {
  return value.length === 36 && value[8] === '-' && value[13] === '-';
}
