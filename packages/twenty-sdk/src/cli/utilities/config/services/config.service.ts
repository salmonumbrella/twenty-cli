import os from 'os';
import path from 'path';
import fs from 'fs-extra';
import { CliError } from '../../errors/cli-error';

export interface WorkspaceConfig {
  apiUrl?: string;
  apiKey?: string;
}

export interface TwentyConfigFile {
  workspaces?: Record<string, WorkspaceConfig>;
  defaultWorkspace?: string;
}

export interface ResolvedConfig {
  apiUrl: string;
  apiKey: string;
  workspace?: string;
}

export interface ConfigOverrides {
  workspace?: string;
  apiUrl?: string;
  apiKey?: string;
}

export class ConfigService {
  private configPath: string;

  constructor(configPath?: string) {
    this.configPath = configPath ?? path.join(os.homedir(), '.twenty', 'config.json');
  }

  async loadConfigFile(): Promise<TwentyConfigFile | null> {
    try {
      const exists = await fs.pathExists(this.configPath);
      if (!exists) return null;
      const content = await fs.readFile(this.configPath, 'utf-8');
      return JSON.parse(content) as TwentyConfigFile;
    } catch (error) {
      throw new CliError(
        `Failed to read config at ${this.configPath}`,
        'INVALID_ARGUMENTS',
        'Check the config file format or remove the file to recreate it.'
      );
    }
  }

  async getConfig(overrides?: ConfigOverrides): Promise<ResolvedConfig> {
    const fileConfig = await this.loadConfigFile();
    const workspace = overrides?.workspace
      ?? process.env.TWENTY_PROFILE
      ?? fileConfig?.defaultWorkspace
      ?? 'default';

    const workspaceConfig = fileConfig?.workspaces?.[workspace] ?? {};

    const apiUrl = overrides?.apiUrl
      ?? process.env.TWENTY_BASE_URL
      ?? workspaceConfig.apiUrl
      ?? 'https://api.twenty.com';

    const apiKey = overrides?.apiKey
      ?? process.env.TWENTY_TOKEN
      ?? workspaceConfig.apiKey
      ?? '';

    if (!apiKey) {
      throw new CliError(
        'Missing API token.',
        'AUTH',
        'Set TWENTY_TOKEN or configure ~/.twenty/config.json with an apiKey.'
      );
    }

    return {
      apiUrl,
      apiKey,
      workspace,
    };
  }
}
