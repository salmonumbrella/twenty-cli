import { describe, it, expect, vi } from 'vitest';
import { SearchService } from '../search.service';

describe('SearchService', () => {
  describe('search', () => {
    it('searches with query', async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            data: {
              search: [
                { recordId: '1', objectNameSingular: 'person', record: { name: 'John' } },
                { recordId: '2', objectNameSingular: 'company', record: { name: 'Acme' } },
              ],
            },
          },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: 'test' });

      expect(mockApi.post).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('query Search'),
        variables: {
          searchInput: 'test',
          limit: undefined,
          includedObjectNameSingulars: undefined,
          excludedObjectNameSingulars: undefined,
        },
      });
      expect(result).toHaveLength(2);
      expect(result[0].recordId).toBe('1');
      expect(result[0].objectNameSingular).toBe('person');
      expect(result[1].recordId).toBe('2');
    });

    it('searches with limit', async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            data: {
              search: [{ recordId: '1', objectNameSingular: 'person', record: { name: 'John' } }],
            },
          },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: 'test', limit: 5 });

      expect(mockApi.post).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('query Search'),
        variables: {
          searchInput: 'test',
          limit: 5,
          includedObjectNameSingulars: undefined,
          excludedObjectNameSingulars: undefined,
        },
      });
      expect(result).toHaveLength(1);
    });

    it('searches with objects filter', async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            data: {
              search: [{ recordId: '1', objectNameSingular: 'person', record: { name: 'John' } }],
            },
          },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: 'test', objects: ['person', 'company'] });

      expect(mockApi.post).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('query Search'),
        variables: {
          searchInput: 'test',
          limit: undefined,
          includedObjectNameSingulars: ['person', 'company'],
          excludedObjectNameSingulars: undefined,
        },
      });
      expect(result).toHaveLength(1);
    });

    it('searches with exclude filter', async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            data: {
              search: [{ recordId: '1', objectNameSingular: 'task', record: { title: 'Task 1' } }],
            },
          },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: 'test', excludeObjects: ['person', 'company'] });

      expect(mockApi.post).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('query Search'),
        variables: {
          searchInput: 'test',
          limit: undefined,
          includedObjectNameSingulars: undefined,
          excludedObjectNameSingulars: ['person', 'company'],
        },
      });
      expect(result).toHaveLength(1);
      expect(result[0].objectNameSingular).toBe('task');
    });

    it('searches with all options combined', async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            data: {
              search: [{ recordId: '1', objectNameSingular: 'person', record: { name: 'John' } }],
            },
          },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({
        query: 'john',
        limit: 10,
        objects: ['person'],
        excludeObjects: ['note'],
      });

      expect(mockApi.post).toHaveBeenCalledWith('/graphql', {
        query: expect.stringContaining('query Search'),
        variables: {
          searchInput: 'john',
          limit: 10,
          includedObjectNameSingulars: ['person'],
          excludedObjectNameSingulars: ['note'],
        },
      });
      expect(result).toHaveLength(1);
    });

    it('returns empty array when no results', async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: { data: { search: [] } },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: 'nonexistent' });

      expect(result).toEqual([]);
    });

    it('handles missing data gracefully', async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({ data: {} }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: 'test' });

      expect(result).toEqual([]);
    });

    it('handles missing search field gracefully', async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({ data: { data: {} } }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: 'test' });

      expect(result).toEqual([]);
    });

    it('propagates API errors', async () => {
      const mockApi = {
        post: vi.fn().mockRejectedValue(new Error('Network error')),
      };

      const service = new SearchService(mockApi as any);

      await expect(service.search({ query: 'test' })).rejects.toThrow('Network error');
    });

    it('propagates GraphQL errors', async () => {
      const mockApi = {
        post: vi.fn().mockRejectedValue(new Error('GraphQL validation error')),
      };

      const service = new SearchService(mockApi as any);

      await expect(service.search({ query: 'test' })).rejects.toThrow('GraphQL validation error');
    });

    it('sends correct GraphQL query structure', async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: { data: { search: [] } },
        }),
      };

      const service = new SearchService(mockApi as any);
      await service.search({ query: 'test' });

      const call = mockApi.post.mock.calls[0];
      const query = call[1].query;

      expect(query).toContain('query Search');
      expect(query).toContain('$searchInput: String!');
      expect(query).toContain('$limit: Int');
      expect(query).toContain('$includedObjectNameSingulars: [String!]');
      expect(query).toContain('$excludedObjectNameSingulars: [String!]');
      expect(query).toContain('recordId');
      expect(query).toContain('objectNameSingular');
      expect(query).toContain('record');
    });
  });
});
