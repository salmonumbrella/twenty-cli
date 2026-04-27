import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";

const ROOT = path.resolve(new URL("..", import.meta.url).pathname);
const README_PATH = path.join(ROOT, "README.md");
const HELP_PATH = path.join(ROOT, "packages", "twenty-sdk", "src", "cli", "help.txt");

const START_MARKER = "<!-- GENERATED:INSTALL_AND_AGENT_CONTRACT:START -->";
const END_MARKER = "<!-- GENERATED:INSTALL_AND_AGENT_CONTRACT:END -->";

function readHelpSections() {
  const contents = readFileSync(HELP_PATH, "utf8");
  const lines = contents.split("\n");
  const sections = new Map();
  let currentSection = null;

  for (const line of lines) {
    if (!line.trim()) {
      continue;
    }

    if (!line.startsWith("  ") && line.endsWith(":")) {
      currentSection = line.slice(0, -1);
      sections.set(currentSection, []);
      continue;
    }

    if (!line.startsWith("  ")) {
      currentSection = null;
      continue;
    }

    if (currentSection) {
      sections.get(currentSection).push(line.trimEnd());
    }
  }

  return sections;
}

function bulletize(lines) {
  return lines.map((line) => `- ${line.trim()}`).join("\n");
}

function formatExitCodes(lines) {
  return ["```text", ...lines.map((line) => line.trim()), "```"].join("\n");
}

function renderSnippet() {
  const sections = readHelpSections();
  const discovery = [
    "twenty --help",
    "twenty --help-json",
    "twenty --hj",
    "twenty roles --help-json",
    "twenty routes invoke --hj",
    "twenty auth list --help-json",
  ];
  const agentUse = sections.get("Agent Use") ?? [];
  const envPrecedence = sections.get("Env Precedence") ?? [];
  const outputGuarantees = sections.get("Output Guarantees") ?? [];
  const exitCodes = sections.get("Exit Codes") ?? [];

  return [
    "## Installation",
    "",
    "Install a standalone release archive, use the Homebrew formula if your account has tap access, or build from source.",
    "",
    "```bash",
    "# Latest macOS ARM64 archive; use linux_amd64, linux_arm64, or darwin_amd64 as needed",
    "gh release download --repo salmonumbrella/twenty-cli --pattern 'twenty_*_darwin_arm64.tar.gz'",
    "tar -xzf twenty_*_darwin_arm64.tar.gz",
    "mkdir -p ~/.local/bin",
    "install -m 0755 twenty ~/.local/bin/twenty",
    "",
    "# Homebrew formula, updated by tagged releases; tap access required",
    "brew install salmonumbrella/tap/twenty-cli",
    "```",
    "",
    "Tagged releases publish standalone `twenty` archives for macOS and Linux and update the maintained Homebrew formula. When `NPM_TOKEN` is configured, releases also publish the scoped npm package.",
    "",
    "```bash",
    "# Build from source",
    "pnpm install",
    "pnpm build",
    "node packages/twenty-sdk/dist/cli/cli.js --help",
    "```",
    "",
    "## Agent Discovery",
    "",
    "The CLI ships with a curated root help contract plus machine-readable help output for agents and automation.",
    "",
    "```bash",
    ...discovery,
    "```",
    "",
    bulletize(agentUse),
    "",
    "### Environment Loading",
    "",
    `${envPrecedence.map((line) => line.trim()).join(" ")}.`,
    "",
    "### Output Guarantees",
    "",
    bulletize(outputGuarantees),
    "",
    "### Exit Codes",
    "",
    formatExitCodes(exitCodes),
  ].join("\n");
}

const readme = readFileSync(README_PATH, "utf8");
const snippet = `${START_MARKER}\n\n${renderSnippet()}\n\n${END_MARKER}`;

if (!readme.includes(START_MARKER) || !readme.includes(END_MARKER)) {
  throw new Error(`README is missing ${START_MARKER} / ${END_MARKER} markers.`);
}

const nextReadme = readme.replace(new RegExp(`${START_MARKER}[\\s\\S]*?${END_MARKER}`), snippet);

writeFileSync(README_PATH, nextReadme, "utf8");
