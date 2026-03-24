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
    expect(help.subcommands.some((command) => command.name === "skills")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "auth")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "dashboards")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "public-domains")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "emailing-domains")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "event-logs")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "postgres-proxy")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "openapi")).toBe(true);
    expect(help.subcommands.some((command) => command.name === "routes")).toBe(true);
    expect(help.exit_codes).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ code: 0 }),
        expect.objectContaining({ code: 5 }),
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
    expect(help.operations.map((operation) => operation.name)).toEqual(
      expect.arrayContaining(["list", "create", "delete", "check-records"]),
    );
    expect(help.examples).toContain("twenty public-domains check-records --domain app.example.com");
  });

  it("builds command help JSON for emailing domain admin commands", () => {
    const program = buildProgram();

    const help = buildHelpJson(program, ["emailing-domains", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "emailing-domains"]);
    expect(help.operations.map((operation) => operation.name)).toEqual(
      expect.arrayContaining(["list", "create", "verify", "delete"]),
    );
    expect(help.examples).toContain("twenty emailing-domains create --domain mail.example.com");
  });

  it("builds command help JSON for Postgres proxy commands", () => {
    const program = buildProgram();

    const help = buildHelpJson(program, ["postgres-proxy", "--help-json"]);

    expect(help.kind).toBe("command");
    expect(help.path).toEqual(["twenty", "postgres-proxy"]);
    expect(help.operations.map((operation) => operation.name)).toEqual(
      expect.arrayContaining(["get", "enable", "disable"]),
    );
    expect(help.examples).toContain("twenty postgres-proxy get");
  });

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
    expect(write.mock.calls[0][0]).toContain("Env Precedence:");
    expect(write.mock.calls[0][0]).toContain(
      ".env then .env.local then the explicit env file; existing environment variables win",
    );
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

  it("does not intercept subcommand detail help", async () => {
    const program = buildProgram();
    const write = vi.fn();

    const handled = await maybeHandleInlineHelp(program, ["auth", "list", "--help"], write);

    expect(handled).toBe(false);
    expect(write).not.toHaveBeenCalled();
  });
});
