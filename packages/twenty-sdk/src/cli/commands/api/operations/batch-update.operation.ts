import path from "path";
import { ApiOperationContext } from "./types";
import { parseArrayPayload, parseBody } from "../../../utilities/shared/body";
import { readJsonInput } from "../../../utilities/shared/io";
import { resolveBulkFilter } from "./bulk-filter";

export async function runBatchUpdateOperation(ctx: ApiOperationContext): Promise<void> {
  if (ctx.options.file) {
    const ext = path.extname(ctx.options.file).toLowerCase();
    if (ext === ".csv") {
      const records = await ctx.services.importer.import(ctx.options.file);
      const response = await ctx.services.records.batchUpdate(ctx.object, records);
      await ctx.services.output.render(response, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
      });
      return;
    }

    const rawPayload = await readJsonInput(ctx.options.data, ctx.options.file);
    if (Array.isArray(rawPayload)) {
      const records = (await parseArrayPayload(ctx.options.data, ctx.options.file)) as Record<
        string,
        unknown
      >[];
      const response = await ctx.services.records.batchUpdate(ctx.object, records);
      await ctx.services.output.render(response, {
        format: ctx.globalOptions.output,
        query: ctx.globalOptions.query,
      });
      return;
    }
  } else if (Array.isArray(await readJsonInput(ctx.options.data, ctx.options.file))) {
    const payload = await parseArrayPayload(ctx.options.data, ctx.options.file);
    const response = await ctx.services.records.batchUpdate(
      ctx.object,
      payload as Record<string, unknown>[],
    );
    await ctx.services.output.render(response, {
      format: ctx.globalOptions.output,
      query: ctx.globalOptions.query,
    });
    return;
  }

  const update = await parseBody(ctx.options.data, ctx.options.file, ctx.options.set);
  const filter = resolveBulkFilter(ctx.options);
  const response = await ctx.services.records.updateMany(ctx.object, update, { filter });
  await ctx.services.output.render(response, {
    format: ctx.globalOptions.output,
    query: ctx.globalOptions.query,
  });
}
