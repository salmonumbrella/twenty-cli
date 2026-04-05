import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "ed912ec5482ff462af38f6fba226a5a263f85006";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
