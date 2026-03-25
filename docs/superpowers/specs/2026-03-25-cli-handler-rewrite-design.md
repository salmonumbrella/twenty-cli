# Twenty CLI Handler Rewrite Design

Date: 2026-03-25

## Goal

Rewrite the Twenty CLI around a single internal command architecture that is consistent, testable, and easier to extend, while deliberately cleaning up awkward command surfaces that are not yet depended on in production.

This rewrite should:

- keep Commander as the CLI framework
- standardize command handlers repo-wide
- consolidate authenticated and public HTTP behavior behind shared services
- fix current help and agent-output contract bugs
- allow intentional CLI breaking changes where they materially improve consistency

## Context

The current CLI already covers a large amount of Twenty functionality: records, metadata, raw REST/GraphQL, roles, files, search, workflows, routes, MCP, and several admin surfaces.

The main problems are no longer about missing commands. They are about inconsistent internal architecture:

- some commands use shared service/context helpers
- some commands instantiate services ad hoc
- some commands bypass the shared transport entirely with raw `axios`
- some shared contract logic treats real commands as special cases because their names start with `help`

That inconsistency is already causing real defects:

- `mcp help-center` initially broke help traversal because shared help logic filtered out any command whose name started with `help`
- `mcp help-center -o agent` still derives the wrong output `kind` because command-path derivation also filters by `help*`
- `routes invoke` and `workflows invoke-webhook` expose shared global flags like `--debug` and `--no-retry`, but currently bypass the shared HTTP transport and therefore do not actually honor those flags consistently

The repo also already contains the beginnings of the right architecture:

- `createCommandContext()` and `createOutputContext()`
- a shared `CliServices` container
- `ApiService`, `McpService`, `RecordsService`, `MetadataService`, and shared output helpers

This rewrite should finish that work rather than replacing the CLI framework itself.

## Problem Statement

The rewrite is intended to resolve three categories of issues.

### High Priority

- Fix shared command-path logic so legitimate `help-*` commands are treated as real commands in all shared layers.
- Ensure agent output `kind` values reflect the actual command path, including `help-*` segments.

### Medium Priority

- Remove raw public/semi-public HTTP calls from command modules and route them through a shared service.
- Ensure any command that exposes shared transport flags actually honors them.
- Standardize error shaping, retry behavior, debug logging, and config resolution for public/semi-public invocations.

### Low Priority

- Remove handler/service drift across the repo.
- Stop recreating services inside command modules when the service container already provides them.
- Reduce bespoke command plumbing in helpers like the shared record-resource command.
- Clean up awkward or inconsistent CLI interfaces while the rewrite is already changing command internals.

## Decision Summary

This design makes five explicit decisions:

1. **Stay on Commander**
2. **Do a repo-wide internal handler rewrite**
3. **Add a shared `PublicHttpService` for public/semi-public transport**
4. **Allow deliberate CLI cleanup, including breaking changes, where the current surface is inconsistent or awkward**
5. **Prefer explicit command contracts over inferred router behavior**

## Why Commander Stays

This rewrite does **not** migrate to `oclif`.

Rationale:

- The current issues are not parser/framework limitations. They are handler, transport, and shared-contract consistency issues.
- The CLI already has a large amount of Commander-based behavior and test coverage, especially around custom help JSON and output contracts.
- A framework migration would not remove the need for the actual architectural work: shared command context, shared service container, shared public transport, and shared contract fixes.
- Rewriting both the framework and the handler architecture at the same time would increase churn without directly solving the bugs that motivated this effort.

This design therefore treats Commander as stable infrastructure and focuses the rewrite on internal architecture and CLI surface quality.

## Non-Goals

This rewrite does not include:

- migrating from Commander to `oclif`, Clipanion, or another framework
- adding a plugin system or external extension API
- redesigning the CLI around a generated manifest/DSL framework
- preserving awkward command surfaces only for backward compatibility
- changing output formats, exit-code classes, or the existence of help JSON as a concept
- changing Twenty server behavior or API semantics

## Design Principles

- **One handler model:** commands should follow one standard action pattern.
- **One transport per request class:** authenticated API calls and public/semi-public HTTP calls should each go through a shared service.
- **Exact command-path semantics:** shared logic must treat real command names literally, including names like `help-center`.
- **Thin command modules:** command files should parse command-specific inputs, delegate to services, and render output.
- **Contracts over nostalgia:** preserve tested contracts that matter; do not preserve internal inconsistency.
- **Deliberate breakage only:** breaking CLI cleanup is allowed, but only where it produces a clearer and more consistent surface.
- **No framework invention:** use Commander directly; add only the minimal internal abstractions needed for consistency.

## Target Architecture

### Command Runtime Model

`createCommandContext()` becomes the default entrypoint for command handlers.

Target shape:

- resolve global options once
- hydrate environment once
- construct a single shared service container
- run command-specific logic against that context
- render output through the shared output service

In practice, most command actions should look conceptually like:

1. parse resource-specific args/options
2. create command context
3. call a domain/service method
4. render output

`createOutputContext()` may remain for the small number of commands that truly need output only, but it should become the exception rather than the default for networked commands.

### Shared Service Container

`CliServices` should become the stable dependency boundary for command modules.

Target contents:

- `config`
- `api`
- `publicHttp`
- `records`
- `metadata`
- `search`
- `mcp`
- `output`
- `importer`
- `exporter`

Current shared services already in place should be reused rather than replaced. The rewrite should expand the container so commands no longer need to instantiate `ConfigService`, `RecordsService`, or other domain services ad hoc.

### Transport Split

The rewrite should formalize two request classes.

#### `ApiService`

Use for authenticated Twenty API traffic:

- REST
- GraphQL
- metadata
- MCP
- admin/resource commands

Responsibilities remain:

- authenticated base URL resolution
- bearer token injection
- retry handling
- debug logging

#### `PublicHttpService`

Add a new shared service for public or semi-public HTTP flows.

Use for commands like:

- `routes invoke`
- `workflows invoke-webhook`
- any future public route/webhook/signed-link style commands

Responsibilities:

- resolve base URL from workspace/env/config
- support explicit auth modes per request or command
- support the same retry/debug semantics as shared authenticated transport
- shape errors consistently into CLI-level failures
- expose a small request interface similar to `ApiService`

This service exists to remove raw `axios` from command modules and to ensure public commands actually honor the global transport flags they already expose.

Required auth modes:

- `none`: never attach bearer auth
- `optional`: attach bearer auth if available; continue without it if absent
- `required`: require bearer auth and fail clearly if unavailable

Planning constraint:

- `PublicHttpService` should not guess auth behavior implicitly from whether a token exists
- command implementations must choose an auth mode deliberately

Expected command usage:

- `routes invoke`: `optional` by default
- `workflows invoke-webhook`: `optional` for the webhook request itself
- workspace-ID discovery for public workflow invocation: `required` unless `--workspace-id` is explicitly provided

### Shared Config Resolution

The rewrite should centralize config/base URL/token resolution logic so command files do not need to know how to:

- load config files
- pick the active workspace
- read env overrides
- decide whether auth is required or optional

The exact implementation can be a small addition to `ConfigService` or a nearby helper, but the design goal is simple:

- command modules should not manually stitch together workspace config and env vars
- commands that need public transport should rely on resolved connection data from shared services rather than encoding token/base-URL fallback rules locally

## Standardized Handler Pattern

Repo-wide, command handlers should follow a single pattern:

- no direct environment hydration in the command file
- no direct `ConfigService` construction in the command file unless the command’s primary purpose is config management itself
- no direct `axios` in the command file
- no ad hoc re-instantiation of services already present in `CliServices`

### What Changes

- commands like `search` should consume `services.search` rather than `new SearchService(services.api)`
- shared helpers like the record-resource command should consume `services.records`
- `auth` subcommands that only read/write local config may still use config-specific helpers, but they should still follow the same command-context rules for env/output handling
- `routes` and public workflow commands should consume `services.publicHttp`

### What Does Not Change

- command registration still happens with Commander
- command modules still own their command trees, descriptions, examples, and arguments
- not every domain needs a large new abstraction layer if the existing service is already sufficient

## Shared Helper Rewrite

The repo-wide rewrite includes shared helper paths, not just leaf commands.

### Record Resource Helper

`registerRecordResourceCommand()` should be rewritten to consume shared command context and shared `records` service rather than instantiating `RecordsService` internally.

Why this matters:

- helpers are where architectural drift gets copied
- if the helper keeps the old pattern, the rewrite will keep reproducing the same inconsistency

### Operation Dispatch Helpers

Operation-style surfaces such as:

- `api`
- `api-metadata`
- similar operation-dispatch resources

should use the same command-context and service patterns as all other commands.

The rewrite may introduce small helper functions for operation dispatch, but it should avoid inventing a new internal CLI framework.

## CLI Surface Standardization

This rewrite is allowed to improve the CLI surface itself, not just the internals.

The target is not “make everything look different.” The target is to remove classes of inconsistency that are currently forcing the help system, output contracts, and command implementations to guess at intent.

### Multi-Operation Command Shape

The current CLI mixes:

- real subcommand trees
- large `resource <operation>` routers

The help system currently has to infer operation metadata from router argument descriptions, which is weaker than explicit command structure.

Preferred end-state:

- multi-operation command families should use real subcommands where that materially improves per-operation contracts, help, and argument specificity

Examples of likely candidates:

- `api`
- `api-metadata`
- `files`
- `serverless`

However, the implementation plan may keep some router-style surfaces if:

- converting them in the first pass would create disproportionate risk
- explicit operation metadata can provide a sufficiently strong contract in the meantime

The planning standard should be:

- explicit subcommands are preferred
- inferred operation contracts are temporary at best

### Global Option Contract

The rewrite should make shared option meanings actually shared.

End-state rule:

- `--query` means output filtering / JMESPath

Implication:

- raw GraphQL request text should no longer use `--query` as its input flag
- a clearer request-document flag such as `--document` should become the canonical GraphQL input option

The implementation plan may choose whether to:

- make the break cleanly and reserve `--query` repo-wide immediately

This spec intentionally does **not** authorize a temporary compatibility alias for GraphQL request text under `--query`.

### Pagination And Payload Vocabulary

Similar commands should stop drifting across:

- `--limit`
- `--cursor`
- `--after`
- `--first`
- bespoke inline/file flag pairs

Target direction:

- standardize a preferred pagination vocabulary for list/search style commands
- standardize a preferred inline/file input vocabulary for JSON payloads
- keep specialized flags only where the payload is genuinely domain-specific

This does not require every command to have identical options, but it does require the repo to stop accumulating arbitrary synonyms without intent.

### Raw Escape Hatch Namespace

Raw escape hatches should be clearly separated from curated command surfaces.

Preferred end-state:

- raw REST and raw GraphQL move under an explicit raw namespace rather than living as top-level peers of curated commands

Example direction:

- `twenty raw rest`
- `twenty raw graphql`

This is an intentionally breaking cleanup that improves root discovery and makes escape hatches visually distinct from supported curated workflows.

### Destructive Command Policy

Destructive commands should follow one explicit safety rule across the repo.

Preferred end-state:

- destructive commands require an explicit `--yes` acknowledgement
- help text and option descriptions accurately describe the real behavior

Interactive confirmation is optional, but if introduced it should be consistent. The simpler and more automation-friendly baseline is explicit `--yes`.

### Explicit Help Metadata

The rewrite should reduce inference in the help contract layer.

End-state direction:

- mutability should be explicit per command/operation where inference is unreliable
- operation metadata should come from declared structure or declared metadata, not just parsed prose

This is especially important for commands whose operation names are not a reliable proxy for whether they mutate server state.

## Help And Output Contract Fixes

This rewrite must fix the underlying shared contract behavior, not patch individual commands.

### Exact Command Paths

Any shared code that derives command paths or resolves commands must use exact command names.

Specifically:

- no shared code should exclude commands because their names merely start with `help`
- only Commander’s built-in `help` command should be treated specially

### Agent Output Kind

`outputKind` must be derived from the full real command path.

Examples:

- `twenty mcp` -> `twenty.mcp`
- `twenty mcp help-center` -> `twenty.mcp.help-center`
- future legitimate `help-*` commands must preserve their full path as well

### Help JSON

The rewrite must preserve:

- root help JSON
- command help JSON
- root text help
- `--help-json` / `--hj`

But the internal logic used to derive path, operations, and subcommands may be rewritten as needed.

## Deliberate CLI Cleanup Policy

Breaking changes are allowed in this rewrite, but they should be **intentional and categorized**, not incidental.

Allowed cleanup categories:

- convert ambiguous router-style surfaces into explicit subcommands
- normalize inconsistent option naming
- normalize argument ordering where current ordering is awkward
- reserve shared global flag names for shared meanings
- move raw escape hatches under a dedicated namespace
- remove or rename low-value aliases that increase inconsistency
- improve command descriptions/examples/help summaries
- standardize destructive command acknowledgement semantics
- make raw/public command behavior line up with what their flags imply

Disallowed cleanup categories for this rewrite:

- large conceptual redesign of the entire command tree for aesthetic reasons alone
- adding parallel compatibility layers for old and new command surfaces
- speculative new features unrelated to handler/transport consistency

The implementation plan should explicitly mark which command-surface changes are:

- behavior-preserving
- additive
- intentionally breaking

## Migration Strategy

This is a repo-wide rewrite, but the design should still support staged execution.

Recommended migration order:

1. fix shared command-path and output-kind logic
2. expand shared service container
3. introduce `PublicHttpService`
4. migrate the shared helper layer
5. migrate the public-route and webhook commands
6. migrate remaining command namespaces to the standardized handler pattern
7. apply CLI cleanup pass where the rewrite exposes awkward surfaces
8. refresh help contracts and examples

This is not the implementation plan. It is the intended sequencing logic for planning.

## Testing Strategy

The rewrite should be verified at four levels.

### 1. Shared Contract Tests

Add focused contract tests for:

- agent output `kind` on a real `help-*` command such as `mcp help-center`
- shared command-path resolution for `help-*` commands
- commands that expose shared transport flags and must honor them after migration

Required examples:

- `mcp help-center -o agent` returns `kind: "twenty.mcp.help-center"`
- public-route/public-webhook commands use shared debug/retry behavior rather than bypassing it

### 2. Existing Test Suite

The existing CLI suite should remain green after the rewrite.

That includes:

- help contract tests
- command registration tests
- service tests
- raw/graphql/rest tests
- command-specific tests across the repo

### 3. New Command/Service Tests

Add or update targeted tests for:

- `PublicHttpService`
- migrated public commands
- rewritten shared helpers
- intentionally changed CLI interfaces

### 4. Real Smoke Tests Against A Configured Production Workspace

The final verification should include smoke tests against the real workspace.

Default smoke scope should prefer non-destructive commands:

- `auth status`
- `search`
- safe raw GraphQL/REST reads
- `mcp status`
- public route/webhook invocations that are known safe

If destructive or mutating smoke tests are needed, they should be explicitly identified and gated during implementation planning.

Known expected prod condition:

- MCP endpoint is reachable
- MCP status currently reports AI disabled for the workspace

That expectation should be treated as a planning-time assumption to revalidate during implementation, not as a timeless truth.

## Risks

### Wide Diff Risk

A repo-wide rewrite will touch many files. That increases review and regression risk.

Mitigation:

- keep the new architecture simple
- avoid speculative abstractions
- preserve tested contracts wherever possible

### Silent Contract Drift

Because the CLI has custom help and agent-output behavior, small internal changes can quietly alter tool-facing contracts.

Mitigation:

- add focused contract tests first
- keep help/output verification prominent in the implementation plan

### Public vs Authenticated Transport Confusion

Some commands are public, some are semi-public, and some optionally use auth to improve behavior.

Mitigation:

- make `PublicHttpService` explicit
- document optional-auth behavior clearly in command tests and help examples

### Breaking Cleanup Creep

Allowing breaking cleanup can expand scope too far.

Mitigation:

- require each intentional surface change to be explicitly justified in the implementation plan
- separate architectural migration from interface cleanup tasks

## Success Criteria

This rewrite is successful when all of the following are true:

- every networked command follows the standardized command-context pattern or has a documented reason not to
- command modules no longer use raw `axios` for CLI traffic that should be shared
- public/semi-public commands use `PublicHttpService`
- `help-*` commands behave correctly in help resolution and agent output kind derivation
- shared helpers no longer recreate services already present in the service container
- the full test suite passes
- new shared contract tests pass
- smoke tests against a configured production workspace pass with expected real-world results
- the resulting CLI surface is more consistent than the current one, even where that required breaking changes

## Planning Readiness

This spec is ready for implementation planning once the user approves it.

The implementation plan should decide:

- the exact end-state shape of `CliServices`
- the exact API surface for `PublicHttpService`
- the prioritized order of command migrations
- the list of intentionally breaking CLI cleanups to include in the first rewrite pass
