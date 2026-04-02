import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "634a5e9599795b1211d00afa3bbacda1cc8f1c2d";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
