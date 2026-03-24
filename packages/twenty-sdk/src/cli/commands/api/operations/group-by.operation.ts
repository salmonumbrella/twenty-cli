import { ApiOperationContext } from "./types";
import { readJsonInput } from "../../../utilities/shared/io";
import { parseKeyValuePairs } from "../../../utilities/shared/parse";

export async function runGroupByOperation(ctx: ApiOperationContext): Promise<void> {
  let payload: unknown | undefined;
  const params = parseKeyValuePairs(ctx.options.param);

  if (ctx.options.data || ctx.options.file) {
    const rawPayload = await readJsonInput(ctx.options.data, ctx.options.file);
    payload = normalizeGroupByPayload(rawPayload);
  } else if (ctx.options.field) {
    payload = { groupBy: [{ [ctx.options.field]: true }] };
  }

  if (ctx.options.filter) {
    payload = mergeGroupByFilter(payload, ctx.options.filter);
  }

  const response = await ctx.services.records.groupBy(
    ctx.object,
    payload,
    Object.keys(params).length ? params : undefined,
  );
  await ctx.services.output.render(response, {
    format: ctx.globalOptions.output,
    query: ctx.globalOptions.query,
  });
}

function normalizeGroupByPayload(payload: unknown): unknown {
  if (Array.isArray(payload)) {
    return { groupBy: payload };
  }

  if (typeof payload !== "object" || payload === null) {
    return payload;
  }

  const record = payload as Record<string, unknown>;
  if (typeof record.groupBy === "string") {
    return {
      ...record,
      groupBy: [{ [record.groupBy]: true }],
    };
  }

  return payload;
}

function mergeGroupByFilter(payload: unknown, filter: string): unknown {
  if (payload == null) {
    return { filter };
  }

  if (typeof payload !== "object" || Array.isArray(payload)) {
    return payload;
  }

  const record = payload as Record<string, unknown>;
  if (record.filter === undefined) {
    return {
      ...record,
      filter,
    };
  }

  return payload;
}
