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
    const upstreamDriftScript = readRepoFile("scripts", "check-upstream-drift.mjs");
    const rootPackage = JSON.parse(readRepoFile("package.json")) as {
      packageManager?: string;
      scripts?: Record<string, string>;
    };
    const sdkPackage = JSON.parse(readRepoFile("packages", "twenty-sdk", "package.json")) as {
      private?: boolean;
      files?: string[];
      dependencies?: Record<string, string>;
      pkg?: {
        scripts?: string[];
      };
    };
    const ciWorkflow = readRepoFile(".github", "workflows", "ci.yml");
    const liveSmokeWorkflow = readRepoFile(".github", "workflows", "live-smoke.yml");
    const releaseWorkflow = readRepoFile(".github", "workflows", "release.yml");
    const driftWorkflow = readRepoFile(".github", "workflows", "upstream-drift.yml");

    expect(rootPackage.packageManager).toMatch(/^pnpm@/);
    expect(rootPackage.scripts).toMatchObject({
      "check:upstream-drift": "node scripts/check-upstream-drift.mjs",
      "check:repo-hygiene": "node scripts/check-repo-hygiene.mjs",
      "readme:generate": "node scripts/render-readme-snippets.mjs",
      "release:build": "node scripts/build-release.mjs",
      "release:metadata": "node scripts/write-release-metadata.mjs",
      "test:smoke:live": "pnpm --filter twenty-sdk test:e2e:live",
      "verify:ci":
        "pnpm build && pnpm test && pnpm test:e2e && pnpm check:repo-hygiene && pnpm exec prek run --all-files",
    });
    expect(sdkPackage.private).toBe(false);
    expect(sdkPackage.files).toEqual(expect.arrayContaining(["dist/**/*"]));
    expect(sdkPackage.dependencies?.["form-data"]).toBeDefined();
    expect(sdkPackage.pkg?.scripts).toEqual(
      expect.arrayContaining(["node_modules/axios/dist/node/axios.cjs"]),
    );
    expect(existsSync(path.join(repoRoot, "packages", "twenty-sdk", "package-lock.json"))).toBe(
      false,
    );

    expect(ciWorkflow).toContain("pnpm/action-setup");
    expect(ciWorkflow).not.toContain("npm ci");
    expect(ciWorkflow).toContain("permissions:");
    expect(ciWorkflow).toContain("contents: read");
    expect(ciWorkflow).toContain("pnpm verify:ci");

    expect(liveSmokeWorkflow).toContain("workflow_run:");
    expect(liveSmokeWorkflow).toContain("workflow_dispatch:");
    expect(liveSmokeWorkflow).toContain("workflow_call:");
    expect(liveSmokeWorkflow).toContain("base_url:");
    expect(liveSmokeWorkflow).toContain("workspace:");
    expect(liveSmokeWorkflow).toContain("TWENTY_LIVE_TOKEN");
    expect(liveSmokeWorkflow).toContain("github.event.workflow_run.head_sha");
    expect(liveSmokeWorkflow).toContain("concurrency:");
    expect(liveSmokeWorkflow).toContain("environment: live-smoke");
    expect(liveSmokeWorkflow).toContain("TWENTY_TOKEN: ${{ secrets.TWENTY_LIVE_TOKEN }}");
    expect(liveSmokeWorkflow).not.toContain("secrets:\n      TWENTY_LIVE_TOKEN:");

    expect(releaseWorkflow).toContain("actions/setup-node");
    expect(releaseWorkflow).not.toContain("setup-go");
    expect(releaseWorkflow).not.toContain("goreleaser");
    expect(releaseWorkflow).toContain("uses: ./.github/workflows/live-smoke.yml");
    expect(releaseWorkflow).toContain("secrets: inherit");
    expect(releaseWorkflow).toContain("needs: [verify, live-smoke]");
    expect(releaseWorkflow).toContain("contents: read");
    expect(releaseWorkflow).toContain("contents: write");
    expect(releaseWorkflow).toContain("pnpm release:build");
    expect(releaseWorkflow).toContain("node scripts/smoke-release-artifact.mjs");
    expect(releaseWorkflow).toContain("actions/upload-artifact@v7");
    expect(releaseWorkflow).toContain("actions/download-artifact@v8");
    expect(releaseWorkflow).toContain("checksums.txt");
    expect(releaseWorkflow).toContain("HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}");
    expect(releaseWorkflow).toContain("if: ${{ env.HOMEBREW_TAP_TOKEN != '' }}");

    expect(driftWorkflow).toContain("pnpm --silent check:upstream-drift");
    expect(upstreamDriftScript).toContain('const AUDIT_SHA = "');
    expect(upstreamDriftScript).not.toContain('"plans"');
    expect(upstreamDriftScript).not.toContain(".md");
    expect(upstreamDriftScript).not.toContain(".sha");
    expect(upstreamDriftScript).not.toContain("path.join(ROOT");

    expect(readme).toContain("<!-- GENERATED:INSTALL_AND_AGENT_CONTRACT:START -->");
    expect(readme).toContain("<!-- GENERATED:INSTALL_AND_AGENT_CONTRACT:END -->");
    expect(readme).toContain("Homebrew formula");
    expect(readme).toContain("twenty --help-json");
    expect(readme).toContain("jsonl");
    expect(readme).toContain("agent");
    expect(readme).not.toContain("npm install -g twenty-sdk");
  });
});
