import { ApiMetadataContext } from './types';
import { CliError } from '../../../utilities/errors/cli-error';

export async function runFieldsGet(ctx: ApiMetadataContext): Promise<void> {
  const id = ctx.arg;
  if (!id) {
    throw new CliError('Missing field ID.', 'INVALID_ARGUMENTS');
  }
  const field = await ctx.services.metadata.getField(id);
  await ctx.services.output.render(field, {
    format: ctx.globalOptions.output,
    query: ctx.globalOptions.query,
  });
}
