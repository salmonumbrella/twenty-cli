import { Command } from 'commander';
import { ConfigService } from '../../utilities/config/services/config.service';
import { OutputService } from '../../utilities/output/services/output.service';
import { TableService } from '../../utilities/output/services/table.service';
import { QueryService } from '../../utilities/output/services/query.service';
import { CliError } from '../../utilities/errors/cli-error';

function maskToken(token: string): string {
  if (token.length <= 8) return '****';
  return token.slice(0, 4) + '****' + token.slice(-4);
}

export function registerAuthCommand(program: Command): void {
  const authCmd = program
    .command('auth')
    .description('Manage authentication and workspaces');

  // auth list
  authCmd
    .command('list')
    .description('List configured workspaces')
    .option('-o, --output <format>', 'Output format (text, json)', 'text')
    .action(async (options: { output: string }) => {
      const configService = new ConfigService();
      const output = new OutputService(new TableService(), new QueryService());

      const workspaces = await configService.listWorkspaces();

      if (workspaces.length === 0) {
        // eslint-disable-next-line no-console
        console.log('No workspaces configured. Use "twenty auth login" to add a workspace.');
        return;
      }

      const displayData = workspaces.map((ws) => ({
        name: ws.name,
        default: ws.isDefault ? 'Y' : '',
        apiUrl: ws.apiUrl ?? '',
      }));

      await output.render(displayData, { format: options.output });
    });

  // auth switch
  authCmd
    .command('switch')
    .description('Set default workspace')
    .argument('<workspace>', 'Workspace name')
    .action(async (workspace: string) => {
      const configService = new ConfigService();
      await configService.setDefaultWorkspace(workspace);
      // eslint-disable-next-line no-console
      console.log(`Switched to workspace "${workspace}".`);
    });

  // auth status
  authCmd
    .command('status')
    .description('Show current authentication status')
    .option('--show-token', 'Show full API token')
    .option('-o, --output <format>', 'Output format (text, json)', 'text')
    .action(async (options: { showToken?: boolean; output: string }) => {
      const configService = new ConfigService();
      const output = new OutputService(new TableService(), new QueryService());

      try {
        const config = await configService.getConfig();
        const statusData = {
          authenticated: true,
          workspace: config.workspace,
          apiUrl: config.apiUrl,
          apiKey: options.showToken ? config.apiKey : maskToken(config.apiKey),
        };

        await output.render(statusData, { format: options.output });
      } catch (error) {
        if (error instanceof CliError && error.code === 'AUTH') {
          const statusData = {
            authenticated: false,
            error: error.message,
          };
          await output.render(statusData, { format: options.output });
        } else {
          throw error;
        }
      }
    });

  // auth login
  authCmd
    .command('login')
    .description('Configure API credentials')
    .requiredOption('--token <token>', 'API token')
    .option('--base-url <url>', 'API base URL', 'https://api.twenty.com')
    .option('--workspace <name>', 'Workspace name', 'default')
    .action(async (options: { token: string; baseUrl: string; workspace: string }) => {
      const configService = new ConfigService();

      await configService.saveWorkspace(options.workspace, {
        apiKey: options.token,
        apiUrl: options.baseUrl,
      });

      // eslint-disable-next-line no-console
      console.log(`Workspace "${options.workspace}" configured.`);
      // eslint-disable-next-line no-console
      console.log(`API URL: ${options.baseUrl}`);
    });

  // auth logout
  authCmd
    .command('logout')
    .description('Remove credentials')
    .option('--workspace <name>', 'Workspace name to remove')
    .option('--all', 'Remove all workspaces')
    .action(async (options: { workspace?: string; all?: boolean }) => {
      const configService = new ConfigService();

      if (options.all) {
        const workspaces = await configService.listWorkspaces();
        for (const ws of workspaces) {
          await configService.removeWorkspace(ws.name);
        }
        // eslint-disable-next-line no-console
        console.log('All workspaces removed.');
        return;
      }

      let workspaceToRemove: string;
      if (options.workspace) {
        workspaceToRemove = options.workspace;
      } else {
        // Get current default workspace
        try {
          const config = await configService.getConfig();
          workspaceToRemove = config.workspace ?? 'default';
        } catch {
          throw new CliError(
            'No workspace specified and no default workspace configured.',
            'INVALID_ARGUMENTS',
            'Use --workspace <name> or --all to specify what to remove.'
          );
        }
      }

      await configService.removeWorkspace(workspaceToRemove);
      // eslint-disable-next-line no-console
      console.log(`Workspace "${workspaceToRemove}" removed.`);
    });
}
