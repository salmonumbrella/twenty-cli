# Twenty CLI Handler Rewrite Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite the Twenty CLI around a single Commander-based handler architecture, add a shared `PublicHttpService`, fix the `help-*` contract bugs, and deliberately clean up inconsistent CLI surfaces repo-wide.

**Architecture:** Keep Commander and the existing help/output model, but standardize all command handlers on `createCommandContext()` plus a shared `CliServices` container. Route authenticated traffic through `ApiService`, route public/semi-public traffic through a new `PublicHttpService`, convert router-style command families toward explicit subcommands, and make help/output contracts derive from declared command structure instead of guesswork.

**Tech Stack:** TypeScript, Commander, Axios/axios-retry, Vitest, existing CLI config/output/error utilities, live smoke tests against a configured production workspace.

---

## Execution Notes

- User explicitly wants this work on the current branch, so do **not** create a worktree.
- Follow TDD strictly inside each task: failing test first, verify the failure, then minimal implementation.
- Commit after each task.
- Force-add plan/spec docs only when committing them because `docs/` is ignored in this repo.
- Preserve the existing custom help JSON, root help text, output formats, and exit-code classes unless a task explicitly calls for a breaking cleanup.
- Treat live smoke-workspace assumptions as revalidation targets, not permanent truths.

## Interface Change Ledger

- **Behavior-preserving:** Tasks 1 through 5 fix shared contracts, expand the shared context/service container, add `PublicHttpService`, and migrate drifted commands/helpers without changing the intended user-facing command contract beyond bug fixes.
- **Intentionally breaking:** Task 6 moves raw escape hatches under `twenty raw` and removes GraphQL request-text use of `--query`.
- **Intentionally breaking:** Tasks 7 through 9 replace router-style operation arguments with real subcommands for the listed command families. No temporary compatibility aliases are retained for the old router form.
- **Intentionally breaking:** Task 10 standardizes destructive confirmation and option vocabulary. The end-state is `--yes` for destructive acknowledgement, `--query` for output filtering only, `--limit` plus `--cursor` for pagination, `--data` plus `--file` for generic JSON payload input, `--document` plus `--file` for raw GraphQL request input, and domain-specific payload flags only where the payload is genuinely a distinct concept such as search filters.
- **Explicit first-pass exceptions:** `roles`, `applications`, and `application-registrations` are not converted to explicit subcommands in this rewrite. They should still adopt shared handler/context conventions if touched, but their router-style command surface stays in place so this pass can focus on the command families already implicated by help/transport drift and the highest-value consistency gaps.

## File Map

### New files

- `packages/twenty-sdk/src/cli/utilities/api/services/public-http.service.ts`
  Shared public/semi-public HTTP transport with explicit auth modes (`none`, `optional`, `required`), shared retry/debug behavior, and CLI-friendly error shaping.
- `packages/twenty-sdk/src/cli/utilities/api/services/__tests__/public-http.service.spec.ts`
  Unit tests for auth-mode handling, config resolution, retry/debug wiring, and transport error classification.
- `packages/twenty-sdk/src/cli/commands/raw/raw.command.ts`
  New parent namespace for raw escape-hatch commands.
- `packages/twenty-sdk/src/cli/commands/raw/__tests__/raw.command.spec.ts`
  Registration/help tests for the new `twenty raw` namespace.
- `packages/twenty-sdk/src/cli/commands/api/__tests__/api.command.spec.ts`
  New command-level tests for `api` once it stops being a single router-argument command.
- `packages/twenty-sdk/src/cli/utilities/shared/confirmation.ts`
  Small shared helper for repo-wide destructive command acknowledgement via `--yes`.
- `packages/twenty-sdk/src/cli/utilities/shared/__tests__/confirmation.spec.ts`
  Unit tests for the destructive-action helper.

### Shared core files to modify

- `packages/twenty-sdk/src/cli/program.ts`
  Re-register command namespaces, including the new `raw` parent and rewritten command trees.
- `packages/twenty-sdk/src/cli/help.ts`
  Simplify help metadata to prefer explicit command structure, fix `help-*` handling, and make mutability explicit where name inference is wrong.
- `packages/twenty-sdk/src/cli/help.txt`
  Refresh root help examples after router cleanup and raw namespace changes.
- `packages/twenty-sdk/src/cli/__tests__/help.spec.ts`
  Contract tests for command paths, subcommands, operation summaries, explicit mutability, and `help-*` output-kind correctness.
- `packages/twenty-sdk/src/cli/utilities/shared/global-options.ts`
  Fix output-kind derivation for real `help-*` commands and standardize global flag meaning.
- `packages/twenty-sdk/src/cli/utilities/shared/context.ts`
  Make `createCommandContext()` the standard handler entrypoint and keep `createOutputContext()` for output-only commands.
- `packages/twenty-sdk/src/cli/utilities/shared/services.ts`
  Expand `CliServices` to include `config`, `publicHttp`, and `search`, and stop leaving shared services out of the container.
- `packages/twenty-sdk/src/cli/utilities/shared/__tests__/utilities.spec.ts`
  Tests for context/service construction and agent-output kind derivation.
- `packages/twenty-sdk/src/cli/utilities/api/services/api.service.ts`
  Keep authenticated transport behavior aligned with the new public transport semantics.
- `packages/twenty-sdk/src/cli/utilities/config/services/config.service.ts`
  Add any helper needed for base URL/token resolution without forcing auth for public traffic.
- `packages/twenty-sdk/src/cli/utilities/errors/error-handler.ts`
  Keep error formatting/exit codes aligned if public transport introduces new wrapped failure shapes.

### Command/service drift files to modify

- `packages/twenty-sdk/src/cli/commands/auth/auth.command.ts`
- `packages/twenty-sdk/src/cli/commands/auth/__tests__/auth.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/search/search.command.ts`
- `packages/twenty-sdk/src/cli/commands/search/__tests__/search.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/connected-accounts/connected-accounts.command.ts`
- `packages/twenty-sdk/src/cli/commands/connected-accounts/__tests__/connected-accounts.command.spec.ts`
- `packages/twenty-sdk/src/cli/utilities/records/commands/register-record-resource-command.ts`
- `packages/twenty-sdk/src/cli/commands/message-channels/__tests__/message-channels.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/calendar-channels/__tests__/calendar-channels.command.spec.ts`
- `packages/twenty-sdk/src/cli/utilities/search/services/search.service.ts`

### Public command files to modify

- `packages/twenty-sdk/src/cli/commands/routes/routes.command.ts`
- `packages/twenty-sdk/src/cli/commands/routes/__tests__/routes.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/workflows/workflows.command.ts`
- `packages/twenty-sdk/src/cli/commands/workflows/__tests__/workflows.command.spec.ts`

### Raw namespace files to modify

- `packages/twenty-sdk/src/cli/commands/raw/rest.command.ts`
- `packages/twenty-sdk/src/cli/commands/raw/graphql.command.ts`
- `packages/twenty-sdk/src/cli/commands/raw/__tests__/rest.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/raw/__tests__/graphql.command.spec.ts`

### Router-family files to modify

- `packages/twenty-sdk/src/cli/commands/api-keys/api-keys.command.ts`
- `packages/twenty-sdk/src/cli/commands/api-keys/__tests__/api-keys.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/webhooks/webhooks.command.ts`
- `packages/twenty-sdk/src/cli/commands/webhooks/__tests__/webhooks.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/route-triggers/route-triggers.command.ts`
- `packages/twenty-sdk/src/cli/commands/route-triggers/__tests__/route-triggers.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/postgres-proxy/postgres-proxy.command.ts`
- `packages/twenty-sdk/src/cli/commands/postgres-proxy/__tests__/postgres-proxy.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/marketplace-apps/marketplace-apps.command.ts`
- `packages/twenty-sdk/src/cli/commands/marketplace-apps/__tests__/marketplace-apps.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/approved-access-domains/approved-access-domains.command.ts`
- `packages/twenty-sdk/src/cli/commands/approved-access-domains/__tests__/approved-access-domains.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/emailing-domains/emailing-domains.command.ts`
- `packages/twenty-sdk/src/cli/commands/emailing-domains/__tests__/emailing-domains.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/public-domains/public-domains.command.ts`
- `packages/twenty-sdk/src/cli/commands/public-domains/__tests__/public-domains.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/skills/skills.command.ts`
- `packages/twenty-sdk/src/cli/commands/skills/__tests__/skills.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/event-logs/event-logs.command.ts`
- `packages/twenty-sdk/src/cli/commands/event-logs/__tests__/event-logs.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/api/api.command.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/batch-create.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/batch-delete.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/batch-update.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/bulk-filter.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/create.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/delete.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/destroy.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/export.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/find-duplicates.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/get.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/group-by.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/import.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/list.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/merge.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/restore.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/types.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/update.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/__tests__/find-duplicates.operation.spec.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/__tests__/group-by.operation.spec.ts`
- `packages/twenty-sdk/src/cli/commands/api/operations/__tests__/operations.spec.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/api-metadata.command.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/__tests__/api-metadata.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-create.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-delete.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-get.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-list.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-update.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-create.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-delete.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-get.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-list.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-update.operation.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/types.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/ui-metadata.operations.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/__tests__/fields-list.operation.spec.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/__tests__/metadata-operations.spec.ts`
- `packages/twenty-sdk/src/cli/commands/api-metadata/operations/__tests__/ui-metadata.operations.spec.ts`
- `packages/twenty-sdk/src/cli/commands/files/files.command.ts`
- `packages/twenty-sdk/src/cli/commands/files/__tests__/files.command.spec.ts`
- `packages/twenty-sdk/src/cli/commands/serverless/serverless.command.ts`
- `packages/twenty-sdk/src/cli/commands/serverless/__tests__/serverless.command.spec.ts`

---

### Task 1: Lock The Shared Contract Bugs In Tests

**Files:**
- Modify: `packages/twenty-sdk/src/cli/__tests__/help.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/shared/__tests__/utilities.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/help.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/shared/global-options.ts`

- [ ] **Step 1: Write failing contract tests for real `help-*` command handling**

Add focused assertions for:

```ts
it("derives agent output kind for mcp help-center from the full command path", () => {
  const command = resolveCommand(buildProgram(), ["mcp", "help-center"]);
  const options = resolveGlobalOptions(command, { /* simulate -o agent */ });
  expect(options.outputKind).toBe("twenty.mcp.help-center");
});

it("marks approved-access-domains validate as mutating in help JSON", () => {
  const help = buildHelpJson(buildProgram(), ["approved-access-domains", "--help-json"]);
  expect(help.operations.find((op) => op.name === "validate")).toEqual(
    expect.objectContaining({ mutates: true }),
  );
});
```

- [ ] **Step 2: Run only the new shared-contract tests and verify they fail**

Run:

```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/__tests__/help.spec.ts src/cli/utilities/shared/__tests__/utilities.spec.ts --reporter verbose
```

Expected:

- FAIL because `help-*` output kind is still collapsed
- FAIL because mutability is still inferred incorrectly for `validate`

- [ ] **Step 3: Implement the minimal shared fixes**

Update shared logic so only the literal built-in `help` command is special-cased:

```ts
if (name && name !== "help") {
  path.unshift(name);
}
```

Add explicit help metadata overrides for operations whose names do not describe mutability correctly:

```ts
"twenty approved-access-domains": {
  operations: [
    { name: "validate", summary: "Validate an approved access domain", mutates: true },
  ],
}
```

- [ ] **Step 4: Re-run the shared-contract tests and verify they pass**

Run the same command from Step 2.

Expected:

- PASS on the new `help-*` output-kind contract
- PASS on explicit mutability override assertions

- [ ] **Step 5: Commit**

```bash
git add packages/twenty-sdk/src/cli/__tests__/help.spec.ts \
  packages/twenty-sdk/src/cli/utilities/shared/__tests__/utilities.spec.ts \
  packages/twenty-sdk/src/cli/help.ts \
  packages/twenty-sdk/src/cli/utilities/shared/global-options.ts
git commit -m "fix: lock shared CLI help contracts"
```

### Task 2: Expand The Shared Command Context And Service Container

**Files:**
- Modify: `packages/twenty-sdk/src/cli/utilities/shared/context.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/shared/services.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/search/services/search.service.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/shared/__tests__/utilities.spec.ts`

- [ ] **Step 1: Write failing tests for the new service container shape**

Add tests that expect `createCommandContext()` to expose the full shared bag:

```ts
it("returns config, api, publicHttp, search, records, metadata, mcp, output, importer, exporter", () => {
  const context = createCommandContext(command);
  expect(context.services).toEqual(
    expect.objectContaining({
      config: expect.anything(),
      api: expect.anything(),
      publicHttp: expect.anything(),
      search: expect.anything(),
      records: expect.anything(),
      metadata: expect.anything(),
      mcp: expect.anything(),
      output: expect.anything(),
    }),
  );
});
```

- [ ] **Step 2: Run the shared utilities test file and verify it fails**

Run:

```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/utilities/shared/__tests__/utilities.spec.ts --reporter verbose
```

Expected:

- FAIL because `CliServices` does not yet include `config`, `publicHttp`, or `search`

- [ ] **Step 3: Implement the minimal context/container expansion**

Update `CliServices` and construction:

```ts
export interface CliServices {
  config: ConfigService;
  api: ApiService;
  publicHttp: PublicHttpService;
  records: RecordsService;
  metadata: MetadataService;
  search: SearchService;
  mcp: McpService;
  output: OutputService;
  importer: ImportService;
  exporter: ExportService;
}
```

Make `createCommandContext()` the default pattern to consume from this point forward.

- [ ] **Step 4: Re-run the shared utilities tests and verify they pass**

Run the same command from Step 2.

Expected:

- PASS on the expanded shared service bag assertions

- [ ] **Step 5: Commit**

```bash
git add packages/twenty-sdk/src/cli/utilities/shared/context.ts \
  packages/twenty-sdk/src/cli/utilities/shared/services.ts \
  packages/twenty-sdk/src/cli/utilities/search/services/search.service.ts \
  packages/twenty-sdk/src/cli/utilities/shared/__tests__/utilities.spec.ts
git commit -m "refactor: expand shared CLI command context"
```

### Task 3: Add PublicHttpService With Explicit Auth Modes

**Files:**
- Create: `packages/twenty-sdk/src/cli/utilities/api/services/public-http.service.ts`
- Create: `packages/twenty-sdk/src/cli/utilities/api/services/__tests__/public-http.service.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/config/services/config.service.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/api/services/api.service.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/errors/error-handler.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/shared/services.ts`

- [ ] **Step 1: Write failing PublicHttpService tests**

Create a dedicated spec with explicit auth-mode coverage:

```ts
it("uses authMode none without attaching Authorization", async () => {});
it("uses authMode optional and continues when no token is configured", async () => {});
it("uses authMode optional and attaches Authorization when a token exists", async () => {});
it("uses authMode required and throws AUTH when no token is configured", async () => {});
it("inherits debug logging and retry settings from global options", async () => {});
it("resolves base URL from explicit workspace/env/config without hard-failing on missing auth", async () => {});
```

- [ ] **Step 2: Run the new PublicHttpService tests and verify they fail**

Run:

```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/utilities/api/services/__tests__/public-http.service.spec.ts --reporter verbose
```

Expected:

- FAIL because the service does not exist yet

- [ ] **Step 3: Implement PublicHttpService and the supporting config helpers**

Create the service with a narrow API:

```ts
type PublicAuthMode = "none" | "optional" | "required";

await publicHttp.request({
  authMode: "optional",
  method: "get",
  path: "/s/public/ping",
  workspace: "smoke",
  params: { source: "cli" },
});
```

Key implementation rules:

- `none`: never attach auth
- `optional`: attach auth only if a token is available
- `required`: fail with `new CliError("Missing API token.", "AUTH", "Set TWENTY_TOKEN or configure an API key for the selected workspace.")` if auth is unavailable
- share retry/debug behavior with `ApiService`
- do not guess auth mode from token presence

- [ ] **Step 4: Re-run the PublicHttpService tests and verify they pass**

Run the same command from Step 2.

Expected:

- PASS on auth-mode, config resolution, retry, and debug assertions

- [ ] **Step 5: Commit**

```bash
git add packages/twenty-sdk/src/cli/utilities/api/services/public-http.service.ts \
  packages/twenty-sdk/src/cli/utilities/api/services/__tests__/public-http.service.spec.ts \
  packages/twenty-sdk/src/cli/utilities/config/services/config.service.ts \
  packages/twenty-sdk/src/cli/utilities/api/services/api.service.ts \
  packages/twenty-sdk/src/cli/utilities/errors/error-handler.ts \
  packages/twenty-sdk/src/cli/utilities/shared/services.ts
git commit -m "feat: add shared public HTTP transport"
```

### Task 4: Migrate Public Commands To PublicHttpService

**Files:**
- Modify: `packages/twenty-sdk/src/cli/commands/routes/routes.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/routes/__tests__/routes.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/workflows/workflows.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/workflows/__tests__/workflows.command.spec.ts`

- [ ] **Step 1: Write failing command tests that assert shared transport usage**

Update route/workflow specs to assert:

```ts
it("routes invoke uses services.publicHttp with authMode optional", async () => {});
it("workflows invoke-webhook uses services.publicHttp with authMode optional", async () => {});
it("workflows workspace-id discovery uses authMode required when --workspace-id is absent", async () => {});
it("public commands honor --debug and --no-retry through shared transport options", async () => {});
```

- [ ] **Step 2: Run the route/workflow command specs and verify they fail**

Run:

```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts src/cli/commands/routes/__tests__/routes.command.spec.ts src/cli/commands/workflows/__tests__/workflows.command.spec.ts --reporter verbose
```

Expected:

- FAIL because command modules still use raw `axios` and local connection helpers

- [ ] **Step 3: Replace raw axios/config logic with shared context + PublicHttpService**

Apply the pattern:

```ts
const { globalOptions, services } = createCommandContext(command);
const response = await services.publicHttp.request({
  authMode: "optional",
  method,
  path: buildRoutePath(routePath),
  params,
  data,
});
```

For workflow webhook invocation:

```ts
const workspaceId = options.workspaceId
  ?? await discoverWorkspaceId(services.publicHttp, { workspace: globalOptions.workspace });
```

- [ ] **Step 4: Re-run the route/workflow command specs and verify they pass**

Run the same command from Step 2.

Expected:

- PASS on transport delegation and auth-mode assertions

- [ ] **Step 5: Commit**

```bash
git add packages/twenty-sdk/src/cli/commands/routes/routes.command.ts \
  packages/twenty-sdk/src/cli/commands/routes/__tests__/routes.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/workflows/workflows.command.ts \
  packages/twenty-sdk/src/cli/commands/workflows/__tests__/workflows.command.spec.ts
git commit -m "refactor: route public commands through shared transport"
```

### Task 5: Remove Handler And Service Drift In Shared Helpers And Local Commands

**Files:**
- Modify: `packages/twenty-sdk/src/cli/commands/auth/auth.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/auth/__tests__/auth.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/search/search.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/search/__tests__/search.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/connected-accounts/connected-accounts.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/connected-accounts/__tests__/connected-accounts.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/utilities/records/commands/register-record-resource-command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/message-channels/__tests__/message-channels.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/calendar-channels/__tests__/calendar-channels.command.spec.ts`

- [ ] **Step 1: Write failing tests for shared-service consumption**

Add assertions that commands/helpers use the shared container rather than recreating services:

```ts
it("search uses services.search", async () => {});
it("connected-accounts uses services.records for record access", async () => {});
it("record-resource helper uses services.records from command context", async () => {});
it("message-channels and calendar-channels still inherit the shared record-resource contract", async () => {});
it("auth local commands still honor env/output handling without bespoke hydration", async () => {});
```

- [ ] **Step 2: Run the affected command/shared-helper tests and verify they fail**

Run:

```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts \
  src/cli/commands/auth/__tests__/auth.command.spec.ts \
  src/cli/commands/search/__tests__/search.command.spec.ts \
  src/cli/commands/connected-accounts/__tests__/connected-accounts.command.spec.ts \
  src/cli/commands/message-channels/__tests__/message-channels.command.spec.ts \
  src/cli/commands/calendar-channels/__tests__/calendar-channels.command.spec.ts \
  src/cli/utilities/shared/__tests__/utilities.spec.ts --reporter verbose
```

Expected:

- FAIL because commands still instantiate `ConfigService`, `SearchService`, or `RecordsService` locally

- [ ] **Step 3: Migrate the drifted commands/helpers to shared context**

Use this pattern consistently:

```ts
const { globalOptions, services } = createCommandContext(command);
const filter = await parseSearchFilter(options.filter, options.filterFile);
const response = await services.search.search({
  query,
  limit: parseInt(options.limit, 10),
  objects: options.objects?.split(","),
  excludeObjects: options.exclude?.split(","),
  after: options.after,
  filter,
});
await services.output.render(response.data, { format: globalOptions.output, query: globalOptions.query });
```

For local-only auth commands, keep local config behavior but stop duplicating environment/output plumbing.

- [ ] **Step 4: Re-run the affected tests and verify they pass**

Run the same command from Step 2.

Expected:

- PASS on shared-service assertions
- PASS on auth env/output coverage

- [ ] **Step 5: Commit**

```bash
git add packages/twenty-sdk/src/cli/commands/auth/auth.command.ts \
  packages/twenty-sdk/src/cli/commands/auth/__tests__/auth.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/search/search.command.ts \
  packages/twenty-sdk/src/cli/commands/search/__tests__/search.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/connected-accounts/connected-accounts.command.ts \
  packages/twenty-sdk/src/cli/commands/connected-accounts/__tests__/connected-accounts.command.spec.ts \
  packages/twenty-sdk/src/cli/utilities/records/commands/register-record-resource-command.ts \
  packages/twenty-sdk/src/cli/commands/message-channels/__tests__/message-channels.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/calendar-channels/__tests__/calendar-channels.command.spec.ts
git commit -m "refactor: standardize shared CLI handler usage"
```

### Task 6: Introduce The Raw Namespace And Reserve `--query` For Output Filtering

**Files:**
- Create: `packages/twenty-sdk/src/cli/commands/raw/raw.command.ts`
- Create: `packages/twenty-sdk/src/cli/commands/raw/__tests__/raw.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/raw/rest.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/raw/graphql.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/raw/__tests__/rest.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/raw/__tests__/graphql.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/program.ts`
- Modify: `packages/twenty-sdk/src/cli/help.ts`
- Modify: `packages/twenty-sdk/src/cli/help.txt`
- Modify: `packages/twenty-sdk/src/cli/__tests__/help.spec.ts`

- [ ] **Step 1: Write failing registration/help tests for the raw namespace and GraphQL flag rename**

Add tests like:

```ts
it("registers raw rest and raw graphql under a raw parent command", () => {});
it("raw graphql uses --document for request text and --query for output filtering", () => {});
it("root help no longer lists rest/graphql as top-level peers", () => {});
```

- [ ] **Step 2: Run the raw/help test files and verify they fail**

Run:

```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts \
  src/cli/commands/raw/__tests__/raw.command.spec.ts \
  src/cli/commands/raw/__tests__/rest.command.spec.ts \
  src/cli/commands/raw/__tests__/graphql.command.spec.ts \
  src/cli/__tests__/help.spec.ts --reporter verbose
```

Expected:

- FAIL because `raw.command.ts` does not exist
- FAIL because `graphql` still uses `--query` for request text

- [ ] **Step 3: Implement the raw namespace and GraphQL input rename**

Move registration to:

```ts
const raw = program.command("raw").description("Escape-hatch raw API commands");
registerRestCommand(raw);
registerGraphqlCommand(raw);
```

Update GraphQL flags to:

```ts
.option("-d, --document <query>", "GraphQL document string")
.option("-f, --file <path>", "GraphQL document file")
```

Use `--query` only for output filtering through `resolveGlobalOptions()`.

- [ ] **Step 4: Re-run the raw/help tests and verify they pass**

Run the same command from Step 2.

Expected:

- PASS on raw namespace registration
- PASS on GraphQL flag semantics
- PASS on root/help JSON expectations

- [ ] **Step 5: Commit**

```bash
git add packages/twenty-sdk/src/cli/commands/raw/raw.command.ts \
  packages/twenty-sdk/src/cli/commands/raw/__tests__/raw.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/raw/rest.command.ts \
  packages/twenty-sdk/src/cli/commands/raw/graphql.command.ts \
  packages/twenty-sdk/src/cli/commands/raw/__tests__/rest.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/raw/__tests__/graphql.command.spec.ts \
  packages/twenty-sdk/src/cli/program.ts \
  packages/twenty-sdk/src/cli/help.ts \
  packages/twenty-sdk/src/cli/help.txt \
  packages/twenty-sdk/src/cli/__tests__/help.spec.ts
git commit -m "refactor: move raw commands under explicit namespace"
```

### Task 7: Convert Simple Router Families To Explicit Subcommands

**Files:**
- Modify: `packages/twenty-sdk/src/cli/commands/api-keys/api-keys.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-keys/__tests__/api-keys.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/webhooks/webhooks.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/webhooks/__tests__/webhooks.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/route-triggers/route-triggers.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/route-triggers/__tests__/route-triggers.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/postgres-proxy/postgres-proxy.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/postgres-proxy/__tests__/postgres-proxy.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/marketplace-apps/marketplace-apps.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/marketplace-apps/__tests__/marketplace-apps.command.spec.ts`

- [ ] **Step 1: Write failing registration/help tests for one full simple-router batch**

For each command family, replace router-argument expectations with explicit subcommand expectations:

```ts
it("api-keys registers list/get/create/update/revoke/assign-role as real subcommands", () => {});
it("webhooks registers list/get/create/update/delete as real subcommands", () => {});
it("postgres-proxy registers get/enable/disable as real subcommands", () => {});
```

- [ ] **Step 2: Run the simple-router command specs and verify they fail**

Run:

```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts \
  src/cli/commands/api-keys/__tests__/api-keys.command.spec.ts \
  src/cli/commands/webhooks/__tests__/webhooks.command.spec.ts \
  src/cli/commands/route-triggers/__tests__/route-triggers.command.spec.ts \
  src/cli/commands/postgres-proxy/__tests__/postgres-proxy.command.spec.ts \
  src/cli/commands/marketplace-apps/__tests__/marketplace-apps.command.spec.ts --reporter verbose
```

Expected:

- FAIL because these commands still expose a single router argument instead of real subcommands

- [ ] **Step 3: Rewrite the simple router families as explicit subcommand trees**

Use a consistent Commander pattern like:

```ts
const apiKeys = program.command("api-keys").description("Manage API keys");
const listCmd = apiKeys.command("list").description("List API keys");
applyGlobalOptions(listCmd);
listCmd.action(async (_opts, command) => {
  const { globalOptions, services } = createCommandContext(command);
  // existing list implementation
});
```

Remove the router argument and move operation-specific args/options onto the specific subcommands.

- [ ] **Step 4: Re-run the simple-router specs and verify they pass**

Run the same command from Step 2.

Expected:

- PASS on subcommand registration and help assertions

- [ ] **Step 5: Commit**

```bash
git add packages/twenty-sdk/src/cli/commands/api-keys/api-keys.command.ts \
  packages/twenty-sdk/src/cli/commands/api-keys/__tests__/api-keys.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/webhooks/webhooks.command.ts \
  packages/twenty-sdk/src/cli/commands/webhooks/__tests__/webhooks.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/route-triggers/route-triggers.command.ts \
  packages/twenty-sdk/src/cli/commands/route-triggers/__tests__/route-triggers.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/postgres-proxy/postgres-proxy.command.ts \
  packages/twenty-sdk/src/cli/commands/postgres-proxy/__tests__/postgres-proxy.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/marketplace-apps/marketplace-apps.command.ts \
  packages/twenty-sdk/src/cli/commands/marketplace-apps/__tests__/marketplace-apps.command.spec.ts
git commit -m "refactor: convert simple router commands to subcommands"
```

### Task 8: Convert Domain Router Families And Single-Purpose Routers

**Files:**
- Modify: `packages/twenty-sdk/src/cli/commands/approved-access-domains/approved-access-domains.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/approved-access-domains/__tests__/approved-access-domains.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/emailing-domains/emailing-domains.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/emailing-domains/__tests__/emailing-domains.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/public-domains/public-domains.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/public-domains/__tests__/public-domains.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/skills/skills.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/skills/__tests__/skills.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/event-logs/event-logs.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/event-logs/__tests__/event-logs.command.spec.ts`

- [ ] **Step 1: Write failing tests for explicit domain subcommands and event-log shape**

Add expectations like:

```ts
it("approved-access-domains registers list/delete/validate as explicit subcommands", () => {});
it("event-logs registers list as an explicit subcommand instead of requiring a literal operation argument", () => {});
```

- [ ] **Step 2: Run the domain-router command specs and verify they fail**

Run:

```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts \
  src/cli/commands/approved-access-domains/__tests__/approved-access-domains.command.spec.ts \
  src/cli/commands/emailing-domains/__tests__/emailing-domains.command.spec.ts \
  src/cli/commands/public-domains/__tests__/public-domains.command.spec.ts \
  src/cli/commands/skills/__tests__/skills.command.spec.ts \
  src/cli/commands/event-logs/__tests__/event-logs.command.spec.ts --reporter verbose
```

Expected:

- FAIL because these commands still depend on a router argument instead of explicit subcommands

- [ ] **Step 3: Rewrite these commands to explicit subcommands**

Key rules:

- move operation-specific args/options onto the subcommand that uses them
- keep explicit metadata for mutating operations such as `validate`
- make `event-logs list` a real subcommand with its own options

- [ ] **Step 4: Re-run the domain-router specs and verify they pass**

Run the same command from Step 2.

Expected:

- PASS on explicit subcommand contracts and help assertions

- [ ] **Step 5: Commit**

```bash
git add packages/twenty-sdk/src/cli/commands/approved-access-domains/approved-access-domains.command.ts \
  packages/twenty-sdk/src/cli/commands/approved-access-domains/__tests__/approved-access-domains.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/emailing-domains/emailing-domains.command.ts \
  packages/twenty-sdk/src/cli/commands/emailing-domains/__tests__/emailing-domains.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/public-domains/public-domains.command.ts \
  packages/twenty-sdk/src/cli/commands/public-domains/__tests__/public-domains.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/skills/skills.command.ts \
  packages/twenty-sdk/src/cli/commands/skills/__tests__/skills.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/event-logs/event-logs.command.ts \
  packages/twenty-sdk/src/cli/commands/event-logs/__tests__/event-logs.command.spec.ts
git commit -m "refactor: convert domain router commands to subcommands"
```

### Task 9: Convert Complex Router Families To Explicit Subcommands

**Files:**
- Modify: `packages/twenty-sdk/src/cli/commands/api/api.command.ts`
- Create: `packages/twenty-sdk/src/cli/commands/api/__tests__/api.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/batch-create.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/batch-delete.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/batch-update.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/bulk-filter.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/create.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/delete.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/destroy.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/export.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/find-duplicates.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/get.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/group-by.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/import.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/list.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/merge.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/restore.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/types.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/update.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/__tests__/find-duplicates.operation.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/__tests__/group-by.operation.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/__tests__/operations.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/api-metadata.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/__tests__/api-metadata.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-create.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-delete.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-get.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-list.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-update.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-create.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-delete.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-get.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-list.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-update.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/types.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/ui-metadata.operations.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/__tests__/fields-list.operation.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/__tests__/metadata-operations.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/__tests__/ui-metadata.operations.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/files/files.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/files/__tests__/files.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/serverless/serverless.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/serverless/__tests__/serverless.command.spec.ts`

- [ ] **Step 1: Write failing command-level tests for explicit complex subcommands**

Add command tests that assert each family is now a real subcommand tree:

```ts
it("api registers list/get/create/update/delete/destroy/batch-create/batch-update/batch-delete/import/export/find-duplicates/group-by/merge/restore as explicit subcommands", () => {});
it("api-metadata registers object and field operations as explicit subcommands", () => {});
it("files registers upload/download/public-asset as explicit subcommands", () => {});
it("serverless registers list/create/create-layer/publish/source/logs as explicit subcommands", () => {});
```

- [ ] **Step 2: Run the complex-family tests and verify they fail**

Run:

```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts \
  src/cli/commands/api/__tests__/api.command.spec.ts \
  src/cli/commands/api/operations/__tests__/operations.spec.ts \
  src/cli/commands/api/operations/__tests__/find-duplicates.operation.spec.ts \
  src/cli/commands/api/operations/__tests__/group-by.operation.spec.ts \
  src/cli/commands/api-metadata/__tests__/api-metadata.command.spec.ts \
  src/cli/commands/api-metadata/operations/__tests__/metadata-operations.spec.ts \
  src/cli/commands/api-metadata/operations/__tests__/fields-list.operation.spec.ts \
  src/cli/commands/api-metadata/operations/__tests__/ui-metadata.operations.spec.ts \
  src/cli/commands/files/__tests__/files.command.spec.ts \
  src/cli/commands/serverless/__tests__/serverless.command.spec.ts --reporter verbose
```

Expected:

- FAIL because these families still depend on operation routers and generic arguments

- [ ] **Step 3: Rewrite the complex families to explicit subcommands while keeping existing operation logic reusable**

Preferred pattern:

```ts
const api = program.command("api").description("Record operations");
const listCmd = api.command("list").argument("<object>");
applyGlobalOptions(listCmd);
listCmd.action(async (object, options, command) => {
  const { globalOptions, services } = createCommandContext(command);
  await runListOperation({ object, options: command.opts(), services, globalOptions });
});
```

Do the same for:

- `api-metadata`
- `files`
- `serverless`

Keep operation implementation files where that remains the cleanest boundary; only the command tree changes.

- [ ] **Step 4: Re-run the complex-family tests and verify they pass**

Run the same command from Step 2.

Expected:

- PASS on explicit subcommand registration
- PASS on existing operation behavior tests

- [ ] **Step 5: Commit**

```bash
git add packages/twenty-sdk/src/cli/commands/api/api.command.ts \
  packages/twenty-sdk/src/cli/commands/api/__tests__/api.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/batch-create.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/batch-delete.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/batch-update.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/bulk-filter.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/create.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/delete.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/destroy.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/export.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/find-duplicates.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/get.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/group-by.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/import.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/list.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/merge.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/restore.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/types.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/update.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/__tests__/find-duplicates.operation.spec.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/__tests__/group-by.operation.spec.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/__tests__/operations.spec.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/api-metadata.command.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/__tests__/api-metadata.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-create.operation.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-delete.operation.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-get.operation.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-list.operation.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/fields-update.operation.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-create.operation.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-delete.operation.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-get.operation.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-list.operation.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/objects-update.operation.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/types.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/ui-metadata.operations.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/__tests__/fields-list.operation.spec.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/__tests__/metadata-operations.spec.ts \
  packages/twenty-sdk/src/cli/commands/api-metadata/operations/__tests__/ui-metadata.operations.spec.ts \
  packages/twenty-sdk/src/cli/commands/files/files.command.ts \
  packages/twenty-sdk/src/cli/commands/files/__tests__/files.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/serverless/serverless.command.ts \
  packages/twenty-sdk/src/cli/commands/serverless/__tests__/serverless.command.spec.ts
git commit -m "refactor: convert complex command families to subcommands"
```

### Task 10: Standardize Destructive Semantics, Pagination, Payload Vocabulary, And Help Metadata

**Files:**
- Create: `packages/twenty-sdk/src/cli/utilities/shared/confirmation.ts`
- Create: `packages/twenty-sdk/src/cli/utilities/shared/__tests__/confirmation.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/api.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/__tests__/api.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/delete.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/destroy.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/api/operations/batch-delete.operation.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/approved-access-domains/approved-access-domains.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/approved-access-domains/__tests__/approved-access-domains.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/route-triggers/route-triggers.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/route-triggers/__tests__/route-triggers.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/public-domains/public-domains.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/public-domains/__tests__/public-domains.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/emailing-domains/emailing-domains.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/emailing-domains/__tests__/emailing-domains.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/search/search.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/search/__tests__/search.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/event-logs/event-logs.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/event-logs/__tests__/event-logs.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/mcp/mcp.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/mcp/__tests__/mcp.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/applications/applications.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/applications/__tests__/applications.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/serverless/serverless.command.ts`
- Modify: `packages/twenty-sdk/src/cli/commands/serverless/__tests__/serverless.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/help.ts`
- Modify: `packages/twenty-sdk/src/cli/__tests__/help.spec.ts`

- [ ] **Step 1: Write failing tests for destructive `--yes` enforcement and canonical vocabulary**

Add tests like:

```ts
it("requires --yes for destructive delete/destroy flows", async () => {});
it("uses consistent pagination naming for list/search style commands", () => {});
it("uses a documented canonical inline/file payload pattern", () => {});
```

- [ ] **Step 2: Run the targeted command/help tests and verify they fail**

Run:

```bash
pnpm --filter twenty-sdk exec vitest run -c vitest.config.ts \
  src/cli/commands/api/__tests__/api.command.spec.ts \
  src/cli/commands/search/__tests__/search.command.spec.ts \
  src/cli/commands/event-logs/__tests__/event-logs.command.spec.ts \
  src/cli/__tests__/help.spec.ts --reporter verbose
```

Expected:

- FAIL because destructive commands still vary widely
- FAIL because pagination/payload vocabulary is not yet normalized

- [ ] **Step 3: Implement the consistency pass**

Introduce a shared helper:

```ts
export function requireYes(options: { yes?: boolean }, action: string): void {
  if (!options.yes) {
    throw new CliError(
      `${action} requires --yes.`,
      "INVALID_ARGUMENTS",
      `Re-run with --yes to confirm ${action.toLowerCase()}.`,
    );
  }
}
```

Apply it repo-wide to destructive commands.

Normalize vocabulary rules:

- `--query` = output filtering
- `--limit` + `--cursor` = canonical pagination flags for list/search commands
- commands currently using `--first` or `--after` migrate to `--limit` and `--cursor`
- generic JSON payload input uses `--data` + `--file`
- raw GraphQL request input uses `--document` + `--file`
- search keeps `--filter` + `--filter-file` because the payload is semantically distinct from the main query string
- destructive acknowledgements use `--yes` only; remove `--force` rather than keeping two confirmation flags
- make help metadata explicit where inference is not reliable

- [ ] **Step 4: Re-run the targeted tests and verify they pass**

Run the same command from Step 2.

Expected:

- PASS on `--yes` enforcement
- PASS on canonical help/option vocabulary assertions

- [ ] **Step 5: Commit**

```bash
git add packages/twenty-sdk/src/cli/utilities/shared/confirmation.ts \
  packages/twenty-sdk/src/cli/utilities/shared/__tests__/confirmation.spec.ts \
  packages/twenty-sdk/src/cli/commands/api/api.command.ts \
  packages/twenty-sdk/src/cli/commands/api/__tests__/api.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/delete.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/destroy.operation.ts \
  packages/twenty-sdk/src/cli/commands/api/operations/batch-delete.operation.ts \
  packages/twenty-sdk/src/cli/commands/approved-access-domains/approved-access-domains.command.ts \
  packages/twenty-sdk/src/cli/commands/approved-access-domains/__tests__/approved-access-domains.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/route-triggers/route-triggers.command.ts \
  packages/twenty-sdk/src/cli/commands/route-triggers/__tests__/route-triggers.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/public-domains/public-domains.command.ts \
  packages/twenty-sdk/src/cli/commands/public-domains/__tests__/public-domains.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/emailing-domains/emailing-domains.command.ts \
  packages/twenty-sdk/src/cli/commands/emailing-domains/__tests__/emailing-domains.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/search/search.command.ts \
  packages/twenty-sdk/src/cli/commands/search/__tests__/search.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/event-logs/event-logs.command.ts \
  packages/twenty-sdk/src/cli/commands/event-logs/__tests__/event-logs.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/mcp/mcp.command.ts \
  packages/twenty-sdk/src/cli/commands/mcp/__tests__/mcp.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/applications/applications.command.ts \
  packages/twenty-sdk/src/cli/commands/applications/__tests__/applications.command.spec.ts \
  packages/twenty-sdk/src/cli/commands/serverless/serverless.command.ts \
  packages/twenty-sdk/src/cli/commands/serverless/__tests__/serverless.command.spec.ts \
  packages/twenty-sdk/src/cli/help.ts
git commit -m "refactor: standardize CLI safety and option semantics"
```

### Task 11: Run Full Verification And Live Smoke Tests

**Files:**
- Test: `package.json`
- Test: `packages/twenty-sdk/src/cli/program.ts`
- Test: `packages/twenty-sdk/src/cli/__tests__/help.spec.ts`
- Test: `packages/twenty-sdk/src/cli/utilities/shared/__tests__/utilities.spec.ts`
- Test: `packages/twenty-sdk/src/cli/utilities/api/services/__tests__/public-http.service.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/raw/__tests__/raw.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/raw/__tests__/rest.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/raw/__tests__/graphql.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/routes/__tests__/routes.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/workflows/__tests__/workflows.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/auth/__tests__/auth.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/search/__tests__/search.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/connected-accounts/__tests__/connected-accounts.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/message-channels/__tests__/message-channels.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/calendar-channels/__tests__/calendar-channels.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/api-keys/__tests__/api-keys.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/webhooks/__tests__/webhooks.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/route-triggers/__tests__/route-triggers.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/postgres-proxy/__tests__/postgres-proxy.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/marketplace-apps/__tests__/marketplace-apps.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/approved-access-domains/__tests__/approved-access-domains.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/emailing-domains/__tests__/emailing-domains.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/public-domains/__tests__/public-domains.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/skills/__tests__/skills.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/event-logs/__tests__/event-logs.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/api/__tests__/api.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/api/operations/__tests__/operations.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/api/operations/__tests__/find-duplicates.operation.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/api/operations/__tests__/group-by.operation.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/api-metadata/__tests__/api-metadata.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/__tests__/metadata-operations.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/__tests__/fields-list.operation.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/api-metadata/operations/__tests__/ui-metadata.operations.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/files/__tests__/files.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/serverless/__tests__/serverless.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/mcp/__tests__/mcp.command.spec.ts`
- Test: `packages/twenty-sdk/src/cli/commands/applications/__tests__/applications.command.spec.ts`

- [ ] **Step 1: Run the full automated verification suite**

Run:

```bash
pnpm test
pnpm build
pnpm lint
```

Expected:

- all tests pass
- build passes
- lint passes or reports only known pre-existing unrelated warnings that are documented before merging

- [ ] **Step 2: Run safe CLI smoke tests against a configured smoke workspace**

Run:

```bash
node packages/twenty-sdk/dist/cli/cli.js auth status --workspace "$TWENTY_SMOKE_WORKSPACE" -o json
node packages/twenty-sdk/dist/cli/cli.js search "Acme" --workspace "$TWENTY_SMOKE_WORKSPACE" -o json
node packages/twenty-sdk/dist/cli/cli.js mcp status --workspace "$TWENTY_SMOKE_WORKSPACE" -o json
node packages/twenty-sdk/dist/cli/cli.js raw graphql query --document 'query CurrentWorkspace { currentWorkspace { id displayName } }' --workspace "$TWENTY_SMOKE_WORKSPACE" -o json
node packages/twenty-sdk/dist/cli/cli.js openapi core --workspace "$TWENTY_SMOKE_WORKSPACE" -o json
```

Expected:

- authenticated commands reach the workspace successfully
- `mcp status` still reports a reachable endpoint with `ai_feature_disabled` unless the workspace configuration has changed

- [ ] **Step 3: Run public-command smoke tests only if safe targets are explicitly configured**

Run:

```bash
if [ -n "${TWENTY_SAFE_ROUTE_PATH:-}" ]; then
  node packages/twenty-sdk/dist/cli/cli.js routes invoke "$TWENTY_SAFE_ROUTE_PATH" --method get --workspace "$TWENTY_SMOKE_WORKSPACE" -o json
else
  echo "Skipping routes smoke test: TWENTY_SAFE_ROUTE_PATH is not set."
fi

if [ -n "${TWENTY_SAFE_WORKFLOW_ID:-}" ] && [ -n "${TWENTY_SAFE_WORKSPACE_ID:-}" ]; then
  node packages/twenty-sdk/dist/cli/cli.js workflows invoke-webhook "$TWENTY_SAFE_WORKFLOW_ID" --workspace-id "$TWENTY_SAFE_WORKSPACE_ID" --method get --workspace "$TWENTY_SMOKE_WORKSPACE" -o json
else
  echo "Skipping workflow webhook smoke test: TWENTY_SAFE_WORKFLOW_ID or TWENTY_SAFE_WORKSPACE_ID is not set."
fi
```

Expected:

- If safe live endpoints are configured, verify `routes` and `workflows invoke-webhook` run through the rewritten public transport.
- If safe live endpoints are not configured, the shell output documents an explicit skip and the command/service test suites remain the primary proof.

- [ ] **Step 4: Address any verification regressions with minimal follow-up commits**

For each failing suite:

- write the smallest missing regression test if one does not already exist
- make the smallest fix
- rerun only the failing target, then rerun the full verification command from Step 1

- [ ] **Step 5: Commit the verification-fix delta and prepare handoff**

```bash
git add -A packages/twenty-sdk/src/cli packages/twenty-sdk/src/cli/utilities
git commit -m "chore: finalize CLI handler rewrite verification"
```

## Final Handoff Checklist

- [ ] All new/updated command specs pass
- [ ] Shared contract tests pass
- [ ] `pnpm test` passes
- [ ] `pnpm build` passes
- [ ] `pnpm lint` passes or any remaining warnings are explicitly documented
- [ ] Configured smoke-workspace tests completed or explicitly skipped with reason
- [ ] Help text/examples updated for rewritten namespaces and flags
- [ ] Breaking CLI changes summarized before merge
