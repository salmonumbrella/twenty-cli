import { describe, it, expect, vi } from 'vitest';
import { RecordsService } from '../records.service';

describe('RecordsService', () => {
  it('lists records with params', async () => {
    const mockApi = {
      get: vi.fn().mockResolvedValue({
        data: { data: { people: [{ id: '1' }] }, totalCount: 1 },
      }),
    };

    const service = new RecordsService(mockApi as any);
    const result = await service.list('people', { limit: 10, filter: 'email[eq]:test@example.com' });

    expect(mockApi.get).toHaveBeenCalledWith('/rest/people', {
      params: { limit: '10', filter: 'email[eq]:test@example.com' },
    });
    expect(result.data).toHaveLength(1);
  });

  it('creates a record', async () => {
    const mockApi = {
      post: vi.fn().mockResolvedValue({
        data: { data: { createPerson: { id: '123', name: 'Test' } } },
      }),
    };

    const service = new RecordsService(mockApi as any);
    const result = await service.create('people', { name: 'Test' });

    expect(mockApi.post).toHaveBeenCalledWith('/rest/people', { name: 'Test' });
    expect((result as any).id).toBe('123');
  });
});
