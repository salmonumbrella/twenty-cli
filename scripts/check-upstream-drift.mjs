import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "3ae63f057419bc3aaffc3e82045dd7980620eb0c";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
