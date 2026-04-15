import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "762fb6fd642e0bd4314af244e41f7796cc706a7c";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
