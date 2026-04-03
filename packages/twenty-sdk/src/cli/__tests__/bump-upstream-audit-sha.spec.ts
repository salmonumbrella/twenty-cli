import { execFileSync } from "node:child_process";
import { readFileSync, copyFileSync, unlinkSync } from "node:fs";
import path from "node:path";
import { describe, expect, it, beforeEach, afterEach } from "vitest";

const repoRoot = path.resolve(__dirname, "../../../../../");
const scriptPath = path.join(repoRoot, "scripts", "bump-upstream-audit-sha.mjs");
const targetPath = path.join(repoRoot, "scripts", "check-upstream-drift.mjs");
const backupPath = `${targetPath}.bak`;

describe("bump-upstream-audit-sha", () => {
  beforeEach(() => {
    copyFileSync(targetPath, backupPath);
  });

  afterEach(() => {
    copyFileSync(backupPath, targetPath);
    try {
      unlinkSync(backupPath);
    } catch {}
  });

  it("rejects invalid SHA arguments", () => {
    expect(() =>
      execFileSync("node", [scriptPath, "not-a-sha"], { encoding: "utf8", stdio: "pipe" }),
    ).toThrow();
  });

  it("rejects missing arguments", () => {
    expect(() => execFileSync("node", [scriptPath], { encoding: "utf8", stdio: "pipe" })).toThrow();
  });

  it("replaces the AUDIT_SHA in the target file", () => {
    const newSha = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
    const output = execFileSync("node", [scriptPath, newSha], {
      encoding: "utf8",
      stdio: "pipe",
    });

    expect(output).toContain(`Bumped AUDIT_SHA to ${newSha}`);

    const contents = readFileSync(targetPath, "utf8");
    expect(contents).toContain(`const AUDIT_SHA = "${newSha}";`);
  });
});
