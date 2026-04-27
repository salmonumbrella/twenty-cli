# Dynamic Record Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make cache-backed `twenty records <object> <operation>` mirror the existing generic record API behavior for collection and destructive operations.

**Architecture:** Keep the command materializer schema-driven, but let a single generated operation support both single-record and bulk forms where the existing `twenty api` command already does. Preserve existing confirmation semantics for destructive commands.

**Tech Stack:** TypeScript, Commander, Vitest, existing `records` service and CLI output service.

---

## File Structure

- Modify `packages/twenty-sdk/src/cli/utilities/schema/schema-command-materializer.ts`
  - Extract collection-level `PATCH`, `DELETE`, and `/restore/<object>` OpenAPI paths.
  - Support bulk `batch-update`, `batch-delete`, `destroy`, and `restore` execution.
  - Keep `--yes` required for destructive dynamic operations.
- Modify `packages/twenty-sdk/src/cli/utilities/schema/__tests__/schema-command-materializer.spec.ts`
  - Cover extraction of collection-level operations.
  - Cover dynamic bulk execution behavior.

## Tasks

- [x] **Task 1: Add parity tests**
  - Update the core OpenAPI fixture to include collection `PATCH`/`DELETE`, single and collection restore, and single `DELETE`.
  - Assert generated `people` operations include `destroy` and `restore`.
  - Assert dynamic `batch-update` uses `records.updateMany` for object payload plus `--ids`.
  - Assert dynamic `restore` uses `records.restoreMany` when no ID is supplied.
  - Assert dynamic `destroy` uses `records.destroyMany` when no ID is supplied and `--yes` is present.
  - Assert dynamic `batch-delete` accepts a JSON array payload as well as `--ids`.

- [x] **Task 2: Implement OpenAPI extraction parity**
  - Map `DELETE /<object>/{id}` to both `delete` and `destroy`.
  - Map `PATCH /<object>` to `batch-update`.
  - Map `DELETE /<object>` to `batch-delete`.
  - Map `PATCH /restore/<object>` to `restore`.

- [x] **Task 3: Implement dynamic execution parity**
  - Let `batch-update` dispatch arrays to `records.batchUpdate`.
  - Let `batch-update` dispatch object payloads plus `--filter`/`--ids` to `records.updateMany`.
  - Let `batch-delete` accept `--ids` or a JSON array from `--data`/`--file`.
  - Let `destroy [id]` use `records.destroy` with an ID and `records.destroyMany` with `--filter`/`--ids`.
  - Let `restore [id]` use `records.restore` with an ID and `records.restoreMany` with `--filter`/`--ids`.

- [x] **Task 4: Verify and commit locally**
  - Run targeted schema materializer tests.
  - Run `pnpm typecheck`, `pnpm lint`, `pnpm build`, and `pnpm test`.
  - Run the built coverage auditor against a local Twenty checkout.
  - Commit locally only.
