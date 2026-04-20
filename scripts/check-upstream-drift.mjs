import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "ade55e293f743b961110874bbaa540c7c7388f4e";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
