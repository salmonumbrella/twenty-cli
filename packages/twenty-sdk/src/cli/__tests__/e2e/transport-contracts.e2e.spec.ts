import { describe, expect, it } from "vitest";
import fs from "node:fs";
import path from "node:path";
import { runCliWithTempHome } from "./helpers/temp-home";

function resolveCliPath(): string {
  return path.resolve(__dirname, "../../../../dist/cli/cli.js");
}

const cliPath = resolveCliPath();
const canRun = fs.existsSync(cliPath);
const describeIf = canRun ? describe : describe.skip;

describeIf("twenty clean-home transport contracts", () => {
  it("openapi core still requires auth in the clean-home red state", () => {
    const result = runCliWithTempHome(["openapi", "core"]);

    expect(result.exitCode).toBe(3);
    expect(result.stderr).toContain("Missing API token.");
    expect(result.stderr).toContain("~/.twenty/config.json");
  });

  it("auth discover does not fail with Missing API token", () => {
    const result = runCliWithTempHome(["auth", "discover", "https://acme.twenty.com"]);

    expect(result.stderr).not.toContain("Missing API token.");
    expect(result.exitCode).not.toBe(3);
  });

  it("files download does not fail with Missing API token", () => {
    const result = runCliWithTempHome([
      "files",
      "download",
      "https://api.twenty.com/file/files-field/file-123?token=signed-token",
    ]);

    expect(result.stderr).not.toContain("Missing API token.");
    expect(result.exitCode).not.toBe(3);
  });

  it("files public-asset does not fail with Missing API token", () => {
    const result = runCliWithTempHome([
      "files",
      "public-asset",
      "images/logo.svg",
      "--workspace-id",
      "ws-123",
      "--application-id",
      "app-123",
    ]);

    expect(result.stderr).not.toContain("Missing API token.");
    expect(result.exitCode).not.toBe(3);
  });
});
