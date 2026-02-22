import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { Command } from 'commander';
import { parseBody, parseArrayPayload } from '../body';
import { applyGlobalOptions, resolveGlobalOptions, GlobalOptions } from '../global-options';
import { readJsonInput, safeJsonParse, readFileOrStdin } from '../io';
import { createServices, CliServices } from '../services';

// Mock fs-extra
vi.mock('fs-extra', () => ({
  default: {
    readFile: vi.fn(),
  },
}));

// Clear fs mock between tests
beforeEach(async () => {
  const fs = await import('fs-extra');
  vi.mocked(fs.default.readFile).mockClear();
});

// Mock the service dependencies
vi.mock('../../api/services/api.service', () => ({
  ApiService: vi.fn().mockImplementation(() => ({
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  })),
}));

vi.mock('../../config/services/config.service', () => ({
  ConfigService: vi.fn().mockImplementation(() => ({
    getConfig: vi.fn(),
  })),
}));

vi.mock('../../records/services/records.service', () => ({
  RecordsService: vi.fn().mockImplementation(() => ({
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  })),
}));

vi.mock('../../metadata/services/metadata.service', () => ({
  MetadataService: vi.fn().mockImplementation(() => ({
    listObjects: vi.fn(),
    getObject: vi.fn(),
  })),
}));

vi.mock('../../output/services/output.service', () => ({
  OutputService: vi.fn().mockImplementation(() => ({
    render: vi.fn(),
  })),
}));

vi.mock('../../output/services/query.service', () => ({
  QueryService: vi.fn().mockImplementation(() => ({
    query: vi.fn(),
  })),
}));

vi.mock('../../output/services/table.service', () => ({
  TableService: vi.fn().mockImplementation(() => ({
    render: vi.fn(),
  })),
}));

vi.mock('../../file/services/export.service', () => ({
  ExportService: vi.fn().mockImplementation(() => ({
    export: vi.fn(),
  })),
}));

vi.mock('../../file/services/import.service', () => ({
  ImportService: vi.fn().mockImplementation(() => ({
    import: vi.fn(),
  })),
}));

describe('body utilities', () => {
  describe('parseBody', () => {
    it('parses JSON from --data string', async () => {
      const result = await parseBody('{"name":"Test","count":42}');
      expect(result).toEqual({ name: 'Test', count: 42 });
    });

    it('parses JSON from --file path', async () => {
      const fs = await import('fs-extra');
      vi.mocked(fs.default.readFile).mockResolvedValue('{"key":"value"}');

      const result = await parseBody(undefined, '/path/to/file.json');
      expect(result).toEqual({ key: 'value' });
      expect(fs.default.readFile).toHaveBeenCalledWith('/path/to/file.json', 'utf-8');
    });

    it('merges --set values into JSON payload', async () => {
      const result = await parseBody('{"existing":"value"}', undefined, ['name=Test', 'count=42']);
      expect(result).toEqual({ existing: 'value', name: 'Test', count: 42 });
    });

    it('creates object from --set values alone', async () => {
      const result = await parseBody(undefined, undefined, ['name=Test', 'count=42']);
      expect(result).toEqual({ name: 'Test', count: 42 });
    });

    it('handles nested --set values', async () => {
      const result = await parseBody(undefined, undefined, ['user.name=John', 'user.age=30']);
      expect(result).toEqual({ user: { name: 'John', age: 30 } });
    });

    it('throws error when payload is not an object', async () => {
      await expect(parseBody('["array"]')).rejects.toThrow('Payload must be a JSON object');
    });

    it('throws error when payload is a primitive', async () => {
      await expect(parseBody('"string"')).rejects.toThrow('Payload must be a JSON object');
    });

    it('throws error when no input is provided', async () => {
      await expect(parseBody()).rejects.toThrow('Missing JSON payload; use --data, --file, or --set');
    });

    it('throws error when data is empty string and no sets', async () => {
      await expect(parseBody('', undefined, [])).rejects.toThrow(
        'Missing JSON payload; use --data, --file, or --set',
      );
    });

    it('throws error when data is whitespace only', async () => {
      await expect(parseBody('   ', undefined, [])).rejects.toThrow(
        'Missing JSON payload; use --data, --file, or --set',
      );
    });

    it('prefers --data over --file when both provided', async () => {
      const fs = await import('fs-extra');
      vi.mocked(fs.default.readFile).mockResolvedValue('{"from":"file"}');

      const result = await parseBody('{"from":"data"}', '/path/to/file.json');
      expect(result).toEqual({ from: 'data' });
      expect(fs.default.readFile).not.toHaveBeenCalled();
    });
  });

  describe('parseArrayPayload', () => {
    it('parses JSON array from --data string', async () => {
      const result = await parseArrayPayload('[{"id":1},{"id":2}]');
      expect(result).toEqual([{ id: 1 }, { id: 2 }]);
    });

    it('parses JSON array from --file path', async () => {
      const fs = await import('fs-extra');
      vi.mocked(fs.default.readFile).mockResolvedValue('[1, 2, 3]');

      const result = await parseArrayPayload(undefined, '/path/to/file.json');
      expect(result).toEqual([1, 2, 3]);
    });

    it('throws error when payload is not an array', async () => {
      await expect(parseArrayPayload('{"not":"array"}')).rejects.toThrow(
        'Batch payload must be a JSON array',
      );
    });

    it('throws error when no input is provided', async () => {
      await expect(parseArrayPayload()).rejects.toThrow('Missing JSON payload; use --data or --file');
    });

    it('throws error when data is empty string', async () => {
      await expect(parseArrayPayload('')).rejects.toThrow('Missing JSON payload; use --data or --file');
    });
  });
});

describe('global-options utilities', () => {
  describe('applyGlobalOptions', () => {
    it('adds output flag to command', () => {
      const command = new Command('test');
      applyGlobalOptions(command);

      const outputOption = command.options.find((opt) => opt.long === '--output');
      expect(outputOption).toBeDefined();
      expect(outputOption?.short).toBe('-o');
    });

    it('adds query flag by default', () => {
      const command = new Command('test');
      applyGlobalOptions(command);

      const queryOption = command.options.find((opt) => opt.long === '--query');
      expect(queryOption).toBeDefined();
    });

    it('excludes query flag when includeQuery is false', () => {
      const command = new Command('test');
      applyGlobalOptions(command, { includeQuery: false });

      const queryOption = command.options.find((opt) => opt.long === '--query');
      expect(queryOption).toBeUndefined();
    });

    it('adds workspace flag to command', () => {
      const command = new Command('test');
      applyGlobalOptions(command);

      const workspaceOption = command.options.find((opt) => opt.long === '--workspace');
      expect(workspaceOption).toBeDefined();
    });

    it('adds debug flag to command', () => {
      const command = new Command('test');
      applyGlobalOptions(command);

      const debugOption = command.options.find((opt) => opt.long === '--debug');
      expect(debugOption).toBeDefined();
    });

    it('adds no-retry flag to command', () => {
      const command = new Command('test');
      applyGlobalOptions(command);

      const noRetryOption = command.options.find((opt) => opt.long === '--no-retry');
      expect(noRetryOption).toBeDefined();
    });
  });

  describe('resolveGlobalOptions', () => {
    let originalEnv: NodeJS.ProcessEnv;

    beforeEach(() => {
      originalEnv = { ...process.env };
      // Clear relevant env vars
      delete process.env.TWENTY_OUTPUT;
      delete process.env.TWENTY_QUERY;
      delete process.env.TWENTY_PROFILE;
      delete process.env.TWENTY_DEBUG;
      delete process.env.TWENTY_NO_RETRY;
    });

    afterEach(() => {
      process.env = originalEnv;
    });

    it('returns default output format as text', () => {
      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test']);

      const options = resolveGlobalOptions(command);
      expect(options.output).toBe('text');
    });

    it('reads output from command option', () => {
      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test', '--output', 'json']);

      const options = resolveGlobalOptions(command);
      expect(options.output).toBe('json');
    });

    it('reads output from TWENTY_OUTPUT env var', () => {
      process.env.TWENTY_OUTPUT = 'csv';

      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test']);

      const options = resolveGlobalOptions(command);
      expect(options.output).toBe('csv');
    });

    it('prefers command option over env var for output', () => {
      process.env.TWENTY_OUTPUT = 'csv';

      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test', '--output', 'json']);

      const options = resolveGlobalOptions(command);
      expect(options.output).toBe('json');
    });

    it('falls back to text for invalid output format', () => {
      process.env.TWENTY_OUTPUT = 'invalid';

      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test']);

      const options = resolveGlobalOptions(command);
      expect(options.output).toBe('text');
    });

    it('reads query from command option', () => {
      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test', '--query', '[0].name']);

      const options = resolveGlobalOptions(command);
      expect(options.query).toBe('[0].name');
    });

    it('reads query from TWENTY_QUERY env var', () => {
      process.env.TWENTY_QUERY = '[0].id';

      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test']);

      const options = resolveGlobalOptions(command);
      expect(options.query).toBe('[0].id');
    });

    it('uses outputQuery override when provided', () => {
      process.env.TWENTY_QUERY = 'env-query';

      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test', '--query', 'cmd-query']);

      const options = resolveGlobalOptions(command, { outputQuery: 'override-query' });
      expect(options.query).toBe('override-query');
    });

    it('reads workspace from command option', () => {
      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test', '--workspace', 'prod']);

      const options = resolveGlobalOptions(command);
      expect(options.workspace).toBe('prod');
    });

    it('reads workspace from TWENTY_PROFILE env var', () => {
      process.env.TWENTY_PROFILE = 'staging';

      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test']);

      const options = resolveGlobalOptions(command);
      expect(options.workspace).toBe('staging');
    });

    it('reads debug from command option', () => {
      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test', '--debug']);

      const options = resolveGlobalOptions(command);
      expect(options.debug).toBe(true);
    });

    it('reads debug from TWENTY_DEBUG env var', () => {
      process.env.TWENTY_DEBUG = 'true';

      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test']);

      const options = resolveGlobalOptions(command);
      expect(options.debug).toBe(true);
    });

    it('defaults debug to false', () => {
      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test']);

      const options = resolveGlobalOptions(command);
      expect(options.debug).toBe(false);
    });

    it('reads noRetry from TWENTY_NO_RETRY env var', () => {
      process.env.TWENTY_NO_RETRY = 'true';

      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test']);

      const options = resolveGlobalOptions(command);
      expect(options.noRetry).toBe(true);
    });

    it('reads noRetry from --no-retry flag', () => {
      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test', '--no-retry']);

      const options = resolveGlobalOptions(command);
      expect(options.noRetry).toBe(true);
    });

    it('defaults noRetry to false', () => {
      const command = new Command('test');
      applyGlobalOptions(command);
      command.parse(['node', 'test']);

      const options = resolveGlobalOptions(command);
      expect(options.noRetry).toBe(false);
    });
  });
});

describe('io utilities', () => {
  describe('safeJsonParse', () => {
    it('parses valid JSON object', () => {
      const result = safeJsonParse('{"key":"value"}');
      expect(result).toEqual({ key: 'value' });
    });

    it('parses valid JSON array', () => {
      const result = safeJsonParse('[1, 2, 3]');
      expect(result).toEqual([1, 2, 3]);
    });

    it('parses JSON primitives', () => {
      expect(safeJsonParse('"string"')).toBe('string');
      expect(safeJsonParse('42')).toBe(42);
      expect(safeJsonParse('true')).toBe(true);
      expect(safeJsonParse('null')).toBe(null);
    });

    it('throws on invalid JSON', () => {
      expect(() => safeJsonParse('{invalid}')).toThrow();
      expect(() => safeJsonParse('undefined')).toThrow();
    });
  });

  describe('readFileOrStdin', () => {
    it('reads from file path', async () => {
      const fs = await import('fs-extra');
      vi.mocked(fs.default.readFile).mockResolvedValue('file content');

      const result = await readFileOrStdin('/path/to/file.txt');
      expect(result).toBe('file content');
      expect(fs.default.readFile).toHaveBeenCalledWith('/path/to/file.txt', 'utf-8');
    });

    // Note: Testing stdin with path='-' requires mocking process.stdin
    // which is complex. The function delegates to readStdin() for that case.
  });

  describe('readJsonInput', () => {
    it('parses data string when provided', async () => {
      const result = await readJsonInput('{"key":"value"}');
      expect(result).toEqual({ key: 'value' });
    });

    it('reads and parses file when filePath provided', async () => {
      const fs = await import('fs-extra');
      vi.mocked(fs.default.readFile).mockResolvedValue('{"from":"file"}');

      const result = await readJsonInput(undefined, '/path/to/file.json');
      expect(result).toEqual({ from: 'file' });
    });

    it('returns undefined when no input provided', async () => {
      const result = await readJsonInput();
      expect(result).toBeUndefined();
    });

    it('returns undefined when data is empty string', async () => {
      const result = await readJsonInput('');
      expect(result).toBeUndefined();
    });

    it('returns undefined when data is whitespace only', async () => {
      const result = await readJsonInput('   ');
      expect(result).toBeUndefined();
    });

    it('returns undefined when file content is empty', async () => {
      const fs = await import('fs-extra');
      vi.mocked(fs.default.readFile).mockResolvedValue('');

      const result = await readJsonInput(undefined, '/path/to/empty.json');
      expect(result).toBeUndefined();
    });

    it('returns undefined when file content is whitespace only', async () => {
      const fs = await import('fs-extra');
      vi.mocked(fs.default.readFile).mockResolvedValue('   \n   ');

      const result = await readJsonInput(undefined, '/path/to/empty.json');
      expect(result).toBeUndefined();
    });

    it('prefers data over filePath when both provided', async () => {
      const fs = await import('fs-extra');
      vi.mocked(fs.default.readFile).mockResolvedValue('{"from":"file"}');

      const result = await readJsonInput('{"from":"data"}', '/path/to/file.json');
      expect(result).toEqual({ from: 'data' });
      expect(fs.default.readFile).not.toHaveBeenCalled();
    });

    it('throws on invalid JSON in data', async () => {
      await expect(readJsonInput('{invalid}')).rejects.toThrow();
    });

    it('throws on invalid JSON in file', async () => {
      const fs = await import('fs-extra');
      vi.mocked(fs.default.readFile).mockResolvedValue('{invalid json}');

      await expect(readJsonInput(undefined, '/path/to/file.json')).rejects.toThrow();
    });

    it('trims filePath before use', async () => {
      const fs = await import('fs-extra');
      vi.mocked(fs.default.readFile).mockResolvedValue('{"key":"value"}');

      const result = await readJsonInput(undefined, '  /path/to/file.json  ');
      expect(result).toEqual({ key: 'value' });
      expect(fs.default.readFile).toHaveBeenCalledWith('/path/to/file.json', 'utf-8');
    });
  });
});

describe('services factory', () => {
  describe('createServices', () => {
    it('returns CliServices object with all required services', () => {
      const globalOptions: GlobalOptions = {
        output: 'json',
        debug: false,
        noRetry: false,
      };

      const services = createServices(globalOptions);

      expect(services).toHaveProperty('api');
      expect(services).toHaveProperty('records');
      expect(services).toHaveProperty('metadata');
      expect(services).toHaveProperty('output');
      expect(services).toHaveProperty('importer');
      expect(services).toHaveProperty('exporter');
    });

    it('passes workspace option to ApiService', async () => {
      const { ApiService } = await import('../../api/services/api.service');

      const globalOptions: GlobalOptions = {
        workspace: 'production',
        debug: false,
        noRetry: false,
      };

      createServices(globalOptions);

      expect(ApiService).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({ workspace: 'production' }),
      );
    });

    it('passes debug option to ApiService', async () => {
      const { ApiService } = await import('../../api/services/api.service');

      const globalOptions: GlobalOptions = {
        debug: true,
        noRetry: false,
      };

      createServices(globalOptions);

      expect(ApiService).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({ debug: true }),
      );
    });

    it('passes noRetry option to ApiService', async () => {
      const { ApiService } = await import('../../api/services/api.service');

      const globalOptions: GlobalOptions = {
        debug: false,
        noRetry: true,
      };

      createServices(globalOptions);

      expect(ApiService).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({ noRetry: true }),
      );
    });

    it('creates services with default options', () => {
      const globalOptions: GlobalOptions = {};

      const services = createServices(globalOptions);

      expect(services.api).toBeDefined();
      expect(services.records).toBeDefined();
      expect(services.metadata).toBeDefined();
      expect(services.output).toBeDefined();
      expect(services.importer).toBeDefined();
      expect(services.exporter).toBeDefined();
    });

    it('creates new service instances on each call', () => {
      const globalOptions: GlobalOptions = {};

      const services1 = createServices(globalOptions);
      const services2 = createServices(globalOptions);

      // Each call should create new instances
      expect(services1.api).not.toBe(services2.api);
      expect(services1.records).not.toBe(services2.records);
    });
  });
});
