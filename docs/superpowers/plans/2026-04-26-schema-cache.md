# Schema Cache Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the local schema cache foundation for dynamic Twenty command discovery.

**Architecture:** Introduce a focused `SchemaCacheService` that resolves the active profile/base URL, fetches core OpenAPI, metadata OpenAPI, and GraphQL introspection through existing authenticated transports, and stores token-free cache entries under `~/.twenty/schema-cache`. Expose it through `twenty schema refresh/status/clear` while keeping `twenty openapi` and `twenty raw graphql` unchanged.

**Tech Stack:** TypeScript, Commander, Vitest, `fs-extra`, Node `crypto`, existing API/config/output services.

---

## File Structure

- Create `packages/twenty-sdk/src/cli/utilities/schema/schema-cache.service.ts`
  - Cache key generation.
  - Token-safe entry serialization.
  - Refresh/status/clear APIs.
  - Shared GraphQL introspection query.
- Create `packages/twenty-sdk/src/cli/utilities/schema/__tests__/schema-cache.service.spec.ts`
  - Cache path, refresh, status, clear, and token-redaction tests.
- Create `packages/twenty-sdk/src/cli/commands/schema/schema.command.ts`
  - Register `schema refresh`, `schema status`, and `schema clear`.
- Create `packages/twenty-sdk/src/cli/commands/schema/__tests__/schema.command.spec.ts`
  - Command registration and output tests.
- Modify `packages/twenty-sdk/src/cli/utilities/shared/services.ts`
  - Add `schemaCache` to `CliServices`.
- Modify `packages/twenty-sdk/src/cli/program.ts`
  - Register schema commands.
- Modify `packages/twenty-sdk/src/cli/help/constants.ts`, `help.txt`, and `src/cli/__tests__/help.spec.ts`
  - Document schema cache commands.

## Tasks

- [x] **Task 1: Write failing service tests**
  - Verify cache entries are written under a hashed profile directory.
  - Verify persisted cache entries include `baseUrl`, `workspace`, `kind`, `fetchedAt`, `contentHash`, and `schema`.
  - Verify token-like query params are redacted from persisted `baseUrl`.
  - Verify `status` reports missing/existing/stale entries.
  - Verify `clear` deletes one kind or all kinds for the active profile.

- [x] **Task 2: Implement `SchemaCacheService`**
  - Use `ConfigService.resolveApiConfig({ requireAuth })`.
  - Fetch:
    - `core-openapi`: `GET /rest/open-api/core`
    - `metadata-openapi`: `GET /rest/open-api/metadata`
    - `graphql`: `POST /graphql` with introspection query
  - Store cache entries with `schemaVersion: 1`.
  - Hash normalized `baseUrl + workspace` for the directory name.

- [x] **Task 3: Write failing command/help tests**
  - `twenty schema --help-json` lists `refresh`, `status`, and `clear`.
  - `refresh` calls service and renders entries.
  - `status` renders cache status without fetching.
  - `clear` renders cleared entries.
  - Root help includes schema cache examples.

- [x] **Task 4: Implement command registration**
  - `twenty schema refresh [kind]`
  - `twenty schema status [kind]`
  - `twenty schema clear [kind]`
  - Valid kinds: `all`, `core-openapi`, `metadata-openapi`, `graphql`, plus aliases `core` and `metadata`.

- [x] **Task 5: Verify and commit locally**
  - Run targeted tests.
  - Run `pnpm typecheck`, `pnpm lint`, `pnpm build`, `pnpm test`.
  - Run built help JSON for `schema`.
  - Commit locally only.
