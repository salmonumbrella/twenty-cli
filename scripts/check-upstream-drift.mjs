import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "a82d07890600ebab33f4586365e946a808e3d6c5";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
