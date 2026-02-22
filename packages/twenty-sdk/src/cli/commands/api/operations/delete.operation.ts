import { ApiOperationContext } from './types';
import { CliError } from '../../../utilities/errors/cli-error';

export async function runDeleteOperation(ctx: ApiOperationContext): Promise<void> {
  const id = ctx.arg;
  if (!id) {
    throw new CliError('Missing record ID.', 'INVALID_ARGUMENTS');
  }
  if (!ctx.options.force) {
    // eslint-disable-next-line no-console
    console.log(`About to delete ${ctx.object} ${id}. Use --force to confirm.`);
    return;
  }
  const response = await ctx.services.records.delete(ctx.object, id);
  if (response == null || (typeof response === 'string' && response === '')) {
    // eslint-disable-next-line no-console
    console.log(`Deleted ${ctx.object} ${id}`);
    return;
  }
  await ctx.services.output.render(response, {
    format: ctx.globalOptions.output,
    query: ctx.globalOptions.query,
  });
}
