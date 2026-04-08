import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "3306d66f5be85d806ffb6402d4de7e1ababb1ff9";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
