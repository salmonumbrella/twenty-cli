import fs from "node:fs";
import { describe, expect, it } from "vitest";
import { resolveBuiltCliPath } from "./cli-runner";
import { runBuiltCliWithTempHome } from "./temp-home";

const cliPath = resolveBuiltCliPath();

if (!fs.existsSync(cliPath)) {
  throw new Error(
    `Missing built CLI artifact at ${cliPath}. Run "pnpm --filter twenty-sdk build" first.`,
  );
}

describe("temp home helper integration", () => {
  it("loads the full root help asset from the built CLI on the sync path", () => {
    const result = runBuiltCliWithTempHome(["--help"]);

    expect(result.exitCode).toBe(0);
    expect(result.stdout).toContain("Auth & Workspace:");
    expect(result.stdout).toContain("Environment:");
  });

  it("keeps openapi core in the clean-home auth red state on the sync path", () => {
    const result = runBuiltCliWithTempHome(["openapi", "core"]);

    expect(result.exitCode).toBe(3);
    expect(result.stderr).toContain("Missing API token.");
    expect(result.stderr).toContain("~/.twenty/config.json");
  });
});
