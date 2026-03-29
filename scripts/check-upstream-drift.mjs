import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "a3f468ef98a83a193ee783addc6dbbeb9bd84a94";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
