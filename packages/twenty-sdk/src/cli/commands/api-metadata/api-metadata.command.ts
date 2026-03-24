import { Command } from "commander";
import { applyGlobalOptions, resolveGlobalOptions } from "../../utilities/shared/global-options";
import { createServices } from "../../utilities/shared/services";
import { CliError } from "../../utilities/errors/cli-error";
import { ApiMetadataContext, ApiMetadataOptions } from "./operations/types";
import { runObjectsList } from "./operations/objects-list.operation";
import { runObjectsGet } from "./operations/objects-get.operation";
import { runObjectsCreate } from "./operations/objects-create.operation";
import { runObjectsUpdate } from "./operations/objects-update.operation";
import { runObjectsDelete } from "./operations/objects-delete.operation";
import { runFieldsList } from "./operations/fields-list.operation";
import { runFieldsGet } from "./operations/fields-get.operation";
import { runFieldsCreate } from "./operations/fields-create.operation";
import { runFieldsUpdate } from "./operations/fields-update.operation";
import { runFieldsDelete } from "./operations/fields-delete.operation";
import {
  runCommandMenuItemsList,
  runCommandMenuItemsGet,
  runCommandMenuItemsCreate,
  runCommandMenuItemsUpdate,
  runCommandMenuItemsDelete,
  runFrontComponentsList,
  runFrontComponentsGet,
  runFrontComponentsCreate,
  runFrontComponentsUpdate,
  runFrontComponentsDelete,
  runNavigationMenuItemsList,
  runNavigationMenuItemsGet,
  runNavigationMenuItemsCreate,
  runNavigationMenuItemsUpdate,
  runNavigationMenuItemsDelete,
  runViewsList,
  runViewsGet,
  runViewsCreate,
  runViewsUpdate,
  runViewsDelete,
  runViewFieldsList,
  runViewFieldsGet,
  runViewFieldsCreate,
  runViewFieldsUpdate,
  runViewFieldsDelete,
  runViewFiltersList,
  runViewFiltersGet,
  runViewFiltersCreate,
  runViewFiltersUpdate,
  runViewFiltersDelete,
  runViewFilterGroupsList,
  runViewFilterGroupsGet,
  runViewFilterGroupsCreate,
  runViewFilterGroupsUpdate,
  runViewFilterGroupsDelete,
  runViewGroupsList,
  runViewGroupsGet,
  runViewGroupsCreate,
  runViewGroupsUpdate,
  runViewGroupsDelete,
  runViewSortsList,
  runViewSortsGet,
  runViewSortsCreate,
  runViewSortsUpdate,
  runViewSortsDelete,
  runPageLayoutsList,
  runPageLayoutsGet,
  runPageLayoutsCreate,
  runPageLayoutsUpdate,
  runPageLayoutsDelete,
  runPageLayoutTabsList,
  runPageLayoutTabsGet,
  runPageLayoutTabsCreate,
  runPageLayoutTabsUpdate,
  runPageLayoutTabsDelete,
  runPageLayoutWidgetsList,
  runPageLayoutWidgetsGet,
  runPageLayoutWidgetsCreate,
  runPageLayoutWidgetsUpdate,
  runPageLayoutWidgetsDelete,
} from "./operations/ui-metadata.operations";

const handlers: Record<string, (ctx: ApiMetadataContext) => Promise<void>> = {
  "objects:list": runObjectsList,
  "objects:get": runObjectsGet,
  "objects:create": runObjectsCreate,
  "objects:update": runObjectsUpdate,
  "objects:delete": runObjectsDelete,
  "fields:list": runFieldsList,
  "fields:get": runFieldsGet,
  "fields:create": runFieldsCreate,
  "fields:update": runFieldsUpdate,
  "fields:delete": runFieldsDelete,
  "command-menu-items:list": runCommandMenuItemsList,
  "command-menu-items:get": runCommandMenuItemsGet,
  "command-menu-items:create": runCommandMenuItemsCreate,
  "command-menu-items:update": runCommandMenuItemsUpdate,
  "command-menu-items:delete": runCommandMenuItemsDelete,
  "front-components:list": runFrontComponentsList,
  "front-components:get": runFrontComponentsGet,
  "front-components:create": runFrontComponentsCreate,
  "front-components:update": runFrontComponentsUpdate,
  "front-components:delete": runFrontComponentsDelete,
  "navigation-menu-items:list": runNavigationMenuItemsList,
  "navigation-menu-items:get": runNavigationMenuItemsGet,
  "navigation-menu-items:create": runNavigationMenuItemsCreate,
  "navigation-menu-items:update": runNavigationMenuItemsUpdate,
  "navigation-menu-items:delete": runNavigationMenuItemsDelete,
  "views:list": runViewsList,
  "views:get": runViewsGet,
  "views:create": runViewsCreate,
  "views:update": runViewsUpdate,
  "views:delete": runViewsDelete,
  "view-fields:list": runViewFieldsList,
  "view-fields:get": runViewFieldsGet,
  "view-fields:create": runViewFieldsCreate,
  "view-fields:update": runViewFieldsUpdate,
  "view-fields:delete": runViewFieldsDelete,
  "view-filters:list": runViewFiltersList,
  "view-filters:get": runViewFiltersGet,
  "view-filters:create": runViewFiltersCreate,
  "view-filters:update": runViewFiltersUpdate,
  "view-filters:delete": runViewFiltersDelete,
  "view-filter-groups:list": runViewFilterGroupsList,
  "view-filter-groups:get": runViewFilterGroupsGet,
  "view-filter-groups:create": runViewFilterGroupsCreate,
  "view-filter-groups:update": runViewFilterGroupsUpdate,
  "view-filter-groups:delete": runViewFilterGroupsDelete,
  "view-groups:list": runViewGroupsList,
  "view-groups:get": runViewGroupsGet,
  "view-groups:create": runViewGroupsCreate,
  "view-groups:update": runViewGroupsUpdate,
  "view-groups:delete": runViewGroupsDelete,
  "view-sorts:list": runViewSortsList,
  "view-sorts:get": runViewSortsGet,
  "view-sorts:create": runViewSortsCreate,
  "view-sorts:update": runViewSortsUpdate,
  "view-sorts:delete": runViewSortsDelete,
  "page-layouts:list": runPageLayoutsList,
  "page-layouts:get": runPageLayoutsGet,
  "page-layouts:create": runPageLayoutsCreate,
  "page-layouts:update": runPageLayoutsUpdate,
  "page-layouts:delete": runPageLayoutsDelete,
  "page-layout-tabs:list": runPageLayoutTabsList,
  "page-layout-tabs:get": runPageLayoutTabsGet,
  "page-layout-tabs:create": runPageLayoutTabsCreate,
  "page-layout-tabs:update": runPageLayoutTabsUpdate,
  "page-layout-tabs:delete": runPageLayoutTabsDelete,
  "page-layout-widgets:list": runPageLayoutWidgetsList,
  "page-layout-widgets:get": runPageLayoutWidgetsGet,
  "page-layout-widgets:create": runPageLayoutWidgetsCreate,
  "page-layout-widgets:update": runPageLayoutWidgetsUpdate,
  "page-layout-widgets:delete": runPageLayoutWidgetsDelete,
};

export function registerApiMetadataCommand(program: Command): void {
  const cmd = program
    .command("api-metadata")
    .description("Schema operations")
    .argument(
      "<type>",
      "Metadata type (objects, fields, command-menu-items, front-components, navigation-menu-items, views, view-fields, view-filters, view-filter-groups, view-groups, view-sorts, page-layouts, page-layout-tabs, or page-layout-widgets)",
    )
    .argument("<operation>", "Operation to perform")
    .argument("[arg]", "Identifier")
    .option("-d, --data <json>", "JSON payload")
    .option("-f, --file <path>", "JSON payload file (use - for stdin)")
    .option("--object <nameOrId>", "Filter by object name or metadata ID")
    .option("--view <id>", "Filter by view ID")
    .option("--page-layout <id>", "Filter by page layout ID")
    .option("--page-layout-tab <id>", "Filter by page layout tab ID")
    .option("--page-layout-type <type>", "Filter page layouts by type (requires --object)");

  applyGlobalOptions(cmd);

  cmd.action(
    async (
      type: string,
      operation: string,
      arg?: string,
      options?: ApiMetadataOptions | Command,
      command?: Command,
    ) => {
      const key = `${type.toLowerCase()}:${operation.toLowerCase()}`;
      const handler = handlers[key];
      if (!handler) {
        throw new CliError(
          `Unknown api-metadata operation ${JSON.stringify(type)} ${JSON.stringify(operation)}.`,
          "INVALID_ARGUMENTS",
        );
      }

      const resolvedCommand = command ?? (options instanceof Command ? options : cmd);
      const globalOptions = resolveGlobalOptions(resolvedCommand);
      const services = createServices(globalOptions);
      const rawOptions = resolvedCommand.opts() as ApiMetadataOptions;

      await handler({
        type,
        operation,
        arg,
        options: rawOptions ?? {},
        services,
        globalOptions,
      });
    },
  );
}
