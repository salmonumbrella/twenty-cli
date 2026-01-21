import { ApiOperationContext } from './types';
import { parseKeyValuePairs } from '../../../utilities/shared/parse';

export async function runListOperation(ctx: ApiOperationContext): Promise<void> {
  const { services, globalOptions } = ctx;
  const limit = ctx.options.limit ? Number(ctx.options.limit) : undefined;
  const params = parseKeyValuePairs(ctx.options.param);

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
