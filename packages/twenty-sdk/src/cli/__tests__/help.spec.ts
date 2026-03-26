import fs from "fs-extra";
import { Command } from "commander";
import { describe, expect, it, vi } from "vitest";
import { buildProgram } from "../program";
import { buildHelpJson, maybeHandleInlineHelp } from "../help";

describe("CLI help contracts", () => {
  it("builds root help JSON with top-level commands", () => {
    const program = buildProgram();

    const help = buildHelpJson(program, []);

    expect(help.kind).toBe("root");
    expect(help.name).toBe("twenty");
    expect(help.aliases).toEqual([]);
    expect(help.subcommands.some((command) => command.name === "raw")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "rest")).toBe(false);
    expect(help.subcommands.some((command) => command.name === "graphql")).toBe(false);
    expect(help.subcommands.some((command) => command.name === "skills")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "auth")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "dashboards")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "public-domains")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "emailing-domains")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "event-logs")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "postgres-proxy")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "openapi")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "routes")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "mcp")).toBe(true);
    expect(help.exit_codes).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ code: 0 }),
        expect.objectContaining({ code: 5 }),
      ]),
    );
  });

  it("includes mcp in the root help JSON contract", () => {
    const help = buildHelpJson(buildProgram(), []);

    expect(help.subcommands.some((command) => command.name === "mcp")).toBe(true);
  });

  it("builds raw help JSON for the escape-hatch namespace", () => {
    const help = buildHelpJson(buildProgram(), ["raw", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "raw"]);
    expect(help.summary).toBe("Escape-hatch raw API commands");
    expect(help.subcommands.map((command) => command.name)).toEqual(
      expect.arrayContaining(["rest", "graphql"]),
    );
    expect(help.examples).toEqual(
      expect.arrayContaining([
        "twenty raw graphql query --document 'query { currentWorkspace { id } }'",
        "twenty raw rest GET /health",
      ]),
    );
  });

  it("builds command help JSON for operation-style commands", () => {
    const program = buildProgram();

    const help = buildHelpJson(program, ["roles", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "roles"]);
    expect(help.operations.map((operation) => operation.name)).toEqual(
      expect.arrayContaining(["list", "get", "assign-agent", "remove-agent"]),
    );
    expect(help.options.some((option) => option.name === "output" && option.global)).toBe(true);
  });

  it("builds command help JSON for nested subcommands", () => {
    const program = buildProgram();

    const help = buildHelpJson(program, ["auth", "list", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "auth", "list"]);
    expect(help.name).toBe("list");
    expect(help.summary).toBe("List configured workspaces");
  });

  it("builds command help JSON for route invocation subcommands", () => {
    const program = buildProgram();

    const help = buildHelpJson(program, ["routes", "invoke", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "routes", "invoke"]);
    expect(help.args.map((argument) => argument.name)).toEqual(["routePath"]);
    expect(help.options.some((option) => option.name === "header")).toBe(true);
    expect(help.options.some((option) => option.name === "output" && option.global)).toBe(true);
    expect(help.output_contract).toEqual(
      expect.objectContaining({
        query_language: "JMESPath",
        query_applies_before_format: true,
        formats: expect.arrayContaining([
          expect.objectContaining({ name: "jsonl" }),
          expect.objectContaining({ name: "agent" }),
        ]),
      }),
    );
  });

  it("builds command help JSON for mcp", () => {
    const help = buildHelpJson(buildProgram(), ["mcp", "--help-json"]);

    expect(help.path).toEqual(["twenty", "mcp"]);
    expect(help.operations.map((operation) => operation.name)).toEqual(
      expect.arrayContaining(["status", "catalog", "learn", "call", "load-skills", "help-center"]),
    );
    expect(help.operations.find((operation) => operation.name === "call")).toEqual(
      expect.objectContaining({
        summary: "Call an MCP tool directly",
        mutates: true,
      }),
    );
  });

  it("describes the mcp call escape hatch options", () => {
    const help = buildHelpJson(buildProgram(), ["mcp", "call", "--help-json"]);

    expect(help.options.some((option) => option.name === "data")).toBe(true);
    expect(help.options.some((option) => option.name === "file")).toBe(true);
    expect(help.options.some((option) => option.name === "args")).toBe(false);
    expect(help.options.some((option) => option.name === "args-file")).toBe(false);
  });

  it("builds command help JSON for mcp help-center", () => {
    const help = buildHelpJson(buildProgram(), ["mcp", "help-center", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "mcp", "help-center"]);
    expect(help.name).toBe("help-center");
    expect(help.args.map((argument) => argument.name)).toEqual(["query"]);
  });

  it("builds command help JSON for approved access domain admin commands", () => {
    const help = buildHelpJson(buildProgram(), ["approved-access-domains", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "approved-access-domains"]);
    expect(help.operations).toEqual([
      expect.objectContaining({ name: "list", mutates: false }),
      expect.objectContaining({ name: "delete", mutates: true }),
      expect.objectContaining({ name: "validate", mutates: true }),
    ]);
  });

  it("builds command help JSON for dashboard duplication", () => {
    const program = buildProgram();

    const help = buildHelpJson(program, ["dashboards", "duplicate", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "dashboards", "duplicate"]);
    expect(help.args.map((argument) => argument.name)).toEqual(["dashboardId"]);
    expect(help.examples).toContain("twenty dashboards duplicate <dashboard-id>");
    expect(help.options.some((option) => option.name === "output" && option.global)).toBe(true);
  });

  it("builds command help JSON for public domain admin commands", () => {
    const program = buildProgram();

    const help = buildHelpJson(program, ["public-domains", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "public-domains"]);
    expect(help.operations).toEqual([
      expect.objectContaining({ name: "list", mutates: false }),
      expect.objectContaining({ name: "create", mutates: true }),
      expect.objectContaining({ name: "delete", mutates: true }),
      expect.objectContaining({ name: "check-records", mutates: false }),
    ]);
    expect(help.examples).toContain("twenty public-domains check-records --domain app.example.com");
  });

  it("builds command help JSON for emailing domain admin commands", () => {
    const program = buildProgram();

    const help = buildHelpJson(program, ["emailing-domains", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "emailing-domains"]);
    expect(help.operations).toEqual([
      expect.objectContaining({ name: "list", mutates: false }),
      expect.objectContaining({ name: "create", mutates: true }),
      expect.objectContaining({ name: "verify", mutates: true }),
      expect.objectContaining({ name: "delete", mutates: true }),
    ]);
    expect(help.examples).toContain("twenty emailing-domains create --domain mail.example.com");
  });

  it("uses canonical pagination naming in help JSON for search and event logs", () => {
    const searchHelp = buildHelpJson(buildProgram(), ["search", "--help-json"]);
    const eventLogsHelp = buildHelpJson(buildProgram(), ["event-logs", "list", "--help-json"]);

    expect(searchHelp.options.some((option) => option.name === "cursor")).toBe(true);
    expect(searchHelp.options.some((option) => option.name === "after")).toBe(false);
    expect(searchHelp.options.some((option) => option.name === "filter-file")).toBe(true);

    expect(eventLogsHelp.options.some((option) => option.name === "limit")).toBe(true);
    expect(eventLogsHelp.options.some((option) => option.name === "cursor")).toBe(true);
    expect(eventLogsHelp.options.some((option) => option.name === "first")).toBe(false);
    expect(eventLogsHelp.options.some((option) => option.name === "after")).toBe(false);
  });

  it("documents canonical destructive confirmation and payload vocabulary", () => {
    const apiDeleteHelp = buildHelpJson(buildProgram(), ["api", "delete", "--help-json"]);
    const apiListHelp = buildHelpJson(buildProgram(), ["api", "list", "--help-json"]);
    const mcpHelp = buildHelpJson(buildProgram(), ["mcp", "--help-json"]);

    expect(apiDeleteHelp.options.some((option) => option.name === "yes")).toBe(true);
    expect(apiDeleteHelp.options.some((option) => option.name === "force")).toBe(false);
    expect(apiListHelp.options.some((option) => option.name === "yes")).toBe(false);

    expect(mcpHelp.examples).toEqual(
      expect.arrayContaining([
        'twenty mcp call execute_tool --data \'{"toolName":"find_companies","arguments":{}}\'',
      ]),
    );
  });

  it("builds command help JSON for skills admin commands", () => {
    const help = buildHelpJson(buildProgram(), ["skills", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "skills"]);
    expect(help.operations).toEqual([
      expect.objectContaining({ name: "list", mutates: false }),
      expect.objectContaining({ name: "get", mutates: false }),
      expect.objectContaining({ name: "create", mutates: true }),
      expect.objectContaining({ name: "update", mutates: true }),
      expect.objectContaining({ name: "delete", mutates: true }),
      expect.objectContaining({ name: "activate", mutates: true }),
      expect.objectContaining({ name: "deactivate", mutates: true }),
    ]);
    expect(help.examples).toContain("twenty skills list");
  });

  it("builds command help JSON for Postgres proxy commands", () => {
    const program = buildProgram();

    const help = buildHelpJson(program, ["postgres-proxy", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "postgres-proxy"]);
    expect(help.operations.map((operation) => operation.name)).toEqual(
      expect.arrayContaining(["get", "enable", "disable"]),
    );
    expect(help.examples).toContain("twenty postgres-proxy get --show-password");
  });

  it.each([
    {
      args: ["api-keys", "--help-json"],
      path: ["twenty", "api-keys"],
      subcommands: ["list", "get", "create", "update", "revoke", "assign-role"],
      options: ["output", "query", "workspace"],
    },
    {
      args: ["webhooks", "--help-json"],
      path: ["twenty", "webhooks"],
      subcommands: ["list", "get", "create", "update", "delete"],
      options: ["output", "query", "workspace"],
    },
    {
      args: ["route-triggers", "--help-json"],
      path: ["twenty", "route-triggers"],
      subcommands: ["list", "get", "create", "update", "delete"],
      options: ["output", "query", "workspace"],
    },
    {
      args: ["marketplace-apps", "--help-json"],
      path: ["twenty", "marketplace-apps"],
      subcommands: ["list", "get", "install"],
      options: ["output", "query", "workspace"],
    },
    {
      args: ["postgres-proxy", "--help-json"],
      path: ["twenty", "postgres-proxy"],
      subcommands: ["get", "enable", "disable"],
      options: ["output", "query", "workspace"],
    },
  ])(
    "builds parent help JSON for $path[1] explicit subcommands",
    ({ args, path, subcommands, options }) => {
      const help = buildHelpJson(buildProgram(), args);

      expect(help.kind).toBe("command");
      expect(help.path).toEqual(path);
      expect(help.subcommands.map((command) => command.name)).toEqual(
        expect.arrayContaining(subcommands),
      );
      expect(help.operations.map((operation) => operation.name)).toEqual(
        expect.arrayContaining(subcommands),
      );
      expect(help.options.map((option) => option.name)).toEqual(expect.arrayContaining(options));
      expect(help.capabilities.supports_output).toBe(true);
      expect(help.capabilities.supports_query).toBe(true);
      expect(help.capabilities.supports_workspace).toBe(true);
      expect(help.capabilities.mutates).toBe(true);
    },
  );

  it("builds command help JSON for event log commands", () => {
    const program = buildProgram();

    const help = buildHelpJson(program, ["event-logs", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "event-logs"]);
    expect(help.operations.map((operation) => operation.name)).toEqual(["list"]);
    expect(help.examples).toContain("twenty event-logs list --table workspace-event");
  });

  it("builds command help JSON for connected account manual mail operations", () => {
    const program = buildProgram();

    const help = buildHelpJson(program, ["connected-accounts", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "connected-accounts"]);
    expect(help.operations.map((operation) => operation.name)).toEqual(
      expect.arrayContaining([
        "list",
        "get",
        "sync",
        "get-imap-smtp-caldav",
        "save-imap-smtp-caldav",
      ]),
    );
    expect(help.examples).toContain(
      "twenty connected-accounts get-imap-smtp-caldav <connected-account-id>",
    );
  });

  it("builds command help JSON for serverless layer operations", () => {
    const program = buildProgram();

    const help = buildHelpJson(program, ["serverless", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "serverless"]);
    expect(help.operations.map((operation) => operation.name)).toEqual(
      expect.arrayContaining(["list", "create", "create-layer", "publish", "source"]),
    );
    expect(help.examples).toContain(
      "twenty serverless create-layer --package-json '{\"dependencies\":{}}' --yarn-lock 'lockfile'",
    );
  });

  it("merges operation metadata overrides for mixed command surfaces", () => {
    const program = buildProgram();

    const filesHelp = buildHelpJson(program, ["files", "--help-json"]);
    const serverlessHelp = buildHelpJson(program, ["serverless", "--help-json"]);
    const appRegistrationsHelp = buildHelpJson(program, [
      "application-registrations",
      "--help-json",
    ]);

    expect(filesHelp.operations.find((operation) => operation.name === "download")?.mutates).toBe(
      false,
    );
    expect(
      filesHelp.operations.find((operation) => operation.name === "public-asset")?.mutates,
    ).toBe(false);
    expect(filesHelp.operations.find((operation) => operation.name === "upload")?.mutates).toBe(
      true,
    );

    expect(serverlessHelp.operations.find((operation) => operation.name === "logs")?.mutates).toBe(
      false,
    );
    expect(
      serverlessHelp.operations.find((operation) => operation.name === "create-layer")?.mutates,
    ).toBe(true);

    expect(
      appRegistrationsHelp.operations.find((operation) => operation.name === "stats")?.mutates,
    ).toBe(false);
    expect(
      appRegistrationsHelp.operations.find((operation) => operation.name === "create")?.mutates,
    ).toBe(true);
  });

  it("uses readable operation summaries for agent-facing help", () => {
    const program = buildProgram();

    const filesHelp = buildHelpJson(program, ["files", "--help-json"]);
    const serverlessHelp = buildHelpJson(program, ["serverless", "--help-json"]);
    const appRegistrationsHelp = buildHelpJson(program, [
      "application-registrations",
      "--help-json",
    ]);

    expect(filesHelp.operations.find((operation) => operation.name === "download")?.summary).toBe(
      "Download a file",
    );
    expect(
      filesHelp.operations.find((operation) => operation.name === "public-asset")?.summary,
    ).toBe("Download a public asset");

    expect(serverlessHelp.operations.find((operation) => operation.name === "get")?.summary).toBe(
      "Get one serverless function",
    );
    expect(
      serverlessHelp.operations.find((operation) => operation.name === "available-packages")
        ?.summary,
    ).toBe("List available packages for a serverless function");

    expect(
      appRegistrationsHelp.operations.find((operation) => operation.name === "tarball-url")
        ?.summary,
    ).toBe("Get the tarball URL for an application registration");
    expect(
      appRegistrationsHelp.operations.find((operation) => operation.name === "create")?.summary,
    ).toBe("Create an application registration");
    expect(
      appRegistrationsHelp.operations.find((operation) => operation.name === "create-variable")
        ?.summary,
    ).toBe("Create a variable for an application registration");
  });

  it("does not expose raw or malformed operation summaries across command help", () => {
    const program = buildProgram();
    const queue = program.commands
      .filter((command) => command.name() !== "help" && !command.name().startsWith("completion"))
      .map((command) => [command.name()]);

    while (queue.length > 0) {
      const path = queue.shift()!;
      const help = buildHelpJson(program, [...path, "--help-json"]);

      for (const subcommand of help.subcommands) {
        queue.push([...path, subcommand.name]);
      }

      for (const operation of help.operations) {
        expect(operation.summary[0]).toBe(operation.summary[0]?.toUpperCase());
        expect(operation.summary).not.toBe(operation.name.replace(/-/g, " "));
        expect(operation.summary).not.toMatch(/\ba application\b/i);
      }
    }
  });

  it("uses descriptive placeholders in help examples", () => {
    const program = buildProgram();
    const queue = program.commands
      .filter((command) => command.name() !== "help" && !command.name().startsWith("completion"))
      .map((command) => [command.name()]);

    while (queue.length > 0) {
      const path = queue.shift()!;
      const help = buildHelpJson(program, [...path, "--help-json"]);

      for (const subcommand of help.subcommands) {
        queue.push([...path, subcommand.name]);
      }

      for (const example of help.examples) {
        expect(example).not.toMatch(/_123\b|abc123\b/);
        expect(example).not.toContain("owner@example.com");
        expect(example).not.toContain("user@example.com");
      }
    }
  });

  it("renders root help text when invoked with no args", async () => {
    const program = buildProgram();
    const write = vi.fn();

    const handled = await maybeHandleInlineHelp(program, [], write);

    expect(handled).toBe(true);
    expect(write).toHaveBeenCalledTimes(1);
    expect(write.mock.calls[0][0]).toContain("twenty - CLI for Twenty");
    expect(write.mock.calls[0][0]).toContain("Discovery:");
    expect(write.mock.calls[0][0]).toContain("Agent Use:");
    expect(write.mock.calls[0][0]).toContain(
      "path, args, options, operations, capabilities, exit_codes, output_contract",
    );
    expect(write.mock.calls[0][0]).toContain("Raw Access:");
    expect(write.mock.calls[0][0]).toContain("twenty raw graphql query --document");
    expect(write.mock.calls[0][0]).toContain("twenty raw rest GET /health");
    expect(write.mock.calls[0][0]).not.toContain("twenty graphql query --query");
    expect(write.mock.calls[0][0]).not.toContain("twenty rest GET /health");
    expect(write.mock.calls[0][0]).toContain("Env Precedence:");
    expect(write.mock.calls[0][0]).toContain(
      ".env then .env.local then the explicit env file; existing environment variables win",
    );
  });

  it("renders root help text when global option values precede --help", async () => {
    for (const args of [
      ["--env-file", ".env.test", "--help"],
      ["-o", "json", "--help"],
      ["--output=json", "--help"],
      ["--query", "items[0]", "--help"],
    ]) {
      const program = buildProgram();
      const write = vi.fn();

      const handled = await maybeHandleInlineHelp(program, args, write);

      expect(handled).toBe(true);
      expect(write).toHaveBeenCalledTimes(1);
      expect(write.mock.calls[0][0]).toContain("twenty - CLI for Twenty");
      expect(write.mock.calls[0][0]).toContain("Environment:");
    }
  });

  it("falls back to bundled root help text when help.txt is unavailable", async () => {
    const program = buildProgram();
    const write = vi.fn();
    const pathExistsSpy = vi.spyOn(fs, "pathExistsSync").mockReturnValue(false);

    try {
      const handled = await maybeHandleInlineHelp(program, [], write);

      expect(handled).toBe(true);
      expect(write).toHaveBeenCalledTimes(1);
      expect(write.mock.calls[0][0]).toContain(
        "Command names are canonical; only --help-json also has the short --hj alias.",
      );
      expect(write.mock.calls[0][0]).toContain(
        "Stable JSON fields: path, args, options, operations, capabilities, exit_codes, output_contract.",
      );
      expect(write.mock.calls[0][0]).not.toContain("Machine-readable root command tree");
    } finally {
      pathExistsSpy.mockRestore();
    }
  });

  it("renders command help JSON without requiring positional operations", async () => {
    const program = buildProgram();
    const write = vi.fn();

    const handled = await maybeHandleInlineHelp(program, ["roles", "--help-json"], write);

    expect(handled).toBe(true);
    expect(write).toHaveBeenCalledTimes(1);

    const payload = JSON.parse(write.mock.calls[0][0] as string) as {
      name: string;
      operations: Array<{ name: string }>;
    };

    expect(payload.name).toBe("roles");
    expect(payload.operations.map((operation) => operation.name)).toContain("list");
  });

  it("accepts the short help-json alias", async () => {
    const program = buildProgram();
    const write = vi.fn();

    const handled = await maybeHandleInlineHelp(program, ["routes", "invoke", "--hj"], write);

    expect(handled).toBe(true);
    expect(write).toHaveBeenCalledTimes(1);

    const payload = JSON.parse(write.mock.calls[0][0] as string) as {
      path: string[];
    };

    expect(payload.path).toEqual(["twenty", "routes", "invoke"]);
  });

  it("accepts boolean-style help-json flags", async () => {
    const program = buildProgram();
    const write = vi.fn();

    const handled = await maybeHandleInlineHelp(program, ["roles", "--help-json=true"], write);

    expect(handled).toBe(true);
    expect(write).toHaveBeenCalledTimes(1);

    const payload = JSON.parse(write.mock.calls[0][0] as string) as {
      name: string;
    };

    expect(payload.name).toBe("roles");
  });

  it("resolves help-json targets even when the flag appears before the command path", () => {
    const help = buildHelpJson(buildProgram(), ["--help-json", "auth", "list"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "auth", "list"]);
    expect(help.name).toBe("list");
  });

  it("serializes option details for command help JSON", () => {
    const help = buildHelpJson(buildProgram(), ["routes", "invoke", "--help-json"]);
    const methodOption = help.options.find((option) => option.name === "method");

    expect(methodOption).toEqual(
      expect.objectContaining({
        flags: "--method <method>",
        type: "string",
        default: '"get"',
        required: false,
        description: "Route method: get, post, put, patch, or delete",
      }),
    );
  });

  it("infers operations from the command argument description", () => {
    const program = new Command();
    program.name("twenty");
    program
      .command("widgets")
      .description("Inspect widgets")
      .argument("<operation>", "list, get, or archive");

    const help = buildHelpJson(program, ["widgets", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "widgets"]);
    expect(help.operations).toEqual([
      expect.objectContaining({
        name: "list",
        summary: "List widgets",
        mutates: false,
      }),
      expect.objectContaining({
        name: "get",
        summary: "Get one widget",
        mutates: false,
      }),
      expect.objectContaining({
        name: "archive",
        summary: "archive",
        mutates: true,
      }),
    ]);
  });

  it("renders only visible subcommands in help JSON", () => {
    const help = buildHelpJson(buildProgram(), ["mcp", "--help-json"]);

    expect(help.subcommands.some((command) => command.name === "help")).toBe(false);
    expect(help.subcommands.some((command) => command.name.startsWith("completion"))).toBe(false);
  });

  it("does not intercept subcommand detail help", async () => {
    const program = buildProgram();
    const write = vi.fn();

    const handled = await maybeHandleInlineHelp(program, ["auth", "list", "--help"], write);

    expect(handled).toBe(false);
    expect(write).not.toHaveBeenCalled();
  });
});
