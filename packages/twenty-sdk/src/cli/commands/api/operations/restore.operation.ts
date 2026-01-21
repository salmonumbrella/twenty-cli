import { ApiOperationContext } from './types';
import { CliError } from '../../../utilities/errors/cli-error';

export async function runRestoreOperation(ctx: ApiOperationContext): Promise<void> {
  const id = ctx.arg;
  if (!id) {
    throw new CliError('Missing record ID.', 'INVALID_ARGUMENTS');
  }
  const response = await ctx.services.records.restore(ctx.object, id);
  await ctx.services.output.render(response, {
    format: ctx.globalOptions.output,
    query: ctx.globalOptions.query,
  });
}
