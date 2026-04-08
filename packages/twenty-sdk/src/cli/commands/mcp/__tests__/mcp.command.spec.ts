import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Command } from "commander";

vi.mock("../../../utilities/shared/services", () => ({
  createServices: vi.fn(),
}));

import { createServices } from "../../../utilities/shared/services";
import { registerMcpCommand } from "../mcp.command";

function createMockServices() {
  return {
    mcp: {
      status: vi.fn(),
      callTool: vi.fn(),
    },
    output: {
      render: vi.fn(),
    },
  };
}

describe("mcp command", () => {
  let program: Command;
  let mockServices: ReturnType<typeof createMockServices>;
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerMcpCommand(program);
    mockServices = createMockServices();
    vi.mocked(createServices).mockReturnValue(mockServices as any);
    consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
    vi.clearAllMocks();
  });

  it("registers the top-level mcp command", () => {
    const command = program.commands.find((candidate) => candidate.name() === "mcp");
    expect(command).toBeDefined();
  });

  it("registers the expected MCP subcommands", () => {
    const command = program.commands.find((candidate) => candidate.name() === "mcp");
    const subcommands = command?.commands.map((candidate) => candidate.name()) ?? [];

    expect(subcommands).toEqual(["status", "catalog", "schema", "exec", "skills", "search"]);
    expect(subcommands).not.toEqual(expect.arrayContaining(["learn", "call", "load-skills", "help-center"]));
  });

  it("runs mcp status and renders the structured status object", async () => {
    mockServices.mcp.status.mockResolvedValue({
      state: "ok",
      available: true,
      reachable: true,
    });

    await program.parseAsync(["node", "test", "mcp", "status"]);

    expect(mockServices.mcp.status).toHaveBeenCalled();
    expect(mockServices.output.render).toHaveBeenCalledWith(
      expect.objectContaining({ state: "ok" }),
      expect.any(Object),
    );
  });

  it("runs mcp catalog via get_tool_catalog", async () => {
    mockServices.mcp.callTool.mockResolvedValue({ categories: [] });

    await program.parseAsync([
      "node",
      "test",
      "mcp",
      "catalog",
      "-o",
      "json",
      "--query",
      "categories",
    ]);

    expect(mockServices.mcp.callTool).toHaveBeenCalledWith("get_tool_catalog", {});
    expect(mockServices.output.render).toHaveBeenCalledWith(
      { categories: [] },
      expect.objectContaining({
        format: "json",
        query: "categories",
      }),
    );
  });

  it("runs mcp schema for exact tool names", async () => {
    mockServices.mcp.callTool.mockResolvedValue({
      tools: ["find_companies", "create_person"],
    });

    await program.parseAsync([
      "node",
      "test",
      "mcp",
      "schema",
      "find_companies",
      "create_person",
      "-o",
      "json",
    ]);

    expect(mockServices.mcp.callTool).toHaveBeenCalledWith("learn_tools", {
      toolNames: ["find_companies", "create_person"],
    });
    expect(mockServices.output.render).toHaveBeenCalledWith(
      { tools: ["find_companies", "create_person"] },
      expect.objectContaining({
        format: "json",
      }),
    );
  });

  it("runs mcp exec with tool arguments from --data", async () => {
    mockServices.mcp.callTool.mockResolvedValue({ ok: true });

    await program.parseAsync([
      "node",
      "test",
      "mcp",
      "exec",
      "find_companies",
      "--data",
      '{"query":"acme"}',
    ]);

    expect(mockServices.mcp.callTool).toHaveBeenCalledWith("execute_tool", {
      toolName: "find_companies",
      arguments: {
        query: "acme",
      },
    });
    expect(mockServices.output.render).toHaveBeenCalledWith(
      { ok: true },
      expect.objectContaining({
        format: "text",
      }),
    );
  });

  it("rejects --data and --file together for mcp exec", async () => {
    await expect(
      program.parseAsync([
        "node",
        "test",
        "mcp",
        "exec",
        "find_companies",
        "--data",
        "{}",
        "--file",
        "/tmp/args.json",
      ]),
    ).rejects.toThrow("provide MCP call arguments via --data or --file");
  });

  it("rejects invalid --data JSON for mcp exec", async () => {
    await expect(
      program.parseAsync(["node", "test", "mcp", "exec", "find_companies", "--data", "{not-json}"]),
    ).rejects.toMatchObject({ code: "INVALID_ARGUMENTS" });
  });

  it("rejects non-object JSON for mcp exec", async () => {
    await expect(
      program.parseAsync([
        "node",
        "test",
        "mcp",
        "exec",
        "find_companies",
        "--data",
        '["not-an-object"]',
      ]),
    ).rejects.toMatchObject({
      code: "INVALID_ARGUMENTS",
      message: "MCP call arguments must be a JSON object.",
    });
  });

  it("rejects unreadable --file for mcp exec", async () => {
    await expect(
      program.parseAsync([
        "node",
        "test",
        "mcp",
        "exec",
        "find_companies",
        "--file",
        "/missing/file.json",
      ]),
    ).rejects.toMatchObject({
      code: "INVALID_ARGUMENTS",
      message: expect.stringContaining("Unable to read MCP call arguments file"),
    });
  });

  it("sends an empty object when mcp exec has no arguments", async () => {
    mockServices.mcp.callTool.mockResolvedValue({ ok: true });

    await program.parseAsync(["node", "test", "mcp", "exec", "get_tool_catalog"]);

    expect(mockServices.mcp.callTool).toHaveBeenCalledWith("execute_tool", {
      toolName: "get_tool_catalog",
      arguments: {},
    });
    expect(mockServices.output.render).toHaveBeenCalledWith(
      { ok: true },
      expect.objectContaining({
        format: "text",
      }),
    );
  });

  it("runs mcp skills", async () => {
    mockServices.mcp.callTool.mockResolvedValue({ loaded: true });

    await program.parseAsync(["node", "test", "mcp", "skills", "workflow-building", "xlsx"]);

    expect(mockServices.mcp.callTool).toHaveBeenCalledWith("load_skills", {
      skillNames: ["workflow-building", "xlsx"],
    });
    expect(mockServices.output.render).toHaveBeenCalledWith(
      { loaded: true },
      expect.objectContaining({
        format: "text",
      }),
    );
  });

  it("runs mcp search", async () => {
    mockServices.mcp.callTool.mockResolvedValue({ matches: [] });

    await program.parseAsync(["node", "test", "mcp", "search", "MCP setup"]);

    expect(mockServices.mcp.callTool).toHaveBeenCalledWith("search_help_center", {
      query: "MCP setup",
    });
    expect(mockServices.output.render).toHaveBeenCalledWith(
      { matches: [] },
      expect.objectContaining({
        format: "text",
      }),
    );
  });

  it("registers exec with explicit tool argument and payload options", () => {
    const command = program.commands.find((candidate) => candidate.name() === "mcp");
    const execCommand = command?.commands.find((candidate) => candidate.name() === "exec");

    expect(execCommand?.registeredArguments?.[0]?.name()).toBe("tool");
    expect(execCommand?.options.find((option) => option.long === "--data")).toBeDefined();
    expect(execCommand?.options.find((option) => option.long === "--file")).toBeDefined();
  });

  it("rejects mcp schema when no tool names are provided", async () => {
    await expect(program.parseAsync(["node", "test", "mcp", "schema"])).rejects.toThrow(
      "missing required argument 'toolNames'",
    );

    expect(mockServices.mcp.callTool).not.toHaveBeenCalled();
    expect(mockServices.output.render).not.toHaveBeenCalled();
  });
});
