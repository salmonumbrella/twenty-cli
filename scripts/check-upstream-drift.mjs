import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "992a7ca12ffdd7007aa39029abb8525ae1c32a5f";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
