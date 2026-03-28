import { readFile, writeFile } from "node:fs/promises";

const newSha = process.argv[2];

if (!/^[0-9a-f]{40}$/.test(newSha ?? "")) {
  console.error("Usage: node scripts/bump-upstream-audit-sha.mjs <40-char-hex-sha>");
  process.exit(1);
}

const filePath = new URL("./check-upstream-drift.mjs", import.meta.url).pathname;
const contents = await readFile(filePath, "utf8");
const pattern = /const AUDIT_SHA = "[0-9a-f]{40}";/;

if (!pattern.test(contents)) {
  console.error("Failed to find AUDIT_SHA pattern in check-upstream-drift.mjs");
  process.exit(1);
}

const updated = contents.replace(pattern, `const AUDIT_SHA = "${newSha}";`);

if (updated !== contents) {
  await writeFile(filePath, updated);
}
console.log(`Bumped AUDIT_SHA to ${newSha}`);
