import { describe, expect, it, vi } from "vitest";
import { DbSearchService } from "../db-search.service";
import { DbConnectionError, UnsupportedDbReadError } from "../../../readbackend/types";

describe("DbSearchService", () => {
  it("uses the resolved databaseUrl directly and maps the existing SearchResponse shape", async () => {
    const metadata = {
      listObjects: vi.fn().mockResolvedValue([
        {
          id: "person-id",
          nameSingular: "person",
          namePlural: "people",
          labelSingular: "Person",
        },
      ]),
    };
    const client = {
      query: vi.fn().mockResolvedValue({
        rows: [
          {
            recordId: "person-1",
            rowData: {
              name: {
                firstName: "Ada",
                lastName: "Lovelace",
              },
              avatarUrl: "https://example.com/ada.png",
            },
            tsRankCD: 0.9,
            tsRank: 0.8,
          },
        ],
      }),
      end: vi.fn().mockResolvedValue(undefined),
    };
    const dbConnection = {
      connect: vi.fn().mockResolvedValue(client),
    };
    const service = new DbSearchService(metadata as never, undefined, dbConnection as never);

    const result = await service.search(
      {
        mode: "db",
        source: "env",
        workspace: "default",
        databaseUrl: "postgresql://reader:secret@db.example.com:5432/twenty?sslmode=require",
      },
      { query: "ada" },
    );

    expect(dbConnection.connect).toHaveBeenCalledWith({
      databaseUrl: "postgresql://reader:secret@db.example.com:5432/twenty?sslmode=require",
    });
    expect(client.query).toHaveBeenCalledWith(expect.stringContaining('from "person"'), [
      "ada",
      21,
    ]);
    expect(result).toEqual({
      data: [
        {
          recordId: "person-1",
          objectNameSingular: "person",
          objectLabelSingular: "Person",
          label: "Ada Lovelace",
          imageUrl: "https://example.com/ada.png",
          tsRankCD: 0.9,
          tsRank: 0.8,
          cursor: "db:0",
        },
      ],
      pageInfo: {
        hasNextPage: false,
        endCursor: "db:0",
      },
    });
  });

  it("rejects filters because DB search does not support them yet", async () => {
    const service = new DbSearchService({
      listObjects: vi.fn(),
    } as never);

    await expect(
      service.search(
        {
          mode: "db",
          source: "env",
          workspace: "default",
          databaseUrl: "postgresql://reader:secret@db.example.com:5432/twenty",
        },
        {
          query: "ada",
          filter: { city: { eq: "London" } },
        },
      ),
    ).rejects.toBeInstanceOf(UnsupportedDbReadError);
  });

  it("maps connection failures to DbConnectionError", async () => {
    const service = new DbSearchService(
      {
        listObjects: vi.fn().mockResolvedValue([
          {
            id: "person-id",
            nameSingular: "person",
            namePlural: "people",
            labelSingular: "Person",
          },
        ]),
      } as never,
      undefined,
      {
        connect: vi.fn().mockRejectedValue(new Error("connect failed")),
      } as never,
    );

    await expect(
      service.search(
        {
          mode: "db",
          source: "env",
          workspace: "default",
          databaseUrl: "postgresql://reader:secret@db.example.com:5432/twenty",
        },
        {
          query: "ada",
        },
      ),
    ).rejects.toBeInstanceOf(DbConnectionError);
  });
});
