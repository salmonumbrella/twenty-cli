import { checkUpstreamDrift } from "./lib/upstream-drift.mjs";

const AUDIT_SHA = "563e27831bfd90af9c1ded2c0a945c588a5ad46c";

const result = await checkUpstreamDrift({
  auditSha: AUDIT_SHA,
  token: process.env.GITHUB_TOKEN,
});

console.log(JSON.stringify(result));
