import path from "node:path";
import { pathToFileURL } from "node:url";
import { describe, expect, it } from "vitest";

const repoRoot = path.resolve(__dirname, "../../../../../");

async function loadUpstreamDriftModule() {
  return import(pathToFileURL(path.join(repoRoot, "scripts", "lib", "upstream-drift.mjs")).href);
}

describe("upstream drift classification", () => {
  it("supports an inline audited sha without reading an audit file", async () => {
    const { checkUpstreamDrift } = await loadUpstreamDriftModule();
    const auditSha = "34b81adce805a042321658783f1e349e1c471b22";

    const result = await checkUpstreamDrift({
      auditSha,
      fetchImpl: async () =>
        ({
          ok: true,
          json: async () => ({ sha: auditSha }),
        }) as Response,
    });

    expect(result).toEqual({
      auditSha,
      latestSha: auditSha,
      relevantFiles: [],
      status: "current",
    });
  });

  it("ignores upstream churn that does not touch audited API surface", async () => {
    const { classifyRelevantUpstreamChanges } = await loadUpstreamDriftModule();

    const relevantFiles = classifyRelevantUpstreamChanges([
      ".github/workflows/ci-test-docker-compose.yaml",
      "packages/twenty-docs/developers/extend/apps/getting-started.mdx",
      "packages/twenty-front/src/modules/ai/components/internal/AIChatEditorFocusEffect.tsx",
      "packages/twenty-sdk/src/cli/commands/build.ts",
      "packages/twenty-server/scripts/ai-sync-models-dev.ts",
    ]);

    expect(relevantFiles).toEqual([]);
  });

  it("flags upstream files that can change the public API surface", async () => {
    const { classifyRelevantUpstreamChanges } = await loadUpstreamDriftModule();

    const relevantFiles = classifyRelevantUpstreamChanges([
      "packages/twenty-server/src/engine/api/common/common-args-processors/data-arg-processor/data-arg-processor.service.ts",
      "packages/twenty-server/src/engine/metadata-modules/ai/ai-agent/agent.resolver.ts",
      "packages/twenty-server/src/modules/dashboard/controllers/dashboard.controller.ts",
      "packages/twenty-server/src/modules/dashboard/resolvers/dashboard.resolver.ts",
    ]);

    expect(relevantFiles).toEqual([
      "packages/twenty-server/src/engine/api/common/common-args-processors/data-arg-processor/data-arg-processor.service.ts",
      "packages/twenty-server/src/engine/metadata-modules/ai/ai-agent/agent.resolver.ts",
      "packages/twenty-server/src/modules/dashboard/controllers/dashboard.controller.ts",
      "packages/twenty-server/src/modules/dashboard/resolvers/dashboard.resolver.ts",
    ]);
  });
});
