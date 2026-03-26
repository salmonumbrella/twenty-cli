# Twenty CLI Transport, Contracts, and Structural Refactor Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make command transport intent explicit, pin hosted-surface command contracts with realistic tests, and split the highest-risk CLI modules into smaller units without changing intended behavior.

**Architecture:** Land the work in three tracks. First, introduce explicit transport intent and contract-style tests around the scoped commands. Second, correct the verified hosted-surface mismatches for `openapi`, `auth renew-token`, and `auth sso-url` while preserving current non-hosted behavior. Third, refactor `help.ts`, `serverless.command.ts`, and repetitive command wiring behind the new test net.

**Tech Stack:** TypeScript, Commander, Vitest, Vitest E2E config, Axios-based CLI services

---

### Task 0: Create an Isolated Worktree and Verify the Baseline

**Files:**
- Modify: none in the repository tree unless `.gitignore` must be updated
- Test: current package baseline

- [ ] **Step 1: Create a dedicated worktree for implementation**

Use superpowers:using-git-worktrees before dispatching any implementation subagents.

Preferred location order:
- existing `.worktrees/`
- existing `worktrees/`
- otherwise ask the user where to create the worktree, per the skill

- [ ] **Step 2: Verify the worktree baseline**

Run from the worktree root:
```bash
pnpm --filter twenty-sdk build
pnpm --filter twenty-sdk test
```

Expected:
- build succeeds
- test suite is green, or any pre-existing failures are documented before implementation starts

- [ ] **Step 3: Stop and resolve baseline uncertainty before coding**

If the baseline is not green:
- document exactly which commands fail
- decide whether to fix baseline first or continue with explicit approval

- [ ] **Step 4: Commit only if `.gitignore` or worktree setup files changed**

If no tracked files changed, skip this step.

### Task 1: Add a Clean-Home Contract Test Harness

**Files:**
- Create: `packages/twenty-sdk/src/cli/__tests__/e2e/helpers/temp-home.ts`
- Create: `packages/twenty-sdk/src/cli/__tests__/e2e/transport-contracts.e2e.spec.ts`
- Modify: `packages/twenty-sdk/vitest.e2e.config.ts`
- Test: `packages/twenty-sdk/src/cli/__tests__/e2e/transport-contracts.e2e.spec.ts`

- [ ] **Step 1: Write the failing clean-home contract tests**

Add E2E-style CLI process tests that run with an empty temporary `HOME` and no ambient Twenty config. Cover at least:
- `twenty openapi core` fails with auth-required behavior on hosted `api.twenty.com`
- `twenty auth discover <origin>` does not fail with “Missing API token”
- `twenty files download <signed-url>` does not fail with “Missing API token”
- `twenty files public-asset ...` does not fail with “Missing API token`

- [ ] **Step 2: Run the new E2E file to verify it fails for the expected reasons**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.e2e.config.ts src/cli/__tests__/e2e/transport-contracts.e2e.spec.ts`

Expected:
- `openapi` expectation may already pass
- `auth discover`, `files download`, and `files public-asset` fail because the current CLI still routes them through auth-required `ApiService`

- [ ] **Step 3: Add a reusable temp-home helper**

Implement a small helper that:
- creates a temporary directory
- sets `HOME` to that directory for a spawned CLI process
- strips `TWENTY_*` env vars unless the test opts into them
- returns exit code, stdout, and stderr for assertions

- [ ] **Step 4: Re-run the failing E2E file and confirm only behavior failures remain**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.e2e.config.ts src/cli/__tests__/e2e/transport-contracts.e2e.spec.ts`

Expected:
- helper/setup works
- failures now point only at current transport behavior

- [ ] **Step 5: Commit**

Run:
```bash
git add packages/twenty-sdk/src/cli/__tests__/e2e/helpers/temp-home.ts \
        packages/twenty-sdk/src/cli/__tests__/e2e/transport-contracts.e2e.spec.ts \
        packages/twenty-sdk/vitest.e2e.config.ts
git commit -m "test: add clean-home CLI contract harness"
```

### Task 2: Make Transport Intent Explicit for Scoped Commands

**Files:**
- Create: `packages/twenty-sdk/src/cli/utilities/shared/request-transport.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/auth/auth.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/openapi/openapi.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/files/files.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/auth/__tests__/auth.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/openapi/__tests__/openapi.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/files/__tests__/files.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/auth/__tests__/auth.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/openapi/__tests__/openapi.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/files/__tests__/files.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/__tests__/e2e/transport-contracts.e2e.spec.ts`

- [ ] **Step 1: Add failing unit-level tests for explicit transport selection**

Update command tests so they assert which transport is used:
- `auth discover` uses public no-auth transport
- `auth renew-token` and `auth sso-url` are routed through explicit public no-auth transport selection
- `openapi` uses private/auth-required transport
- `files download` and `files public-asset` use public no-auth transport

- [ ] **Step 2: Run the affected command suites to verify failure**

Run:
```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts \
  src/cli/commands/auth/__tests__/auth.command.spec.ts \
  src/cli/commands/openapi/__tests__/openapi.command.spec.ts \
  src/cli/commands/files/__tests__/files.command.spec.ts
```

Expected:
- assertions fail because these commands currently use `services.api`

- [ ] **Step 3: Implement a thin request-transport helper**

Add a small helper with explicit request intent, for example:
- private request via `services.api`
- public request with auth mode and shared option mapping via `services.publicHttp`

The helper must stay thin and must not absorb command parsing or output formatting.

- [ ] **Step 4: Migrate the scoped commands to explicit transport intent**

Use the helper in:
- `auth discover`
- `auth renew-token`
- `auth sso-url`
- `openapi`
- `files download`
- `files public-asset`

Do not migrate unrelated commands in this task.

- [ ] **Step 5: Re-run the unit suites and E2E contract suite**

Run:
```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts \
  src/cli/commands/auth/__tests__/auth.command.spec.ts \
  src/cli/commands/openapi/__tests__/openapi.command.spec.ts \
  src/cli/commands/files/__tests__/files.command.spec.ts
pnpm --filter twenty-sdk exec vitest run -c vitest.e2e.config.ts \
  src/cli/__tests__/e2e/transport-contracts.e2e.spec.ts
```

Expected:
- unit tests pass
- clean-home contract tests no longer fail on missing token for the public no-auth commands

- [ ] **Step 6: Commit**

Run:
```bash
git add packages/twenty-sdk/src/cli/utilities/shared/request-transport.ts \
        packages/twenty-sdk/src/cli/commands/auth/auth.command.ts \
        packages/twenty-sdk/src/cli/commands/openapi/openapi.command.ts \
        packages/twenty-sdk/src/cli/commands/files/files.command.ts \
        packages/twenty-sdk/src/cli/commands/auth/__tests__/auth.command.spec.ts \
        packages/twenty-sdk/src/cli/commands/openapi/__tests__/openapi.command.spec.ts \
        packages/twenty-sdk/src/cli/commands/files/__tests__/files.command.spec.ts
git commit -m "refactor: make CLI transport intent explicit"
```

### Task 3: Correct Hosted-Surface Contract Behavior for OpenAPI and Auth Helpers

**Files:**
- Modify: `packages/twenty-sdk/src/cli/commands/auth/auth.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/openapi/openapi.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/auth/__tests__/auth.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/openapi/__tests__/openapi.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/__tests__/e2e/transport-contracts.e2e.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/auth/__tests__/auth.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/openapi/__tests__/openapi.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/__tests__/e2e/transport-contracts.e2e.spec.ts`

- [ ] **Step 1: Add failing tests for the verified hosted-surface contract**

Add explicit tests for:
- hosted `openapi` remains auth-required
- hosted `auth renew-token` targets `/metadata`
- hosted `auth sso-url` targets `/metadata`
- hosted `auth renew-token` pins the `/metadata` response-shape handling that the command expects after endpoint selection is corrected
- hosted `auth sso-url` pins the `/metadata` response-shape handling that the command expects after endpoint selection is corrected
- non-hosted surfaces preserve current behavior for `renew-token` and `sso-url`

- [ ] **Step 2: Run the focused suites to verify failure**

Run:
```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts \
  src/cli/commands/auth/__tests__/auth.command.spec.ts \
  src/cli/commands/openapi/__tests__/openapi.command.spec.ts
pnpm --filter twenty-sdk exec vitest run -c vitest.e2e.config.ts \
  src/cli/__tests__/e2e/transport-contracts.e2e.spec.ts
```

Expected:
- hosted endpoint assertions fail because `renew-token` and `sso-url` currently target `/graphql`

- [ ] **Step 3: Implement hosted versus non-hosted endpoint selection**

Implement the narrowest logic that satisfies the spec:
- if base URL host is `api.twenty.com`, route `renew-token` and `sso-url` to `/metadata`
- otherwise preserve current endpoint behavior
- avoid generic multi-surface fallback machinery in this task

- [ ] **Step 4: Align hosted response parsing with the `/metadata` contract**

After the endpoint switch is in place, update `renew-token` and `sso-url` parsing only as needed to satisfy the schema-shape tests added in Step 1.
Do not broaden scope into speculative schema normalization for non-hosted surfaces.

- [ ] **Step 5: Re-run focused suites and then the broader auth/openapi coverage**

Run:
```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts \
  src/cli/commands/auth/__tests__/auth.command.spec.ts \
  src/cli/commands/openapi/__tests__/openapi.command.spec.ts
pnpm --filter twenty-sdk exec vitest run -c vitest.e2e.config.ts \
  src/cli/__tests__/e2e/transport-contracts.e2e.spec.ts
```

Expected:
- hosted and non-hosted expectations pass

- [ ] **Step 6: Commit**

Run:
```bash
git add packages/twenty-sdk/src/cli/commands/auth/auth.command.ts \
        packages/twenty-sdk/src/cli/commands/openapi/openapi.command.ts \
        packages/twenty-sdk/src/cli/commands/auth/__tests__/auth.command.spec.ts \
        packages/twenty-sdk/src/cli/commands/openapi/__tests__/openapi.command.spec.ts \
        packages/twenty-sdk/src/cli/__tests__/e2e/transport-contracts.e2e.spec.ts
git commit -m "fix: align hosted auth helper and openapi contracts"
```

### Task 4: Split the Help System into Focused Modules

**Files:**
- Create: `packages/twenty-sdk/src/cli/help/types.ts`
- Create: `packages/twenty-sdk/src/cli/help/constants.ts`
- Create: `packages/twenty-sdk/src/cli/help/command-resolution.ts`
- Create: `packages/twenty-sdk/src/cli/help/options.ts`
- Create: `packages/twenty-sdk/src/cli/help/operations.ts`
- Create: `packages/twenty-sdk/src/cli/help/document-builder.ts`
- Modify: `packages/twenty-sdk/src/cli/help.ts`
- Modify: `packages/twenty-sdk/src/cli/__tests__/help.spec.ts`
- Test: `packages/twenty-sdk/src/cli/__tests__/help.spec.ts`

- [ ] **Step 1: Capture current help behavior with focused failing or expanded tests where needed**

Before extraction, add any missing assertions that protect:
- root help fallback behavior
- help-json target resolution
- option serialization
- operation inference
- visible subcommand rendering

- [ ] **Step 2: Run help tests to verify the safety net**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/__tests__/help.spec.ts`

Expected:
- green baseline before extraction

- [ ] **Step 3: Extract types and pure helpers first**

Move only pure types, constants, and helper functions into new files without changing behavior.

- [ ] **Step 4: Extract command resolution and document builder**

Refactor `help.ts` so it becomes a small composition layer that wires together the extracted modules.

- [ ] **Step 5: Re-run help tests and build**

Run:
```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/__tests__/help.spec.ts
pnpm --filter twenty-sdk build
```

Expected:
- help tests pass
- build succeeds with the new module layout

- [ ] **Step 6: Commit**

Run:
```bash
git add packages/twenty-sdk/src/cli/help.ts \
        packages/twenty-sdk/src/cli/help \
        packages/twenty-sdk/src/cli/__tests__/help.spec.ts
git commit -m "refactor: split CLI help system"
```

### Task 5: Split Serverless Command Registration from Operations

**Files:**
- Create: `packages/twenty-sdk/src/cli/commands/serverless/serverless.types.ts`
- Create: `packages/twenty-sdk/src/cli/commands/serverless/serverless.shared.ts`
- Create: `packages/twenty-sdk/src/cli/commands/serverless/serverless.registration.ts`
- Create: `packages/twenty-sdk/src/cli/commands/serverless/operations/list.operation.ts`
- Create: `packages/twenty-sdk/src/cli/commands/serverless/operations/get.operation.ts`
- Create: `packages/twenty-sdk/src/cli/commands/serverless/operations/create.operation.ts`
- Create: `packages/twenty-sdk/src/cli/commands/serverless/operations/update.operation.ts`
- Create: `packages/twenty-sdk/src/cli/commands/serverless/operations/delete.operation.ts`
- Create: `packages/twenty-sdk/src/cli/commands/serverless/operations/execute.operation.ts`
- Create: `packages/twenty-sdk/src/cli/commands/serverless/operations/source.operation.ts`
- Create: `packages/twenty-sdk/src/cli/commands/serverless/operations/logs.operation.ts`
- Create: `packages/twenty-sdk/src/cli/commands/serverless/operations/publish.operation.ts`
- Create: `packages/twenty-sdk/src/cli/commands/serverless/operations/create-layer.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/serverless/serverless.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/serverless/__tests__/serverless.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/serverless/__tests__/serverless.command.spec.ts`

- [ ] **Step 1: Add or tighten tests that lock current serverless command registration and operation behavior**

Protect:
- command registration
- legacy/current compatibility behavior
- logs execution behavior
- create/update payload shaping

- [ ] **Step 2: Run the focused serverless suite to confirm the baseline**

Run: `pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/commands/serverless/__tests__/serverless.command.spec.ts`

Expected:
- green baseline before extraction

- [ ] **Step 3: Extract shared types and helpers**

Move:
- option types
- compatibility helpers
- payload-building helpers
- render helpers

Keep the original exports stable while moving internals.

- [ ] **Step 4: Extract operation modules and slim down the top-level command file**

Make `serverless.command.ts` a thin registration/composition layer. Keep the command surface unchanged.

- [ ] **Step 5: Re-run the focused suite and build**

Run:
```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/commands/serverless/__tests__/serverless.command.spec.ts
pnpm --filter twenty-sdk build
```

Expected:
- serverless tests pass
- build succeeds

- [ ] **Step 6: Commit**

Run:
```bash
git add packages/twenty-sdk/src/cli/commands/serverless \
        packages/twenty-sdk/src/cli/commands/serverless/serverless.command.ts
git commit -m "refactor: split serverless command operations"
```

### Task 6: Introduce a Minimal Command-Registration Helper

**Files:**
- Create: `packages/twenty-sdk/src/cli/utilities/shared/register-command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/api.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/mcp/mcp.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/__tests__/operations.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/mcp/__tests__/mcp.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/api/operations/__tests__/operations.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/mcp/__tests__/mcp.command.spec.ts`

- [ ] **Step 1: Add or tighten tests that lock command registration and action behavior for API and MCP**

Protect the current subcommand names, option parsing, and action dispatch shape before abstraction.

- [ ] **Step 2: Run the focused suites to confirm the baseline**

Run:
```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts \
  src/cli/commands/api/operations/__tests__/operations.spec.ts \
  src/cli/commands/mcp/__tests__/mcp.command.spec.ts
```

Expected:
- green baseline before introducing the helper

- [ ] **Step 3: Implement the smallest helper that removes duplicated command wiring**

Requirements:
- no DSL
- no dynamic reflection
- still obvious at the call site what command, args, and action are being registered

- [ ] **Step 4: Apply the helper to at least two concrete call sites**

Refactor:
- `api.command.ts`
- `mcp.command.ts`

Only keep the helper if both files become clearer.

- [ ] **Step 5: Re-run focused suites and build**

Run:
```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts \
  src/cli/commands/api/operations/__tests__/operations.spec.ts \
  src/cli/commands/mcp/__tests__/mcp.command.spec.ts
pnpm --filter twenty-sdk build
```

Expected:
- focused suites pass
- build succeeds

- [ ] **Step 6: Commit**

Run:
```bash
git add packages/twenty-sdk/src/cli/utilities/shared/register-command.ts \
        packages/twenty-sdk/src/cli/commands/api/api.command.ts \
        packages/twenty-sdk/src/cli/commands/mcp/mcp.command.ts \
        packages/twenty-sdk/src/cli/commands/api/operations/__tests__/operations.spec.ts \
        packages/twenty-sdk/src/cli/commands/mcp/__tests__/mcp.command.spec.ts
git commit -m "refactor: reduce repeated command registration"
```

### Task 7: Final Verification and Integration Review

**Files:**
- Modify: any files touched by prior tasks
- Test: `packages/twenty-sdk`

- [ ] **Step 1: Run the combined targeted suites**

Run:
```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts \
  src/cli/commands/auth/__tests__/auth.command.spec.ts \
  src/cli/commands/openapi/__tests__/openapi.command.spec.ts \
  src/cli/commands/files/__tests__/files.command.spec.ts \
  src/cli/__tests__/help.spec.ts \
  src/cli/commands/serverless/__tests__/serverless.command.spec.ts \
  src/cli/commands/api/operations/__tests__/operations.spec.ts \
  src/cli/commands/mcp/__tests__/mcp.command.spec.ts
pnpm --filter twenty-sdk exec vitest run -c vitest.e2e.config.ts \
  src/cli/__tests__/e2e/transport-contracts.e2e.spec.ts
```

Expected:
- all targeted suites pass

- [ ] **Step 2: Run build and type-oriented verification**

Run:
```bash
pnpm --filter twenty-sdk typecheck
pnpm --filter twenty-sdk build
```

Expected:
- both commands exit successfully

- [ ] **Step 3: Run the full package test suite if the targeted suites are green**

Run: `pnpm --filter twenty-sdk test`

Expected:
- package test suite passes

- [ ] **Step 4: Review git diff for unintended scope creep**
 
- [ ] **Step 4: Review git diff for unintended scope creep**

Run:
```bash
git status --short
git diff --stat "$(git merge-base HEAD origin/HEAD 2>/dev/null || git rev-list --max-parents=0 HEAD | tail -n 1)"..HEAD
```

Expected:
- only the planned files are affected

- [ ] **Step 5: Commit any final cleanup**

Run:
```bash
git add -A
git commit -m "chore: finalize CLI refactor integration"
```
