import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "90661cc821b19e561540e6264793865a0d94e352";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
