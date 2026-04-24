import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "a36c67a5004dcbf3e7849960c68625cea2e7b5f9";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
