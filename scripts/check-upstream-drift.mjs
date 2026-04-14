import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "ed9dd3c27557fb90de694fd27f95077304aafc3e";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
