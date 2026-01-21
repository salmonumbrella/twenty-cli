import { describe, it, expect, vi } from 'vitest';
import { MetadataService } from '../metadata.service';

describe('MetadataService', () => {
  describe('listObjects', () => {
    it('returns array of objects', async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({
          data: { data: { objects: [{ id: '1', nameSingular: 'person' }] } },
        }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.listObjects();

      expect(mockApi.get).toHaveBeenCalledWith('/rest/metadata/objects');
      expect(result).toHaveLength(1);
      expect(result[0].nameSingular).toBe('person');
    });

    it('returns empty array when no objects', async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({
          data: { data: { objects: [] } },
        }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.listObjects();

      expect(result).toEqual([]);
    });

    it('handles missing data gracefully', async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({ data: {} }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.listObjects();

      expect(result).toEqual([]);
    });
  });

  describe('getObject', () => {
    it('fetches by ID when UUID provided', async () => {
      const uuid = '12345678-1234-5678-9012-123456789012';
      const mockApi = {
        get: vi.fn().mockResolvedValue({
          data: { data: { object: { id: uuid, nameSingular: 'person' } } },
        }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.getObject(uuid);

      expect(mockApi.get).toHaveBeenCalledWith(`/rest/metadata/objects/${uuid}`);
      expect(result.nameSingular).toBe('person');
    });

    it('looks up by nameSingular when non-UUID provided', async () => {
      const mockApi = {
        get: vi.fn()
          .mockResolvedValueOnce({
            data: { data: { objects: [{ id: 'obj-id', nameSingular: 'person', namePlural: 'people' }] } },
          })
          .mockResolvedValueOnce({
            data: { data: { object: { id: 'obj-id', nameSingular: 'person', fields: [] } } },
          }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.getObject('person');

      expect(mockApi.get).toHaveBeenCalledWith('/rest/metadata/objects');
      expect(mockApi.get).toHaveBeenCalledWith('/rest/metadata/objects/obj-id');
      expect(result.id).toBe('obj-id');
    });

    it('looks up by namePlural when non-UUID provided', async () => {
      const mockApi = {
        get: vi.fn()
          .mockResolvedValueOnce({
            data: { data: { objects: [{ id: 'obj-id', nameSingular: 'person', namePlural: 'people' }] } },
          })
          .mockResolvedValueOnce({
            data: { data: { object: { id: 'obj-id', nameSingular: 'person' } } },
          }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.getObject('people');

      expect(mockApi.get).toHaveBeenCalledWith('/rest/metadata/objects');
      expect(mockApi.get).toHaveBeenCalledWith('/rest/metadata/objects/obj-id');
    });

    it('throws when object not found by name', async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({
          data: { data: { objects: [{ id: '1', nameSingular: 'task' }] } },
        }),
      };

      const service = new MetadataService(mockApi as any);

      await expect(service.getObject('nonexistent')).rejects.toThrow('Object not found: nonexistent');
    });

    it('handles fallback data structure for direct ID lookup', async () => {
      const uuid = '12345678-1234-5678-9012-123456789012';
      const mockApi = {
        get: vi.fn().mockResolvedValue({
          data: { data: { id: uuid, nameSingular: 'widget' } },
        }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.getObject(uuid);

      expect(result.id).toBe(uuid);
      expect(result.nameSingular).toBe('widget');
    });
  });

  describe('listFields', () => {
    it('returns array of fields', async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({
          data: { data: { fields: [{ id: '1', name: 'email' }] } },
        }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.listFields();

      expect(mockApi.get).toHaveBeenCalledWith('/rest/metadata/fields');
      expect(result).toHaveLength(1);
      expect(result[0].id).toBe('1');
    });

    it('returns empty array when no fields', async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({
          data: { data: { fields: [] } },
        }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.listFields();

      expect(result).toEqual([]);
    });

    it('handles missing data gracefully', async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({ data: {} }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.listFields();

      expect(result).toEqual([]);
    });
  });

  describe('getField', () => {
    it('fetches field by ID', async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({
          data: { data: { field: { id: 'f1', name: 'email' } } },
        }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.getField('f1');

      expect(mockApi.get).toHaveBeenCalledWith('/rest/metadata/fields/f1');
      expect(result.id).toBe('f1');
    });

    it('handles fallback data structure', async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({
          data: { data: { id: 'f2', name: 'phone' } },
        }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.getField('f2');

      expect(result.id).toBe('f2');
    });
  });

  describe('createObject', () => {
    it('posts new object definition', async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: { data: { createObject: { id: 'new-id' } } },
        }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.createObject({ nameSingular: 'widget', namePlural: 'widgets' });

      expect(mockApi.post).toHaveBeenCalledWith('/rest/metadata/objects', {
        nameSingular: 'widget',
        namePlural: 'widgets',
      });
      expect(result).toEqual({ data: { createObject: { id: 'new-id' } } });
    });

    it('returns null when response has no data', async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({}),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.createObject({ nameSingular: 'test' });

      expect(result).toBeNull();
    });
  });

  describe('createField', () => {
    it('posts new field definition', async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: { data: { createField: { id: 'new-field-id' } } },
        }),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.createField({ objectMetadataId: 'obj-1', name: 'rating', type: 'NUMBER' });

      expect(mockApi.post).toHaveBeenCalledWith('/rest/metadata/fields', {
        objectMetadataId: 'obj-1',
        name: 'rating',
        type: 'NUMBER',
      });
      expect(result).toEqual({ data: { createField: { id: 'new-field-id' } } });
    });

    it('returns null when response has no data', async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({}),
      };

      const service = new MetadataService(mockApi as any);
      const result = await service.createField({ name: 'test' });

      expect(result).toBeNull();
    });
  });
});
