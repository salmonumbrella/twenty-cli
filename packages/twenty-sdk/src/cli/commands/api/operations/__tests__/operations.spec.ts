import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { runCreateOperation } from '../create.operation';
import { runUpdateOperation } from '../update.operation';
import { runDeleteOperation } from '../delete.operation';
import { runGetOperation } from '../get.operation';
import { runListOperation } from '../list.operation';
import { runDestroyOperation } from '../destroy.operation';
import { runRestoreOperation } from '../restore.operation';
import { runImportOperation } from '../import.operation';
import { runExportOperation } from '../export.operation';
import { runMergeOperation } from '../merge.operation';
import { runBatchCreateOperation } from '../batch-create.operation';
import { runBatchUpdateOperation } from '../batch-update.operation';
import { runBatchDeleteOperation } from '../batch-delete.operation';
import { CliError } from '../../../../utilities/errors/cli-error';
import { ApiOperationContext } from '../types';

// Mock the body utility
vi.mock('../../../../utilities/shared/body', () => ({
  parseBody: vi.fn().mockImplementation(async (data: string | undefined) => {
    if (data) return JSON.parse(data);
    return {};
  }),
  parseArrayPayload: vi.fn().mockImplementation(async (data: string | undefined) => {
    if (data) return JSON.parse(data);
    return [];
  }),
}));

// Mock the io utility
vi.mock('../../../../utilities/shared/io', () => ({
  readJsonInput: vi.fn().mockImplementation(async (data: string | undefined) => {
    if (data) return JSON.parse(data);
    return undefined;
  }),
}));

function createMockContext(overrides: Partial<ApiOperationContext> = {}): ApiOperationContext {
  return {
    object: 'people',
    arg: undefined,
    options: {},
    globalOptions: { output: 'json' },
    services: {
      api: {} as any,
      records: {
        create: vi.fn().mockResolvedValue({ id: 'test-id', name: 'Test' }),
        update: vi.fn().mockResolvedValue({ id: 'test-id', name: 'Updated' }),
        delete: vi.fn().mockResolvedValue(undefined),
        get: vi.fn().mockResolvedValue({ id: 'test-id', name: 'Test' }),
        list: vi.fn().mockResolvedValue({ data: [{ id: '1' }, { id: '2' }] }),
        listAll: vi.fn().mockResolvedValue({ data: [{ id: '1' }, { id: '2' }, { id: '3' }] }),
        destroy: vi.fn().mockResolvedValue(undefined),
        restore: vi.fn().mockResolvedValue({ id: 'test-id', deletedAt: null }),
        merge: vi.fn().mockResolvedValue({ id: 'merged-id' }),
        batchCreate: vi.fn().mockResolvedValue([{ id: '1' }, { id: '2' }]),
        batchUpdate: vi.fn().mockResolvedValue([{ id: '1' }, { id: '2' }]),
        batchDelete: vi.fn().mockResolvedValue({ deleted: 2 }),
        findDuplicates: vi.fn(),
        groupBy: vi.fn(),
      } as any,
      metadata: {} as any,
      output: {
        render: vi.fn(),
      } as any,
      importer: {
        import: vi.fn().mockResolvedValue([{ name: 'Test1' }, { name: 'Test2' }]),
      } as any,
      exporter: {
        export: vi.fn(),
      } as any,
    },
    ...overrides,
  } as ApiOperationContext;
}

describe('API Operations', () => {
  let consoleSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
  });

  afterEach(() => {
    consoleSpy.mockRestore();
    vi.clearAllMocks();
  });

  // ==================== CREATE OPERATION ====================
  describe('runCreateOperation', () => {
    it('creates record with --data and renders output', async () => {
      const ctx = createMockContext({
        options: { data: '{"name":"Test Person"}' },
      });

      await runCreateOperation(ctx);

      expect(ctx.services.records.create).toHaveBeenCalledWith('people', { name: 'Test Person' });
      expect(ctx.services.output.render).toHaveBeenCalledWith(
        { id: 'test-id', name: 'Test' },
        { format: 'json', query: undefined }
      );
    });

    it('propagates error when create fails', async () => {
      const ctx = createMockContext({
        options: { data: '{"name":"Test"}' },
      });
      (ctx.services.records.create as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('API error')
      );

      await expect(runCreateOperation(ctx)).rejects.toThrow('API error');
    });
  });

  // ==================== UPDATE OPERATION ====================
  describe('runUpdateOperation', () => {
    it('updates record with ID and renders output', async () => {
      const ctx = createMockContext({
        arg: 'record-123',
        options: { data: '{"name":"Updated Name"}' },
      });

      await runUpdateOperation(ctx);

      expect(ctx.services.records.update).toHaveBeenCalledWith('people', 'record-123', {
        name: 'Updated Name',
      });
      expect(ctx.services.output.render).toHaveBeenCalled();
    });

    it('throws CliError when ID is missing', async () => {
      const ctx = createMockContext({
        arg: undefined,
        options: { data: '{"name":"Test"}' },
      });

      await expect(runUpdateOperation(ctx)).rejects.toThrow(CliError);
      await expect(runUpdateOperation(ctx)).rejects.toThrow('Missing record ID');
    });
  });

  // ==================== DELETE OPERATION ====================
  describe('runDeleteOperation', () => {
    it('deletes record when --force is provided', async () => {
      const ctx = createMockContext({
        arg: 'record-123',
        options: { force: true },
      });

      await runDeleteOperation(ctx);

      expect(ctx.services.records.delete).toHaveBeenCalledWith('people', 'record-123');
      expect(consoleSpy).toHaveBeenCalledWith('Deleted people record-123');
    });

    it('prints confirmation message without --force', async () => {
      const ctx = createMockContext({
        arg: 'record-123',
        options: { force: false },
      });

      await runDeleteOperation(ctx);

      expect(ctx.services.records.delete).not.toHaveBeenCalled();
      expect(consoleSpy).toHaveBeenCalledWith(
        'About to delete people record-123. Use --force to confirm.'
      );
    });

    it('throws CliError when ID is missing', async () => {
      const ctx = createMockContext({
        arg: undefined,
        options: { force: true },
      });

      await expect(runDeleteOperation(ctx)).rejects.toThrow(CliError);
      await expect(runDeleteOperation(ctx)).rejects.toThrow('Missing record ID');
    });

    it('renders response if delete returns data', async () => {
      const ctx = createMockContext({
        arg: 'record-123',
        options: { force: true },
      });
      (ctx.services.records.delete as ReturnType<typeof vi.fn>).mockResolvedValue({
        id: 'record-123',
        deleted: true,
      });

      await runDeleteOperation(ctx);

      expect(ctx.services.output.render).toHaveBeenCalledWith(
        { id: 'record-123', deleted: true },
        expect.any(Object)
      );
    });
  });

  // ==================== GET OPERATION ====================
  describe('runGetOperation', () => {
    it('gets record by ID and renders output', async () => {
      const ctx = createMockContext({
        arg: 'record-123',
        options: { include: 'company,notes' },
      });

      await runGetOperation(ctx);

      expect(ctx.services.records.get).toHaveBeenCalledWith('people', 'record-123', {
        include: 'company,notes',
      });
      expect(ctx.services.output.render).toHaveBeenCalled();
    });

    it('throws CliError when ID is missing', async () => {
      const ctx = createMockContext({
        arg: undefined,
      });

      await expect(runGetOperation(ctx)).rejects.toThrow(CliError);
      await expect(runGetOperation(ctx)).rejects.toThrow('Missing record ID');
    });
  });

  // ==================== LIST OPERATION ====================
  describe('runListOperation', () => {
    it('lists records with pagination options', async () => {
      const ctx = createMockContext({
        options: {
          limit: '10',
          cursor: 'abc123',
          filter: 'name[eq]=Test',
          sort: 'createdAt',
          order: 'desc',
        },
      });

      await runListOperation(ctx);

      expect(ctx.services.records.list).toHaveBeenCalledWith('people', {
        limit: 10,
        cursor: 'abc123',
        filter: 'name[eq]=Test',
        include: undefined,
        sort: 'createdAt',
        order: 'desc',
        fields: undefined,
        params: {},
      });
      expect(ctx.services.output.render).toHaveBeenCalled();
    });

    it('uses listAll when --all is provided', async () => {
      const ctx = createMockContext({
        options: { all: true },
      });

      await runListOperation(ctx);

      expect(ctx.services.records.listAll).toHaveBeenCalled();
      expect(ctx.services.records.list).not.toHaveBeenCalled();
    });

    it('parses key-value params correctly', async () => {
      const ctx = createMockContext({
        options: {
          param: ['depth=2', 'include_deleted=true'],
        },
      });

      await runListOperation(ctx);

      expect(ctx.services.records.list).toHaveBeenCalledWith(
        'people',
        expect.objectContaining({
          params: { depth: ['2'], include_deleted: ['true'] },
        })
      );
    });
  });

  // ==================== DESTROY OPERATION ====================
  describe('runDestroyOperation', () => {
    it('destroys record when --force is provided', async () => {
      const ctx = createMockContext({
        arg: 'record-123',
        options: { force: true },
      });

      await runDestroyOperation(ctx);

      expect(ctx.services.records.destroy).toHaveBeenCalledWith('people', 'record-123');
      expect(consoleSpy).toHaveBeenCalledWith('Destroyed people record-123');
    });

    it('prints confirmation message without --force', async () => {
      const ctx = createMockContext({
        arg: 'record-123',
        options: { force: false },
      });

      await runDestroyOperation(ctx);

      expect(ctx.services.records.destroy).not.toHaveBeenCalled();
      expect(consoleSpy).toHaveBeenCalledWith(
        'About to destroy people record-123. Use --force to confirm.'
      );
    });

    it('throws CliError when ID is missing', async () => {
      const ctx = createMockContext({
        arg: undefined,
        options: { force: true },
      });

      await expect(runDestroyOperation(ctx)).rejects.toThrow(CliError);
    });
  });

  // ==================== RESTORE OPERATION ====================
  describe('runRestoreOperation', () => {
    it('restores record by ID and renders output', async () => {
      const ctx = createMockContext({
        arg: 'record-123',
      });

      await runRestoreOperation(ctx);

      expect(ctx.services.records.restore).toHaveBeenCalledWith('people', 'record-123');
      expect(ctx.services.output.render).toHaveBeenCalled();
    });

    it('throws CliError when ID is missing', async () => {
      const ctx = createMockContext({
        arg: undefined,
      });

      await expect(runRestoreOperation(ctx)).rejects.toThrow(CliError);
      await expect(runRestoreOperation(ctx)).rejects.toThrow('Missing record ID');
    });
  });

  // ==================== IMPORT OPERATION ====================
  describe('runImportOperation', () => {
    it('imports records from file and batch creates them', async () => {
      const ctx = createMockContext({
        arg: '/path/to/data.csv',
        options: {},
      });

      await runImportOperation(ctx);

      expect(ctx.services.importer.import).toHaveBeenCalledWith('/path/to/data.csv', {
        dryRun: undefined,
      });
      expect(ctx.services.records.batchCreate).toHaveBeenCalledWith('people', [
        { name: 'Test1' },
        { name: 'Test2' },
      ]);
      expect(consoleSpy).toHaveBeenCalledWith('Import complete: 2 imported.');
    });

    it('throws CliError when file path is missing', async () => {
      const ctx = createMockContext({
        arg: undefined,
      });

      await expect(runImportOperation(ctx)).rejects.toThrow(CliError);
      await expect(runImportOperation(ctx)).rejects.toThrow('Missing import file path');
    });

    it('does not create records in dry-run mode', async () => {
      const ctx = createMockContext({
        arg: '/path/to/data.csv',
        options: { dryRun: true },
      });

      await runImportOperation(ctx);

      expect(ctx.services.importer.import).toHaveBeenCalledWith('/path/to/data.csv', {
        dryRun: true,
      });
      expect(ctx.services.records.batchCreate).not.toHaveBeenCalled();
    });

    it('handles empty import gracefully', async () => {
      const ctx = createMockContext({
        arg: '/path/to/empty.csv',
      });
      (ctx.services.importer.import as ReturnType<typeof vi.fn>).mockResolvedValue([]);

      await runImportOperation(ctx);

      expect(consoleSpy).toHaveBeenCalledWith('No records to import.');
      expect(ctx.services.records.batchCreate).not.toHaveBeenCalled();
    });

    it('continues on error when --continue-on-error is set', async () => {
      const ctx = createMockContext({
        arg: '/path/to/data.csv',
        options: { continueOnError: true, batchSize: '1' },
      });
      (ctx.services.importer.import as ReturnType<typeof vi.fn>).mockResolvedValue([
        { name: 'Test1' },
        { name: 'Test2' },
      ]);
      (ctx.services.records.batchCreate as ReturnType<typeof vi.fn>)
        .mockRejectedValueOnce(new Error('Batch 1 failed'))
        .mockResolvedValueOnce([{ id: '2' }]);

      await runImportOperation(ctx);

      expect(consoleSpy).toHaveBeenCalledWith('Import complete: 1 imported, 1 failed.');
    });

    it('caps batch size at 60', async () => {
      const ctx = createMockContext({
        arg: '/path/to/data.csv',
        options: { batchSize: '100' },
      });
      const manyRecords = Array.from({ length: 120 }, (_, i) => ({ name: `Test${i}` }));
      (ctx.services.importer.import as ReturnType<typeof vi.fn>).mockResolvedValue(manyRecords);

      await runImportOperation(ctx);

      // Should have been called twice with batches of 60
      expect(ctx.services.records.batchCreate).toHaveBeenCalledTimes(2);
      expect((ctx.services.records.batchCreate as ReturnType<typeof vi.fn>).mock.calls[0][1]).toHaveLength(60);
      expect((ctx.services.records.batchCreate as ReturnType<typeof vi.fn>).mock.calls[1][1]).toHaveLength(60);
    });
  });

  // ==================== EXPORT OPERATION ====================
  describe('runExportOperation', () => {
    it('exports records to JSON format', async () => {
      const ctx = createMockContext({
        options: { format: 'json', outputFile: '/path/to/output.json' },
      });

      await runExportOperation(ctx);

      expect(ctx.services.records.list).toHaveBeenCalled();
      expect(ctx.services.exporter.export).toHaveBeenCalledWith(
        [{ id: '1' }, { id: '2' }],
        { format: 'json', output: '/path/to/output.json' }
      );
    });

    it('exports records to CSV format', async () => {
      const ctx = createMockContext({
        options: { format: 'csv' },
      });

      await runExportOperation(ctx);

      expect(ctx.services.exporter.export).toHaveBeenCalledWith(expect.any(Array), {
        format: 'csv',
        output: undefined,
      });
    });

    it('throws CliError for unsupported format', async () => {
      const ctx = createMockContext({
        options: { format: 'xml' },
      });

      await expect(runExportOperation(ctx)).rejects.toThrow(CliError);
      await expect(runExportOperation(ctx)).rejects.toThrow('Unsupported export format');
    });

    it('uses listAll when --all is provided', async () => {
      const ctx = createMockContext({
        options: { all: true, format: 'json' },
      });

      await runExportOperation(ctx);

      expect(ctx.services.records.listAll).toHaveBeenCalled();
      expect(ctx.services.records.list).not.toHaveBeenCalled();
    });

    it('uses --output as file path when not a format', async () => {
      const ctx = createMockContext({
        options: { output: '/path/to/file.json', format: 'json' },
      });

      await runExportOperation(ctx);

      expect(ctx.services.exporter.export).toHaveBeenCalledWith(expect.any(Array), {
        format: 'json',
        output: '/path/to/file.json',
      });
    });
  });

  // ==================== MERGE OPERATION ====================
  describe('runMergeOperation', () => {
    it('merges records using --source and --target', async () => {
      const ctx = createMockContext({
        options: { source: 'id-1', target: 'id-2' },
      });

      await runMergeOperation(ctx);

      expect(ctx.services.records.merge).toHaveBeenCalledWith('people', {
        ids: ['id-1', 'id-2'],
        conflictPriorityIndex: 0,
      });
      expect(ctx.services.output.render).toHaveBeenCalled();
    });

    it('merges records using --ids', async () => {
      const ctx = createMockContext({
        options: { ids: 'id-1,id-2,id-3', priority: '1' },
      });

      await runMergeOperation(ctx);

      expect(ctx.services.records.merge).toHaveBeenCalledWith('people', {
        ids: ['id-1', 'id-2', 'id-3'],
        conflictPriorityIndex: 1,
      });
    });

    it('merges records using --data JSON payload', async () => {
      const ctx = createMockContext({
        options: { data: '{"ids":["a","b"],"conflictPriorityIndex":0}' },
      });

      await runMergeOperation(ctx);

      expect(ctx.services.records.merge).toHaveBeenCalledWith('people', {
        ids: ['a', 'b'],
        conflictPriorityIndex: 0,
      });
    });

    it('throws CliError when only --source is provided', async () => {
      const ctx = createMockContext({
        options: { source: 'id-1' },
      });

      await expect(runMergeOperation(ctx)).rejects.toThrow(CliError);
      await expect(runMergeOperation(ctx)).rejects.toThrow('Both --source and --target are required');
    });

    it('throws CliError when no payload is provided', async () => {
      const ctx = createMockContext({
        options: {},
      });

      await expect(runMergeOperation(ctx)).rejects.toThrow(CliError);
      await expect(runMergeOperation(ctx)).rejects.toThrow('Missing payload');
    });

    it('adds dryRun flag when --dry-run is provided', async () => {
      const ctx = createMockContext({
        options: { ids: 'id-1,id-2', dryRun: true },
      });

      await runMergeOperation(ctx);

      expect(ctx.services.records.merge).toHaveBeenCalledWith('people', {
        ids: ['id-1', 'id-2'],
        conflictPriorityIndex: 0,
        dryRun: true,
      });
    });
  });

  // ==================== BATCH CREATE OPERATION ====================
  describe('runBatchCreateOperation', () => {
    it('batch creates records from --data JSON array', async () => {
      const ctx = createMockContext({
        options: { data: '[{"name":"A"},{"name":"B"}]' },
      });

      await runBatchCreateOperation(ctx);

      expect(ctx.services.records.batchCreate).toHaveBeenCalledWith('people', [
        { name: 'A' },
        { name: 'B' },
      ]);
      expect(ctx.services.output.render).toHaveBeenCalled();
    });

    it('batch creates records from CSV file', async () => {
      const ctx = createMockContext({
        options: { file: '/path/to/data.csv' },
      });

      await runBatchCreateOperation(ctx);

      expect(ctx.services.importer.import).toHaveBeenCalledWith('/path/to/data.csv');
      expect(ctx.services.records.batchCreate).toHaveBeenCalled();
    });

    it('propagates error when batch create fails', async () => {
      const ctx = createMockContext({
        options: { data: '[{"name":"A"}]' },
      });
      (ctx.services.records.batchCreate as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('Batch failed')
      );

      await expect(runBatchCreateOperation(ctx)).rejects.toThrow('Batch failed');
    });
  });

  // ==================== BATCH UPDATE OPERATION ====================
  describe('runBatchUpdateOperation', () => {
    it('batch updates records from --data JSON array', async () => {
      const ctx = createMockContext({
        options: { data: '[{"id":"1","name":"Updated A"},{"id":"2","name":"Updated B"}]' },
      });

      await runBatchUpdateOperation(ctx);

      expect(ctx.services.records.batchUpdate).toHaveBeenCalledWith('people', [
        { id: '1', name: 'Updated A' },
        { id: '2', name: 'Updated B' },
      ]);
      expect(ctx.services.output.render).toHaveBeenCalled();
    });

    it('batch updates records from CSV file', async () => {
      const ctx = createMockContext({
        options: { file: '/path/to/updates.csv' },
      });

      await runBatchUpdateOperation(ctx);

      expect(ctx.services.importer.import).toHaveBeenCalledWith('/path/to/updates.csv');
      expect(ctx.services.records.batchUpdate).toHaveBeenCalled();
    });

    it('propagates error when batch update fails', async () => {
      const ctx = createMockContext({
        options: { data: '[{"id":"1"}]' },
      });
      (ctx.services.records.batchUpdate as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('Update failed')
      );

      await expect(runBatchUpdateOperation(ctx)).rejects.toThrow('Update failed');
    });
  });

  // ==================== BATCH DELETE OPERATION ====================
  describe('runBatchDeleteOperation', () => {
    it('batch deletes records with --ids when --force is provided', async () => {
      const ctx = createMockContext({
        options: { ids: 'id-1,id-2,id-3', force: true },
      });

      await runBatchDeleteOperation(ctx);

      expect(ctx.services.records.batchDelete).toHaveBeenCalledWith('people', [
        'id-1',
        'id-2',
        'id-3',
      ]);
      expect(ctx.services.output.render).toHaveBeenCalled();
    });

    it('batch deletes records with --data JSON array', async () => {
      const ctx = createMockContext({
        options: { data: '["id-a","id-b"]', force: true },
      });

      await runBatchDeleteOperation(ctx);

      expect(ctx.services.records.batchDelete).toHaveBeenCalledWith('people', ['id-a', 'id-b']);
    });

    it('prints confirmation message without --force', async () => {
      const ctx = createMockContext({
        options: { ids: 'id-1,id-2' },
      });

      await runBatchDeleteOperation(ctx);

      expect(ctx.services.records.batchDelete).not.toHaveBeenCalled();
      expect(consoleSpy).toHaveBeenCalledWith(
        'About to batch delete people. Use --force to confirm.'
      );
    });

    it('throws CliError when no IDs provided', async () => {
      const ctx = createMockContext({
        options: { force: true },
      });

      await expect(runBatchDeleteOperation(ctx)).rejects.toThrow(CliError);
      await expect(runBatchDeleteOperation(ctx)).rejects.toThrow('Missing JSON payload');
    });

    it('throws CliError when payload is not an array', async () => {
      const ctx = createMockContext({
        options: { data: '{"id":"1"}', force: true },
      });

      await expect(runBatchDeleteOperation(ctx)).rejects.toThrow(CliError);
      await expect(runBatchDeleteOperation(ctx)).rejects.toThrow('Batch payload must be a JSON array');
    });

    it('throws CliError when IDs array is empty', async () => {
      const ctx = createMockContext({
        options: { ids: '  ,  , ', force: true },
      });

      await expect(runBatchDeleteOperation(ctx)).rejects.toThrow(CliError);
      await expect(runBatchDeleteOperation(ctx)).rejects.toThrow('No valid IDs provided');
    });
  });
});
