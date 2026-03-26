import { Command } from "commander";
import {
  createServerlessOperationContext,
  streamLogicFunctionLogs,
} from "../serverless.shared";

export async function runServerlessLogsOperation(
  id: string | undefined,
  command: Command,
): Promise<void> {
  const context = createServerlessOperationContext(command);
  await streamLogicFunctionLogs(id, context.options, context);
}
