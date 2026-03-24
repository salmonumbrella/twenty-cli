import { ApiOperationContext } from "./types";
import { CliError } from "../../../utilities/errors/cli-error";
import { resolveBulkFilter } from "./bulk-filter";

export async function runDestroyOperation(ctx: ApiOperationContext): Promise<void> {
  const id = ctx.arg;
  if (!ctx.options.force) {
    if (!id && !ctx.options.filter && !ctx.options.ids) {
      throw new CliError("Missing record ID.", "INVALID_ARGUMENTS");
    }

    const target = id ? `${ctx.object} ${id}` : `${ctx.object} matching the provided filter`;
    // eslint-disable-next-line no-console
    console.log(`About to destroy ${target}. Use --force to confirm.`);
    return;
  }

  if (id) {
    const response = await ctx.services.records.destroy(ctx.object, id);
    if (response == null || (typeof response === "string" && response === "")) {
      // eslint-disable-next-line no-console
      console.log(`Destroyed ${ctx.object} ${id}`);
      return;
    }
    await ctx.services.output.render(response, {
      format: ctx.globalOptions.output,
      query: ctx.globalOptions.query,
    });
    return;
  }

  const filter = resolveBulkFilter(ctx.options);
  const response = await ctx.services.records.destroyMany(ctx.object, { filter });
  await ctx.services.output.render(response, {
    format: ctx.globalOptions.output,
    query: ctx.globalOptions.query,
  });
}
