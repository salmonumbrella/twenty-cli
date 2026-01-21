import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { Command } from 'commander';
import { registerAuthCommand } from '../auth.command';
import { ConfigService, WorkspaceInfo, ResolvedConfig } from '../../../utilities/config/services/config.service';
import { CliError } from '../../../utilities/errors/cli-error';

vi.mock('../../../utilities/config/services/config.service');

describe('auth commands', () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerAuthCommand(program);
    consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
  });

  afterEach(() => {
    consoleSpy.mockRestore();
    vi.clearAllMocks();
  });

  describe('auth list', () => {
    it('shows message when no workspaces configured', async () => {
      vi.mocked(ConfigService.prototype.listWorkspaces).mockResolvedValue([]);

      await program.parseAsync(['node', 'test', 'auth', 'list']);

      expect(consoleSpy).toHaveBeenCalledWith(
        'No workspaces configured. Use "twenty auth login" to add a workspace.'
      );
    });

    it('lists configured workspaces with default marker', async () => {
      const workspaces: WorkspaceInfo[] = [
        { name: 'production', isDefault: true, apiUrl: 'https://api.twenty.com' },
        { name: 'staging', isDefault: false, apiUrl: 'https://staging.twenty.com' },
      ];
      vi.mocked(ConfigService.prototype.listWorkspaces).mockResolvedValue(workspaces);

      await program.parseAsync(['node', 'test', 'auth', 'list', '-o', 'json']);

      expect(ConfigService.prototype.listWorkspaces).toHaveBeenCalled();
      // JSON output will contain the formatted workspace data
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual([
        { name: 'production', default: 'Y', apiUrl: 'https://api.twenty.com' },
        { name: 'staging', default: '', apiUrl: 'https://staging.twenty.com' },
      ]);
    });
  });

  describe('auth switch', () => {
    it('switches default workspace successfully', async () => {
      vi.mocked(ConfigService.prototype.setDefaultWorkspace).mockResolvedValue(undefined);

      await program.parseAsync(['node', 'test', 'auth', 'switch', 'production']);

      expect(ConfigService.prototype.setDefaultWorkspace).toHaveBeenCalledWith('production');
      expect(consoleSpy).toHaveBeenCalledWith('Switched to workspace "production".');
    });
  });

  describe('auth login', () => {
    it('saves workspace with token and default URL/workspace', async () => {
      vi.mocked(ConfigService.prototype.saveWorkspace).mockResolvedValue(undefined);

      await program.parseAsync(['node', 'test', 'auth', 'login', '--token', 'my-api-token']);

      expect(ConfigService.prototype.saveWorkspace).toHaveBeenCalledWith('default', {
        apiKey: 'my-api-token',
        apiUrl: 'https://api.twenty.com',
      });
      expect(consoleSpy).toHaveBeenCalledWith('Workspace "default" configured.');
      expect(consoleSpy).toHaveBeenCalledWith('API URL: https://api.twenty.com');
    });

    it('saves workspace with custom base-url and workspace name', async () => {
      vi.mocked(ConfigService.prototype.saveWorkspace).mockResolvedValue(undefined);

      await program.parseAsync([
        'node', 'test', 'auth', 'login',
        '--token', 'custom-token',
        '--base-url', 'https://custom.twenty.com',
        '--workspace', 'production',
      ]);

      expect(ConfigService.prototype.saveWorkspace).toHaveBeenCalledWith('production', {
        apiKey: 'custom-token',
        apiUrl: 'https://custom.twenty.com',
      });
      expect(consoleSpy).toHaveBeenCalledWith('Workspace "production" configured.');
      expect(consoleSpy).toHaveBeenCalledWith('API URL: https://custom.twenty.com');
    });
  });

  describe('auth logout', () => {
    it('removes specified workspace', async () => {
      vi.mocked(ConfigService.prototype.removeWorkspace).mockResolvedValue(undefined);

      await program.parseAsync(['node', 'test', 'auth', 'logout', '--workspace', 'staging']);

      expect(ConfigService.prototype.removeWorkspace).toHaveBeenCalledWith('staging');
      expect(consoleSpy).toHaveBeenCalledWith('Workspace "staging" removed.');
    });

    it('removes all workspaces with --all flag', async () => {
      const workspaces: WorkspaceInfo[] = [
        { name: 'production', isDefault: true, apiUrl: 'https://api.twenty.com' },
        { name: 'staging', isDefault: false, apiUrl: 'https://staging.twenty.com' },
      ];
      vi.mocked(ConfigService.prototype.listWorkspaces).mockResolvedValue(workspaces);
      vi.mocked(ConfigService.prototype.removeWorkspace).mockResolvedValue(undefined);

      await program.parseAsync(['node', 'test', 'auth', 'logout', '--all']);

      expect(ConfigService.prototype.listWorkspaces).toHaveBeenCalled();
      expect(ConfigService.prototype.removeWorkspace).toHaveBeenCalledWith('production');
      expect(ConfigService.prototype.removeWorkspace).toHaveBeenCalledWith('staging');
      expect(consoleSpy).toHaveBeenCalledWith('All workspaces removed.');
    });

    it('removes default workspace when no option specified', async () => {
      const config: ResolvedConfig = {
        apiUrl: 'https://api.twenty.com',
        apiKey: 'test-token',
        workspace: 'current-workspace',
      };
      vi.mocked(ConfigService.prototype.getConfig).mockResolvedValue(config);
      vi.mocked(ConfigService.prototype.removeWorkspace).mockResolvedValue(undefined);

      await program.parseAsync(['node', 'test', 'auth', 'logout']);

      expect(ConfigService.prototype.getConfig).toHaveBeenCalled();
      expect(ConfigService.prototype.removeWorkspace).toHaveBeenCalledWith('current-workspace');
      expect(consoleSpy).toHaveBeenCalledWith('Workspace "current-workspace" removed.');
    });
  });

  describe('auth status', () => {
    it('shows authenticated status with masked token', async () => {
      const config: ResolvedConfig = {
        apiUrl: 'https://api.twenty.com',
        apiKey: 'abcd1234efgh5678',
        workspace: 'production',
      };
      vi.mocked(ConfigService.prototype.getConfig).mockResolvedValue(config);

      await program.parseAsync(['node', 'test', 'auth', 'status', '-o', 'json']);

      expect(ConfigService.prototype.getConfig).toHaveBeenCalled();
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual({
        authenticated: true,
        workspace: 'production',
        apiUrl: 'https://api.twenty.com',
        apiKey: 'abcd****5678',
      });
    });

    it('shows full token when --show-token used', async () => {
      const config: ResolvedConfig = {
        apiUrl: 'https://api.twenty.com',
        apiKey: 'abcd1234efgh5678',
        workspace: 'production',
      };
      vi.mocked(ConfigService.prototype.getConfig).mockResolvedValue(config);

      await program.parseAsync(['node', 'test', 'auth', 'status', '--show-token', '-o', 'json']);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual({
        authenticated: true,
        workspace: 'production',
        apiUrl: 'https://api.twenty.com',
        apiKey: 'abcd1234efgh5678',
      });
    });

    it('shows unauthenticated status when no config exists', async () => {
      const authError = new CliError(
        'Missing API token.',
        'AUTH',
        'Set TWENTY_TOKEN or configure ~/.twenty/config.json with an apiKey.'
      );
      vi.mocked(ConfigService.prototype.getConfig).mockRejectedValue(authError);

      await program.parseAsync(['node', 'test', 'auth', 'status', '-o', 'json']);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual({
        authenticated: false,
        error: 'Missing API token.',
      });
    });

    it('masks short tokens properly', async () => {
      const config: ResolvedConfig = {
        apiUrl: 'https://api.twenty.com',
        apiKey: 'short',
        workspace: 'production',
      };
      vi.mocked(ConfigService.prototype.getConfig).mockResolvedValue(config);

      await program.parseAsync(['node', 'test', 'auth', 'status', '-o', 'json']);

      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.apiKey).toBe('****');
    });
  });
});
