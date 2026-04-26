import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "499067ae1429b8fbcb8e0547a8800f7e74cd47b8";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
