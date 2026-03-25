import { Command } from "commander";
import { applyGlobalOptions } from "../../utilities/shared/global-options";
import { createCommandContext } from "../../utilities/shared/context";
import { ApiCommandOptions, ApiOperationContext } from "./operations/types";
import { runListOperation } from "./operations/list.operation";
import { runGetOperation } from "./operations/get.operation";
import { runCreateOperation } from "./operations/create.operation";
import { runUpdateOperation } from "./operations/update.operation";
import { runDeleteOperation } from "./operations/delete.operation";
import { runDestroyOperation } from "./operations/destroy.operation";
import { runRestoreOperation } from "./operations/restore.operation";
import { runBatchCreateOperation } from "./operations/batch-create.operation";
import { runBatchUpdateOperation } from "./operations/batch-update.operation";
import { runBatchDeleteOperation } from "./operations/batch-delete.operation";
import { runImportOperation } from "./operations/import.operation";
import { runExportOperation } from "./operations/export.operation";
import { runGroupByOperation } from "./operations/group-by.operation";
import { runFindDuplicatesOperation } from "./operations/find-duplicates.operation";
import { runMergeOperation } from "./operations/merge.operation";

function applyApiOptions(command: Command): void {
  command
    .option("--limit <number>", "Limit number of records")
    .option("--all", "Fetch all records")
    .option("--filter <expression>", "Filter expression")
    .option("--include <relations>", "Include related records")
    .option("--cursor <cursor>", "Pagination cursor")
    .option("--sort <field>", "Sort field")
    .option("--order <direction>", "Sort order (asc or desc)")
    .option("--param <key=value>", "Additional query params", collect)
    .option("-d, --data <json>", "JSON payload")
    .option("-f, --file <path>", "JSON/CSV file payload (use - for stdin)")
    .option("--set <key=value>", "Set a field value", collect)
    .option("--ids <ids>", "Comma-separated IDs")
    .option("--format <format>", "Export format (json or csv)")
    .option("--output-file <path>", "Output file path")
    .option("--batch-size <number>", "Batch size (import)")
    .option("--dry-run", "Preview without executing")
    .option("--continue-on-error", "Continue on batch errors")
    .option("--field <field>", "Group-by field")
    .option("--source <id>", "Source record ID (merge)")
    .option("--target <id>", "Target record ID (merge)")
    .option("--priority <index>", "Conflict priority index (merge)");
}

function applyApiDestructiveOptions(command: Command): void {
  command.option("--yes", "Confirm destructive operations");
}

function createApiOperationContext(
  command: Command,
  object: string,
  arg?: string,
  arg2?: string,
): ApiOperationContext {
  const { globalOptions, services } = createCommandContext(command);
  const rawOptions = command.opts() as ApiCommandOptions;

  return {
    object,
    arg,
    arg2,
    options: rawOptions,
    services,
    globalOptions,
  };
}

export function registerApiCommand(program: Command): void {
  const api = program.command("api").description("Record operations");
  applyGlobalOptions(api);

  const listCmd = api.command("list").description("List records").argument("<object>", "Object name (plural)");
  applyApiOptions(listCmd);
  applyGlobalOptions(listCmd);
  listCmd.action(async (object: string, _options: unknown, command: Command) => {
    await runListOperation(createApiOperationContext(command, object));
  });

  const getCmd = api
    .command("get")
    .description("Get a record")
    .argument("<object>", "Object name (plural)")
    .argument("[id]", "Record ID");
  applyApiOptions(getCmd);
  applyGlobalOptions(getCmd);
  getCmd.action(async (object: string, id: string | undefined, _options: unknown, command: Command) => {
    await runGetOperation(createApiOperationContext(command, object, id));
  });

  const createCmd = api.command("create").description("Create a record").argument("<object>", "Object name (plural)");
  applyApiOptions(createCmd);
  applyGlobalOptions(createCmd);
  createCmd.action(async (object: string, _options: unknown, command: Command) => {
    await runCreateOperation(createApiOperationContext(command, object));
  });

  const updateCmd = api
    .command("update")
    .description("Update a record")
    .argument("<object>", "Object name (plural)")
    .argument("[id]", "Record ID");
  applyApiOptions(updateCmd);
  applyGlobalOptions(updateCmd);
  updateCmd.action(async (object: string, id: string | undefined, _options: unknown, command: Command) => {
    await runUpdateOperation(createApiOperationContext(command, object, id));
  });

  const deleteCmd = api
    .command("delete")
    .description("Delete a record")
    .argument("<object>", "Object name (plural)")
    .argument("[id]", "Record ID");
  applyApiOptions(deleteCmd);
  applyApiDestructiveOptions(deleteCmd);
  applyGlobalOptions(deleteCmd);
  deleteCmd.action(async (object: string, id: string | undefined, _options: unknown, command: Command) => {
    await runDeleteOperation(createApiOperationContext(command, object, id));
  });

  const destroyCmd = api
    .command("destroy")
    .description("Permanently destroy a record")
    .argument("<object>", "Object name (plural)")
    .argument("[id]", "Record ID");
  applyApiOptions(destroyCmd);
  applyApiDestructiveOptions(destroyCmd);
  applyGlobalOptions(destroyCmd);
  destroyCmd.action(async (object: string, id: string | undefined, _options: unknown, command: Command) => {
    await runDestroyOperation(createApiOperationContext(command, object, id));
  });

  const restoreCmd = api
    .command("restore")
    .description("Restore a deleted record")
    .argument("<object>", "Object name (plural)")
    .argument("[id]", "Record ID");
  applyApiOptions(restoreCmd);
  applyGlobalOptions(restoreCmd);
  restoreCmd.action(async (object: string, id: string | undefined, _options: unknown, command: Command) => {
    await runRestoreOperation(createApiOperationContext(command, object, id));
  });

  const batchCreateCmd = api
    .command("batch-create")
    .description("Create many records")
    .argument("<object>", "Object name (plural)");
  applyApiOptions(batchCreateCmd);
  applyGlobalOptions(batchCreateCmd);
  batchCreateCmd.action(async (object: string, _options: unknown, command: Command) => {
    await runBatchCreateOperation(createApiOperationContext(command, object));
  });

  const batchUpdateCmd = api
    .command("batch-update")
    .description("Update many records")
    .argument("<object>", "Object name (plural)");
  applyApiOptions(batchUpdateCmd);
  applyGlobalOptions(batchUpdateCmd);
  batchUpdateCmd.action(async (object: string, _options: unknown, command: Command) => {
    await runBatchUpdateOperation(createApiOperationContext(command, object));
  });

  const batchDeleteCmd = api
    .command("batch-delete")
    .description("Delete many records")
    .argument("<object>", "Object name (plural)");
  applyApiOptions(batchDeleteCmd);
  applyApiDestructiveOptions(batchDeleteCmd);
  applyGlobalOptions(batchDeleteCmd);
  batchDeleteCmd.action(async (object: string, _options: unknown, command: Command) => {
    await runBatchDeleteOperation(createApiOperationContext(command, object));
  });

  const importCmd = api
    .command("import")
    .description("Import records from a file")
    .argument("<object>", "Object name (plural)")
    .argument("[filePath]", "Import file path");
  applyApiOptions(importCmd);
  applyGlobalOptions(importCmd);
  importCmd.action(async (object: string, filePath: string | undefined, _options: unknown, command: Command) => {
    await runImportOperation(createApiOperationContext(command, object, filePath));
  });

  const exportCmd = api.command("export").description("Export records").argument("<object>", "Object name (plural)");
  applyApiOptions(exportCmd);
  applyGlobalOptions(exportCmd);
  exportCmd.action(async (object: string, _options: unknown, command: Command) => {
    await runExportOperation(createApiOperationContext(command, object));
  });

  const findDuplicatesCmd = api
    .command("find-duplicates")
    .description("Find duplicate records")
    .argument("<object>", "Object name (plural)");
  applyApiOptions(findDuplicatesCmd);
  applyGlobalOptions(findDuplicatesCmd);
  findDuplicatesCmd.action(async (object: string, _options: unknown, command: Command) => {
    await runFindDuplicatesOperation(createApiOperationContext(command, object));
  });

  const groupByCmd = api
    .command("group-by")
    .description("Group records")
    .argument("<object>", "Object name (plural)");
  applyApiOptions(groupByCmd);
  applyGlobalOptions(groupByCmd);
  groupByCmd.action(async (object: string, _options: unknown, command: Command) => {
    await runGroupByOperation(createApiOperationContext(command, object));
  });

  const mergeCmd = api.command("merge").description("Merge records").argument("<object>", "Object name (plural)");
  applyApiOptions(mergeCmd);
  applyGlobalOptions(mergeCmd);
  mergeCmd.action(async (object: string, _options: unknown, command: Command) => {
    await runMergeOperation(createApiOperationContext(command, object));
  });
}

function collect(value: string, previous: string[] = []): string[] {
  return previous.concat([value]);
}
