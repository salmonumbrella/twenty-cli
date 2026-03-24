import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";

const repoRoot = path.resolve(__dirname, "../../../../../");

function readRepoFile(...segments: string[]) {
  return readFileSync(path.join(repoRoot, ...segments), "utf8");
}

describe("repo release consistency", () => {
  it("keeps package metadata, workflows, and install docs aligned", () => {
    const readme = readRepoFile("README.md");
    const rootPackage = JSON.parse(readRepoFile("package.json")) as {
      packageManager?: string;
      scripts?: Record<string, string>;
    };
    const sdkPackage = JSON.parse(readRepoFile("packages", "twenty-sdk", "package.json")) as {
      private?: boolean;
      files?: string[];
      dependencies?: Record<string, string>;
    };
    const ciWorkflow = readRepoFile(".github", "workflows", "ci.yml");
    const releaseWorkflow = readRepoFile(".github", "workflows", "release.yml");
    const driftWorkflow = readRepoFile(".github", "workflows", "upstream-drift.yml");

    expect(rootPackage.packageManager).toMatch(/^pnpm@/);
    expect(rootPackage.scripts).toMatchObject({
      "check:upstream-drift": "node scripts/check-upstream-drift.mjs",
      "readme:generate": "node scripts/render-readme-snippets.mjs",
      "release:build": "node scripts/build-release.mjs",
      "release:metadata": "node scripts/write-release-metadata.mjs",
    });
    expect(sdkPackage.private).toBe(false);
    expect(sdkPackage.files).toEqual(expect.arrayContaining(["dist/**/*"]));
    expect(sdkPackage.dependencies?.["form-data"]).toBeDefined();
    expect(existsSync(path.join(repoRoot, "packages", "twenty-sdk", "package-lock.json"))).toBe(
      false,
    );

    expect(ciWorkflow).toContain("pnpm/action-setup");
    expect(ciWorkflow).not.toContain("npm ci");
    expect(ciWorkflow).toContain("pnpm exec prek run --all-files");

    expect(releaseWorkflow).toContain("actions/setup-node");
    expect(releaseWorkflow).not.toContain("setup-go");
    expect(releaseWorkflow).not.toContain("goreleaser");
    expect(releaseWorkflow).toContain("pnpm release:build");
    expect(releaseWorkflow).toContain("actions/upload-artifact@v4");
    expect(releaseWorkflow).toContain("checksums.txt");

    expect(driftWorkflow).toContain("pnpm check:upstream-drift");

    expect(readme).toContain("<!-- GENERATED:INSTALL_AND_AGENT_CONTRACT:START -->");
    expect(readme).toContain("<!-- GENERATED:INSTALL_AND_AGENT_CONTRACT:END -->");
    expect(readme).toContain("Homebrew formula");
    expect(readme).toContain("twenty --help-json");
    expect(readme).toContain("jsonl");
    expect(readme).toContain("agent");
    expect(readme).not.toContain("npm install -g twenty-sdk");
  });
});
