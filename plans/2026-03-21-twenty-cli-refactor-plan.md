# Twenty CLI Refactor Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert `twenty-cli` into a clean TypeScript workspace with proper `.env` support, `pnpm` + OXC + `prek` tooling, stable verification flows, and a tracked audit of upstream Twenty API coverage.

**Architecture:** Keep the CLI implementation in `packages/twenty-sdk`, but move repo orchestration to the root with a real package manager workspace and shared tooling. Clean the git surface by removing tracked dependency/build trash and documenting upstream coverage before expanding missing commands.

**Tech Stack:** TypeScript, Node.js, pnpm workspace, Vitest, Oxlint, Oxfmt, Prek

---

### Task 1: Add Root Workspace and Tooling

**Files:**

- Create: `package.json`
- Create: `pnpm-workspace.yaml`
- Create: `oxlint.config.ts`
- Create: `.pre-commit-config.yaml`
- Delete: `lefthook.yml`
- Modify: `.gitignore`

- [ ] **Step 1: Add root package manager manifest**

Create root scripts for `build`, `test`, `lint`, `format`, `format:check`, `setup`, and `smoke`.

- [ ] **Step 2: Add workspace declaration**

Point the root workspace at `packages/*`.

- [ ] **Step 3: Replace stale Go hook commands**

Use the requested `oxlint` and `oxfmt` flow through `@j178/prek`, adapted for recursive TypeScript repo globs and repo-level scripts.

- [ ] **Step 4: Add ignore rules**

Ignore `node_modules`, caches, coverage output, `.env`, and generated artifacts that should not be committed.

- [ ] **Step 5: Verify root tooling**

Run root-level install and confirm `pnpm` can discover the workspace and binaries.

- [ ] **Step 6: Commit**

Commit message: `chore: add pnpm workspace and prek hooks`

### Task 2: Clean Tracked Generated and Dependency Artifacts

**Files:**

- Modify: `.gitignore`
- Remove tracked content from git index under `packages/twenty-sdk/node_modules`
- Remove or normalize tracked generated build artifacts under `packages/twenty-sdk/dist`

- [ ] **Step 1: Inventory tracked garbage**

Confirm tracked `node_modules`, cache files, and stale generated files.

- [ ] **Step 2: Remove tracked dependency/cache artifacts**

Untrack `packages/twenty-sdk/node_modules` and any test-result garbage from git.

- [ ] **Step 3: Decide and implement `dist` strategy**

Either untrack `dist` and build it on demand, or make the packaging flow regenerate it deterministically. Prefer the option that stops local build drift.

- [ ] **Step 4: Verify clean build behavior**

Run `build` and `test` and confirm the repo does not acquire unexpected dirt.

- [ ] **Step 5: Commit**

Commit message: `chore: remove tracked generated artifacts`

### Task 3: Add Proper `.env` Support

**Files:**

- Create: `.env.example`
- Create or modify: runtime env loader under `packages/twenty-sdk/src/cli/utilities/config/` or `packages/twenty-sdk/src/cli/utilities/shared/`
- Modify: `packages/twenty-sdk/src/cli/cli.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/config/services/config.service.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/shared/global-options.ts`
- Test: config/auth-related tests

- [ ] **Step 1: Write failing tests**

Add tests for default `.env` loading and explicit `--env-file`.

- [ ] **Step 2: Implement env loader**

Load `.env` before config resolution, with CLI overrides taking precedence.

- [ ] **Step 3: Expose env-file control**

Add `--env-file` to the CLI entry path or global option handling.

- [ ] **Step 4: Update auth/config help text**

Make the CLI explain `.env` usage and config precedence clearly.

- [ ] **Step 5: Run tests**

Run targeted config/auth tests, then the full suite.

- [ ] **Step 6: Commit**

Commit message: `feat: add dotenv-based config loading`

### Task 4: Add Root Verification and Smoke-Test Workflow

**Files:**

- Modify: `README.md`
- Modify: root `package.json`
- Modify: `packages/twenty-sdk/package.json`
- Modify or create: smoke-test support under `packages/twenty-sdk`

- [ ] **Step 1: Add smoke-test script**

Use the configured hosted workspace without printing secrets.

- [ ] **Step 2: Add setup/build/test docs**

Document root commands and `.env` flow.

- [ ] **Step 3: Ensure repo commands are ergonomic**

A new contributor should be able to run setup and checks from repo root.

- [ ] **Step 4: Commit**

Commit message: `docs: add root workflow and smoke test guidance`

### Task 5: Write the Twenty API Coverage Audit

**Files:**

- Create: `plans/2026-03-21-twenty-api-coverage-audit.md`

- [ ] **Step 1: Inventory upstream Twenty capability**

Use official repo/docs only.

- [ ] **Step 2: Map current CLI coverage**

Group by records, metadata, raw access, auth, search, files, webhooks, API keys, serverless, and other upstream domains.

- [ ] **Step 3: Mark gaps**

For each gap, note source proof, likely protocol (`REST` or `GraphQL`), and implementation priority.

- [ ] **Step 4: Commit**

Commit message: `docs: add twenty api coverage audit`

### Task 6: Implement Missing CLI Coverage in Priority Order

**Files:**

- Modify or create command/service/test files under `packages/twenty-sdk/src/cli/`

- [ ] **Step 1: Select first missing domain from the audit**

Prefer gaps with clear upstream API shape and low integration risk.

- [ ] **Step 2: Add failing tests**

Cover command registration, request shape, and output behavior.

- [ ] **Step 3: Implement minimal support**

Use existing command and service patterns.

- [ ] **Step 4: Verify targeted tests**

Run only the relevant test files first.

- [ ] **Step 5: Repeat by domain**

Batch work into coherent, reviewable commits rather than one giant change.

### Task 7: Final Verification

**Files:**

- Review all modified files

- [ ] **Step 1: Run root checks**

Run build, tests, lint, and format checks from repo root.

- [ ] **Step 2: Run smoke tests**

Run the hosted-instance smoke path without exposing tokens.

- [ ] **Step 3: Verify git cleanliness expectations**

Normal verification commands should not leave noisy drift.

- [ ] **Step 4: Summarize any remaining gaps**

If upstream coverage is not fully implemented yet, list the remainder explicitly rather than implying completion.
