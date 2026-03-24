import { describe, it, expect, vi } from "vitest";
import { RecordsService } from "../records.service";

describe("RecordsService", () => {
  describe("list", () => {
    it("lists records with params", async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({
          data: { data: { people: [{ id: "1" }] }, totalCount: 1 },
        }),
      };

      const service = new RecordsService(mockApi as any);
      const result = await service.list("people", {
        limit: 10,
        filter: "email[eq]:test@example.com",
        sort: "createdAt",
        order: "desc",
      });

      expect(mockApi.get).toHaveBeenCalledWith("/rest/people", {
        params: {
          limit: "10",
          filter: "email[eq]:test@example.com",
          order_by: "createdAt[DescNullsLast]",
        },
      });
      expect(result.data).toHaveLength(1);
    });

    it("returns pageInfo when available", async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({
          data: {
            data: { people: [{ id: "1" }] },
            pageInfo: { hasNextPage: true, endCursor: "abc123" },
          },
        }),
      };

      const service = new RecordsService(mockApi as any);
      const result = await service.list("people");

      expect(result.pageInfo?.hasNextPage).toBe(true);
      expect(result.pageInfo?.endCursor).toBe("abc123");
    });
  });

  describe("listAll", () => {
    it("fetches all pages until hasNextPage is false", async () => {
      const mockApi = {
        get: vi
          .fn()
          .mockResolvedValueOnce({
            data: {
              data: { people: [{ id: "1" }] },
              pageInfo: { hasNextPage: true, endCursor: "cursor1" },
              totalCount: 2,
            },
          })
          .mockResolvedValueOnce({
            data: {
              data: { people: [{ id: "2" }] },
              pageInfo: { hasNextPage: false },
              totalCount: 2,
            },
          }),
      };

      const service = new RecordsService(mockApi as any);
      const result = await service.listAll("people");

      expect(mockApi.get).toHaveBeenCalledTimes(2);
      expect(result.data).toHaveLength(2);
      expect(result.totalCount).toBe(2);
    });
  });

  describe("create", () => {
    it("creates a record", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({
          data: { data: { createPerson: { id: "123", name: "Test" } } },
        }),
      };

      const service = new RecordsService(mockApi as any);
      const result = await service.create("people", { name: "Test" });

      expect(mockApi.post).toHaveBeenCalledWith("/rest/people", { name: "Test" });
      expect((result as any).id).toBe("123");
    });
  });

  describe("get", () => {
    it("gets a single record", async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({
          data: { data: { person: { id: "1", name: "Test" } } },
        }),
      };

      const service = new RecordsService(mockApi as any);
      const result = await service.get("people", "1");

      expect(mockApi.get).toHaveBeenCalledWith("/rest/people/1", { params: {} });
      expect((result as any).name).toBe("Test");
    });

    it("includes depth param when include option set", async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({
          data: { data: { person: { id: "1" } } },
        }),
      };

      const service = new RecordsService(mockApi as any);
      await service.get("people", "1", { include: "company" });

      expect(mockApi.get).toHaveBeenCalledWith("/rest/people/1", { params: { depth: "1" } });
    });
  });

  describe("update", () => {
    it("updates a record with PATCH", async () => {
      const mockApi = {
        patch: vi.fn().mockResolvedValue({
          data: { data: { updatePerson: { id: "1", name: "Updated" } } },
        }),
      };

      const service = new RecordsService(mockApi as any);
      const result = await service.update("people", "1", { name: "Updated" });

      expect(mockApi.patch).toHaveBeenCalledWith("/rest/people/1", { name: "Updated" });
      expect((result as any).name).toBe("Updated");
    });
  });

  describe("delete", () => {
    it("soft deletes a record", async () => {
      const mockApi = {
        delete: vi.fn().mockResolvedValue({ data: { deleteTask: { id: "1" } } }),
      };

      const service = new RecordsService(mockApi as any);
      await service.delete("tasks", "1");

      expect(mockApi.delete).toHaveBeenCalledWith("/rest/tasks/1", {
        params: { soft_delete: "true" },
      });
    });
  });

  describe("destroy", () => {
    it("hard deletes a record", async () => {
      const mockApi = {
        delete: vi.fn().mockResolvedValue({ data: { destroyTask: { id: "1" } } }),
      };

      const service = new RecordsService(mockApi as any);
      await service.destroy("tasks", "1");

      expect(mockApi.delete).toHaveBeenCalledWith("/rest/tasks/1");
    });
  });

  describe("restore", () => {
    it("restores a soft-deleted record", async () => {
      const mockApi = {
        patch: vi.fn().mockResolvedValue({ data: { restoreTask: { id: "1" } } }),
      };

      const service = new RecordsService(mockApi as any);
      await service.restore("tasks", "1");

      expect(mockApi.patch).toHaveBeenCalledWith("/rest/restore/tasks/1");
    });
  });

  describe("batch operations", () => {
    it("batch creates records", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({ data: { createTasks: [{ id: "1" }, { id: "2" }] } }),
      };

      const service = new RecordsService(mockApi as any);
      const records = [{ title: "Task 1" }, { title: "Task 2" }];
      await service.batchCreate("tasks", records);

      expect(mockApi.post).toHaveBeenCalledWith("/rest/batch/tasks", records);
    });

    it("batch updates records", async () => {
      const mockApi = {
        patch: vi.fn().mockResolvedValue({ data: { updateTasks: [{ id: "1" }] } }),
      };

      const service = new RecordsService(mockApi as any);
      const records = [{ id: "1", title: "Updated" }];
      await service.batchUpdate("tasks", records);

      expect(mockApi.patch).toHaveBeenCalledWith("/rest/tasks/1", { title: "Updated" });
    });

    it("batch deletes records", async () => {
      const mockApi = {
        delete: vi.fn().mockResolvedValue({ data: { deleteTasks: [{ id: "1" }] } }),
      };

      const service = new RecordsService(mockApi as any);
      await service.batchDelete("tasks", ["1", "2"]);

      expect(mockApi.delete).toHaveBeenCalledWith("/rest/tasks", {
        params: {
          filter: "id[in]:[1,2]",
          soft_delete: "true",
        },
      });
    });

    it("updates many records with a collection PATCH and filter", async () => {
      const mockApi = {
        patch: vi.fn().mockResolvedValue({ data: { data: { updateTasks: [{ id: "1" }] } } }),
      };

      const service = new RecordsService(mockApi as any);
      await service.updateMany("tasks", { status: "DONE" }, { filter: "status[eq]:TODO" });

      expect(mockApi.patch).toHaveBeenCalledWith(
        "/rest/tasks",
        { status: "DONE" },
        {
          params: {
            filter: "status[eq]:TODO",
          },
        },
      );
    });

    it("restores many records with a collection restore route and filter", async () => {
      const mockApi = {
        patch: vi.fn().mockResolvedValue({ data: { data: { restoreTasks: [{ id: "1" }] } } }),
      };

      const service = new RecordsService(mockApi as any);
      await service.restoreMany("tasks", { filter: "id[in]:[1,2]" });

      expect(mockApi.patch).toHaveBeenCalledWith("/rest/restore/tasks", undefined, {
        params: {
          filter: "id[in]:[1,2]",
        },
      });
    });

    it("destroys many records with a collection DELETE and filter", async () => {
      const mockApi = {
        delete: vi.fn().mockResolvedValue({ data: { data: { deleteTasks: [{ id: "1" }] } } }),
      };

      const service = new RecordsService(mockApi as any);
      await service.destroyMany("tasks", { filter: "status[eq]:DONE" });

      expect(mockApi.delete).toHaveBeenCalledWith("/rest/tasks", {
        params: {
          filter: "status[eq]:DONE",
        },
      });
    });
  });

  describe("groupBy", () => {
    it("sends GET request with params when no payload", async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({ data: { groups: [] } }),
      };

      const service = new RecordsService(mockApi as any);
      await service.groupBy("people", { groupBy: [{ city: true }] });

      expect(mockApi.get).toHaveBeenCalledWith("/rest/people/groupBy", {
        params: {
          group_by: '[{"city":true}]',
        },
      });
    });

    it("serializes payload into query params when provided", async () => {
      const mockApi = {
        get: vi.fn().mockResolvedValue({ data: { groups: [] } }),
      };

      const service = new RecordsService(mockApi as any);
      await service.groupBy("people", {
        groupBy: [{ status: true }],
        filter: {
          city: {
            eq: "NYC",
          },
        },
      });

      expect(mockApi.get).toHaveBeenCalledWith("/rest/people/groupBy", {
        params: {
          group_by: '[{"status":true}]',
          filter: "city[eq]:NYC",
        },
      });
    });
  });

  describe("findDuplicates", () => {
    it("sends POST request with fields payload", async () => {
      const mockApi = {
        post: vi.fn().mockResolvedValue({ data: { duplicates: [] } }),
      };

      const service = new RecordsService(mockApi as any);
      await service.findDuplicates("people", { fields: ["email", "phone"] });

      expect(mockApi.post).toHaveBeenCalledWith("/rest/people/duplicates", {
        fields: ["email", "phone"],
      });
    });
  });

  describe("merge", () => {
    it("sends PATCH request to merge endpoint", async () => {
      const mockApi = {
        patch: vi.fn().mockResolvedValue({ data: { mergePerson: { id: "1" } } }),
      };

      const service = new RecordsService(mockApi as any);
      await service.merge("people", { primaryId: "1", secondaryIds: ["2"] });

      expect(mockApi.patch).toHaveBeenCalledWith("/rest/people/merge", {
        primaryId: "1",
        secondaryIds: ["2"],
      });
    });
  });
});
