import { describe, expect, it, vi } from "vitest";
import { ReadBackendService } from "../read-backend.service";
import { DbConnectionError, UnsupportedDbReadError } from "../types";

describe("ReadBackendService", () => {
  it("uses API search when the resolved target is not db", async () => {
    const apiSearch = {
      search: vi.fn().mockResolvedValue({ data: [{ recordId: "api-1" }] }),
    };
    const resolver = {
      resolve: vi.fn().mockResolvedValue({ mode: "api", source: "none", workspace: "ws" }),
    };
    const dbSearch = {
      search: vi.fn(),
    };

    const service = new ReadBackendService(resolver as never, apiSearch as never, undefined, {
      search: dbSearch as never,
    });
    const result = await service.runSearch({ query: "john" });

    expect(resolver.resolve).toHaveBeenCalledWith({ workspace: undefined });
    expect(apiSearch.search).toHaveBeenCalledWith({ query: "john" });
    expect(dbSearch.search).not.toHaveBeenCalled();
    expect(result).toEqual({ data: [{ recordId: "api-1" }] });
  });

  it("uses DB search when the resolved target is db", async () => {
    const target = {
      mode: "db",
      source: "env",
      workspace: "ws",
      databaseUrl: "postgresql://db.example.com:5432/twenty",
    };
    const apiSearch = {
      search: vi.fn(),
    };
    const resolver = {
      resolve: vi.fn().mockResolvedValue(target),
    };
    const dbSearch = {
      search: vi.fn().mockResolvedValue({ data: [{ recordId: "db-1" }] }),
    };

    const service = new ReadBackendService(resolver as never, apiSearch as never, undefined, {
      search: dbSearch as never,
    });
    const result = await service.runSearch({ query: "john" });

    expect(dbSearch.search).toHaveBeenCalledWith(target, { query: "john" });
    expect(apiSearch.search).not.toHaveBeenCalled();
    expect(result).toEqual({ data: [{ recordId: "db-1" }] });
  });

  it("falls back to API search when the DB read is unsupported", async () => {
    const target = {
      mode: "db",
      source: "profile",
      workspace: "ws",
      databaseUrl: "postgresql://db.example.com:5432/twenty",
      profileName: "readonly",
    };
    const apiSearch = {
      search: vi.fn().mockResolvedValue({ data: [{ recordId: "api-1" }] }),
    };
    const resolver = {
      resolve: vi.fn().mockResolvedValue(target),
    };
    const dbSearch = {
      search: vi.fn().mockRejectedValue(new UnsupportedDbReadError("search not implemented")),
    };

    const service = new ReadBackendService(resolver as never, apiSearch as never, undefined, {
      search: dbSearch as never,
    });
    const result = await service.runSearch({ query: "john" });

    expect(dbSearch.search).toHaveBeenCalledWith(target, { query: "john" });
    expect(apiSearch.search).toHaveBeenCalledWith({ query: "john" });
    expect(result).toEqual({ data: [{ recordId: "api-1" }] });
  });

  it("falls back to API search when the DB connection fails", async () => {
    const target = {
      mode: "db",
      source: "env",
      workspace: "ws",
      databaseUrl: "postgresql://db.example.com:5432/twenty",
    };
    const apiSearch = {
      search: vi.fn().mockResolvedValue({ data: [{ recordId: "api-1" }] }),
    };
    const resolver = {
      resolve: vi.fn().mockResolvedValue(target),
    };
    const dbSearch = {
      search: vi.fn().mockRejectedValue(new DbConnectionError("database unavailable")),
    };

    const service = new ReadBackendService(resolver as never, apiSearch as never, undefined, {
      search: dbSearch as never,
    });
    const result = await service.runSearch({ query: "john" });

    expect(dbSearch.search).toHaveBeenCalledWith(target, { query: "john" });
    expect(apiSearch.search).toHaveBeenCalledWith({ query: "john" });
    expect(result).toEqual({ data: [{ recordId: "api-1" }] });
  });

  it("uses DB records reads when the resolved target is db", async () => {
    const target = {
      mode: "db",
      source: "env",
      workspace: "ws",
      databaseUrl: "postgresql://db.example.com:5432/twenty",
    };
    const resolver = {
      resolve: vi.fn().mockResolvedValue(target),
    };
    const apiSearch = {
      search: vi.fn(),
    };
    const apiRecords = {
      list: vi.fn(),
      listAll: vi.fn(),
      get: vi.fn(),
      groupBy: vi.fn(),
    };
    const dbRecords = {
      list: vi.fn().mockResolvedValue({ data: [{ id: "db-1" }] }),
    };

    const service = new ReadBackendService(
      resolver as never,
      apiSearch as never,
      apiRecords as never,
      { records: dbRecords as never },
    );
    const result = await service.list("people", { limit: 1 });

    expect(dbRecords.list).toHaveBeenCalledWith(target, "people", { limit: 1 });
    expect(apiRecords.list).not.toHaveBeenCalled();
    expect(result).toEqual({ data: [{ id: "db-1" }] });
  });

  it("falls back to API records reads when the DB read is unsupported", async () => {
    const target = {
      mode: "db",
      source: "profile",
      workspace: "ws",
      databaseUrl: "postgresql://db.example.com:5432/twenty",
      profileName: "readonly",
    };
    const resolver = {
      resolve: vi.fn().mockResolvedValue(target),
    };
    const apiSearch = {
      search: vi.fn(),
    };
    const apiRecords = {
      list: vi.fn().mockResolvedValue({ data: [{ id: "api-1" }] }),
      listAll: vi.fn(),
      get: vi.fn(),
      groupBy: vi.fn(),
    };
    const dbRecords = {
      list: vi.fn().mockRejectedValue(new UnsupportedDbReadError("unsupported include")),
    };

    const service = new ReadBackendService(
      resolver as never,
      apiSearch as never,
      apiRecords as never,
      { records: dbRecords as never },
    );
    const result = await service.list("people", { include: "company" });

    expect(dbRecords.list).toHaveBeenCalledWith(target, "people", { include: "company" });
    expect(apiRecords.list).toHaveBeenCalledWith("people", { include: "company" });
    expect(result).toEqual({ data: [{ id: "api-1" }] });
  });
});
