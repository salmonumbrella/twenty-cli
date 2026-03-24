import { readFileSync } from "node:fs";
import path from "node:path";

const ROOT = path.resolve(new URL("..", import.meta.url).pathname);
const AUDIT_PATH = path.join(ROOT, "plans", "2026-03-21-twenty-api-coverage-audit.md");

function readAuditReference() {
  const contents = readFileSync(AUDIT_PATH, "utf8");
  const match = contents.match(/Upstream reference:\s+`twentyhq\/twenty`\s+`([0-9a-f]{40})`/);

  if (!match) {
    throw new Error(`Could not find upstream reference in ${AUDIT_PATH}`);
  }

  return match[1];
}

async function fetchLatestSha() {
  const headers = {
    Accept: "application/vnd.github+json",
    "User-Agent": "twenty-cli-upstream-drift-check",
  };

  if (process.env.GITHUB_TOKEN) {
    headers.Authorization = `Bearer ${process.env.GITHUB_TOKEN}`;
  }

  const response = await fetch("https://api.github.com/repos/twentyhq/twenty/commits/main", {
    headers,
  });

  if (!response.ok) {
    throw new Error(`GitHub API request failed: ${response.status} ${response.statusText}`);
  }

  const body = await response.json();
  return body.sha;
}

const auditSha = readAuditReference();
const latestSha = await fetchLatestSha();

if (latestSha !== auditSha) {
  console.error(
    `Upstream drift detected: audit references ${auditSha}, but twentyhq/twenty main is ${latestSha}.`,
  );
  process.exit(1);
}

console.log(`Upstream reference is current at ${auditSha}.`);
