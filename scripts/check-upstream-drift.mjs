import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "1df99424167579a76b092231fcc234e2027af658";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
