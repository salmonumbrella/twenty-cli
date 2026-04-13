import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "bf5cc68f253a21298cf816b77c17791fe1cc4350";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
