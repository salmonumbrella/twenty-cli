import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "7f1814805deb332ff69a42370cbcba9b07812d5e";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
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
