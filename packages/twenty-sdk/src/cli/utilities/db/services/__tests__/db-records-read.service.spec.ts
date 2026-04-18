import { describe, expect, it, vi } from "vitest";
import { DbRecordsReadService } from "../db-records-read.service";
import { UnsupportedDbReadError } from "../../../readbackend/types";

const DB_TARGET = {
  mode: "db",
  source: "env",
  workspace: "default",
  databaseUrl: "postgresql://reader:secret@db.example.com:5432/twenty?sslmode=require",
} as const;

describe("DbRecordsReadService", () => {
  it("returns the existing ListResponse shape for db-backed list reads", async () => {
    const planner = {
      planObject: vi.fn().mockResolvedValue({
        objectMetadata: { id: "person-id", nameSingular: "person", namePlural: "people" },
        tableName: "person",
        includes: [],
      }),
    };
    const filterCompiler = {
      compile: vi.fn().mockReturnValue([
        { field: "status", operator: "eq", value: "ACTIVE" },
        { field: "score", operator: "gte", value: 10 },
      ]),
    };
    const client = {
      query: vi
        .fn()
        .mockResolvedValueOnce({ rows: [{ totalCount: "3" }] })
        .mockResolvedValueOnce({
          rows: [
            { rowData: { id: "person-1", name: "Ada", status: "ACTIVE" } },
            { rowData: { id: "person-2", name: "Grace", status: "ACTIVE" } },
          ],
        }),
      end: vi.fn().mockResolvedValue(undefined),
    };
    const dbConnection = {
      connect: vi.fn().mockResolvedValue(client),
    };
    const service = new DbRecordsReadService(
      {} as never,
      planner as never,
      filterCompiler as never,
      dbConnection as never,
    );

    const result = await service.list(DB_TARGET, "people", {
      limit: 2,
      filter: "status[eq]:ACTIVE;score[gte]:10",
      sort: "createdAt",
      order: "desc",
    });

    expect(dbConnection.connect).toHaveBeenCalledWith({
      databaseUrl: "postgresql://reader:secret@db.example.com:5432/twenty?sslmode=require",
    });
    expect(planner.planObject).toHaveBeenCalledWith("people", { include: undefined });
    expect(filterCompiler.compile).toHaveBeenCalledWith("status[eq]:ACTIVE;score[gte]:10");

    const queries = client.query.mock.calls.map(([sql, params]) => ({
      sql: String(sql),
      params,
    }));

    expect(queries).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          sql: expect.stringContaining('from "person" as t'),
          params: ["ACTIVE", 10],
        }),
        expect.objectContaining({
          sql: expect.stringContaining('order by t."createdAt" desc nulls last, t."id" asc'),
          params: ["ACTIVE", 10, 2, 0],
        }),
      ]),
    );
    expect(result).toEqual({
      data: [
        { id: "person-1", name: "Ada", status: "ACTIVE" },
        { id: "person-2", name: "Grace", status: "ACTIVE" },
      ],
      totalCount: 3,
      pageInfo: {
        hasNextPage: true,
        endCursor: "db:1",
      },
    });
  });

  it("accumulates rows across db pages for listAll", async () => {
    const planner = {
      planObject: vi.fn().mockResolvedValue({
        objectMetadata: { id: "person-id", nameSingular: "person", namePlural: "people" },
        tableName: "person",
        includes: [],
      }),
    };
    const filterCompiler = {
      compile: vi.fn().mockReturnValue([]),
    };
    const client = {
      query: vi
        .fn()
        .mockResolvedValueOnce({ rows: [{ totalCount: "2" }] })
        .mockResolvedValueOnce({ rows: [{ rowData: { id: "person-1", name: "Ada" } }] })
        .mockResolvedValueOnce({ rows: [{ totalCount: "2" }] })
        .mockResolvedValueOnce({ rows: [{ rowData: { id: "person-2", name: "Grace" } }] }),
      end: vi.fn().mockResolvedValue(undefined),
    };
    const dbConnection = {
      connect: vi.fn().mockResolvedValue(client),
    };
    const service = new DbRecordsReadService(
      {} as never,
      planner as never,
      filterCompiler as never,
      dbConnection as never,
    );

    const result = await service.listAll(DB_TARGET, "people", { limit: 1 });

    expect(result).toEqual({
      data: [
        { id: "person-1", name: "Ada" },
        { id: "person-2", name: "Grace" },
      ],
      totalCount: 2,
      pageInfo: {
        hasNextPage: false,
        endCursor: "db:1",
      },
    });
  });

  it("hydrates MANY_TO_ONE includes for get", async () => {
    const planner = {
      planObject: vi.fn().mockResolvedValue({
        objectMetadata: { id: "person-id", nameSingular: "person", namePlural: "people" },
        tableName: "person",
        includes: [
          {
            relationName: "company",
            joinColumnName: "companyId",
            fieldMetadata: { id: "field-1" },
            objectMetadata: { id: "company-id", nameSingular: "company", namePlural: "companies" },
            tableName: "company",
          },
        ],
      }),
    };
    const client = {
      query: vi
        .fn()
        .mockResolvedValueOnce({ rows: [{ rowData: { id: "person-1", companyId: "company-1" } }] })
        .mockResolvedValueOnce({ rows: [{ rowData: { id: "company-1", name: "Acme" } }] }),
      end: vi.fn().mockResolvedValue(undefined),
    };
    const dbConnection = {
      connect: vi.fn().mockResolvedValue(client),
    };
    const service = new DbRecordsReadService(
      {} as never,
      planner as never,
      undefined,
      dbConnection as never,
    );

    const result = await service.get(DB_TARGET, "people", "person-1", { include: "company" });

    expect(result).toEqual({
      id: "person-1",
      companyId: "company-1",
      company: {
        id: "company-1",
        name: "Acme",
      },
    });
  });

  it("maps DB groupBy rows into the existing response shape", async () => {
    const planner = {
      planObject: vi.fn().mockResolvedValue({
        objectMetadata: { id: "person-id", nameSingular: "person", namePlural: "people" },
        tableName: "person",
        includes: [],
      }),
    };
    const filterCompiler = {
      compile: vi.fn().mockReturnValue([{ field: "status", operator: "eq", value: "ACTIVE" }]),
    };
    const client = {
      query: vi.fn().mockResolvedValue({
        rows: [{ group_0: "London", countNotEmptyId: "2" }],
      }),
      end: vi.fn().mockResolvedValue(undefined),
    };
    const dbConnection = {
      connect: vi.fn().mockResolvedValue(client),
    };
    const service = new DbRecordsReadService(
      {} as never,
      planner as never,
      filterCompiler as never,
      dbConnection as never,
    );

    const result = await service.groupBy(
      DB_TARGET,
      "people",
      { groupBy: [{ city: true }], filter: "status[eq]:ACTIVE" },
      { aggregate: ["totalCount"], limit: ["5"] },
    );

    expect(result).toEqual([
      {
        groupByDimensionValues: ["London"],
        countNotEmptyId: "2",
      },
    ]);
  });

  it("rejects include hydration on list reads", async () => {
    const service = new DbRecordsReadService({} as never);

    await expect(service.list(DB_TARGET, "people", { include: "company" })).rejects.toBeInstanceOf(
      UnsupportedDbReadError,
    );
  });
});
