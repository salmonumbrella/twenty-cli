import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { Command } from "commander";
import { registerSearchCommand } from "../search.command";
import {
  SearchResponse,
  SearchResult,
  SearchService,
} from "../../../utilities/search/services/search.service";
import { mockConstructor } from "../../../test-utils/mock-constructor";
import { CliError } from "../../../utilities/errors/cli-error";

vi.mock("../../../utilities/search/services/search.service");
vi.mock("../../../utilities/api/services/api.service");
vi.mock("../../../utilities/shared/io", () => ({
  readJsonInput: vi.fn(),
}));
vi.mock("../../../utilities/config/services/config.service", () => ({
  ConfigService: vi.fn(function MockConfigService() {
    return {
      getConfig: vi.fn().mockResolvedValue({
        apiUrl: "https://api.twenty.com",
        apiKey: "test-token",
        workspace: "default",
      }),
    };
  }),
}));

import { readJsonInput } from "../../../utilities/shared/io";

describe("search command", () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;
  let mockSearch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerSearchCommand(program);
    consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});
    mockSearch = vi.fn();
    vi.mocked(readJsonInput).mockResolvedValue(undefined);
    vi.mocked(SearchService).mockImplementation(
      mockConstructor(
        () =>
          ({
            search: mockSearch,
          }) as unknown as SearchService,
      ),
    );
  });

  afterEach(() => {
    consoleSpy.mockRestore();
    vi.clearAllMocks();
  });

  describe("command registration", () => {
    it("registers search command with correct name and description", () => {
      const searchCmd = program.commands.find((cmd) => cmd.name() === "search");
      expect(searchCmd).toBeDefined();
      expect(searchCmd?.description()).toBe("Full-text search across all records");
    });

    it("has required query argument", () => {
      const searchCmd = program.commands.find((cmd) => cmd.name() === "search");
      const args = searchCmd?.registeredArguments ?? [];
      expect(args.length).toBe(1);
      expect(args[0].name()).toBe("query");
      expect(args[0].required).toBe(true);
    });

    it("has --limit option with default value", () => {
      const searchCmd = program.commands.find((cmd) => cmd.name() === "search");
      const opts = searchCmd?.options ?? [];
      const limitOpt = opts.find((o) => o.long === "--limit");
      expect(limitOpt).toBeDefined();
      expect(limitOpt?.defaultValue).toBe("20");
    });

    it("has --objects option", () => {
      const searchCmd = program.commands.find((cmd) => cmd.name() === "search");
      const opts = searchCmd?.options ?? [];
      const objectsOpt = opts.find((o) => o.long === "--objects");
      expect(objectsOpt).toBeDefined();
    });

    it("has --exclude option", () => {
      const searchCmd = program.commands.find((cmd) => cmd.name() === "search");
      const opts = searchCmd?.options ?? [];
      const excludeOpt = opts.find((o) => o.long === "--exclude");
      expect(excludeOpt).toBeDefined();
    });

    it("has pagination and filter options", () => {
      const searchCmd = program.commands.find((cmd) => cmd.name() === "search");
      const opts = searchCmd?.options ?? [];
      expect(opts.find((o) => o.long === "--after")).toBeDefined();
      expect(opts.find((o) => o.long === "--include-page-info")).toBeDefined();
      expect(opts.find((o) => o.long === "--filter")).toBeDefined();
      expect(opts.find((o) => o.long === "--filter-file")).toBeDefined();
    });

    it("has global options applied", () => {
      const searchCmd = program.commands.find((cmd) => cmd.name() === "search");
      const opts = searchCmd?.options ?? [];
      const outputOpt = opts.find((o) => o.long === "--output");
      const queryOpt = opts.find((o) => o.long === "--query");
      const workspaceOpt = opts.find((o) => o.long === "--workspace");
      expect(outputOpt).toBeDefined();
      expect(queryOpt).toBeDefined();
      expect(workspaceOpt).toBeDefined();
    });
  });

  describe("search execution", () => {
    it("performs search with default limit", async () => {
      mockSearch.mockResolvedValue({ data: [buildSearchResult("1", "John")] });

      await program.parseAsync(["node", "test", "search", "john", "-o", "json"]);

      expect(mockSearch).toHaveBeenCalledWith({
        query: "john",
        limit: 20,
        objects: undefined,
        excludeObjects: undefined,
        after: undefined,
        filter: undefined,
      });
    });

    it("performs search with custom limit", async () => {
      mockSearch.mockResolvedValue({ data: [] });

      await program.parseAsync(["node", "test", "search", "test", "--limit", "50", "-o", "json"]);

      expect(mockSearch).toHaveBeenCalledWith({
        query: "test",
        limit: 50,
        objects: undefined,
        excludeObjects: undefined,
        after: undefined,
        filter: undefined,
      });
    });

    it("filters by included objects", async () => {
      mockSearch.mockResolvedValue({ data: [] });

      await program.parseAsync([
        "node",
        "test",
        "search",
        "query",
        "--objects",
        "person,company",
        "-o",
        "json",
      ]);

      expect(mockSearch).toHaveBeenCalledWith({
        query: "query",
        limit: 20,
        objects: ["person", "company"],
        excludeObjects: undefined,
        after: undefined,
        filter: undefined,
      });
    });

    it("filters by excluded objects", async () => {
      mockSearch.mockResolvedValue({ data: [] });

      await program.parseAsync([
        "node",
        "test",
        "search",
        "query",
        "--exclude",
        "note,task",
        "-o",
        "json",
      ]);

      expect(mockSearch).toHaveBeenCalledWith({
        query: "query",
        limit: 20,
        objects: undefined,
        excludeObjects: ["note", "task"],
        after: undefined,
        filter: undefined,
      });
    });

    it("combines all options", async () => {
      mockSearch.mockResolvedValue({ data: [] });

      await program.parseAsync([
        "node",
        "test",
        "search",
        "hello",
        "--limit",
        "10",
        "--objects",
        "person",
        "--exclude",
        "company",
        "-o",
        "json",
      ]);

      expect(mockSearch).toHaveBeenCalledWith({
        query: "hello",
        limit: 10,
        objects: ["person"],
        excludeObjects: ["company"],
        after: undefined,
        filter: undefined,
      });
    });

    it("passes after cursor and filter JSON to the search service", async () => {
      mockSearch.mockResolvedValue({ data: [] });
      vi.mocked(readJsonInput).mockResolvedValue({
        id: { eq: "rec-1" },
      });

      await program.parseAsync([
        "node",
        "test",
        "search",
        "hello",
        "--after",
        "cursor-1",
        "--filter",
        '{"id":{"eq":"rec-1"}}',
        "-o",
        "json",
      ]);

      expect(readJsonInput).toHaveBeenCalledWith('{"id":{"eq":"rec-1"}}', undefined);
      expect(mockSearch).toHaveBeenCalledWith({
        query: "hello",
        limit: 20,
        objects: undefined,
        excludeObjects: undefined,
        after: "cursor-1",
        filter: {
          id: { eq: "rec-1" },
        },
      });
    });

    it("loads filter input from file", async () => {
      mockSearch.mockResolvedValue({ data: [] });
      vi.mocked(readJsonInput).mockResolvedValue({
        and: [{ createdAt: { gte: "2026-01-01T00:00:00.000Z" } }],
      });

      await program.parseAsync([
        "node",
        "test",
        "search",
        "hello",
        "--filter-file",
        "filters.json",
        "-o",
        "json",
      ]);

      expect(readJsonInput).toHaveBeenCalledWith(undefined, "filters.json");
      expect(mockSearch).toHaveBeenCalledWith({
        query: "hello",
        limit: 20,
        objects: undefined,
        excludeObjects: undefined,
        after: undefined,
        filter: {
          and: [{ createdAt: { gte: "2026-01-01T00:00:00.000Z" } }],
        },
      });
    });

    it("outputs search results", async () => {
      const results: SearchResult[] = [
        buildSearchResult("rec-1", "Alice"),
        buildSearchResult("rec-2", "Acme", "company", "Company"),
      ];
      mockSearch.mockResolvedValue({
        data: results,
        pageInfo: { hasNextPage: true, endCursor: "cursor-2" },
      });

      await program.parseAsync(["node", "test", "search", "a", "-o", "json"]);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toHaveLength(2);
      expect(parsed[0].recordId).toBe("rec-1");
      expect(parsed[1].recordId).toBe("rec-2");
      expect(parsed[0].label).toBe("Alice");
    });

    it("outputs pageInfo when explicitly requested", async () => {
      const response: SearchResponse = {
        data: [buildSearchResult("rec-1", "Alice")],
        pageInfo: {
          hasNextPage: true,
          endCursor: "cursor-1",
        },
      };
      mockSearch.mockResolvedValue(response);

      await program.parseAsync([
        "node",
        "test",
        "search",
        "a",
        "--include-page-info",
        "-o",
        "json",
      ]);

      const output = consoleSpy.mock.calls[0][0] as string;
      expect(JSON.parse(output)).toEqual(response);
    });

    it("handles empty results", async () => {
      mockSearch.mockResolvedValue({ data: [] });

      await program.parseAsync(["node", "test", "search", "nonexistent", "-o", "json"]);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual([]);
    });
  });

  describe("error handling", () => {
    it("requires query argument", async () => {
      await expect(program.parseAsync(["node", "test", "search"])).rejects.toThrow();
    });

    it("rejects non-object filter input", async () => {
      vi.mocked(readJsonInput).mockResolvedValue([]);

      await expect(
        program.parseAsync(["node", "test", "search", "hello", "--filter", "[]", "-o", "json"]),
      ).rejects.toThrow(CliError);
    });
  });
});

function buildSearchResult(
  recordId: string,
  label: string,
  objectNameSingular = "person",
  objectLabelSingular = "Person",
): SearchResult {
  return {
    recordId,
    objectNameSingular,
    objectLabelSingular,
    label,
    imageUrl: null,
    tsRankCD: 0.9,
    tsRank: 0.8,
  };
}
