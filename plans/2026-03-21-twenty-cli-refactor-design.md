# Twenty CLI Refactor Design

Date: 2026-03-21

## Goal

Turn `twenty-cli` into a normal TypeScript CLI repository with:

- a real root workspace/tooling layer
- clean local builds and tests
- first-class `.env` support
- TypeScript-native git hooks and verification
- an explicit audit of Twenty upstream API coverage

## Current Problems

- The repo is half-converted from an earlier Go layout.
- Root tooling is stale: `lefthook.yml` and `release.yml` still use Go conventions.
- There is no real root Node workspace, so repo-level automation is weak.
- Running the normal TypeScript build dirties the worktree because generated artifacts are tracked or stale.
- `packages/twenty-sdk/node_modules` is tracked, which is not acceptable for a mature TypeScript repo.
- The CLI reads env vars directly but has no proper `.env` workflow.
- The public repo shape and the local filesystem diverge because ignored files and generated files are mixed together.

## Reference Direction

The target structure should be closer to mature CLIs in both camps:

- Chatwoot CLI for developer workflow, smoke-test discipline, and auth/config ergonomics
- Roam Tools for TypeScript repo shape: root package manager, root scripts/config, shared TypeScript config, and package-oriented layout

## Proposed Architecture

### 1. Root Workspace Layer

Add a root-managed TypeScript workspace with:

- `package.json`
- `pnpm-workspace.yaml`
- root scripts for `build`, `test`, `lint`, `format`, `setup`, and smoke checks
- root lint/format config
- root `pnpm`/hook config (`.pre-commit-config.yaml`, repo scripts)

`packages/twenty-sdk` remains the actual CLI package.

### 2. Build Artifact Strategy

Stop treating generated local output as source-controlled project state.

Desired end state:

- `node_modules` is ignored and untracked
- test/cache output is ignored and untracked
- `dist` is either built on demand and ignored, or regenerated in a controlled packaging flow rather than drifting in git

The key requirement is that `build` and `test` do not make the repo look broken.

### 3. Configuration and `.env`

Support config resolution in this order:

1. explicit CLI flags
2. environment variables, including values loaded from `.env` or `--env-file`
3. persisted workspace config in `~/.twenty/config.json`

This preserves current behavior while making local development and smoke testing much easier.

### 4. Verification Workflow

Replace the stale Go hooks with a TypeScript flow based on:

- `pnpm`
- `oxlint`
- `oxfmt`
- `@j178/prek`

This is a better fit than retaining `lefthook` in a TypeScript repo and follows the `roam-tools` direction the user requested.

### 5. Coverage Audit

Create a tracked markdown audit that maps:

- current Twenty upstream API capability
- current CLI coverage
- missing or partial coverage
- implementation priority

This document becomes the source of truth for the expansion work.

## Implementation Phases

### Phase 1. Repository Foundation

- root workspace
- root scripts
- TypeScript `prek`
- ignore hygiene
- remove tracked dependency/cache garbage

### Phase 2. Runtime and Config Maturity

- `.env` loading
- `--env-file`
- README and examples
- smoke-test command(s)

### Phase 3. API Coverage Audit

- upstream capability inventory
- CLI gap matrix
- tracked markdown artifact

### Phase 4. CLI Coverage Expansion

Implement missing areas in priority order, focusing first on features already exposed upstream and already consistent with the current command model.

## Non-Goals

- Rewriting the CLI into a different framework just for style
- Copying Chatwoot CLI or Roam Tools wholesale
- Rebuilding already-covered functionality without evidence of a gap

## Success Criteria

- `pnpm install`, `pnpm build`, `pnpm test`, `pnpm lint`, and `pnpm format:check` work from repo root
- normal build/test runs do not create confusing repo drift
- `.env` works predictably
- hooks are TypeScript-native
- coverage gaps are documented in a tracked markdown file
- smoke tests can run against the configured hosted Twenty workspace without exposing secrets
