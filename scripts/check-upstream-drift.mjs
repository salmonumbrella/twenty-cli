import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "6e23ca35e643154c9e20ed8ba1c010f662230e55";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
