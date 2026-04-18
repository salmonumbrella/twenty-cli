import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "fd495ee61bf40969ee40af7f64b13806f4261c60";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
