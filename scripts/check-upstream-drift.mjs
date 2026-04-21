import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "7947b1b84383b33aa3835eaffc83838ad11dc5b2";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
