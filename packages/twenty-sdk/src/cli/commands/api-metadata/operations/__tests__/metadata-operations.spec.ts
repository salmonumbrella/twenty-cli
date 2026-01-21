import { describe, it, expect, vi } from 'vitest';
import { runObjectsList } from '../objects-list.operation';
import { runObjectsGet } from '../objects-get.operation';
import { runObjectsCreate } from '../objects-create.operation';
import { runFieldsGet } from '../fields-get.operation';
import { runFieldsCreate } from '../fields-create.operation';
import { CliError } from '../../../../utilities/errors/cli-error';
import { ApiMetadataContext } from '../types';

// Mock the body utility
vi.mock('../../../../utilities/shared/body', () => ({
  parseBody: vi.fn().mockImplementation(async (data: string | undefined) => {
    if (data) return JSON.parse(data);
    return {};
  }),
}));

function createMockContext(overrides: Partial<ApiMetadataContext> = {}): ApiMetadataContext {
  return {
    type: 'metadata',
    operation: 'test',
    arg: undefined,
    options: {},
    globalOptions: { output: 'json' },
    services: {
      api: {} as any,
      records: {} as any,
      metadata: {
        listObjects: vi.fn().mockResolvedValue([
          { id: 'obj-1', nameSingular: 'person', namePlural: 'people' },
          { id: 'obj-2', nameSingular: 'company', namePlural: 'companies' },
        ]),
        getObject: vi.fn().mockResolvedValue({
          id: 'obj-1',
          nameSingular: 'person',
          namePlural: 'people',
          fields: [
            { id: 'f1', name: 'firstName' },
            { id: 'f2', name: 'lastName' },
          ],
        }),
        createObject: vi.fn().mockResolvedValue({
          id: 'new-obj',
          nameSingular: 'customObject',
          namePlural: 'customObjects',
        }),
        listFields: vi.fn().mockResolvedValue([
          { id: 'f1', name: 'email', objectMetadataId: 'obj-1' },
          { id: 'f2', name: 'phone', objectMetadataId: 'obj-2' },
        ]),
        getField: vi.fn().mockResolvedValue({
          id: 'field-123',
          name: 'email',
          type: 'TEXT',
          objectMetadataId: 'obj-1',
        }),
        createField: vi.fn().mockResolvedValue({
          id: 'new-field',
          name: 'newField',
          type: 'TEXT',
          objectMetadataId: 'obj-1',
        }),
      } as any,
      output: {
        render: vi.fn(),
      } as any,
      importer: {} as any,
      exporter: {} as any,
    },
    ...overrides,
  } as ApiMetadataContext;
}

describe('Metadata Operations', () => {
  // ==================== OBJECTS LIST OPERATION ====================
  describe('runObjectsList', () => {
    it('returns list of objects and renders output', async () => {
      const ctx = createMockContext();

      await runObjectsList(ctx);

      expect(ctx.services.metadata.listObjects).toHaveBeenCalled();
      expect(ctx.services.output.render).toHaveBeenCalledWith(
        [
          { id: 'obj-1', nameSingular: 'person', namePlural: 'people' },
          { id: 'obj-2', nameSingular: 'company', namePlural: 'companies' },
        ],
        { format: 'json', query: undefined }
      );
    });

    it('propagates error when listObjects fails', async () => {
      const ctx = createMockContext();
      (ctx.services.metadata.listObjects as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('API connection failed')
      );

      await expect(runObjectsList(ctx)).rejects.toThrow('API connection failed');
    });

    it('handles empty objects list', async () => {
      const ctx = createMockContext();
      (ctx.services.metadata.listObjects as ReturnType<typeof vi.fn>).mockResolvedValue([]);

      await runObjectsList(ctx);

      expect(ctx.services.output.render).toHaveBeenCalledWith(
        [],
        { format: 'json', query: undefined }
      );
    });
  });

  // ==================== OBJECTS GET OPERATION ====================
  describe('runObjectsGet', () => {
    it('returns single object with fields by ID', async () => {
      const ctx = createMockContext({
        arg: 'obj-1',
      });

      await runObjectsGet(ctx);

      expect(ctx.services.metadata.getObject).toHaveBeenCalledWith('obj-1');
      expect(ctx.services.output.render).toHaveBeenCalledWith(
        {
          id: 'obj-1',
          nameSingular: 'person',
          namePlural: 'people',
          fields: [
            { id: 'f1', name: 'firstName' },
            { id: 'f2', name: 'lastName' },
          ],
        },
        { format: 'json', query: undefined }
      );
    });

    it('returns single object by name', async () => {
      const ctx = createMockContext({
        arg: 'person',
      });

      await runObjectsGet(ctx);

      expect(ctx.services.metadata.getObject).toHaveBeenCalledWith('person');
      expect(ctx.services.output.render).toHaveBeenCalled();
    });

    it('throws CliError when object identifier is missing', async () => {
      const ctx = createMockContext({
        arg: undefined,
      });

      await expect(runObjectsGet(ctx)).rejects.toThrow(CliError);
      await expect(runObjectsGet(ctx)).rejects.toThrow('Missing object identifier');
    });

    it('propagates error when object is not found', async () => {
      const ctx = createMockContext({
        arg: 'nonexistent',
      });
      (ctx.services.metadata.getObject as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('Object not found: nonexistent')
      );

      await expect(runObjectsGet(ctx)).rejects.toThrow('Object not found: nonexistent');
    });
  });

  // ==================== OBJECTS CREATE OPERATION ====================
  describe('runObjectsCreate', () => {
    it('creates new object with --data and renders output', async () => {
      const ctx = createMockContext({
        options: { data: '{"nameSingular":"customObject","namePlural":"customObjects"}' },
      });

      await runObjectsCreate(ctx);

      expect(ctx.services.metadata.createObject).toHaveBeenCalledWith({
        nameSingular: 'customObject',
        namePlural: 'customObjects',
      });
      expect(ctx.services.output.render).toHaveBeenCalledWith(
        {
          id: 'new-obj',
          nameSingular: 'customObject',
          namePlural: 'customObjects',
        },
        { format: 'json', query: undefined }
      );
    });

    it('propagates error when createObject fails', async () => {
      const ctx = createMockContext({
        options: { data: '{"nameSingular":"invalid"}' },
      });
      (ctx.services.metadata.createObject as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('Validation error: namePlural is required')
      );

      await expect(runObjectsCreate(ctx)).rejects.toThrow('Validation error: namePlural is required');
    });

    it('creates object with minimal payload', async () => {
      const ctx = createMockContext({
        options: { data: '{}' },
      });

      await runObjectsCreate(ctx);

      expect(ctx.services.metadata.createObject).toHaveBeenCalledWith({});
      expect(ctx.services.output.render).toHaveBeenCalled();
    });

    it('handles complex object payload with fields', async () => {
      const payload = {
        nameSingular: 'project',
        namePlural: 'projects',
        labelSingular: 'Project',
        labelPlural: 'Projects',
        icon: 'IconFolder',
      };
      const ctx = createMockContext({
        options: { data: JSON.stringify(payload) },
      });

      await runObjectsCreate(ctx);

      expect(ctx.services.metadata.createObject).toHaveBeenCalledWith(payload);
    });
  });

  // ==================== FIELDS GET OPERATION ====================
  describe('runFieldsGet', () => {
    it('returns single field by ID and renders output', async () => {
      const ctx = createMockContext({
        arg: 'field-123',
      });

      await runFieldsGet(ctx);

      expect(ctx.services.metadata.getField).toHaveBeenCalledWith('field-123');
      expect(ctx.services.output.render).toHaveBeenCalledWith(
        {
          id: 'field-123',
          name: 'email',
          type: 'TEXT',
          objectMetadataId: 'obj-1',
        },
        { format: 'json', query: undefined }
      );
    });

    it('throws CliError when field ID is missing', async () => {
      const ctx = createMockContext({
        arg: undefined,
      });

      await expect(runFieldsGet(ctx)).rejects.toThrow(CliError);
      await expect(runFieldsGet(ctx)).rejects.toThrow('Missing field ID');
    });

    it('propagates error when field is not found', async () => {
      const ctx = createMockContext({
        arg: 'nonexistent-field',
      });
      (ctx.services.metadata.getField as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('Field not found')
      );

      await expect(runFieldsGet(ctx)).rejects.toThrow('Field not found');
    });

    it('returns field with all metadata properties', async () => {
      const fullField = {
        id: 'field-456',
        name: 'status',
        type: 'SELECT',
        objectMetadataId: 'obj-2',
        label: 'Status',
        isCustom: true,
        isActive: true,
        options: [
          { label: 'Active', value: 'ACTIVE', color: 'green' },
          { label: 'Inactive', value: 'INACTIVE', color: 'gray' },
        ],
      };
      const ctx = createMockContext({
        arg: 'field-456',
      });
      (ctx.services.metadata.getField as ReturnType<typeof vi.fn>).mockResolvedValue(fullField);

      await runFieldsGet(ctx);

      expect(ctx.services.output.render).toHaveBeenCalledWith(
        fullField,
        { format: 'json', query: undefined }
      );
    });
  });

  // ==================== FIELDS CREATE OPERATION ====================
  describe('runFieldsCreate', () => {
    it('creates new field with --data and renders output', async () => {
      const ctx = createMockContext({
        options: { data: '{"name":"newField","type":"TEXT","objectMetadataId":"obj-1"}' },
      });

      await runFieldsCreate(ctx);

      expect(ctx.services.metadata.createField).toHaveBeenCalledWith({
        name: 'newField',
        type: 'TEXT',
        objectMetadataId: 'obj-1',
      });
      expect(ctx.services.output.render).toHaveBeenCalledWith(
        {
          id: 'new-field',
          name: 'newField',
          type: 'TEXT',
          objectMetadataId: 'obj-1',
        },
        { format: 'json', query: undefined }
      );
    });

    it('propagates error when createField fails', async () => {
      const ctx = createMockContext({
        options: { data: '{"name":"invalid"}' },
      });
      (ctx.services.metadata.createField as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('Validation error: type is required')
      );

      await expect(runFieldsCreate(ctx)).rejects.toThrow('Validation error: type is required');
    });

    it('creates field with SELECT type and options', async () => {
      const payload = {
        name: 'priority',
        type: 'SELECT',
        objectMetadataId: 'obj-1',
        label: 'Priority',
        options: [
          { label: 'High', value: 'HIGH', color: 'red' },
          { label: 'Medium', value: 'MEDIUM', color: 'yellow' },
          { label: 'Low', value: 'LOW', color: 'green' },
        ],
      };
      const ctx = createMockContext({
        options: { data: JSON.stringify(payload) },
      });

      await runFieldsCreate(ctx);

      expect(ctx.services.metadata.createField).toHaveBeenCalledWith(payload);
    });

    it('creates field with minimal required properties', async () => {
      const ctx = createMockContext({
        options: { data: '{}' },
      });

      await runFieldsCreate(ctx);

      expect(ctx.services.metadata.createField).toHaveBeenCalledWith({});
      expect(ctx.services.output.render).toHaveBeenCalled();
    });
  });
});
