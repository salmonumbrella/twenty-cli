import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { Command } from 'commander';
import { registerFilesCommand } from '../files.command';
import { ApiService } from '../../../utilities/api/services/api.service';
import { CliError } from '../../../utilities/errors/cli-error';
import fs from 'fs-extra';
import FormData from 'form-data';

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
vi.mock('fs-extra');
vi.mock('form-data');

describe('files command', () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;
  let mockPost: ReturnType<typeof vi.fn>;
  let mockGet: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerFilesCommand(program);
    consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    mockPost = vi.fn();
    mockGet = vi.fn();
    vi.mocked(ApiService).mockImplementation(() => ({
      post: mockPost,
      get: mockGet,
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
    it('registers files command with correct name and description', () => {
      const filesCmd = program.commands.find(cmd => cmd.name() === 'files');
      expect(filesCmd).toBeDefined();
      expect(filesCmd?.description()).toBe('Manage file attachments');
    });

    it('has required operation argument', () => {
      const filesCmd = program.commands.find(cmd => cmd.name() === 'files');
      const args = filesCmd?.registeredArguments ?? [];
      expect(args.length).toBe(2);
      expect(args[0].name()).toBe('operation');
      expect(args[0].required).toBe(true);
    });

    it('has optional path-or-id argument', () => {
      const filesCmd = program.commands.find(cmd => cmd.name() === 'files');
      const args = filesCmd?.registeredArguments ?? [];
      expect(args[1].name()).toBe('path-or-id');
      expect(args[1].required).toBe(false);
    });

    it('has --output-file option', () => {
      const filesCmd = program.commands.find(cmd => cmd.name() === 'files');
      const opts = filesCmd?.options ?? [];
      const outputFileOpt = opts.find(o => o.long === '--output-file');
      expect(outputFileOpt).toBeDefined();
    });

    it('has global options applied', () => {
      const filesCmd = program.commands.find(cmd => cmd.name() === 'files');
      const opts = filesCmd?.options ?? [];
      const outputOpt = opts.find(o => o.long === '--output');
      const queryOpt = opts.find(o => o.long === '--query');
      const workspaceOpt = opts.find(o => o.long === '--workspace');
      expect(outputOpt).toBeDefined();
      expect(queryOpt).toBeDefined();
      expect(workspaceOpt).toBeDefined();
    });
  });

  describe('upload operation', () => {
    it('uploads a file successfully', async () => {
      vi.mocked(fs.pathExists).mockResolvedValue(true as never);
      vi.mocked(fs.createReadStream).mockReturnValue({ pipe: vi.fn() } as never);

      const mockFormData = {
        append: vi.fn(),
        getHeaders: vi.fn().mockReturnValue({ 'content-type': 'multipart/form-data' }),
      };
      vi.mocked(FormData).mockImplementation(() => mockFormData as unknown as FormData);

      const uploadResponse = { id: 'file-123', name: 'test.txt', size: 1024 };
      mockPost.mockResolvedValue({ data: uploadResponse });

      await program.parseAsync(['node', 'test', 'files', 'upload', '/path/to/test.txt', '-o', 'json']);

      expect(fs.pathExists).toHaveBeenCalledWith('/path/to/test.txt');
      expect(fs.createReadStream).toHaveBeenCalledWith('/path/to/test.txt');
      expect(mockFormData.append).toHaveBeenCalledWith('file', expect.anything(), 'test.txt');
      expect(mockPost).toHaveBeenCalledWith('/files', mockFormData, {
        headers: { 'content-type': 'multipart/form-data' },
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.id).toBe('file-123');
    });

    it('throws error when file path is missing', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'files', 'upload'])
      ).rejects.toThrow(CliError);
    });

    it('throws error when file does not exist', async () => {
      vi.mocked(fs.pathExists).mockResolvedValue(false as never);

      await expect(
        program.parseAsync(['node', 'test', 'files', 'upload', '/nonexistent/file.txt'])
      ).rejects.toThrow('File not found: /nonexistent/file.txt');
    });
  });

  describe('download operation', () => {
    it('downloads a file with default output path', async () => {
      const fileContent = Buffer.from('file content');
      mockGet.mockResolvedValue({ data: fileContent });
      vi.mocked(fs.writeFile).mockResolvedValue(undefined as never);

      await program.parseAsync(['node', 'test', 'files', 'download', 'file-123']);

      expect(mockGet).toHaveBeenCalledWith('/files/file-123', { responseType: 'arraybuffer' });
      expect(fs.writeFile).toHaveBeenCalledWith('file-123', fileContent);
      expect(consoleSpy).toHaveBeenCalledWith('Downloaded to file-123');
    });

    it('downloads a file with custom output path', async () => {
      const fileContent = Buffer.from('file content');
      mockGet.mockResolvedValue({ data: fileContent });
      vi.mocked(fs.writeFile).mockResolvedValue(undefined as never);

      await program.parseAsync(['node', 'test', 'files', 'download', 'file-123', '--output-file', '/downloads/myfile.txt']);

      expect(mockGet).toHaveBeenCalledWith('/files/file-123', { responseType: 'arraybuffer' });
      expect(fs.writeFile).toHaveBeenCalledWith('/downloads/myfile.txt', fileContent);
      expect(consoleSpy).toHaveBeenCalledWith('Downloaded to /downloads/myfile.txt');
    });

    it('throws error when file ID is missing', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'files', 'download'])
      ).rejects.toThrow(CliError);
    });
  });

  describe('delete operation', () => {
    it('deletes a file by ID', async () => {
      mockPost.mockResolvedValue({ data: { data: { deleteFile: true } } });

      await program.parseAsync(['node', 'test', 'files', 'delete', 'file-123']);

      expect(mockPost).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('deleteFile'),
        variables: { id: 'file-123' },
      });
      expect(consoleSpy).toHaveBeenCalledWith('File file-123 deleted.');
    });

    it('throws error when file ID is missing', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'files', 'delete'])
      ).rejects.toThrow(CliError);
    });
  });

  describe('error handling', () => {
    it('requires operation argument', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'files'])
      ).rejects.toThrow();
    });

    it('throws error for unknown operation', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'files', 'unknown', 'some-id'])
      ).rejects.toThrow(CliError);
    });

    it('handles case insensitive operations', async () => {
      mockPost.mockResolvedValue({ data: { data: { deleteFile: true } } });

      await program.parseAsync(['node', 'test', 'files', 'DELETE', 'file-123']);

      expect(mockPost).toHaveBeenCalled();
      expect(consoleSpy).toHaveBeenCalledWith('File file-123 deleted.');
    });
  });
});
