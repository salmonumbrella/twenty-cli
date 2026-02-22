import { Command } from 'commander';
import { applyGlobalOptions, resolveGlobalOptions } from '../../utilities/shared/global-options';
import { createServices } from '../../utilities/shared/services';
import { SearchService } from '../../utilities/search/services/search.service';

export function registerSearchCommand(program: Command): void {
  const cmd = program
    .command('search')
    .description('Full-text search across all records')
    .argument('<query>', 'Search query')
    .option('--limit <number>', 'Maximum results', '20')
    .option('--objects <list>', 'Comma-separated object names to include')
    .option('--exclude <list>', 'Comma-separated object names to exclude');

  applyGlobalOptions(cmd);

  cmd.action(async (query: string, options: SearchOptions, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const searchService = new SearchService(services.api);

    const results = await searchService.search({
      query,
      limit: parseInt(options.limit, 10),
      objects: options.objects?.split(','),
      excludeObjects: options.exclude?.split(','),
    });

    await services.output.render(results, {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });
}

interface SearchOptions {
  limit: string;
  objects?: string;
  exclude?: string;
}
