import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "81fc9607125a6ade1a4e1f82b917913f4b9fddf4";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
