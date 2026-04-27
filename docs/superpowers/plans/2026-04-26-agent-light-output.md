# Agent-Light Output Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Twenty CLI default to compact agent-friendly JSON with `--light/--li`, `--full`, and `--agent-mode/--ai`, while removing stale `agent` output envelopes.

**Architecture:** Keep behavior centralized. `resolveGlobalOptions` owns mode resolution, `OutputService` owns full/light rendering, a compact alias registry owns JSON key aliases, and command alias helpers decorate Commander command trees after registration.

**Tech Stack:** TypeScript, Commander, Vitest, existing JMESPath query service, existing help JSON builder.

---

### Task 1: Runtime Contract

**Files:**
- Modify: `packages/twenty-sdk/src/cli/utilities/shared/global-options.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/shared/__tests__/utilities.spec.ts`

- [x] Add failing tests proving default output is `json`, `agent` is rejected, `--light/--li`, `--full`, and `--agent-mode/--ai` resolve correctly, and `--light` conflicts with `--full`.
- [x] Implement global option definitions and mode resolution.
- [x] Run `pnpm --filter ./packages/twenty-sdk test -- src/cli/utilities/shared/__tests__/utilities.spec.ts`.
- [x] Commit with `feat: add agent-light runtime options`.

### Task 2: Light Renderer

**Files:**
- Create: `packages/twenty-sdk/src/cli/utilities/output/services/compact-aliases.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/output/services/output.service.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/output/services/__tests__/output.service.spec.ts`

- [x] Add failing tests proving JSON is compact, full output keeps canonical keys, light output rewrites aliases, unknown keys pass through, and alias collisions fail.
- [x] Implement the compact alias registry and recursive light projection.
- [x] Remove agent envelope rendering from `OutputService`.
- [x] Run `pnpm --filter ./packages/twenty-sdk test -- src/cli/utilities/output/services/__tests__/output.service.spec.ts`.
- [x] Commit with `feat: add compact light output renderer`.

### Task 3: Command Aliases

**Files:**
- Create: `packages/twenty-sdk/src/cli/utilities/shared/command-aliases.ts`
- Modify: `packages/twenty-sdk/src/cli/program.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/schema/schema-command-materializer.ts`
- Modify: `packages/twenty-sdk/src/cli/__tests__/help.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/schema/__tests__/schema-command-materializer.spec.ts`

- [x] Add failing tests for `r people ls`, `md api-keys ls`, and aliases reported in help JSON.
- [x] Add alias helpers for static root commands and common operation names.
- [x] Apply aliases after command registration.
- [x] Make dynamic resources kebab-case commands while preserving schema names internally.
- [x] Run help and schema materializer tests.
- [x] Commit with `feat: add short command aliases`.

### Task 4: Help And Docs Cleanup

**Files:**
- Modify: `packages/twenty-sdk/src/cli/help/types.ts`
- Modify: `packages/twenty-sdk/src/cli/help/constants.ts`
- Modify: `packages/twenty-sdk/src/cli/help.txt`
- Modify: `README.md`
- Modify: `packages/twenty-sdk/src/cli/__tests__/release-contract.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/__tests__/help.spec.ts`

- [x] Add failing assertions that help output formats exclude `agent` and include the new mode flags.
- [x] Update help constants, curated help, and README references.
- [x] Run release/help tests.
- [x] Commit with `docs: update agent-light output contract`.

### Task 5: Verification

**Files:**
- No planned source changes.

- [x] Run `pnpm typecheck`.
- [x] Run `pnpm lint`.
- [x] Run `pnpm build`.
- [x] Run `pnpm test`.
- [x] Run built CLI read-only smoke: `auth status`, `schema status all`, `metadata api-keys list --query 'length(@)'`, and `records people list --limit 1 --query 'keys(@)'`.
- [x] Confirm `git status --short --branch` is clean.
