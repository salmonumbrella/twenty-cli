import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "ec283b8f2d59e198228112eb19236606c2e5cbe1";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
