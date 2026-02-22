import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { Command } from 'commander';
import { registerRestCommand } from '../rest.command';

// Mock the services module
vi.mock('../../../utilities/shared/services', () => ({
  createServices: vi.fn(),
}));

// Mock the io utility
vi.mock('../../../utilities/shared/io', () => ({
  readJsonInput: vi.fn(),
}));

import { createServices } from '../../../utilities/shared/services';
import { readJsonInput } from '../../../utilities/shared/io';

function createMockServices() {
  return {
    api: {
      request: vi.fn().mockResolvedValue({ data: { id: 'test-id' } }),
    },
    output: {
      render: vi.fn(),
    },
  };
}

describe('rest command', () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;
  let mockServices: ReturnType<typeof createMockServices>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerRestCommand(program);
    consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

    mockServices = createMockServices();
    vi.mocked(createServices).mockReturnValue(mockServices as any);
    vi.mocked(readJsonInput).mockResolvedValue(undefined);
  });

  afterEach(() => {
    consoleSpy.mockRestore();
    vi.clearAllMocks();
  });

  describe('GET request', () => {
    it('makes GET request to specified path', async () => {
      await program.parseAsync(['node', 'test', 'rest', 'GET', '/people']);

      expect(mockServices.api.request).toHaveBeenCalledWith({
        method: 'get',
        url: '/people',
        params: undefined,
        data: undefined,
      });
      expect(mockServices.output.render).toHaveBeenCalled();
    });

    it('makes GET request with query params', async () => {
      await program.parseAsync([
        'node', 'test', 'rest', 'GET', '/people',
        '--param', 'limit=10',
        '--param', 'offset=20',
      ]);

      expect(mockServices.api.request).toHaveBeenCalledWith({
        method: 'get',
        url: '/people',
        params: { limit: ['10'], offset: ['20'] },
        data: undefined,
      });
    });

    it('adds leading slash to path if missing', async () => {
      await program.parseAsync(['node', 'test', 'rest', 'GET', 'people']);

      expect(mockServices.api.request).toHaveBeenCalledWith(
        expect.objectContaining({ url: '/people' })
      );
    });
  });

  describe('POST request', () => {
    it('makes POST request with JSON data', async () => {
      vi.mocked(readJsonInput).mockResolvedValue({ name: 'John Doe' });

      await program.parseAsync([
        'node', 'test', 'rest', 'POST', '/people',
        '--data', '{"name":"John Doe"}',
      ]);

      expect(mockServices.api.request).toHaveBeenCalledWith({
        method: 'post',
        url: '/people',
        params: undefined,
        data: { name: 'John Doe' },
      });
    });

    it('makes POST request with data from file', async () => {
      vi.mocked(readJsonInput).mockResolvedValue({ email: 'test@example.com' });

      await program.parseAsync([
        'node', 'test', 'rest', 'POST', '/people',
        '--file', '/path/to/data.json',
      ]);

      expect(readJsonInput).toHaveBeenCalledWith(undefined, '/path/to/data.json');
      expect(mockServices.api.request).toHaveBeenCalledWith({
        method: 'post',
        url: '/people',
        params: undefined,
        data: { email: 'test@example.com' },
      });
    });
  });

  describe('PATCH request', () => {
    it('makes PATCH request with JSON data', async () => {
      vi.mocked(readJsonInput).mockResolvedValue({ name: 'Updated Name' });

      await program.parseAsync([
        'node', 'test', 'rest', 'PATCH', '/people/123',
        '--data', '{"name":"Updated Name"}',
      ]);

      expect(mockServices.api.request).toHaveBeenCalledWith({
        method: 'patch',
        url: '/people/123',
        params: undefined,
        data: { name: 'Updated Name' },
      });
    });
  });

  describe('DELETE request', () => {
    it('makes DELETE request to specified path', async () => {
      await program.parseAsync(['node', 'test', 'rest', 'DELETE', '/people/123']);

      expect(mockServices.api.request).toHaveBeenCalledWith({
        method: 'delete',
        url: '/people/123',
        params: undefined,
        data: undefined,
      });
    });
  });

  describe('output format', () => {
    it('passes output format to render', async () => {
      await program.parseAsync([
        'node', 'test', 'rest', 'GET', '/people',
        '-o', 'json',
      ]);

      expect(mockServices.output.render).toHaveBeenCalledWith(
        { id: 'test-id' },
        expect.objectContaining({ format: 'json' })
      );
    });

    it('passes query filter to render', async () => {
      await program.parseAsync([
        'node', 'test', 'rest', 'GET', '/people',
        '--query', 'data[0]',
      ]);

      expect(mockServices.output.render).toHaveBeenCalledWith(
        { id: 'test-id' },
        expect.objectContaining({ query: 'data[0]' })
      );
    });
  });

  describe('error handling', () => {
    it('propagates API errors', async () => {
      mockServices.api.request.mockRejectedValue(new Error('API request failed'));

      await expect(
        program.parseAsync(['node', 'test', 'rest', 'GET', '/people'])
      ).rejects.toThrow('API request failed');
    });

    it('handles invalid JSON data gracefully', async () => {
      vi.mocked(readJsonInput).mockRejectedValue(new SyntaxError('Invalid JSON'));

      await expect(
        program.parseAsync([
          'node', 'test', 'rest', 'POST', '/people',
          '--data', 'invalid json',
        ])
      ).rejects.toThrow('Invalid JSON');
    });
  });

  describe('method case insensitivity', () => {
    it('converts method to lowercase', async () => {
      await program.parseAsync(['node', 'test', 'rest', 'Get', '/people']);

      expect(mockServices.api.request).toHaveBeenCalledWith(
        expect.objectContaining({ method: 'get' })
      );
    });

    it('handles uppercase method', async () => {
      await program.parseAsync(['node', 'test', 'rest', 'PUT', '/people/123']);

      expect(mockServices.api.request).toHaveBeenCalledWith(
        expect.objectContaining({ method: 'put' })
      );
    });
  });
});
