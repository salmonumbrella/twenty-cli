import { describe, it, expect } from 'vitest';
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

function resolveConfig(): { token?: string; baseUrl?: string; workspace?: string } | null {
  const envToken = process.env.TWENTY_TOKEN;
  const envBaseUrl = process.env.TWENTY_BASE_URL;
  const envWorkspace = process.env.TWENTY_PROFILE;

  if (envToken) {
    return { token: envToken, baseUrl: envBaseUrl, workspace: envWorkspace };
  }

  const configPath = path.join(os.homedir(), '.twenty', 'config.json');
  if (!fs.existsSync(configPath)) {
    return null;
  }

  try {
    const raw = fs.readFileSync(configPath, 'utf-8');
    const parsed = JSON.parse(raw) as any;
    const workspace = envWorkspace ?? parsed.defaultWorkspace ?? 'default';
    const workspaceConfig = parsed.workspaces?.[workspace];
    if (!workspaceConfig?.apiKey) {
      return null;
    }
    return {
      token: workspaceConfig.apiKey,
      baseUrl: workspaceConfig.apiUrl,
      workspace,
    };
  } catch {
    return null;
  }
}

function resolveCliPath(): string {
  return path.resolve(__dirname, '../../../../dist/cli/cli.js');
}

const config = resolveConfig();
const cliPath = resolveCliPath();
const canRun = !!config && fs.existsSync(cliPath);

const describeIf = canRun ? describe : describe.skip;

describeIf('twenty api e2e', () => {
  it('lists people as json', () => {
    const env = {
      ...process.env,
      ...(config?.token ? { TWENTY_TOKEN: config.token } : {}),
      ...(config?.baseUrl ? { TWENTY_BASE_URL: config.baseUrl } : {}),
      ...(config?.workspace ? { TWENTY_PROFILE: config.workspace } : {}),
    };

    const output = execFileSync('node', [cliPath, 'api', 'people', 'list', '--limit', '1', '--output', 'json'], {
      env,
      encoding: 'utf-8',
    });

    const parsed = JSON.parse(output);
    expect(Array.isArray(parsed)).toBe(true);
  });
});

const describeMutations = process.env.TWENTY_E2E_MUTATION === 'true' ? describe : describe.skip;

describeMutations('twenty api e2e mutations', () => {
  it('creates and deletes a person', () => {
    if (!config) {
      throw new Error('Missing config');
    }

    const env = {
      ...process.env,
      ...(config.token ? { TWENTY_TOKEN: config.token } : {}),
      ...(config.baseUrl ? { TWENTY_BASE_URL: config.baseUrl } : {}),
      ...(config.workspace ? { TWENTY_PROFILE: config.workspace } : {}),
    };

    const createPayload = JSON.stringify({ name: { firstName: 'E2E', lastName: 'Test' } });
    const createdRaw = execFileSync('node', [cliPath, 'api', 'people', 'create', '--data', createPayload, '--output', 'json'], {
      env,
      encoding: 'utf-8',
    });
    const created = JSON.parse(createdRaw);
    expect(created.id).toBeTruthy();

    execFileSync('node', [cliPath, 'api', 'people', 'delete', created.id, '--force'], {
      env,
      encoding: 'utf-8',
    });
  });
});
