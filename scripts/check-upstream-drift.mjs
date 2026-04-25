import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "41571ea37766326b4134118d333860114f5b8b86";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
