import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "36dece43c735b37d685895d2e48a89f936435bf8";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
