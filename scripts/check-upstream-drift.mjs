import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "789f8aba5d19ae36562e0bfa5939f0d85f0777d2";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
