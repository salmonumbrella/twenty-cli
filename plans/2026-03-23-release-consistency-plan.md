# Twenty CLI Release Consistency Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the stale Go-shaped release path with a Node-native release pipeline, unify CI around the repo's pnpm workflow, and add lightweight contract checks so packaging, docs, and upstream drift stay honest.

**Architecture:** Keep the CLI implementation in `packages/twenty-sdk`, but move release orchestration to root-level scripts and workflows. Build standalone `twenty` executables with `@yao-pkg/pkg` so GitHub releases and Homebrew still receive the four platform archives the tap generator expects. Add repo-contract tests and generation scripts so README/install/help claims stay aligned with runtime behavior.

**Tech Stack:** TypeScript, Node.js 20, pnpm workspace, Vitest, GitHub Actions, `@yao-pkg/pkg`

---

### Task 1: Add Release And Docs Contract Tests

**Files:**

- Create: `packages/twenty-sdk/src/cli/__tests__/release-contract.spec.ts`
- Modify: `packages/twenty-sdk/package.json`
- Test: `packages/twenty-sdk/src/cli/__tests__/release-contract.spec.ts`

- [ ] **Step 1: Write the failing repo-contract test**

Add a Vitest file that reads the root `README.md`, root `package.json`, package `package.json`, and workflow YAML files. Assert:

- the package is not marked `private`
- CI references `pnpm`, not `npm ci`
- release workflow references Node, not GoReleaser
- install docs do not claim a publish path the package metadata cannot satisfy

- [ ] **Step 2: Run the targeted test and verify it fails**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/__tests__/release-contract.spec.ts`
Expected: FAIL because the package is still private and the workflows still use npm/GoReleaser.

- [ ] **Step 3: Add a second failing docs-sync assertion**

Extend the same test file so it also checks for README markers around the generated install/help contract section. This should fail until the README is normalized.

- [ ] **Step 4: Re-run the targeted test**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/__tests__/release-contract.spec.ts`
Expected: FAIL with both release-contract and README-marker failures.

- [ ] **Step 5: Commit**

```bash
git add packages/twenty-sdk/src/cli/__tests__/release-contract.spec.ts packages/twenty-sdk/package.json
git commit -m "test: add release consistency contract coverage"
```

### Task 2: Make The Package Releasable And Build Standalone Artifacts

**Files:**

- Create: `scripts/build-release.mjs`
- Create: `scripts/write-release-metadata.mjs`
- Modify: `package.json`
- Modify: `packages/twenty-sdk/package.json`
- Modify: `.gitignore`
- Test: `packages/twenty-sdk/src/cli/__tests__/release-contract.spec.ts`

- [ ] **Step 1: Write the failing package-layout assertion**

Extend `release-contract.spec.ts` with expectations for:

- `packages/twenty-sdk/package.json` has `files`
- a release script exists in the root manifest
- release output naming follows `twenty_<version>_<os>_<arch>.tar.gz`

- [ ] **Step 2: Run the targeted test and verify it fails**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/__tests__/release-contract.spec.ts`
Expected: FAIL because there is no release build script or package file whitelist yet.

- [ ] **Step 3: Implement minimal release packaging**

Add:

- root scripts for `release:build` and `release:metadata`
- `@yao-pkg/pkg` as a dev dependency
- a package `files` whitelist and package metadata needed for packaging
- `scripts/build-release.mjs` to:
  - build the CLI
  - produce `twenty` standalone executables for `darwin_arm64`, `darwin_amd64`, `linux_arm64`, and `linux_amd64`
  - archive them under the naming contract Homebrew expects
  - emit `checksums.txt`
- `scripts/write-release-metadata.mjs` to derive the version from the git tag and write a small JSON metadata file for the workflow dispatch step

- [ ] **Step 4: Verify the targeted test passes**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/__tests__/release-contract.spec.ts`
Expected: PASS.

- [ ] **Step 5: Smoke the release builder locally**

Run: `pnpm release:build -- --version 0.0.0-test --targets linux-x64`
Expected: PASS and create a single archive plus checksum entry under `dist/release/`.

- [ ] **Step 6: Commit**

```bash
git add package.json packages/twenty-sdk/package.json .gitignore scripts/build-release.mjs scripts/write-release-metadata.mjs packages/twenty-sdk/src/cli/__tests__/release-contract.spec.ts
git commit -m "feat: add node-native release packaging"
```

### Task 3: Unify CI And Replace GoReleaser With A Node Release Workflow

**Files:**

- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/release.yml`
- Modify: `README.md`
- Test: `packages/twenty-sdk/src/cli/__tests__/release-contract.spec.ts`

- [ ] **Step 1: Write the failing workflow assertions**

Extend `release-contract.spec.ts` so it expects:

- CI uses `pnpm/action-setup` plus `actions/setup-node` with pnpm cache
- CI runs root `pnpm build`, `pnpm test`, and `pnpm exec prek run --all-files`
- release workflow uses Node and the new release scripts
- release workflow uploads archives and `checksums.txt`

- [ ] **Step 2: Run the targeted test and verify it fails**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/__tests__/release-contract.spec.ts`
Expected: FAIL because the workflows are still npm/Go-based.

- [ ] **Step 3: Implement the workflow changes**

Update:

- `ci.yml` to run from repo root with pnpm and full verification
- `release.yml` to:
  - set up pnpm and Node 20
  - install dependencies from the workspace root
  - run the release builder
  - create a GitHub release or upload assets to the tag release
  - dispatch the Homebrew formula update with the same deterministic archive contract used by the tap generator

- [ ] **Step 4: Make installation docs honest**

Normalize the README install section to match the new consistent state:

- describe repo-local pnpm setup for contributors
- keep public install instructions aligned with package metadata
- note that Homebrew consumes GitHub release archives, not Go binaries

- [ ] **Step 5: Verify the targeted test passes**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/__tests__/release-contract.spec.ts`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add .github/workflows/ci.yml .github/workflows/release.yml README.md packages/twenty-sdk/src/cli/__tests__/release-contract.spec.ts
git commit -m "chore: replace goreleaser workflows with node release"
```

### Task 4: Add Lightweight Upstream Drift Detection

**Files:**

- Create: `scripts/check-upstream-drift.mjs`
- Create: `.github/workflows/upstream-drift.yml`
- Modify: `plans/2026-03-21-twenty-api-coverage-audit.md`
- Test: `packages/twenty-sdk/src/cli/__tests__/release-contract.spec.ts`

- [ ] **Step 1: Write the failing drift assertion**

Extend `release-contract.spec.ts` so it expects:

- a drift workflow exists
- the audit file exposes an upstream commit reference
- the drift script reads that reference and compares it to current upstream `twentyhq/twenty`

- [ ] **Step 2: Run the targeted test and verify it fails**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/__tests__/release-contract.spec.ts`
Expected: FAIL because the drift script and workflow do not exist yet.

- [ ] **Step 3: Implement the minimal drift check**

Add:

- `scripts/check-upstream-drift.mjs` that parses the audit's upstream reference and compares it to the latest `twentyhq/twenty` default-branch commit via the GitHub API
- a scheduled/manual workflow that runs the script and fails loudly when upstream has moved

- [ ] **Step 4: Verify the targeted test passes**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/__tests__/release-contract.spec.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add scripts/check-upstream-drift.mjs .github/workflows/upstream-drift.yml plans/2026-03-21-twenty-api-coverage-audit.md packages/twenty-sdk/src/cli/__tests__/release-contract.spec.ts
git commit -m "chore: add upstream drift checks"
```

### Task 5: Generate The README Install And Agent Contract Section

**Files:**

- Create: `scripts/render-readme-snippets.mjs`
- Modify: `README.md`
- Modify: `packages/twenty-sdk/src/cli/help.ts`
- Test: `packages/twenty-sdk/src/cli/__tests__/release-contract.spec.ts`
- Test: `packages/twenty-sdk/src/cli/__tests__/help.spec.ts`

- [ ] **Step 1: Write the failing generation assertion**

Extend `release-contract.spec.ts` to require explicit README markers, and add or update a help test that confirms the generated snippet still mentions `--help-json`, `--hj`, `agent`, `jsonl`, and stable exit codes.

- [ ] **Step 2: Run targeted tests and verify they fail**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/__tests__/release-contract.spec.ts src/cli/__tests__/help.spec.ts`
Expected: FAIL until the generator and markers exist.

- [ ] **Step 3: Implement the generator**

Add a small root script that:

- renders the install/agent-contract block from one source of truth
- rewrites only the marked README section
- keeps the generated content stable so tests can assert it

- [ ] **Step 4: Update the README markers and content**

Replace the hand-maintained section with generated markers and render the fresh output.

- [ ] **Step 5: Verify targeted tests pass**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/__tests__/release-contract.spec.ts src/cli/__tests__/help.spec.ts`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add scripts/render-readme-snippets.mjs README.md packages/twenty-sdk/src/cli/help.ts packages/twenty-sdk/src/cli/__tests__/release-contract.spec.ts packages/twenty-sdk/src/cli/__tests__/help.spec.ts
git commit -m "docs: generate install and agent contract snippets"
```

### Task 6: Final Verification

**Files:**

- Review all modified files

- [ ] **Step 1: Run the full local verification suite**

Run: `pnpm test`
Expected: PASS.

- [ ] **Step 2: Run formatting, linting, and typecheck**

Run: `pnpm exec prek run --all-files`
Expected: PASS.

- [ ] **Step 3: Run end-to-end tests**

Run: `pnpm test:e2e`
Expected: PASS.

- [ ] **Step 4: Smoke the compiled CLI and release builder**

Run:

- `node packages/twenty-sdk/dist/cli/cli.js --help-json`
- `node packages/twenty-sdk/dist/cli/cli.js auth status -o agent`
- `pnpm release:build -- --version 0.0.0-test --targets linux-x64`

Expected: PASS and a release archive/checksum written under `dist/release/`.

- [ ] **Step 5: Verify git cleanliness**

Run: `git status --short`
Expected: only intentional tracked changes before the final commit, then clean after commit.

- [ ] **Step 6: Commit**

```bash
git add .
git commit -m "chore: align release, ci, and docs contracts"
```
