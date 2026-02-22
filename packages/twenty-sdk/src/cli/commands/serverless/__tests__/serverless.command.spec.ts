import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { Command } from 'commander';
import { registerServerlessCommand } from '../serverless.command';
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

describe('serverless command', () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;
  let mockPost: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerServerlessCommand(program);
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
    it('registers serverless command with correct name and description', () => {
      const serverlessCmd = program.commands.find(cmd => cmd.name() === 'serverless');
      expect(serverlessCmd).toBeDefined();
      expect(serverlessCmd?.description()).toBe('Manage serverless functions');
    });

    it('has required operation argument', () => {
      const serverlessCmd = program.commands.find(cmd => cmd.name() === 'serverless');
      const args = serverlessCmd?.registeredArguments ?? [];
      expect(args.length).toBe(2);
      expect(args[0].name()).toBe('operation');
      expect(args[0].required).toBe(true);
    });

    it('has optional id argument', () => {
      const serverlessCmd = program.commands.find(cmd => cmd.name() === 'serverless');
      const args = serverlessCmd?.registeredArguments ?? [];
      expect(args[1].name()).toBe('id');
      expect(args[1].required).toBe(false);
    });

    it('has --data option', () => {
      const serverlessCmd = program.commands.find(cmd => cmd.name() === 'serverless');
      const opts = serverlessCmd?.options ?? [];
      const dataOpt = opts.find(o => o.long === '--data');
      expect(dataOpt).toBeDefined();
    });

    it('has --file option', () => {
      const serverlessCmd = program.commands.find(cmd => cmd.name() === 'serverless');
      const opts = serverlessCmd?.options ?? [];
      const fileOpt = opts.find(o => o.long === '--file');
      expect(fileOpt).toBeDefined();
    });

    it('has --name option', () => {
      const serverlessCmd = program.commands.find(cmd => cmd.name() === 'serverless');
      const opts = serverlessCmd?.options ?? [];
      const nameOpt = opts.find(o => o.long === '--name');
      expect(nameOpt).toBeDefined();
    });

    it('has --description option', () => {
      const serverlessCmd = program.commands.find(cmd => cmd.name() === 'serverless');
      const opts = serverlessCmd?.options ?? [];
      const descOpt = opts.find(o => o.long === '--description');
      expect(descOpt).toBeDefined();
    });

    it('has global options applied', () => {
      const serverlessCmd = program.commands.find(cmd => cmd.name() === 'serverless');
      const opts = serverlessCmd?.options ?? [];
      const outputOpt = opts.find(o => o.long === '--output');
      const queryOpt = opts.find(o => o.long === '--query');
      const workspaceOpt = opts.find(o => o.long === '--workspace');
      expect(outputOpt).toBeDefined();
      expect(queryOpt).toBeDefined();
      expect(workspaceOpt).toBeDefined();
    });
  });

  describe('list operation', () => {
    it('lists serverless functions', async () => {
      const functions = [
        { id: 'fn-1', name: 'Function 1', description: 'Test 1', syncStatus: 'READY', createdAt: '2024-01-01', updatedAt: '2024-01-02' },
        { id: 'fn-2', name: 'Function 2', description: 'Test 2', syncStatus: 'PENDING', createdAt: '2024-01-03', updatedAt: '2024-01-04' },
      ];
      mockPost.mockResolvedValue({ data: { data: { findManyServerlessFunctions: functions } } });

      await program.parseAsync(['node', 'test', 'serverless', 'list', '-o', 'json']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('findManyServerlessFunctions'),
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toHaveLength(2);
      expect(parsed[0].id).toBe('fn-1');
    });

    it('handles empty functions list', async () => {
      mockPost.mockResolvedValue({ data: { data: { findManyServerlessFunctions: [] } } });

      await program.parseAsync(['node', 'test', 'serverless', 'list', '-o', 'json']);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual([]);
    });

    it('handles null functions response', async () => {
      mockPost.mockResolvedValue({ data: { data: { findManyServerlessFunctions: null } } });

      await program.parseAsync(['node', 'test', 'serverless', 'list', '-o', 'json']);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual([]);
    });
  });

  describe('get operation', () => {
    it('gets a single function by ID', async () => {
      const func = { id: 'fn-1', name: 'Function 1', description: 'Test', syncStatus: 'READY', sourceCodeFullPath: '/path/to/code', createdAt: '2024-01-01', updatedAt: '2024-01-02' };
      mockPost.mockResolvedValue({ data: { data: { findOneServerlessFunction: func } } });

      await program.parseAsync(['node', 'test', 'serverless', 'get', 'fn-1', '-o', 'json']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('findOneServerlessFunction(id: $id)'),
        variables: { id: 'fn-1' },
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.id).toBe('fn-1');
      expect(parsed.sourceCodeFullPath).toBe('/path/to/code');
    });

    it('throws error when ID is missing', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'serverless', 'get'])
      ).rejects.toThrow(CliError);
    });
  });

  describe('create operation', () => {
    it('creates a function with --name option', async () => {
      const newFunc = { id: 'fn-new', name: 'NewFunction', description: 'New function' };
      mockPost.mockResolvedValue({ data: { data: { createOneServerlessFunction: newFunc } } });

      await program.parseAsync(['node', 'test', 'serverless', 'create', '--name', 'NewFunction', '--description', 'New function', '-o', 'json']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('createOneServerlessFunction'),
        variables: { name: 'NewFunction', description: 'New function' },
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.id).toBe('fn-new');
    });

    it('creates a function without description', async () => {
      const newFunc = { id: 'fn-new', name: 'NewFunction', description: null };
      mockPost.mockResolvedValue({ data: { data: { createOneServerlessFunction: newFunc } } });

      await program.parseAsync(['node', 'test', 'serverless', 'create', '--name', 'NewFunction', '-o', 'json']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('createOneServerlessFunction'),
        variables: { name: 'NewFunction', description: undefined },
      });
    });

    it('throws error when --name is missing', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'serverless', 'create'])
      ).rejects.toThrow(CliError);
    });
  });

  describe('update operation', () => {
    it('updates a function with JSON data', async () => {
      const updatedFunc = { id: 'fn-1', name: 'UpdatedFunction', description: 'Updated description' };
      mockPost.mockResolvedValue({ data: { data: { updateOneServerlessFunction: updatedFunc } } });
      const payload = { name: 'UpdatedFunction', description: 'Updated description' };

      await program.parseAsync(['node', 'test', 'serverless', 'update', 'fn-1', '-d', JSON.stringify(payload), '-o', 'json']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('updateOneServerlessFunction'),
        variables: { id: 'fn-1', ...payload },
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.id).toBe('fn-1');
      expect(parsed.name).toBe('UpdatedFunction');
    });

    it('updates a function without data', async () => {
      const updatedFunc = { id: 'fn-1', name: 'Function', description: 'Desc' };
      mockPost.mockResolvedValue({ data: { data: { updateOneServerlessFunction: updatedFunc } } });

      await program.parseAsync(['node', 'test', 'serverless', 'update', 'fn-1', '-o', 'json']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('updateOneServerlessFunction'),
        variables: { id: 'fn-1' },
      });
    });

    it('throws error when ID is missing', async () => {
      const payload = { name: 'UpdatedFunction' };
      await expect(
        program.parseAsync(['node', 'test', 'serverless', 'update', '-d', JSON.stringify(payload)])
      ).rejects.toThrow(CliError);
    });
  });

  describe('delete operation', () => {
    it('deletes a function by ID', async () => {
      mockPost.mockResolvedValue({ data: { data: { deleteOneServerlessFunction: { id: 'fn-1' } } } });

      await program.parseAsync(['node', 'test', 'serverless', 'delete', 'fn-1']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('deleteOneServerlessFunction'),
        variables: { id: 'fn-1' },
      });
      expect(consoleSpy).toHaveBeenCalledWith('Serverless function fn-1 deleted.');
    });

    it('throws error when ID is missing', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'serverless', 'delete'])
      ).rejects.toThrow(CliError);
    });
  });

  describe('execute operation', () => {
    it('executes a function with payload', async () => {
      const result = { data: { result: 'success' }, status: 'OK', duration: 150 };
      mockPost.mockResolvedValue({ data: { data: { executeOneServerlessFunction: result } } });
      const payload = { input: 'test data' };

      await program.parseAsync(['node', 'test', 'serverless', 'execute', 'fn-1', '-d', JSON.stringify(payload), '-o', 'json']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('executeOneServerlessFunction'),
        variables: { id: 'fn-1', payload },
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.status).toBe('OK');
      expect(parsed.duration).toBe(150);
    });

    it('executes a function without payload', async () => {
      const result = { data: null, status: 'OK', duration: 50 };
      mockPost.mockResolvedValue({ data: { data: { executeOneServerlessFunction: result } } });

      await program.parseAsync(['node', 'test', 'serverless', 'execute', 'fn-1', '-o', 'json']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('executeOneServerlessFunction'),
        variables: { id: 'fn-1', payload: {} },
      });
    });

    it('throws error when ID is missing', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'serverless', 'execute'])
      ).rejects.toThrow(CliError);
    });
  });

  describe('publish operation', () => {
    it('publishes a function by ID', async () => {
      const result = { id: 'fn-1', syncStatus: 'READY' };
      mockPost.mockResolvedValue({ data: { data: { publishServerlessFunction: result } } });

      await program.parseAsync(['node', 'test', 'serverless', 'publish', 'fn-1', '-o', 'json']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('publishServerlessFunction'),
        variables: { id: 'fn-1' },
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.id).toBe('fn-1');
      expect(parsed.syncStatus).toBe('READY');
    });

    it('throws error when ID is missing', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'serverless', 'publish'])
      ).rejects.toThrow(CliError);
    });
  });

  describe('source operation', () => {
    it('gets source code for a function', async () => {
      const sourceCode = 'export default function handler() { return "Hello"; }';
      mockPost.mockResolvedValue({ data: { data: { getServerlessFunctionSourceCode: sourceCode } } });

      await program.parseAsync(['node', 'test', 'serverless', 'source', 'fn-1']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('getServerlessFunctionSourceCode'),
        variables: { id: 'fn-1' },
      });
      expect(consoleSpy).toHaveBeenCalledWith(sourceCode);
    });

    it('handles null source code', async () => {
      mockPost.mockResolvedValue({ data: { data: { getServerlessFunctionSourceCode: null } } });

      await program.parseAsync(['node', 'test', 'serverless', 'source', 'fn-1']);

      expect(consoleSpy).toHaveBeenCalledWith('');
    });

    it('throws error when ID is missing', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'serverless', 'source'])
      ).rejects.toThrow(CliError);
    });
  });

  describe('error handling', () => {
    it('requires operation argument', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'serverless'])
      ).rejects.toThrow();
    });

    it('throws error for unknown operation', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'serverless', 'unknown'])
      ).rejects.toThrow(CliError);
    });

    it('handles case insensitive operations', async () => {
      const functions = [{ id: 'fn-1', name: 'Function', description: '', syncStatus: 'READY', createdAt: '', updatedAt: '' }];
      mockPost.mockResolvedValue({ data: { data: { findManyServerlessFunctions: functions } } });

      await program.parseAsync(['node', 'test', 'serverless', 'LIST', '-o', 'json']);

      expect(mockPost).toHaveBeenCalled();
    });
  });
});
