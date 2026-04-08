import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { OutputService } from "../output.service";
import { QueryService } from "../query.service";
import { TableService } from "../table.service";

describe("OutputService", () => {
  let outputService: OutputService;
  let consoleSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    const queryService = new QueryService();
    const tableService = new TableService();
    outputService = new OutputService(tableService, queryService);
    consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});
  });

  afterEach(() => {
    consoleSpy.mockRestore();
  });

  describe("CSV output with nested objects", () => {
    it("serializes nested objects to JSON strings", async () => {
      const data = [
        {
          id: "1",
          name: { firstName: "John", lastName: "Doe" },
          emails: { primaryEmail: "john@example.com", additionalEmails: [] },
        },
      ];

      await outputService.render(data, { format: "csv" });

      const output = consoleSpy.mock.calls[0][0];
      expect(output).toContain("id,name,emails");
      // CSV escapes inner quotes by doubling them: "" inside quoted field
      expect(output).toContain('"{""firstName"":""John"",""lastName"":""Doe""}"');
      expect(output).not.toContain("[object Object]");
    });

    it("handles arrays in CSV output", async () => {
      const data = [{ id: "1", tags: ["a", "b", "c"] }];

      await outputService.render(data, { format: "csv" });

      const output = consoleSpy.mock.calls[0][0];
      // CSV escapes inner quotes by doubling them: "" inside quoted field
      expect(output).toContain('"[""a"",""b"",""c""]"');
      expect(output).not.toContain("[object Object]");
    });

    it("handles null and primitive values correctly", async () => {
      const data = [{ id: "1", name: "Test", count: 42, active: true, deleted: null }];

      await outputService.render(data, { format: "csv" });

      const output = consoleSpy.mock.calls[0][0];
      expect(output).toContain("Test");
      expect(output).toContain("42");
      expect(output).toContain("true");
    });
  });

  describe("text output with CLI diagnostics", () => {
    it("prints a CLI note and omits _cli from the rendered table", async () => {
      await outputService.render(
        {
          message:
            "No skills found with names: xlsx. Available skills: workflow-building, data-manipulation, xlsx, pdf.",
          skills: [],
          _cli: {
            diagnosis: "workspace_skills_unavailable",
            message:
              "The MCP server advertised skill names but returned no loaded skills for this workspace. This is likely a workspace configuration issue, not a CLI transport failure.",
          },
        },
        { format: "text" },
      );

      expect(consoleSpy).toHaveBeenCalledWith(
        "Note: The MCP server advertised skill names but returned no loaded skills for this workspace. This is likely a workspace configuration issue, not a CLI transport failure.",
      );
      expect(consoleSpy.mock.calls[1]?.[0]).toContain("MESSAGE");
      expect(consoleSpy.mock.calls[1]?.[0]).toContain("SKILLS");
      expect(consoleSpy.mock.calls[1]?.[0]).not.toContain("_CLI");
    });
  });

  describe("JSONL output", () => {
    it("writes arrays as newline-delimited JSON objects", async () => {
      await outputService.render(
        [
          { id: "1", name: "Ada" },
          { id: "2", name: "Linus" },
        ],
        { format: "jsonl" },
      );

      expect(consoleSpy).toHaveBeenCalledTimes(1);
      expect(consoleSpy.mock.calls[0][0]).toBe(
        '{"id":"1","name":"Ada"}\n{"id":"2","name":"Linus"}',
      );
    });

    it("writes singleton values as one JSON line", async () => {
      await outputService.render({ ok: true }, { format: "jsonl" });

      expect(consoleSpy).toHaveBeenCalledWith('{"ok":true}');
    });
  });

  describe("agent output", () => {
    it("wraps arrays in a list envelope", async () => {
      await outputService.render([{ id: "1" }], {
        format: "agent",
        kind: "applications.list",
      });

      expect(JSON.parse(consoleSpy.mock.calls[0][0] as string)).toEqual({
        kind: "applications.list",
        items: [{ id: "1" }],
      });
    });

    it("wraps objects in an item envelope", async () => {
      await outputService.render(
        { id: "fn-1", name: "BuildFunction" },
        {
          format: "agent",
          kind: "serverless.get",
        },
      );

      expect(JSON.parse(consoleSpy.mock.calls[0][0] as string)).toEqual({
        kind: "serverless.get",
        item: { id: "fn-1", name: "BuildFunction" },
      });
    });

    it("wraps primitives in a data envelope", async () => {
      await outputService.render(true, {
        format: "agent",
        kind: "applications.install",
      });

      expect(JSON.parse(consoleSpy.mock.calls[0][0] as string)).toEqual({
        kind: "applications.install",
        data: true,
      });
    });

    it("applies queries before wrapping the agent envelope", async () => {
      await outputService.render(
        {
          records: [
            { id: "1", name: "Ada" },
            { id: "2", name: "Linus" },
          ],
        },
        {
          format: "agent",
          kind: "people.list",
          query: "records",
        },
      );

      expect(JSON.parse(consoleSpy.mock.calls[0][0] as string)).toEqual({
        kind: "people.list",
        items: [
          { id: "1", name: "Ada" },
          { id: "2", name: "Linus" },
        ],
      });
    });
  });
});
