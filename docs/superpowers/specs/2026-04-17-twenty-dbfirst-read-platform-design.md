# Twenty DB-First Read Platform Design

Date: 2026-04-17
Status: Proposed
Scope: Implicit DB-first read architecture for `twenty-cli`

## Summary

`twenty-cli` should adopt a DB-first read architecture for self-hosted and
proxy-enabled Twenty workspaces, using the current `postgres-proxy`
capability as the preferred bootstrap path and keeping all write actions on
the official API.

The design should follow the successful shape of `chatwoot-cli`, not
`beeper-cli`.

What to copy from `chatwoot-cli`:

- one backend abstraction owned by the runtime rather than command flags
- DB-first reads when a usable database configuration exists
- narrow fallback rules based on capability gaps
- adapters that preserve existing output contracts
- mutations staying API-only

What not to copy from `chatwoot-cli`:

- public `api | db | auto` mode switches on user-facing read commands

What to copy from `beeper-cli`:

- strong diagnostics for operators

What not to copy from `beeper-cli`:

- local filesystem database discovery

This differs from Chatwoot in one important way: Twenty already exposes a
first-class `postgres-proxy` workflow that can return workspace-scoped
database credentials. Chatwoot does not expose an equivalent product feature,
which is why `chatwoot-cli` had to rely on an operator-provided
`CHATWOOT_DATABASE_URL`. Twenty should take advantage of the product feature it
already has.

## Goals

- Make read-only record retrieval DB-first whenever a usable DB profile or
  explicit DB target is configured.
- Keep backend choice implicit on normal commands.
- Prefer `postgres-proxy get` as the standard credential bootstrap path.
- Support an opt-in cached DB profile store with multiple named profiles per
  workspace such as `prod`, `staging`, and `local`.
- Keep `.env` support first-class so the feature works equally well for plain
  env users and 1Password-backed env injection.
- Preserve current output contracts for `text`, `json`, `jsonl`, `agent`, and
  `csv`.
- Cover relationship-heavy reads, not only flat record lookups.
- Keep all mutations and control-plane actions on the official Twenty API.

## Non-Goals

- Adding public `--read-source`, `--backend`, `db`, `api`, or `auto` switches
  to normal read commands.
- Replacing `postgres-proxy` with a second credential system.
- Rewriting the raw GraphQL or raw REST commands to bypass the public API.
- Moving mutation flows to direct SQL.
- Promising DB parity for every control-plane, metadata, workflow, or admin
  command in the CLI.
- Requiring users to run Tailscale or manually provision raw database access
  just to benefit from DB-first reads.

## Selected Architecture

### Runtime policy

Normal read commands should not expose backend selection.

The runtime should decide the backend automatically:

- if a usable DB target exists for the active workspace, use DB-first reads
- otherwise use the current API path
- if a DB-backed surface is not yet supported safely, fall back to the API
  through the shared read backend
- all mutation commands remain API-only even when DB-first reads are active

This keeps the command UX clean and matches the desired `chatwoot-cli` style:
the backend is an implementation detail, not a user-facing mode.

### Narrow internal override

There should still be one internal support/debug escape hatch, but it should
not be part of the normal CLI UX or help contract.

Recommended internal override:

- `TWENTY_INTERNAL_READ_BACKEND=api`
- `TWENTY_INTERNAL_READ_BACKEND=db`

Rules:

- undocumented in normal README/help output
- used only for support, troubleshooting, and test coverage
- never required for normal operation

### Bootstrap model

The standard bootstrap path should be:

1. user configures an endpoint target
2. CLI fetches or refreshes workspace-scoped credentials through
   `twenty postgres-proxy get`
3. CLI caches that configuration only if the user explicitly opted into a DB
   profile
4. supported read commands then use DB-first automatically

Important upstream constraint:

- current Twenty returns `user`, `password`, and `workspaceId` from
  `getPostgresCredentials`
- current Twenty does not return proxy `host`, `port`, `database`, or
  `sslmode`

Because of that, the CLI still needs one explicit endpoint target input in
addition to proxy-fetched credentials.

### Primary endpoint input

Use one explicit env-backed target variable as the primary input:

- `TWENTY_DB_PROXY_URL`

Recommended shape:

- `postgresql://HOST:PORT/DATABASE?sslmode=require`

This is better than split host/port variables as the primary design because it:

- matches Postgres tooling conventions
- keeps `.env` usage simple
- is easier to cache and reason about
- reduces partially configured states

Split endpoint variables can be added later as compatibility helpers if a real
need appears, but they should not be the primary model.

### DB profile model

DB profiles should be explicit, durable, and multi-profile per workspace.

Recommended public command family:

- `twenty db profile init <name>`
- `twenty db profile list`
- `twenty db profile show [name]`
- `twenty db profile use <name>`
- `twenty db profile test [name]`
- `twenty db profile refresh-creds [name]`
- `twenty db profile remove <name>`
- `twenty db status`

Profiles are scoped under the active Twenty workspace and support multiple
named entries such as:

- `prod`
- `staging`
- `local`

Recommended profile fields:

- `name`
- `workspace`
- `workspaceId`
- `proxyUrl`
- `credentialSource`
- `cachedUser`
- `cachedPassword`
- `lastRefreshedAt`
- `lastValidatedAt`
- `notes` (optional operator annotation)

Profile storage should be opt-in. Normal reads must not silently create durable
local config as a side effect.

### Cached credentials

The cached profile store should persist credentials when the user explicitly
opts into DB profiles.

Rationale:

- faster startup because the CLI can skip a proxy credential fetch in the
  common case
- better fit for frequent lookup-heavy usage
- closer to how famous CLIs handle explicit durable setup

This cache only improves startup and dependency on the API at process start. It
does not make actual DB queries faster after connection establishment.

### Env compatibility

`.env` support remains first-class.

Reasons:

- it matches the current CLI environment model
- it works naturally with 1Password-managed env injection
- it works for users who do not use a secret manager
- it keeps automation and CI straightforward

Recommended precedence:

1. internal debug override
2. explicit env target such as `TWENTY_DB_PROXY_URL`
3. explicit env workspace DB profile selection such as `TWENTY_DB_PROFILE`
4. cached DB profile selected for the active workspace
5. API-only fallback

Credentials should resolve in this order:

1. explicit manual env credential override if supported in a later phase
2. cached profile credentials
3. live `postgres-proxy get` refresh when API auth is available

### Defaulting

Defaulting should be simple:

- if the active workspace has a usable DB profile or an explicit DB proxy URL,
  supported read commands default to DB-first
- if not, they default to the API

No user-facing read command should need to say "use the DB" or "use the API".

### Diagnostics

Diagnostics matter, but they belong in status surfaces rather than mode flags.

Recommended additions:

- extend `twenty auth status` with DB-first visibility
- add `twenty db status`

Suggested status fields:

- whether a DB target is configured
- active DB profile name
- whether cached credentials exist
- whether credentials were refreshed from `postgres-proxy`
- whether the DB is reachable
- which read surfaces are currently DB-capable
- whether supported reads will currently use DB-first or API-first

### Package shape

The TypeScript package layout should mirror the `chatwoot-cli` separation of
concerns, adapted to the current repo:

- `packages/twenty-sdk/src/cli/utilities/db`
  - profile/config loading
  - connection assembly
  - capability detection
  - row adapters
  - SQL-backed read services
- `packages/twenty-sdk/src/cli/utilities/readbackend`
  - backend selection
  - DB-first fallback harness
  - per-surface orchestration
- existing command files
  - unchanged public UX where possible
  - continue owning flag parsing and output rendering

The important boundary is:

- commands stay stable
- services decide whether a read comes from DB or API
- adapters preserve existing response shapes

### Data model strategy

The design should treat DB-first reads as a generic record-read platform, not a
hard-coded set of only `people`, `companies`, and `opportunities`.

Scope should include:

- standard objects
- custom objects
- relationship-heavy reads exposed through current REST depth/include behavior

That means the runtime needs object metadata awareness so it can resolve:

- object names
- relation paths
- field projections
- table/query capability

Metadata reads themselves do not need to become DB-first in phase 1, but
metadata-aware planning for record reads is required.

### Fallback policy

Fallback should be narrow and explicit inside the runtime.

Fallback to the API is allowed when:

- no usable DB target exists
- DB connection cannot be opened
- cached credentials are invalid and live proxy refresh is unavailable
- required tables, columns, or relation mappings are unsupported
- the command shape requests a read surface not yet implemented safely in SQL

Fallback to the API is not allowed for:

- mutation commands
- raw commands that explicitly promise API semantics

Once a read surface reaches parity, an empty DB result should be treated as
authoritative rather than causing fallback.

### Output invariants

DB-backed commands must preserve the same output contracts the CLI already
advertises today.

Invariants:

- command names and normal flags remain stable
- output shape remains stable
- query filtering still runs before formatting
- renderers do not fork by backend
- JSON payload shape does not expose internal backend details

## Exhaustive Command Surface Inventory

This section lists the current command families in the repo and explicitly
marks whether they belong in the DB-first read platform.

### Included command families

| Command family | Status | Scope |
| --- | --- | --- |
| `twenty search` | In scope | DB-first full-text and relationship-aware record search |
| `twenty api` | Partially in scope | Read-only record operations become DB-first where safe |
| `twenty postgres-proxy` | Supporting capability | Credential bootstrap and refresh path, not the read path itself |
| `twenty db` | New command family | Explicit DB profile and diagnostics management |

### Excluded command families

| Command family | Status | Reason |
| --- | --- | --- |
| `twenty auth` | API/control plane | Authentication and workspace config stay API/config-backed; only status gains DB diagnostics |
| `twenty api-metadata` | Out of scope | Metadata/control-plane surface, not phase-1 DB-first target |
| `twenty application-registrations` | Out of scope | Control plane |
| `twenty applications` | Out of scope | Control plane |
| `twenty approved-access-domains` | Out of scope | Control plane |
| `twenty calendar-channels` | Out of scope | Integration/config surface |
| `twenty connected-accounts` | Out of scope | Integration/config surface |
| `twenty dashboards` | Out of scope | UI/control-plane surface |
| `twenty emailing-domains` | Out of scope | Control plane |
| `twenty event-logs` | Out of scope | Operational/analytics surface |
| `twenty files` | Out of scope | File/storage workflow, not record-read platform |
| `twenty marketplace-apps` | Out of scope | Control plane |
| `twenty mcp` | Out of scope | MCP discovery/execution contract should remain API-backed |
| `twenty message-channels` | Out of scope | Integration/config surface |
| `twenty openapi` | Out of scope | Documentation/schema retrieval |
| `twenty public-domains` | Out of scope | Control plane |
| `twenty raw` | Out of scope | Raw commands must keep explicit API semantics |
| `twenty roles` | Out of scope | Control plane |
| `twenty route-triggers` | Out of scope | Control plane |
| `twenty routes` | Out of scope | HTTP route invocation should stay API/HTTP-backed |
| `twenty serverless` | Out of scope | Control plane and execution surface |
| `twenty skills` | Out of scope | Control plane |
| `twenty webhooks` | Out of scope | Control plane |
| `twenty workflows` | Out of scope | Control plane and execution surface |

## Detailed `twenty api` Inventory

`twenty api` is mixed-scope. Only read-only record operations should become
DB-first in this design.

### `twenty api` operations in scope

| Operation | Status | Notes |
| --- | --- | --- |
| `list` | In scope | Core DB-first surface for all supported record objects |
| `get` | In scope | Core DB-first surface, including relationship-heavy expansion |
| `export` | In scope | Should reuse DB-first list/query pipeline for large read volume |
| `group-by` | In scope | DB-backed aggregation where semantics can be matched safely |

### `twenty api` operations partially in scope

| Operation | Status | Notes |
| --- | --- | --- |
| `list --all` | In scope via `list` | DB paging replaces API pagination fan-out |
| `get --include <rels>` | In scope via `get` | Relationship-heavy expansion is a primary DB-first target |
| `export --format json/csv` | In scope via `export` | Rendering stays unchanged; retrieval becomes DB-first |

### `twenty api` operations out of scope

| Operation | Status | Reason |
| --- | --- | --- |
| `create` | Out of scope | Mutation |
| `update` | Out of scope | Mutation |
| `delete` | Out of scope | Mutation |
| `destroy` | Out of scope | Mutation |
| `restore` | Out of scope | Mutation |
| `batch-create` | Out of scope | Mutation |
| `batch-update` | Out of scope | Mutation |
| `batch-delete` | Out of scope | Mutation |
| `import` | Out of scope | Mutation/import workflow |
| `merge` | Out of scope | Mutation and server-side business logic |

### `twenty api` deferred read-like operation

| Operation | Status | Reason |
| --- | --- | --- |
| `find-duplicates` | Deferred | Read-only on paper, but dedupe semantics are business-logic-heavy and better left API-backed until parity is proven |

## Included Data Surfaces

The DB-first read platform should apply to record objects generally, not only
to a short allowlist.

Included object scope:

- standard CRM objects such as people, companies, opportunities, tasks,
  notes, favorites, attachments, and activity-like records where the CLI is
  reading workspace record data
- custom objects created in the active workspace
- related records reachable through current `get`/`list` include-depth behavior

Included relationship-heavy read scope:

- person to company
- company to people
- opportunity relations
- notes/tasks/attachments/favorites/activity relations
- custom object relations
- any relation surfaced through current record `depth` semantics once SQL
  capability exists

The principle is:

- if a command is reading workspace record data and not mutating it, it should
  be evaluated for DB-first support
- if it is a control-plane or admin/config command, it remains API-backed

## Phasing Recommendation

Although the architectural scope is broad, implementation should still phase
the work.

Recommended order:

1. core runtime and profile management
2. `search`
3. `api list`
4. `api get` with relationship-heavy reads
5. `api export`
6. `api group-by`
7. deferred read-like surfaces such as `find-duplicates` only if parity is
   worth the added complexity

This keeps the public design broad while keeping the first implementation cuts
safe.

## Main Risks And Mitigations

### Endpoint metadata gap in upstream proxy support

Risk:

- `postgres-proxy get` does not currently return proxy host/port/database
  metadata

Mitigation:

- require one explicit endpoint target such as `TWENTY_DB_PROXY_URL`
- keep that target cacheable in DB profiles

### Output drift between DB and API paths

Risk:

- DB rows may not line up perfectly with current command payloads

Mitigation:

- adapt DB rows back into current service payloads
- keep renderers unchanged
- add contract tests for text and machine outputs

### Credential staleness

Risk:

- cached proxy credentials may be revoked or rotated

Mitigation:

- refresh through `postgres-proxy get` when API auth is available
- expose refresh/test/status commands
- keep a narrow internal override for support debugging

### Relationship complexity

Risk:

- include-depth and custom-object relations can drift from simple table reads

Mitigation:

- phase relationship-heavy reads after the core runtime exists
- gate unsupported relation paths behind runtime capability checks
- fall back through the shared backend only for unsupported read shapes

### Scope creep into control-plane commands

Risk:

- broad "read everything from DB" ambition pulls the project into metadata,
  workflows, and admin APIs

Mitigation:

- keep the boundary explicit: record-read platform only
- preserve the excluded command-family list in this spec as the governing scope

## Success Criteria

The design is successful when:

- supported read-only record commands use DB-first automatically when a usable
  DB profile or DB target exists
- normal commands do not expose public backend-mode flags
- `postgres-proxy` is the preferred bootstrap path rather than manual raw DB
  setup
- multiple named DB profiles per workspace are supported
- `.env` remains first-class
- mutations remain API-only
- the exhaustive included/excluded command inventory remains clear and stable
