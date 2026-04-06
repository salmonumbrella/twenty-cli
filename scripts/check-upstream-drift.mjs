import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "aec43da1e2ba64f426a0077f9d674261976a0c2c";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
