import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "119014f86d2eeceb64e9aea2b40f47e8fb7bc5d7";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
