# Dynamic Twenty Coverage and Command Cache Design

## Goal

Build the CLI toward Google Workspace CLI-style dynamic command discovery for Twenty, where the command surface is derived from the Twenty instance the CLI is pointed at. The first shipped slice is a local coverage auditor: `twenty coverage compare --upstream <path>`.

## Context

Twenty exposes multiple public surfaces:

- Core REST OpenAPI at `/rest/open-api/core`.
- Metadata REST OpenAPI at `/rest/open-api/metadata`.
- GraphQL SDL/introspection for fixed and generated resolvers.
- Raw REST, raw GraphQL, and MCP escape hatches.

Unlike Google Discovery Service, Twenty's core REST schema is workspace-dependent. Standard objects, custom objects, app-installed objects, and feature flags can change the available operations. A static CLI cannot honestly claim 100% first-class coverage for every self-hosted instance. A dynamic command cache can.

## First Slice: Coverage Auditor

`twenty coverage compare --upstream <path>` compares this CLI's current command/help surface with a local Twenty checkout. It does not contact a live Twenty server.

Inputs:

- Upstream checkout root.
- Upstream OpenAPI generator sources.
- Upstream generated GraphQL schema at `packages/twenty-client-sdk/src/metadata/generated/schema.graphql`.
- This CLI's Commander help metadata.
- This CLI source tree for GraphQL operation literals.

Output:

- Stable JSON/text/agent-renderable object.
- `status: "ok"` when all audited first-class surfaces are covered.
- `status: "missing_coverage"` when first-class gaps exist.
- Summary counts per surface.
- Flat missing items with `surface`, `name`, `upstreamPath`, and `suggestedCommand`.
- Notes that distinguish first-class gaps from raw escape-hatch coverage.

Audited surfaces:

- Core REST operation patterns:
  - `GET /{object}`
  - `POST /{object}`
  - `DELETE /{object}`
  - `PATCH /{object}`
  - `POST /batch/{object}`
  - `GET /{object}/{id}`
  - `DELETE /{object}/{id}`
  - `PATCH /{object}/{id}`
  - `POST /{object}/duplicates`
  - `PATCH /restore/{object}/{id}`
  - `PATCH /restore/{object}`
  - `PATCH /{object}/merge`
  - `GET /{object}/groupBy`
  - `POST /dashboards/{id}/duplicate`
- Metadata REST resources from `OpenApiService.generateMetaDataSchema`.
- GraphQL `Query` and `Mutation` operation names from the generated SDL.

Coverage rules:

- Core REST is covered by generic `twenty api` operations and the `dashboards duplicate` command.
- Metadata REST is covered by first-class metadata/admin command families when the operation maps directly to the upstream resource and action.
- GraphQL is first-class covered when the operation name appears in CLI source. Raw GraphQL is recorded as escape-hatch coverage, but does not hide first-class gaps.

## Dynamic Command Cache Target Architecture

The CLI should eventually boot in this order:

1. Load static core commands needed before discovery: `auth`, `coverage`, `openapi`, `raw`, and cache management.
2. Resolve active Twenty profile: base URL, workspace/profile name, auth mode, and token.
3. Load a schema cache entry keyed by:
   - normalized base URL,
   - workspace/profile,
   - schema kind (`core-openapi`, `metadata-openapi`, `graphql`),
   - Twenty server version when available,
   - object metadata collection hash when available.
4. If the cache is stale or missing, fetch schemas from the live instance.
5. Build dynamic Commander commands from cached schemas.
6. Keep raw REST/GraphQL/MCP as escape hatches.

Cache constraints:

- Do not persist token-bearing schema URLs.
- Redact examples or descriptions that include `token=` values.
- Prefer auth headers for schema fetching.
- Invalidate on metadata collection hash changes when Twenty exposes them.
- Default TTL should be short enough for self-hosted development, likely 24 hours with `--refresh` and `cache clear`.

## Breaking Change Direction

This CLI can break current command layout. The long-term preferred layout is schema-driven:

- `twenty records <object> <operation>` or dynamic `<object> <operation>` for core REST records.
- `twenty metadata <resource> <operation>` for metadata REST.
- `twenty graphql <operation>` for fixed GraphQL operations when generated command wrappers add value.
- Raw commands stay available for one-off and newly introduced operations.

The migration does not need backwards compatibility, but the auditor should keep reporting old command mappings until the dynamic command cache replaces them.

## Non-Goals For The First Slice

- No live Twenty schema fetch.
- No dynamic command generation.
- No AST parsing of TypeScript resolver decorators.
- No mutation of the upstream checkout.
- No GitHub network access.

## Success Criteria

- `twenty coverage compare --upstream /path/to/twenty -o json` runs locally.
- The command reports current first-class gaps instead of claiming false 100% coverage.
- Help JSON documents the `coverage` command and its read-only `compare` operation.
- Unit tests cover extraction, comparison, command registration, and help contracts.
