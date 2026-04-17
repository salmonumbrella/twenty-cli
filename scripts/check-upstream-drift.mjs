import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "3268a86f4b75da702636a5abce3f620642f802ed";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
