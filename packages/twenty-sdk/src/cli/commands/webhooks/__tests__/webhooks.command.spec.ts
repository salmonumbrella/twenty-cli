import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { Command } from 'commander';
import { registerWebhooksCommand } from '../webhooks.command';
import { ApiService } from '../../../utilities/api/services/api.service';
import { CliError } from '../../../utilities/errors/cli-error';

vi.mock('../../../utilities/api/services/api.service');
vi.mock('../../../utilities/config/services/config.service', () => ({
  ConfigService: vi.fn().mockImplementation(() => ({
    getConfig: vi.fn().mockResolvedValue({
      apiUrl: 'https://api.twenty.com',
      apiKey: 'test-token',
      workspace: 'default',
    }),
  })),
}));

describe('webhooks command', () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;
  let mockPost: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerWebhooksCommand(program);
    consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    mockPost = vi.fn();
    vi.mocked(ApiService).mockImplementation(() => ({
      post: mockPost,
      get: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
      request: vi.fn(),
    }) as unknown as ApiService);
  });

  afterEach(() => {
    consoleSpy.mockRestore();
    vi.clearAllMocks();
  });

  describe('command registration', () => {
    it('registers webhooks command with correct name and description', () => {
      const webhooksCmd = program.commands.find(cmd => cmd.name() === 'webhooks');
      expect(webhooksCmd).toBeDefined();
      expect(webhooksCmd?.description()).toBe('Manage webhooks');
    });

    it('has required operation argument', () => {
      const webhooksCmd = program.commands.find(cmd => cmd.name() === 'webhooks');
      const args = webhooksCmd?.registeredArguments ?? [];
      expect(args.length).toBe(2);
      expect(args[0].name()).toBe('operation');
      expect(args[0].required).toBe(true);
    });

    it('has optional id argument', () => {
      const webhooksCmd = program.commands.find(cmd => cmd.name() === 'webhooks');
      const args = webhooksCmd?.registeredArguments ?? [];
      expect(args[1].name()).toBe('id');
      expect(args[1].required).toBe(false);
    });

    it('has --data option', () => {
      const webhooksCmd = program.commands.find(cmd => cmd.name() === 'webhooks');
      const opts = webhooksCmd?.options ?? [];
      const dataOpt = opts.find(o => o.long === '--data');
      expect(dataOpt).toBeDefined();
    });

    it('has --file option', () => {
      const webhooksCmd = program.commands.find(cmd => cmd.name() === 'webhooks');
      const opts = webhooksCmd?.options ?? [];
      const fileOpt = opts.find(o => o.long === '--file');
      expect(fileOpt).toBeDefined();
    });

    it('has global options applied', () => {
      const webhooksCmd = program.commands.find(cmd => cmd.name() === 'webhooks');
      const opts = webhooksCmd?.options ?? [];
      const outputOpt = opts.find(o => o.long === '--output');
      const queryOpt = opts.find(o => o.long === '--query');
      const workspaceOpt = opts.find(o => o.long === '--workspace');
      expect(outputOpt).toBeDefined();
      expect(queryOpt).toBeDefined();
      expect(workspaceOpt).toBeDefined();
    });
  });

  describe('list operation', () => {
    it('lists webhooks', async () => {
      const webhooks = [
        { id: 'wh-1', targetUrl: 'https://example.com/hook1', operations: ['create'], description: 'Test 1', createdAt: '2024-01-01' },
        { id: 'wh-2', targetUrl: 'https://example.com/hook2', operations: ['update'], description: 'Test 2', createdAt: '2024-01-02' },
      ];
      mockPost.mockResolvedValue({ data: { data: { webhooks } } });

      await program.parseAsync(['node', 'test', 'webhooks', 'list', '-o', 'json']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('webhooks'),
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toHaveLength(2);
      expect(parsed[0].id).toBe('wh-1');
    });

    it('handles empty webhooks list', async () => {
      mockPost.mockResolvedValue({ data: { data: { webhooks: [] } } });

      await program.parseAsync(['node', 'test', 'webhooks', 'list', '-o', 'json']);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual([]);
    });

    it('handles null webhooks response', async () => {
      mockPost.mockResolvedValue({ data: { data: { webhooks: null } } });

      await program.parseAsync(['node', 'test', 'webhooks', 'list', '-o', 'json']);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual([]);
    });
  });

  describe('get operation', () => {
    it('gets a single webhook by ID', async () => {
      const webhook = { id: 'wh-1', targetUrl: 'https://example.com/hook', operations: ['create'], description: 'Test', secret: 'secret123', createdAt: '2024-01-01', updatedAt: '2024-01-02' };
      mockPost.mockResolvedValue({ data: { data: { webhook } } });

      await program.parseAsync(['node', 'test', 'webhooks', 'get', 'wh-1', '-o', 'json']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('webhook(id: $id)'),
        variables: { id: 'wh-1' },
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.id).toBe('wh-1');
      expect(parsed.secret).toBe('secret123');
    });

    it('throws error when ID is missing', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'webhooks', 'get'])
      ).rejects.toThrow(CliError);
    });
  });

  describe('create operation', () => {
    it('creates a webhook with JSON data', async () => {
      const newWebhook = { id: 'wh-new', targetUrl: 'https://example.com/new', operations: ['create'], description: 'New webhook' };
      mockPost.mockResolvedValue({ data: { data: { createWebhook: newWebhook } } });
      const payload = { targetUrl: 'https://example.com/new', operations: ['create'] };

      await program.parseAsync(['node', 'test', 'webhooks', 'create', '-d', JSON.stringify(payload), '-o', 'json']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('createWebhook'),
        variables: { data: payload },
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.id).toBe('wh-new');
    });

    it('throws error when data is missing', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'webhooks', 'create'])
      ).rejects.toThrow('Missing JSON payload');
    });
  });

  describe('update operation', () => {
    it('updates a webhook with JSON data', async () => {
      const updatedWebhook = { id: 'wh-1', targetUrl: 'https://example.com/updated', operations: ['update'], description: 'Updated webhook' };
      mockPost.mockResolvedValue({ data: { data: { updateWebhook: updatedWebhook } } });
      const payload = { targetUrl: 'https://example.com/updated' };

      await program.parseAsync(['node', 'test', 'webhooks', 'update', 'wh-1', '-d', JSON.stringify(payload), '-o', 'json']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('updateWebhook'),
        variables: { id: 'wh-1', data: payload },
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.id).toBe('wh-1');
      expect(parsed.targetUrl).toBe('https://example.com/updated');
    });

    it('throws error when ID is missing', async () => {
      const payload = { targetUrl: 'https://example.com/updated' };
      await expect(
        program.parseAsync(['node', 'test', 'webhooks', 'update', '-d', JSON.stringify(payload)])
      ).rejects.toThrow(CliError);
    });

    it('throws error when data is missing', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'webhooks', 'update', 'wh-1'])
      ).rejects.toThrow('Missing JSON payload');
    });
  });

  describe('delete operation', () => {
    it('deletes a webhook by ID', async () => {
      mockPost.mockResolvedValue({ data: { data: { deleteWebhook: { id: 'wh-1' } } } });

      await program.parseAsync(['node', 'test', 'webhooks', 'delete', 'wh-1']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('deleteWebhook'),
        variables: { id: 'wh-1' },
      });
      expect(consoleSpy).toHaveBeenCalledWith('Webhook wh-1 deleted.');
    });

    it('throws error when ID is missing', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'webhooks', 'delete'])
      ).rejects.toThrow(CliError);
    });
  });

  describe('error handling', () => {
    it('requires operation argument', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'webhooks'])
      ).rejects.toThrow();
    });

    it('throws error for unknown operation', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'webhooks', 'unknown'])
      ).rejects.toThrow(CliError);
    });

    it('handles case insensitive operations', async () => {
      const webhooks = [{ id: 'wh-1', targetUrl: 'https://example.com', operations: [], description: '', createdAt: '' }];
      mockPost.mockResolvedValue({ data: { data: { webhooks } } });

      await program.parseAsync(['node', 'test', 'webhooks', 'LIST', '-o', 'json']);

      expect(mockPost).toHaveBeenCalled();
    });
  });
});
