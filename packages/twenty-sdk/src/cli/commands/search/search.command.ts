import { Command } from "commander";
import { applyGlobalOptions, resolveGlobalOptions } from "../../utilities/shared/global-options";
import { createServices } from "../../utilities/shared/services";
import { SearchService } from "../../utilities/search/services/search.service";
import { readJsonInput } from "../../utilities/shared/io";
import { CliError } from "../../utilities/errors/cli-error";

export function registerSearchCommand(program: Command): void {
  const cmd = program
    .command("search")
    .description("Full-text search across all records")
    .argument("<query>", "Search query")
    .option("--limit <number>", "Maximum results", "20")
    .option("--objects <list>", "Comma-separated object names to include")
    .option("--exclude <list>", "Comma-separated object names to exclude")
    .option("--after <cursor>", "Pagination cursor for the next page")
    .option("--include-page-info", "Include top-level pageInfo in output")
    .option("--filter <json>", "JSON filter object")
    .option("--filter-file <path>", "JSON filter file (use - for stdin)");

  applyGlobalOptions(cmd);

  cmd.action(async (query: string, options: SearchOptions, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const searchService = new SearchService(services.api);
    const filter = await parseSearchFilter(options.filter, options.filterFile);

    const response = await searchService.search({
      query,
      limit: parseInt(options.limit, 10),
      objects: options.objects?.split(","),
      excludeObjects: options.exclude?.split(","),
      after: options.after,
      filter,
    });

    const output = options.includePageInfo ? response : response.data;

    await services.output.render(output, {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });
}

interface SearchOptions {
  limit: string;
  objects?: string;
  exclude?: string;
  after?: string;
  includePageInfo?: boolean;
  filter?: string;
  filterFile?: string;
}

async function parseSearchFilter(
  data?: string,
  filePath?: string,
): Promise<Record<string, unknown> | undefined> {
  const filter = await readJsonInput(data, filePath);
  if (filter == null) {
    return undefined;
  }

  if (typeof filter !== "object" || Array.isArray(filter)) {
    throw new CliError("Search filter must be a JSON object.", "INVALID_ARGUMENTS");
  }

  return filter as Record<string, unknown>;
}
