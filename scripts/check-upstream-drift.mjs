import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "282ee9ac42f5d6aaa7a31189289e47710e180622";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
