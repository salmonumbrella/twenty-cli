import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "c0fef0be08567fff930bd2d2a5ded3e12e3c35b5";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
