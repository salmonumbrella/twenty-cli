import { ApiOperationContext } from "./types";
import { parseKeyValuePairs } from "../../../utilities/shared/parse";
import { CliError } from "../../../utilities/errors/cli-error";

const OUTPUT_FORMATS = new Set(["json", "csv", "text"]);

export async function runExportOperation(ctx: ApiOperationContext): Promise<void> {
  const format = (ctx.options.format ?? "json").toLowerCase();
  if (format !== "json" && format !== "csv") {
    throw new CliError(`Unsupported export format ${JSON.stringify(format)}.`, "INVALID_ARGUMENTS");
  }
  if (ctx.options.fields) {
    throw new CliError(
      "--fields is not supported for export. Twenty REST find-many only supports depth-based field expansion.",
      "INVALID_ARGUMENTS",
    );
  }

  const params = parseKeyValuePairs(ctx.options.param);
  const limit = ctx.options.limit ? Number(ctx.options.limit) : 200;
  const listOptions = {
    limit: Number.isNaN(limit) ? 200 : limit,
    cursor: ctx.options.cursor,
    filter: ctx.options.filter,
    include: ctx.options.include,
    sort: ctx.options.sort,
    order: ctx.options.order,
    params,
  };

  const shouldAll = ctx.options.all === true;
  const response = shouldAll
    ? await ctx.services.records.listAll(ctx.object, listOptions)
    : await ctx.services.records.list(ctx.object, listOptions);

  let outputFile = ctx.options.outputFile;
  if (!outputFile && ctx.options.output && !OUTPUT_FORMATS.has(ctx.options.output)) {
    outputFile = ctx.options.output;
  }

  await ctx.services.exporter.export(response.data as Record<string, unknown>[], {
    format: format as "json" | "csv",
    output: outputFile,
  });
}
