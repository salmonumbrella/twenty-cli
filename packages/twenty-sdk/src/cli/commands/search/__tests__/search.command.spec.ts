import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { Command } from 'commander';
import { registerSearchCommand } from '../search.command';
import { SearchService, SearchResult } from '../../../utilities/search/services/search.service';
import { ApiService } from '../../../utilities/api/services/api.service';

vi.mock('../../../utilities/search/services/search.service');
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

describe('search command', () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;
  let mockSearch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerSearchCommand(program);
    consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    mockSearch = vi.fn();
    vi.mocked(SearchService).mockImplementation(() => ({
      search: mockSearch,
    }) as unknown as SearchService);
  });

  afterEach(() => {
    consoleSpy.mockRestore();
    vi.clearAllMocks();
  });

  describe('command registration', () => {
    it('registers search command with correct name and description', () => {
      const searchCmd = program.commands.find(cmd => cmd.name() === 'search');
      expect(searchCmd).toBeDefined();
      expect(searchCmd?.description()).toBe('Full-text search across all records');
    });

    it('has required query argument', () => {
      const searchCmd = program.commands.find(cmd => cmd.name() === 'search');
      const args = searchCmd?.registeredArguments ?? [];
      expect(args.length).toBe(1);
      expect(args[0].name()).toBe('query');
      expect(args[0].required).toBe(true);
    });

    it('has --limit option with default value', () => {
      const searchCmd = program.commands.find(cmd => cmd.name() === 'search');
      const opts = searchCmd?.options ?? [];
      const limitOpt = opts.find(o => o.long === '--limit');
      expect(limitOpt).toBeDefined();
      expect(limitOpt?.defaultValue).toBe('20');
    });

    it('has --objects option', () => {
      const searchCmd = program.commands.find(cmd => cmd.name() === 'search');
      const opts = searchCmd?.options ?? [];
      const objectsOpt = opts.find(o => o.long === '--objects');
      expect(objectsOpt).toBeDefined();
    });

    it('has --exclude option', () => {
      const searchCmd = program.commands.find(cmd => cmd.name() === 'search');
      const opts = searchCmd?.options ?? [];
      const excludeOpt = opts.find(o => o.long === '--exclude');
      expect(excludeOpt).toBeDefined();
    });

    it('has global options applied', () => {
      const searchCmd = program.commands.find(cmd => cmd.name() === 'search');
      const opts = searchCmd?.options ?? [];
      const outputOpt = opts.find(o => o.long === '--output');
      const queryOpt = opts.find(o => o.long === '--query');
      const workspaceOpt = opts.find(o => o.long === '--workspace');
      expect(outputOpt).toBeDefined();
      expect(queryOpt).toBeDefined();
      expect(workspaceOpt).toBeDefined();
    });
  });

  describe('search execution', () => {
    it('performs search with default limit', async () => {
      const results: SearchResult[] = [
        { recordId: '1', objectNameSingular: 'person', record: { name: 'John' } },
      ];
      mockSearch.mockResolvedValue(results);

      await program.parseAsync(['node', 'test', 'search', 'john', '-o', 'json']);

      expect(mockSearch).toHaveBeenCalledWith({
        query: 'john',
        limit: 20,
        objects: undefined,
        excludeObjects: undefined,
      });
    });

    it('performs search with custom limit', async () => {
      mockSearch.mockResolvedValue([]);

      await program.parseAsync(['node', 'test', 'search', 'test', '--limit', '50', '-o', 'json']);

      expect(mockSearch).toHaveBeenCalledWith({
        query: 'test',
        limit: 50,
        objects: undefined,
        excludeObjects: undefined,
      });
    });

    it('filters by included objects', async () => {
      mockSearch.mockResolvedValue([]);

      await program.parseAsync(['node', 'test', 'search', 'query', '--objects', 'person,company', '-o', 'json']);

      expect(mockSearch).toHaveBeenCalledWith({
        query: 'query',
        limit: 20,
        objects: ['person', 'company'],
        excludeObjects: undefined,
      });
    });

    it('filters by excluded objects', async () => {
      mockSearch.mockResolvedValue([]);

      await program.parseAsync(['node', 'test', 'search', 'query', '--exclude', 'note,task', '-o', 'json']);

      expect(mockSearch).toHaveBeenCalledWith({
        query: 'query',
        limit: 20,
        objects: undefined,
        excludeObjects: ['note', 'task'],
      });
    });

    it('combines all options', async () => {
      mockSearch.mockResolvedValue([]);

      await program.parseAsync([
        'node', 'test', 'search', 'hello',
        '--limit', '10',
        '--objects', 'person',
        '--exclude', 'company',
        '-o', 'json',
      ]);

      expect(mockSearch).toHaveBeenCalledWith({
        query: 'hello',
        limit: 10,
        objects: ['person'],
        excludeObjects: ['company'],
      });
    });

    it('outputs search results', async () => {
      const results: SearchResult[] = [
        { recordId: 'rec-1', objectNameSingular: 'person', record: { name: 'Alice' } },
        { recordId: 'rec-2', objectNameSingular: 'company', record: { name: 'Acme' } },
      ];
      mockSearch.mockResolvedValue(results);

      await program.parseAsync(['node', 'test', 'search', 'a', '-o', 'json']);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toHaveLength(2);
      expect(parsed[0].recordId).toBe('rec-1');
      expect(parsed[1].recordId).toBe('rec-2');
    });

    it('handles empty results', async () => {
      mockSearch.mockResolvedValue([]);

      await program.parseAsync(['node', 'test', 'search', 'nonexistent', '-o', 'json']);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual([]);
    });
  });

  describe('error handling', () => {
    it('requires query argument', async () => {
      await expect(
        program.parseAsync(['node', 'test', 'search'])
      ).rejects.toThrow();
    });
  });
});
