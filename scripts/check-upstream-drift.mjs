import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "3b55026452f8d845cc1ed280c3bd97402e272bff";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
