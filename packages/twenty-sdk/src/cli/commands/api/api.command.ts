import { Command } from "commander";
import { applyGlobalOptions } from "../../utilities/shared/global-options";
import { createCommandContext } from "../../utilities/shared/context";
import { registerCommand } from "../../utilities/shared/register-command";
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

  registerCommand(api, "list", "List records", (command) => {
    command.argument("<object>", "Object name (plural)");
    applyApiOptions(command);
    applyGlobalOptions(command);
    command.action(async (object: string, _options: unknown, actionCommand: Command) => {
      await runListOperation(createApiOperationContext(actionCommand, object));
    });
  });

  registerCommand(api, "get", "Get a record", (command) => {
    command.argument("<object>", "Object name (plural)");
    command.argument("[id]", "Record ID");
    applyApiOptions(command);
    applyGlobalOptions(command);
    command.action(
      async (object: string, id: string | undefined, _options: unknown, actionCommand: Command) => {
        await runGetOperation(createApiOperationContext(actionCommand, object, id));
      },
    );
  });

  registerCommand(api, "create", "Create a record", (command) => {
    command.argument("<object>", "Object name (plural)");
    applyApiOptions(command);
    applyGlobalOptions(command);
    command.action(async (object: string, _options: unknown, actionCommand: Command) => {
      await runCreateOperation(createApiOperationContext(actionCommand, object));
    });
  });

  registerCommand(api, "update", "Update a record", (command) => {
    command.argument("<object>", "Object name (plural)");
    command.argument("[id]", "Record ID");
    applyApiOptions(command);
    applyGlobalOptions(command);
    command.action(
      async (object: string, id: string | undefined, _options: unknown, actionCommand: Command) => {
        await runUpdateOperation(createApiOperationContext(actionCommand, object, id));
      },
    );
  });

  registerCommand(api, "delete", "Delete a record", (command) => {
    command.argument("<object>", "Object name (plural)");
    command.argument("[id]", "Record ID");
    applyApiOptions(command);
    applyApiDestructiveOptions(command);
    applyGlobalOptions(command);
    command.action(
      async (object: string, id: string | undefined, _options: unknown, actionCommand: Command) => {
        await runDeleteOperation(createApiOperationContext(actionCommand, object, id));
      },
    );
  });

  registerCommand(api, "destroy", "Permanently destroy a record", (command) => {
    command.argument("<object>", "Object name (plural)");
    command.argument("[id]", "Record ID");
    applyApiOptions(command);
    applyApiDestructiveOptions(command);
    applyGlobalOptions(command);
    command.action(
      async (object: string, id: string | undefined, _options: unknown, actionCommand: Command) => {
        await runDestroyOperation(createApiOperationContext(actionCommand, object, id));
      },
    );
  });

  registerCommand(api, "restore", "Restore a deleted record", (command) => {
    command.argument("<object>", "Object name (plural)");
    command.argument("[id]", "Record ID");
    applyApiOptions(command);
    applyGlobalOptions(command);
    command.action(
      async (object: string, id: string | undefined, _options: unknown, actionCommand: Command) => {
        await runRestoreOperation(createApiOperationContext(actionCommand, object, id));
      },
    );
  });

  registerCommand(api, "batch-create", "Create many records", (command) => {
    command.argument("<object>", "Object name (plural)");
    applyApiOptions(command);
    applyGlobalOptions(command);
    command.action(async (object: string, _options: unknown, actionCommand: Command) => {
      await runBatchCreateOperation(createApiOperationContext(actionCommand, object));
    });
  });

  registerCommand(api, "batch-update", "Update many records", (command) => {
    command.argument("<object>", "Object name (plural)");
    applyApiOptions(command);
    applyGlobalOptions(command);
    command.action(async (object: string, _options: unknown, actionCommand: Command) => {
      await runBatchUpdateOperation(createApiOperationContext(actionCommand, object));
    });
  });

  registerCommand(api, "batch-delete", "Delete many records", (command) => {
    command.argument("<object>", "Object name (plural)");
    applyApiOptions(command);
    applyApiDestructiveOptions(command);
    applyGlobalOptions(command);
    command.action(async (object: string, _options: unknown, actionCommand: Command) => {
      await runBatchDeleteOperation(createApiOperationContext(actionCommand, object));
    });
  });

  registerCommand(api, "import", "Import records from a file", (command) => {
    command.argument("<object>", "Object name (plural)");
    command.argument("[filePath]", "Import file path");
    applyApiOptions(command);
    applyGlobalOptions(command);
    command.action(
      async (
        object: string,
        filePath: string | undefined,
        _options: unknown,
        actionCommand: Command,
      ) => {
        await runImportOperation(createApiOperationContext(actionCommand, object, filePath));
      },
    );
  });

  registerCommand(api, "export", "Export records", (command) => {
    command.argument("<object>", "Object name (plural)");
    applyApiOptions(command);
    applyGlobalOptions(command);
    command.action(async (object: string, _options: unknown, actionCommand: Command) => {
      await runExportOperation(createApiOperationContext(actionCommand, object));
    });
  });

  registerCommand(api, "find-duplicates", "Find duplicate records", (command) => {
    command.argument("<object>", "Object name (plural)");
    applyApiOptions(command);
    applyGlobalOptions(command);
    command.action(async (object: string, _options: unknown, actionCommand: Command) => {
      await runFindDuplicatesOperation(createApiOperationContext(actionCommand, object));
    });
  });

  registerCommand(api, "group-by", "Group records", (command) => {
    command.argument("<object>", "Object name (plural)");
    applyApiOptions(command);
    applyGlobalOptions(command);
    command.action(async (object: string, _options: unknown, actionCommand: Command) => {
      await runGroupByOperation(createApiOperationContext(actionCommand, object));
    });
  });

  registerCommand(api, "merge", "Merge records", (command) => {
    command.argument("<object>", "Object name (plural)");
    applyApiOptions(command);
    applyGlobalOptions(command);
    command.action(async (object: string, _options: unknown, actionCommand: Command) => {
      await runMergeOperation(createApiOperationContext(actionCommand, object));
    });
  });
}

function collect(value: string, previous: string[] = []): string[] {
  return previous.concat([value]);
}
