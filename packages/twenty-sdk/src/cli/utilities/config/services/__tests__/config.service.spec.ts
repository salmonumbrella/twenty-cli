import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ConfigService, WorkspaceInfo, TwentyConfigFile } from '../config.service';
import { CliError } from '../../../errors/cli-error';
import fs from 'fs-extra';
import os from 'os';

vi.mock('fs-extra');
vi.mock('os');

describe('ConfigService', () => {
  const mockHomedir = '/home/testuser';
  const mockConfigPath = `${mockHomedir}/.twenty/config.json`;

  beforeEach(() => {
    vi.mocked(os.homedir).mockReturnValue(mockHomedir);
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('listWorkspaces', () => {
    it('returns empty array when no config exists', async () => {
      vi.mocked(fs.pathExists).mockResolvedValue(false as never);

      const service = new ConfigService();
      const result = await service.listWorkspaces();

      expect(result).toEqual([]);
    });

    it('returns empty array when config has no workspaces', async () => {
      vi.mocked(fs.pathExists).mockResolvedValue(true as never);
      vi.mocked(fs.readFile).mockResolvedValue(JSON.stringify({}) as never);

      const service = new ConfigService();
      const result = await service.listWorkspaces();

      expect(result).toEqual([]);
    });

    it('returns workspace info with isDefault marked', async () => {
      const config: TwentyConfigFile = {
        workspaces: {
          prod: { apiUrl: 'https://api.twenty.com', apiKey: 'key1' },
          staging: { apiUrl: 'https://staging.twenty.com', apiKey: 'key2' },
        },
        defaultWorkspace: 'prod',
      };
      vi.mocked(fs.pathExists).mockResolvedValue(true as never);
      vi.mocked(fs.readFile).mockResolvedValue(JSON.stringify(config) as never);

      const service = new ConfigService();
      const result = await service.listWorkspaces();

      expect(result).toHaveLength(2);
      expect(result).toContainEqual({
        name: 'prod',
        isDefault: true,
        apiUrl: 'https://api.twenty.com',
      });
      expect(result).toContainEqual({
        name: 'staging',
        isDefault: false,
        apiUrl: 'https://staging.twenty.com',
      });
    });

    it('returns workspaces without apiUrl when not set', async () => {
      const config: TwentyConfigFile = {
        workspaces: {
          minimal: { apiKey: 'key1' },
        },
        defaultWorkspace: 'minimal',
      };
      vi.mocked(fs.pathExists).mockResolvedValue(true as never);
      vi.mocked(fs.readFile).mockResolvedValue(JSON.stringify(config) as never);

      const service = new ConfigService();
      const result = await service.listWorkspaces();

      expect(result).toEqual([
        { name: 'minimal', isDefault: true, apiUrl: undefined },
      ]);
    });
  });

  describe('setDefaultWorkspace', () => {
    it('throws if workspace does not exist', async () => {
      const config: TwentyConfigFile = {
        workspaces: {
          prod: { apiKey: 'key1' },
        },
        defaultWorkspace: 'prod',
      };
      vi.mocked(fs.pathExists).mockResolvedValue(true as never);
      vi.mocked(fs.readFile).mockResolvedValue(JSON.stringify(config) as never);

      const service = new ConfigService();

      await expect(service.setDefaultWorkspace('nonexistent')).rejects.toThrow(CliError);
      await expect(service.setDefaultWorkspace('nonexistent')).rejects.toThrow(
        "Workspace 'nonexistent' does not exist"
      );
    });

    it('throws if no config exists', async () => {
      vi.mocked(fs.pathExists).mockResolvedValue(false as never);

      const service = new ConfigService();

      await expect(service.setDefaultWorkspace('prod')).rejects.toThrow(CliError);
      await expect(service.setDefaultWorkspace('prod')).rejects.toThrow(
        "Workspace 'prod' does not exist"
      );
    });

    it('updates and saves config with new default', async () => {
      const config: TwentyConfigFile = {
        workspaces: {
          prod: { apiKey: 'key1' },
          staging: { apiKey: 'key2' },
        },
        defaultWorkspace: 'prod',
      };
      vi.mocked(fs.pathExists).mockResolvedValue(true as never);
      vi.mocked(fs.readFile).mockResolvedValue(JSON.stringify(config) as never);
      vi.mocked(fs.outputFile).mockResolvedValue(undefined as never);

      const service = new ConfigService();
      await service.setDefaultWorkspace('staging');

      expect(fs.outputFile).toHaveBeenCalledWith(
        mockConfigPath,
        expect.stringContaining('"defaultWorkspace": "staging"'),
        'utf-8'
      );
    });
  });

  describe('saveWorkspace', () => {
    it('creates config if not exists and sets as default', async () => {
      vi.mocked(fs.pathExists).mockResolvedValue(false as never);
      vi.mocked(fs.outputFile).mockResolvedValue(undefined as never);

      const service = new ConfigService();
      await service.saveWorkspace('prod', {
        apiUrl: 'https://api.twenty.com',
        apiKey: 'key1',
      });

      expect(fs.outputFile).toHaveBeenCalledWith(
        mockConfigPath,
        expect.any(String),
        'utf-8'
      );

      const savedConfig = JSON.parse(
        vi.mocked(fs.outputFile).mock.calls[0][1] as string
      ) as TwentyConfigFile;
      expect(savedConfig.workspaces?.prod).toEqual({
        apiUrl: 'https://api.twenty.com',
        apiKey: 'key1',
      });
      expect(savedConfig.defaultWorkspace).toBe('prod');
    });

    it('adds to existing config without changing default', async () => {
      const existingConfig: TwentyConfigFile = {
        workspaces: {
          prod: { apiKey: 'key1' },
        },
        defaultWorkspace: 'prod',
      };
      vi.mocked(fs.pathExists).mockResolvedValue(true as never);
      vi.mocked(fs.readFile).mockResolvedValue(JSON.stringify(existingConfig) as never);
      vi.mocked(fs.outputFile).mockResolvedValue(undefined as never);

      const service = new ConfigService();
      await service.saveWorkspace('staging', {
        apiUrl: 'https://staging.twenty.com',
        apiKey: 'key2',
      });

      const savedConfig = JSON.parse(
        vi.mocked(fs.outputFile).mock.calls[0][1] as string
      ) as TwentyConfigFile;
      expect(savedConfig.workspaces?.staging).toEqual({
        apiUrl: 'https://staging.twenty.com',
        apiKey: 'key2',
      });
      expect(savedConfig.defaultWorkspace).toBe('prod');
    });

    it('updates existing workspace config', async () => {
      const existingConfig: TwentyConfigFile = {
        workspaces: {
          prod: { apiUrl: 'https://old.twenty.com', apiKey: 'oldkey' },
        },
        defaultWorkspace: 'prod',
      };
      vi.mocked(fs.pathExists).mockResolvedValue(true as never);
      vi.mocked(fs.readFile).mockResolvedValue(JSON.stringify(existingConfig) as never);
      vi.mocked(fs.outputFile).mockResolvedValue(undefined as never);

      const service = new ConfigService();
      await service.saveWorkspace('prod', {
        apiUrl: 'https://new.twenty.com',
        apiKey: 'newkey',
      });

      const savedConfig = JSON.parse(
        vi.mocked(fs.outputFile).mock.calls[0][1] as string
      ) as TwentyConfigFile;
      expect(savedConfig.workspaces?.prod).toEqual({
        apiUrl: 'https://new.twenty.com',
        apiKey: 'newkey',
      });
    });
  });

  describe('removeWorkspace', () => {
    it('removes workspace from config', async () => {
      const config: TwentyConfigFile = {
        workspaces: {
          prod: { apiKey: 'key1' },
          staging: { apiKey: 'key2' },
        },
        defaultWorkspace: 'prod',
      };
      vi.mocked(fs.pathExists).mockResolvedValue(true as never);
      vi.mocked(fs.readFile).mockResolvedValue(JSON.stringify(config) as never);
      vi.mocked(fs.outputFile).mockResolvedValue(undefined as never);

      const service = new ConfigService();
      await service.removeWorkspace('staging');

      const savedConfig = JSON.parse(
        vi.mocked(fs.outputFile).mock.calls[0][1] as string
      ) as TwentyConfigFile;
      expect(savedConfig.workspaces?.staging).toBeUndefined();
      expect(savedConfig.workspaces?.prod).toEqual({ apiKey: 'key1' });
      expect(savedConfig.defaultWorkspace).toBe('prod');
    });

    it('clears default if removing last workspace', async () => {
      const config: TwentyConfigFile = {
        workspaces: {
          prod: { apiKey: 'key1' },
        },
        defaultWorkspace: 'prod',
      };
      vi.mocked(fs.pathExists).mockResolvedValue(true as never);
      vi.mocked(fs.readFile).mockResolvedValue(JSON.stringify(config) as never);
      vi.mocked(fs.outputFile).mockResolvedValue(undefined as never);

      const service = new ConfigService();
      await service.removeWorkspace('prod');

      const savedConfig = JSON.parse(
        vi.mocked(fs.outputFile).mock.calls[0][1] as string
      ) as TwentyConfigFile;
      expect(savedConfig.workspaces?.prod).toBeUndefined();
      expect(savedConfig.defaultWorkspace).toBeUndefined();
    });

    it('throws if workspace does not exist', async () => {
      const config: TwentyConfigFile = {
        workspaces: {
          prod: { apiKey: 'key1' },
        },
        defaultWorkspace: 'prod',
      };
      vi.mocked(fs.pathExists).mockResolvedValue(true as never);
      vi.mocked(fs.readFile).mockResolvedValue(JSON.stringify(config) as never);

      const service = new ConfigService();

      await expect(service.removeWorkspace('nonexistent')).rejects.toThrow(CliError);
      await expect(service.removeWorkspace('nonexistent')).rejects.toThrow(
        "Workspace 'nonexistent' does not exist"
      );
    });

    it('throws if no config exists', async () => {
      vi.mocked(fs.pathExists).mockResolvedValue(false as never);

      const service = new ConfigService();

      await expect(service.removeWorkspace('prod')).rejects.toThrow(CliError);
    });

    it('sets next available workspace as default when removing default', async () => {
      const config: TwentyConfigFile = {
        workspaces: {
          alpha: { apiKey: 'key1' },
          beta: { apiKey: 'key2' },
          gamma: { apiKey: 'key3' },
        },
        defaultWorkspace: 'beta',
      };
      vi.mocked(fs.pathExists).mockResolvedValue(true as never);
      vi.mocked(fs.readFile).mockResolvedValue(JSON.stringify(config) as never);
      vi.mocked(fs.outputFile).mockResolvedValue(undefined as never);

      const service = new ConfigService();
      await service.removeWorkspace('beta');

      const savedConfig = JSON.parse(
        vi.mocked(fs.outputFile).mock.calls[0][1] as string
      ) as TwentyConfigFile;
      expect(savedConfig.workspaces?.beta).toBeUndefined();
      // Should pick one of the remaining workspaces as default
      expect(['alpha', 'gamma']).toContain(savedConfig.defaultWorkspace);
    });
  });
});
