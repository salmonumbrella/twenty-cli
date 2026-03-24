import { fileURLToPath } from "node:url";
import path from "node:path";

import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const AUDIT_PATH = path.join(ROOT, "plans", "2026-03-21-twenty-api-coverage-audit.md");

const result = await checkUpstreamDrift({
  auditPath: AUDIT_PATH,
  token: process.env.GITHUB_TOKEN,
});

if (result.status === "current") {
  console.log(`Upstream reference is current at ${result.auditSha}.`);
  process.exit(0);
}

if (result.status === "non_api_drift") {
  console.log(
    `Upstream moved from ${result.auditSha} to ${result.latestSha}, but no audited API-surface paths changed.`,
  );
  process.exit(0);
}

console.error(
  `Upstream API drift detected: audit references ${result.auditSha}, but twentyhq/twenty main is ${result.latestSha}.`,
);

for (const file of result.relevantFiles) {
  console.error(`- ${file}`);
}

process.exit(1);
