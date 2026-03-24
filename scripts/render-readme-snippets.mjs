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
    "Tagged releases publish standalone `twenty` archives for macOS and Linux, and the release workflow updates the maintained Homebrew formula from those archives.",
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
const snippet = `${START_MARKER}\n${renderSnippet()}\n${END_MARKER}`;

if (!readme.includes(START_MARKER) || !readme.includes(END_MARKER)) {
  throw new Error(`README is missing ${START_MARKER} / ${END_MARKER} markers.`);
}

const nextReadme = readme.replace(new RegExp(`${START_MARKER}[\\s\\S]*?${END_MARKER}`), snippet);

writeFileSync(README_PATH, nextReadme, "utf8");
