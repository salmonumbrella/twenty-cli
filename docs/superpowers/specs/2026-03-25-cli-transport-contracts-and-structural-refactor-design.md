# Twenty CLI Transport, Contract, and Structural Refactor Design

**Date:** 2026-03-25

## Goal

Refactor the CLI in three ordered tracks so command transport behavior is explicit, command/API contracts are pinned by tests, and the largest high-risk modules are split into smaller units without changing user-visible behavior unnecessarily.

## Why This Work Exists

The current CLI has three related problems:

1. Commands inherit authentication behavior implicitly from the client they happen to use.
2. Tests mock low enough in the stack that auth and endpoint contract regressions can pass locally.
3. A few large files concentrate too many responsibilities, making future changes risky.

The recent investigation confirmed concrete examples:

- Public-looking and signed-URL flows still route through auth-required `ApiService`.
- Some README claims were stronger than the real hosted API behavior.
- `openapi`, `auth renew-token`, and `auth sso-url` need their runtime assumptions revalidated.
- `help.ts` and `serverless.command.ts` are already large enough to justify decomposition.

## Scope

This design covers five requested refactors grouped into three tracks:

1. Transport boundary cleanup
2. Contract and coverage improvements
3. Structural cleanup for help generation, serverless command layout, and repetitive command wiring

## Non-Goals

- Rewriting every command to a new abstraction in one pass
- Merging `ApiService` and `PublicHttpService` into a single client
- Broad API-surface cleanup beyond the commands already investigated
- Large visual or UX changes to CLI output
- Unrelated documentation changes outside the already-corrected README claims

## Execution Order

The work must proceed in this order:

1. Track 1: transport and contracts
2. Track 2: behavior and coverage
3. Track 3: structural cleanup

This order is mandatory because Track 3 moves code around and should be protected by tests that already pin the intended behavior.

## Track 1: Transport and Contracts

### Objective

Make command auth behavior explicit instead of implicit.

### Design

Keep the two existing low-level clients:

- `ApiService` for auth-required workspace API calls
- `PublicHttpService` for requests that may be unauthenticated, optionally authenticated, or explicitly require auth

Add a thin shared helper layer that makes transport intent obvious at command call sites. Each affected command must choose one of:

- `private`
- `public optional auth`
- `public no auth`

The helper should reduce accidental misuse, but it should not hide which path a command is taking.

### Initial Command Scope

The first migration wave is limited to:

- `auth discover`
- `auth renew-token`
- `auth sso-url`
- `openapi`
- `files download`
- `files public-asset`

### Command Transport Matrix

The initial migration wave must assign explicit transport intent as follows:

| Command | Intended transport intent | Notes |
|---------|----------------------------|-------|
| `auth discover` | public no auth | This command is for public workspace auth discovery and should work without ambient credentials. |
| `auth renew-token` | public no auth | On the hosted `api.twenty.com` surface, the verified target is `/metadata`, not `/graphql`. Availability may differ on other API surfaces. |
| `auth sso-url` | public no auth | On the hosted `api.twenty.com` surface, the verified target is `/metadata`, not `/graphql`. Availability may differ on other API surfaces. |
| `openapi` | private | On the hosted `api.twenty.com` surface, `/rest/open-api/*` currently requires authentication. |
| `files download` | public no auth | Signed URLs and tokenized `/file/*` downloads should not require workspace credentials. |
| `files public-asset` | public no auth | Public asset downloads should work without ambient credentials. |

### Constraints

- Do not convert unrelated commands during this track.
- Do not redesign the whole service container.
- Favor small helpers over a generalized framework.

## Track 2: Behavior and Coverage

### Objective

Pin behavior with realistic tests, then fix the verified command/API mismatches.

### Test Strategy

Add command-level tests that exercise the CLI without inheriting credentials from the developer machine.

This track needs a reusable clean-home test helper so tests can simulate:

- no `~/.twenty/config.json`
- no ambient token
- explicit auth when required
- no-auth behavior when allowed

### Coverage Targets

Add or strengthen coverage around:

- auth-required versus no-auth command execution
- signed URL download behavior
- public asset behavior
- openapi hosted-surface auth expectations
- auth helper endpoint and schema expectations

### Runtime Corrections

This track may correct command behavior where live verification already showed mismatches:

- `openapi` assumptions on hosted `api.twenty.com`
- `renew-token` endpoint/schema assumptions
- `sso-url` endpoint/schema assumptions

Changes in this track should be evidence-driven. If a behavior is not verified, do not redesign it speculatively.

### Verified Hosted-Surface Contract Targets

The following hosted-surface findings are part of the planning baseline and should be reflected in tests and implementation:

| Command | Hosted target path | Hosted auth expectation | Verified hosted behavior to pin |
|---------|--------------------|-------------------------|---------------------------------|
| `openapi` | `/rest/open-api/core` and `/rest/open-api/metadata` | auth required | Unauthenticated requests on `api.twenty.com` return `403 Missing authentication token`. |
| `auth discover` | `/metadata` | no auth required | `getPublicWorkspaceDataByDomain` is callable without auth and returns business-level GraphQL errors rather than auth rejection when the workspace is not found. |
| `auth renew-token` | `/metadata` | no auth required on hosted surface | The mutation is not available on `/graphql` on hosted `api.twenty.com`; hosted planning should target `/metadata`. Response-shape expectations must be pinned from verified schema-compatible tests, not from the current implementation. |
| `auth sso-url` | `/metadata` | no auth required on hosted surface | The mutation is not available on `/graphql` on hosted `api.twenty.com`; hosted planning should target `/metadata`. |

For `renew-token` and `sso-url`, tests should pin command transport selection and endpoint choice first, then pin the schema shape that is actually implemented after the red-green cycle confirms the correct hosted contract.

## Track 3: Structural Cleanup

### Objective

Reduce future regression risk by decomposing the largest CLI modules after earlier tracks are green.

### Help System Split

Refactor `packages/twenty-sdk/src/cli/help.ts` into a focused `help/` module set. The target split is:

- `types`
- `constants` or `metadata`
- command resolution
- argument and option serialization
- operation inference
- help document building
- top-level CLI entry helpers

The resulting top-level `help.ts` should become a small composition layer.

### Serverless Split

Refactor `packages/twenty-sdk/src/cli/commands/serverless/serverless.command.ts` into a `serverless/` folder where the top-level registration file is thin and operation logic is grouped by responsibility.

Target structure:

- registration file
- shared compatibility and GraphQL helpers
- per-operation modules or closely related operation groups

### Repetitive Command Wiring

Introduce a small command-registration helper only where it clearly improves readability for repetitive patterns such as `api.command.ts`.

This is not permission to create a generic DSL. Prefer minimal helpers that remove duplication while preserving obvious control flow.

## File Boundary Guidance

### Likely New or Modified Areas

- `packages/twenty-sdk/src/cli/utilities/api/services/`
- `packages/twenty-sdk/src/cli/utilities/shared/`
- `packages/twenty-sdk/src/cli/commands/auth/`
- `packages/twenty-sdk/src/cli/commands/openapi/`
- `packages/twenty-sdk/src/cli/commands/files/`
- `packages/twenty-sdk/src/cli/help/`
- `packages/twenty-sdk/src/cli/commands/serverless/`
- affected command test suites
- shared CLI test helpers for clean-home execution

### Boundary Rules

- Transport helpers must not also own output formatting or command parsing.
- Contract tests should exercise command behavior, not internal implementation details.
- Extraction work in Track 3 must preserve import direction clarity and avoid circular dependencies.
- New helpers should be introduced only when at least two concrete call sites benefit immediately.

## Testing and Verification

### TDD Requirement

Tracks 1 and 2 follow strict TDD:

1. write failing test
2. verify failure is for the expected reason
3. implement the minimal fix
4. rerun focused tests
5. rerun broader affected suites

Track 3 refactors are protected by pre-existing green tests plus any new focused tests needed for extracted seams.

### Verification Levels

For each task:

- focused test file or command suite
- broader package tests for the touched subsystem
- typecheck and build when structural changes affect many imports

Before claiming the overall program is complete:

- package test suite relevant to changed commands
- package build
- any additional targeted smoke checks created during implementation

## Subagent Execution Model

Implementation will use subagent-driven development.

Rules:

- one implementer subagent per task
- no parallel implementers touching the same files
- spec-compliance review before code-quality review
- same implementer fixes reviewer findings, then gets re-reviewed

Parallelism is only allowed for disjoint write scopes, such as separate structural extractions after earlier tracks are stable.

## Rollback and Stopping Points

The work is intentionally staged so it can stop safely after each track:

- after Track 1: explicit transport intent and reduced auth ambiguity
- after Track 2: stronger tests and corrected verified command behavior
- after Track 3: smaller modules with behavior already pinned

Each extraction in Track 3 should also be small enough to revert independently if needed.

## Risks

- Live API behavior may differ between hosted and self-hosted surfaces.
- Over-abstracting command registration could make the CLI harder to read.
- Splitting large files can accidentally move behavior across modules without preserving test coverage.

## Mitigations

- Keep hosted-surface assumptions explicit in tests and docs.
- Prefer narrow helpers over framework-like abstractions.
- Land structural work only behind green behavior tests.

## Success Criteria

This design is successful when:

- affected commands declare auth behavior explicitly
- tests catch clean-home auth regressions
- verified API-surface mismatches are corrected or intentionally guarded
- `help.ts` and `serverless.command.ts` are substantially smaller and more focused
- command wiring duplication is reduced without obscuring behavior
