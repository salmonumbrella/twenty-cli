# Dynamic Schema Commands Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Materialize cached Twenty schemas into first-class CLI command trees.

**Architecture:** Keep startup network-free and synchronous. At `buildProgram()` time, read any matching token-safe schema cache entries from `~/.twenty/schema-cache`, parse OpenAPI paths, and register `records` and `metadata` command trees. `twenty schema refresh` remains the explicit network step that populates or updates the cache.

**Tech Stack:** TypeScript, Commander, Vitest, existing records/API/output services.

---

## File Structure

- Create `packages/twenty-sdk/src/cli/utilities/schema/schema-cache-reader.ts`
  - Synchronous cache context resolution from env/config file.
  - Synchronous read of cached entries for startup command registration.
- Create `packages/twenty-sdk/src/cli/utilities/schema/schema-command-materializer.ts`
  - Parse OpenAPI paths into resource operation specs.
  - Register `records <object> <operation>` for cached core OpenAPI objects.
  - Register `metadata <resource> <operation>` for cached metadata OpenAPI resources.
- Create tests under `packages/twenty-sdk/src/cli/utilities/schema/__tests__/`.
- Modify `packages/twenty-sdk/src/cli/program.ts`
  - Register cached schema commands after static bootstrap commands.
- Modify help docs/tests.

## Tasks

- [x] **Task 1: Add cache reader tests**
  - Resolve active workspace/base URL from env and config file.
  - Read cached schema entries from the same hash directory used by `SchemaCacheService`.
  - Return no entries when no cache exists.

- [x] **Task 2: Implement cache reader**
  - Use synchronous `fs` reads only.
  - Reuse the same normalization and hashing rules as `SchemaCacheService`.
  - Do not throw on missing or malformed cache; startup should stay resilient.

- [x] **Task 3: Add materializer tests**
  - Parse core paths into object operations.
  - Parse metadata paths into resource operations.
  - Register `records people list/get/create/update/delete`.
  - Register `metadata views list/get/create/update/delete`.
  - Verify actions call the existing services with the correct REST/record paths.

- [x] **Task 4: Implement materializer**
  - Records operations should delegate to existing `records` service for list/get/create/update/delete and use API paths for special operations as needed.
  - Metadata operations should use `/rest/metadata/<resource>` endpoints.
  - Unknown or absent cached schemas should skip registration cleanly.

- [x] **Task 5: Wire into program and help**
  - `buildProgram()` should call `registerCachedSchemaCommands(program)`.
  - Help should document `records` and `metadata` only as cache-backed command families.

- [x] **Task 6: Verify and commit locally**
  - Run targeted tests.
  - Run `pnpm typecheck`, `pnpm lint`, `pnpm build`, `pnpm test`.
  - Confirm existing coverage audit remains green.
