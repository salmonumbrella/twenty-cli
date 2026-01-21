import { ApiOperationContext } from './types';
import { readJsonInput } from '../../../utilities/shared/io';
import { parseKeyValuePairs } from '../../../utilities/shared/parse';

export async function runGroupByOperation(ctx: ApiOperationContext): Promise<void> {
  let payload: unknown | undefined;

  if (ctx.options.data || ctx.options.file) {
    payload = await readJsonInput(ctx.options.data, ctx.options.file);
  } else if (ctx.options.field) {
    payload = { groupBy: ctx.options.field };
  }

  const params = parseKeyValuePairs(ctx.options.param);
  if (ctx.options.filter) {
    params.filter = [ctx.options.filter];
  }

  const response = await ctx.services.records.groupBy(ctx.object, payload, Object.keys(params).length ? params : undefined);
  await ctx.services.output.render(response, {
    format: ctx.globalOptions.output,
    query: ctx.globalOptions.query,
  });
}
