import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "b8374f5531cec0a3cd7d531e6f989b6d64937693";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
