import { Command } from "commander";
import { describe, expect, it } from "vitest";
import { resolveTargetCommand } from "../../../help/command-resolution";
import { applyCommandAliases, resolveOperationAlias } from "../command-aliases";

describe("command aliases", () => {
  it("adds short aliases across static, dynamic resource, and operation commands", () => {
    const program = new Command("twenty");
    const records = program.command("records");
    const resource = records.command("approved-access-domains");
    resource.command("list");

    applyCommandAliases(program);

    expect(records.aliases()).toContain("r");
    expect(resource.aliases()).toContain("aad");
    expect(resource.commands.find((command) => command.name() === "list")?.aliases()).toContain(
      "ls",
    );
    expect(resolveTargetCommand(program, ["r", "aad", "ls"]).path).toEqual([
      "twenty",
      "records",
      "approved-access-domains",
      "list",
    ]);
  });

  it("does not add an alias that collides with a sibling command", () => {
    const program = new Command("twenty");
    const parent = program.command("parent");
    parent.command("bet");
    parent.command("buildExportTask");

    applyCommandAliases(program);

    expect(
      parent.commands.find((command) => command.name() === "buildExportTask")?.aliases(),
    ).not.toContain("bet");
  });

  it("resolves operation-style argument aliases", () => {
    expect(resolveOperationAlias("ls", ["list", "get", "update"])).toBe("list");
    expect(resolveOperationAlias("up", ["list", "get", "update"])).toBe("update");
    expect(resolveOperationAlias("aa", ["assign-agent", "remove-agent"])).toBe("assign-agent");
    expect(resolveOperationAlias("unknown", ["list"])).toBe("unknown");
  });
});
