import { describe, it, expect, vi } from "vitest";
import { SearchService } from "../search.service";

describe("SearchService", () => {
  describe("search", () => {
    it("searches with query", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            data: {
              search: {
                edges: [
                  {
                    cursor: "cursor-1",
                    node: {
                      recordId: "1",
                      objectNameSingular: "person",
                      objectLabelSingular: "Person",
                      label: "John",
                      imageUrl: null,
                      tsRankCD: 0.9,
                      tsRank: 0.8,
                    },
                  },
                  {
                    cursor: "cursor-2",
                    node: {
                      recordId: "2",
                      objectNameSingular: "company",
                      objectLabelSingular: "Company",
                      label: "Acme",
                      imageUrl: null,
                      tsRankCD: 0.7,
                      tsRank: 0.6,
                    },
                  },
                ],
              },
            },
          },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: "test" });

      expect(mockApi.post).toHaveBeenCalledWith("/graphql", {
        query: expect.stringContaining("query Search"),
        variables: {
          searchInput: "test",
          limit: 20,
          after: undefined,
          filter: undefined,
          includedObjectNameSingulars: undefined,
          excludedObjectNameSingulars: undefined,
        },
      });
      expect(result.data).toHaveLength(2);
      expect(result.data[0].recordId).toBe("1");
      expect(result.data[0].objectNameSingular).toBe("person");
      expect(result.data[0].cursor).toBe("cursor-1");
      expect(result.data[1].recordId).toBe("2");
    });

    it("searches with limit", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            data: {
              search: {
                edges: [
                  {
                    cursor: "cursor-1",
                    node: {
                      recordId: "1",
                      objectNameSingular: "person",
                      objectLabelSingular: "Person",
                      label: "John",
                      imageUrl: null,
                      tsRankCD: 0.9,
                      tsRank: 0.8,
                    },
                  },
                ],
              },
            },
          },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: "test", limit: 5 });

      expect(mockApi.post).toHaveBeenCalledWith("/graphql", {
        query: expect.stringContaining("query Search"),
        variables: {
          searchInput: "test",
          limit: 5,
          after: undefined,
          filter: undefined,
          includedObjectNameSingulars: undefined,
          excludedObjectNameSingulars: undefined,
        },
      });
      expect(result.data).toHaveLength(1);
    });

    it("searches with objects filter", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            data: {
              search: {
                edges: [
                  {
                    cursor: "cursor-1",
                    node: {
                      recordId: "1",
                      objectNameSingular: "person",
                      objectLabelSingular: "Person",
                      label: "John",
                      imageUrl: null,
                      tsRankCD: 0.9,
                      tsRank: 0.8,
                    },
                  },
                ],
              },
            },
          },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: "test", objects: ["person", "company"] });

      expect(mockApi.post).toHaveBeenCalledWith("/graphql", {
        query: expect.stringContaining("query Search"),
        variables: {
          searchInput: "test",
          limit: 20,
          after: undefined,
          filter: undefined,
          includedObjectNameSingulars: ["person", "company"],
          excludedObjectNameSingulars: undefined,
        },
      });
      expect(result.data).toHaveLength(1);
    });

    it("searches with exclude filter", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            data: {
              search: {
                edges: [
                  {
                    cursor: "cursor-1",
                    node: {
                      recordId: "1",
                      objectNameSingular: "task",
                      objectLabelSingular: "Task",
                      label: "Task 1",
                      imageUrl: null,
                      tsRankCD: 0.9,
                      tsRank: 0.8,
                    },
                  },
                ],
              },
            },
          },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: "test", excludeObjects: ["person", "company"] });

      expect(mockApi.post).toHaveBeenCalledWith("/graphql", {
        query: expect.stringContaining("query Search"),
        variables: {
          searchInput: "test",
          limit: 20,
          after: undefined,
          filter: undefined,
          includedObjectNameSingulars: undefined,
          excludedObjectNameSingulars: ["person", "company"],
        },
      });
      expect(result.data).toHaveLength(1);
      expect(result.data[0].objectNameSingular).toBe("task");
    });

    it("searches with all options combined", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            data: {
              search: {
                edges: [
                  {
                    cursor: "cursor-1",
                    node: {
                      recordId: "1",
                      objectNameSingular: "person",
                      objectLabelSingular: "Person",
                      label: "John",
                      imageUrl: null,
                      tsRankCD: 0.9,
                      tsRank: 0.8,
                    },
                  },
                ],
              },
            },
          },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({
        query: "john",
        limit: 10,
        objects: ["person"],
        excludeObjects: ["note"],
      });

      expect(mockApi.post).toHaveBeenCalledWith("/graphql", {
        query: expect.stringContaining("query Search"),
        variables: {
          searchInput: "john",
          limit: 10,
          after: undefined,
          filter: undefined,
          includedObjectNameSingulars: ["person"],
          excludedObjectNameSingulars: ["note"],
        },
      });
      expect(result.data).toHaveLength(1);
    });

    it("searches with after cursor and filter", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            data: {
              search: {
                edges: [
                  {
                    cursor: "cursor-2",
                    node: {
                      recordId: "2",
                      objectNameSingular: "person",
                      objectLabelSingular: "Person",
                      label: "Jane",
                      imageUrl: null,
                      tsRankCD: 0.8,
                      tsRank: 0.7,
                    },
                  },
                ],
              },
            },
          },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({
        query: "jane",
        after: "cursor-1",
        filter: {
          id: { eq: "rec-2" },
        },
      });

      expect(mockApi.post).toHaveBeenCalledWith("/graphql", {
        query: expect.stringContaining("query Search"),
        variables: {
          searchInput: "jane",
          limit: 20,
          after: "cursor-1",
          filter: {
            id: { eq: "rec-2" },
          },
          includedObjectNameSingulars: undefined,
          excludedObjectNameSingulars: undefined,
        },
      });
      expect(result.data).toHaveLength(1);
      expect(result.data[0].cursor).toBe("cursor-2");
    });

    it("returns empty data when no results", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: { data: { search: { edges: [] } } },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: "nonexistent" });

      expect(result).toEqual({ data: [], pageInfo: undefined });
    });

    it("handles missing data gracefully", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({ data: {} }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: "test" });

      expect(result).toEqual({ data: [], pageInfo: undefined });
    });

    it("handles missing search field gracefully", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({ data: { data: {} } }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: "test" });

      expect(result).toEqual({ data: [], pageInfo: undefined });
    });

    it("returns pageInfo when available", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            data: {
              search: {
                edges: [],
                pageInfo: {
                  hasNextPage: true,
                  endCursor: "cursor-2",
                },
              },
            },
          },
        }),
      };

      const service = new SearchService(mockApi as any);
      const result = await service.search({ query: "test" });

      expect(result.pageInfo).toEqual({
        hasNextPage: true,
        endCursor: "cursor-2",
      });
    });

    it("propagates API errors", async () => {
      const mockApi = {
        post: vi.fn().mockRejectedValue(new Error("Network error")),
      };

      const service = new SearchService(mockApi as any);

      await expect(service.search({ query: "test" })).rejects.toThrow("Network error");
    });

    it("propagates GraphQL errors", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            errors: [
              {
                message: 'Variable "$limit" of type "Int" used in position expecting type "Int!".',
              },
            ],
          },
        }),
      };

      const service = new SearchService(mockApi as any);

      await expect(service.search({ query: "test" })).rejects.toThrow(
        'Variable "$limit" of type "Int" used in position expecting type "Int!".',
      );
    });

    it("sends correct GraphQL query structure", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: { data: { search: [] } },
        }),
      };

      const service = new SearchService(mockApi as any);
      await service.search({ query: "test" });

      const call = mockApi.post.mock.calls[0];
      const query = call[1].query;

      expect(query).toContain("query Search");
      expect(query).toContain("$searchInput: String!");
      expect(query).toContain("$limit: Int!");
      expect(query).toContain("$after: String");
      expect(query).toContain("$filter: ObjectRecordFilterInput");
      expect(query).toContain("$includedObjectNameSingulars: [String!]");
      expect(query).toContain("$excludedObjectNameSingulars: [String!]");
      expect(query).toContain("after: $after");
      expect(query).toContain("filter: $filter");
      expect(query).toContain("edges");
      expect(query).toContain("cursor");
      expect(query).toContain("recordId");
      expect(query).toContain("objectNameSingular");
      expect(query).toContain("objectLabelSingular");
      expect(query).toContain("label");
    });

    it("defaults limit to 20 when omitted", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: {
            data: {
              search: {
                edges: [],
              },
            },
          },
        }),
      };

      const service = new SearchService(mockApi as any);
      await service.search({ query: "test" });

      expect(mockApi.post).toHaveBeenCalledWith("/graphql", {
        query: expect.stringContaining("query Search"),
        variables: expect.objectContaining({
          searchInput: "test",
          limit: 20,
        }),
      });
    });
  });
});
