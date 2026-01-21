import { ApiOperationContext } from './types';
import { readJsonInput } from '../../../utilities/shared/io';
import { CliError } from '../../../utilities/errors/cli-error';

export async function runFindDuplicatesOperation(ctx: ApiOperationContext): Promise<void> {
  let payload: unknown | undefined;

  if (ctx.options.data || ctx.options.file) {
    payload = await readJsonInput(ctx.options.data, ctx.options.file);
  } else if (ctx.options.fields) {
    const fields = ctx.options.fields.split(',').map((field) => field.trim()).filter(Boolean);
    if (fields.length === 0) {
      throw new CliError('No fields provided for duplicate detection.', 'INVALID_ARGUMENTS');
    }
    payload = { fields };
  }

  if (!payload) {
    throw new CliError('Missing payload; use --fields, --data, or --file.', 'INVALID_ARGUMENTS');
  }

  const response = await ctx.services.records.findDuplicates(ctx.object, payload);
  await ctx.services.output.render(response, {
    format: ctx.globalOptions.output,
    query: ctx.globalOptions.query,
  });
}
