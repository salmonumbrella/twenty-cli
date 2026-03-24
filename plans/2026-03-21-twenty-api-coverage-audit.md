# Twenty API Coverage Audit

Date: 2026-03-23
Upstream reference: `twentyhq/twenty` `0ded15b3633cbf7b15d0a3ae1eecd020da54f4fc`
Local CLI reference: `feat/chatwoot-parity-and-api-coverage`

Audit maintenance note:

- `0ded15b3633cbf7b15d0a3ae1eecd020da54f4fc` was verified against the prior audited reference on 2026-03-23. The intervening changes were UI/CI fixes plus one workspace-schema enum side-effect fix, with no newly exposed public API surface to add here.

Legend:

- `[x]` covered in this CLI, either first-class or via a documented raw escape hatch
- `[blocked]` intentionally not counted as covered in the current API-key-first CLI because upstream requires user/session/browser auth or exposes no public endpoint

Important upstream nuance:

- Current metadata GraphQL is mounted at `/metadata`, not `/graphql`.
- The local CLI already has escape hatches in `raw rest` and `raw graphql`, so some items are intentionally covered without dedicated first-class wrappers.

## Discovery And Transport

- `[x]` Raw REST execution via `twenty rest`
- `[x]` Raw GraphQL execution via `twenty graphql`
- `[x]` Metadata GraphQL raw execution via `twenty graphql --endpoint metadata`
- `[x]` OpenAPI discovery endpoints `/rest/open-api/core` and `/rest/open-api/metadata` now have first-class `twenty openapi core|metadata`
- `[x]` GraphQL introspection for `/graphql` and `/metadata` now has first-class coverage via `twenty graphql schema [--endpoint graphql|metadata]`

## Agent Discovery And Ergonomics

- `[x]` Root `twenty --help` now renders a curated static help contract from `packages/twenty-sdk/src/cli/help.txt`
- `[x]` Root and command-level `--help-json` now expose machine-readable args, options, subcommands, and operation contracts without tripping required-argument validation
- `[x]` Discovery now accepts `--hj` and boolean-style `--help-json=true|1`, which is materially easier for agent callers and wrapper scripts
- `[x]` The machine-readable help contract now includes `aliases`, stable `exit_codes`, and an `output_contract` documenting query-before-format behavior plus `text|json|jsonl|agent|csv` guarantees
- `[x]` The help contract now includes `skills` and the current top-level command tree, which makes the CLI materially easier for agent callers to inspect before executing mutations
- `[x]` Stable agent-oriented rendered output now includes `jsonl` and `agent` envelopes with command-derived `kind` values

## Core Records: REST

Upstream REST surface per object:

- `GET /rest/{plural}`
- `POST /rest/{plural}`
- `GET /rest/{plural}/{id}`
- `PATCH /rest/{plural}/{id}`
- `DELETE /rest/{plural}/{id}`
- `POST /rest/batch/{plural}`
- `PATCH /rest/{plural}`
- `DELETE /rest/{plural}`
- `POST /rest/{plural}/duplicates`
- `PATCH /rest/restore/{plural}/{id}`
- `PATCH /rest/restore/{plural}`
- `PATCH /rest/{plural}/merge`
- `GET /rest/{plural}/groupBy`

Status:

- `[x]` List records: first-class and aligned with current upstream find-many query parameters
- `[x]` Get single record
- `[x]` Create record
- `[x]` Update record
- `[x]` Soft delete single record
- `[x]` Hard delete single record
- `[x]` Restore single record
- `[x]` Batch create
- `[x]` Batch update: first-class collection `PATCH /rest/{plural}` via `--filter`/`--ids` plus legacy array fan-out compatibility
- `[x]` Batch delete: first-class soft delete many via collection `DELETE /rest/{plural}?soft_delete=true`
- `[x]` Find duplicates: first-class via `--ids` or raw `{ ids|data }` payloads against `/rest/{plural}/duplicates`
- `[x]` Group by: first-class and current, including upstream `group_by` JSON-string serialization and `--field` shorthand
- `[x]` Merge
- `[x]` Restore many: first-class via `twenty api <object> restore --filter ...` or `--ids ...`
- `[x]` Dashboard duplicate: first-class via `twenty dashboards duplicate <id>`
- `[x]` Hard destroy many: first-class via `twenty api <object> destroy --filter ... --force` or `--ids ... --force`

## Core Records: GraphQL

Upstream GraphQL families per object:

- queries: `{plural}`, `{singular}`, `{singular}Duplicates`, `{plural}GroupBy`
- mutations: `create{Singular}`, `create{Plural}`, `update{Singular}`, `update{Plural}`, `delete{Singular}`, `delete{Plural}`, `destroy{Singular}`, `destroy{Plural}`, `restore{Singular}`, `restore{Plural}`, `merge{Plural}`

Status:

- `[x]` Entire core GraphQL surface is available through raw GraphQL
- `[x]` Core GraphQL-only flows like richer relationship queries or custom mutation shapes remain intentionally covered through `twenty graphql` instead of duplicated first-class wrappers

## Search

Upstream GraphQL surface:

- `search(searchInput, limit, filter, includedObjectNameSingulars, excludedObjectNameSingulars, after)` returning a connection with `edges` and `pageInfo`

Status:

- `[x]` First-class `twenty search` matches current upstream query variables for include/exclude object lists, `filter`, and `after`
- `[x]` Per-edge pagination cursors are preserved in first-class output
- `[x]` Connection `pageInfo` is now available in first-class output via `twenty search --include-page-info`

Remaining gap:

- Default output remains a flattened edge list for CLI ergonomics; use `--include-page-info` when you need structured pagination metadata

## Metadata: Objects And Fields

Upstream surface:

- REST and metadata GraphQL for objects and fields
- Queries and mutations include `object(s)`, `field(s)`, `createOneObject`, `updateOneObject`, `deleteOneObject`, `createOneField`, `updateOneField`, `deleteOneField`

Status:

- `[x]` Object list/get/create/update/delete
- `[x]` Field list/get/create/update/delete

## Metadata: Views, Layout, And UI Schema

Upstream families:

- views
- view fields
- view filters
- view sorts
- view groups
- view filter groups
- page layouts
- page layout tabs
- page layout widgets
- navigation and command menu items
- front components

Status:

- `[x]` First-class `twenty api-metadata` coverage for views, view fields, view filters, view sorts, view groups, view filter groups, page layouts, page layout tabs, and page layout widgets
- `[x]` First-class `twenty api-metadata` coverage for command menu items, front components, and navigation menu items via `/metadata` GraphQL

## Metadata: Roles, Permissions, Skills, Admin, Dashboards

Upstream families:

- roles and permissions
- dashboards and dashboard duplication
- admin panel actions
- skills / AI agents

Status:

- `[x]` First-class `twenty roles` now covers role list/get/create/update/delete plus permission-flag/object-permission/field-permission upserts
- `[x]` First-class `twenty roles list/get` can include assigned workspace members, agents, API keys, and nested permission details
- `[x]` First-class `twenty roles` now covers agent role assignment/removal when the upstream AI feature flag is enabled
- `[x]` First-class `twenty skills` now covers `skills`, `skill`, `createSkill`, `updateSkill`, `deleteSkill`, `activateSkill`, and `deactivateSkill`
- `[x]` Dashboard duplication now has first-class REST coverage via `twenty dashboards duplicate`
- `[x]` Other dashboard actions remain covered through raw GraphQL or raw REST
- `[blocked]` Admin-panel queue and metrics actions are intentionally out of scope for the API-key-first CLI because upstream guards them with `UserAuthGuard` and `AdminPanelGuard`

## API Keys

Upstream surface:

- REST and metadata GraphQL
- list, get, create, update, revoke
- assign role to API key
- `generateApiKeyToken(apiKeyId, expiresAt)`

Status:

- `[x]` First-class `twenty api-keys` matches the current upstream metadata GraphQL surface
- `[x]` First-class update operation exists
- `[x]` First-class role assignment exists
- `[x]` Token generation passes the required `expiresAt`

Concrete local drift:

- No current drift identified in the maintained first-class command surface

## Webhooks

Upstream surface:

- REST and metadata GraphQL
- list, get, create, update, delete

Status:

- `[x]` First-class `twenty webhooks` now matches the current upstream core-module GraphQL surface

Concrete local drift:

- No current drift identified in the maintained first-class command surface

## Applications And Variables

Upstream surface:

- `findManyApplications`
- `findOneApplication(id)`
- `syncApplication(manifest)`
- `createDevelopmentApplication(universalIdentifier, name)`
- `generateApplicationToken(applicationId)`
- `uninstallApplication(universalIdentifier)`
- `updateOneApplicationVariable(key, value, applicationId)`

Status:

- `[x]` First-class `twenty applications` now covers list/get/sync/uninstall/update-variable against the current `/metadata` schema, with a legacy sync fallback for older workspaces
- `[x]` First-class `twenty applications` now covers `create-development` and `generate-token`
- `[x]` First-class `twenty approved-access-domains` now covers list/delete/validate
- `[blocked]` `createApprovedAccessDomain` is intentionally out of scope for the API-key-first CLI because upstream requires `@AuthUser()`
- `[x]` First-class `twenty public-domains` now covers list/create/delete/check-records
- `[x]` First-class `twenty emailing-domains` now covers list/create/delete/verify
- `[x]` First-class `twenty postgres-proxy` now covers get/enable/disable, with password masking by default
- `[x]` API key role assignment is first-class under `twenty api-keys assign-role`
- `[blocked]` Workspace-member role assignment is intentionally out of scope for the API-key-first CLI because upstream requires `UserAuthGuard`

## Application Registrations And Marketplace

Upstream surface:

- `findManyApplicationRegistrations`
- `findOneApplicationRegistration`
- `findApplicationRegistrationStats`
- `findApplicationRegistrationVariables`
- `createApplicationRegistration`
- `updateApplicationRegistration`
- `deleteApplicationRegistration`
- `createApplicationRegistrationVariable`
- `updateApplicationRegistrationVariable`
- `deleteApplicationRegistrationVariable`
- `rotateApplicationRegistrationClientSecret`
- `transferApplicationRegistrationOwnership`
- `findManyMarketplaceApps`
- `findOneMarketplaceApp`
- `installMarketplaceApp`

Status:

- `[x]` First-class `twenty application-registrations` now covers list/get/stats/list-variables/create/update/delete/create-variable/update-variable/delete-variable/rotate-secret/transfer-ownership
- `[x]` First-class `twenty application-registrations tarball-url` now covers signed tarball discovery
- `[x]` First-class `twenty marketplace-apps` now covers list/get/install

## Logic Functions, Routes, And Workflows

Upstream surface:

- serverless function list/get/create/update/delete/execute/publish
- serverless layer creation
- available packages discovery
- source retrieval
- `logicFunctionLogs` subscription
- public route triggers under `/s/*`
- workflow webhook triggers
- workflow activate/deactivate/run/stop controls

Status:

- `[x]` First-class `twenty serverless` now prefers current `ServerlessFunction*` list/get/create/update/delete/execute/publish/packages/source operations and falls back to legacy `LogicFunction*` schemas on older workspaces
- `[x]` First-class `twenty serverless create-layer` now covers current upstream `createOneServerlessFunctionLayer`
- `[x]` First-class `twenty serverless logs` now covers the `logicFunctionLogs` subscription over metadata SSE
- `[x]` First-class `twenty routes invoke` now covers public `GET|POST|PUT|PATCH|DELETE /s/*` route execution with query params, JSON bodies, and custom request headers
- `[x]` First-class `twenty route-triggers` now covers route-trigger definition list/get/create/update/delete
- `[x]` First-class `twenty workflows invoke-webhook` now covers the public `GET|POST /webhooks/workflows/:workspaceId/:workflowId` trigger path
- `[x]` First-class `twenty workflows activate|deactivate|run|stop-run` now covers the private workflow control mutations when authenticated with a user bearer token that satisfies upstream workspace/user/workflows guards
- `[x]` Public route and workflow invocation helpers remain first-class where upstream exposes them

Upstream note:

- Current upstream exposes `createOneServerlessFunctionLayer` as the dedicated layer lifecycle mutation; broader layer listing or reuse would require new upstream surface area rather than a missing CLI wrapper

## Files

Upstream surface:

- REST downloads: `GET /file/:fileFolder/:id`
- REST public assets: `GET /public-assets/:workspaceId/:applicationId/*path`
- GraphQL uploads:
  - `uploadAIChatFile`
  - `uploadWorkflowFile`
  - `uploadWorkspaceLogo`
  - `uploadWorkspaceMemberProfilePicture`
  - `uploadFilesFieldFile`
  - `uploadFilesFieldFileByUniversalIdentifier`

Status:

- `[x]` First-class `twenty files` now covers verified domain-specific GraphQL uploads plus signed/public download routes
- `[x]` First-class `twenty files upload --target app-tarball` now covers `uploadAppTarball`
- `[x]` First-class `twenty files upload --target application-file` now covers current application-development file uploads, including source, built assets, and public assets
- `[blocked]` Generic file delete is intentionally out of scope because upstream does not expose a public generic delete surface

## Event Streams And Subscriptions

Upstream surface:

- `eventLogs`
- `logicFunctionLogs`
- `onEventSubscription(eventStreamId)`
- `addQueryToEventStream`
- `removeQueryFromEventStream`

Status:

- `[x]` First-class `twenty event-logs list` now covers the enterprise `eventLogs` query
- `[x]` First-class `twenty serverless logs` now covers `logicFunctionLogs`
- `[blocked]` `onEventSubscription`, `addQueryToEventStream`, and `removeQueryFromEventStream` are intentionally out of scope for the API-key-first CLI because upstream guards them with `UserAuthGuard`

## Auth, Workspaces, Invitations, SSO, And Two-Factor Auth

Upstream surface:

- sign-in and sign-up flows
- token renewal
- password reset
- workspace activate/update/delete
- workspace invitations
- public workspace lookup and custom-domain validation
- SSO authorization URL and IdP management
- two-factor auth enrollment and verification

Status:

- `[x]` Local `twenty auth` manages saved workspace config and now covers first-class current-workspace/public-discovery queries plus public token renewal and SSO authorization URL discovery
- `[blocked]` Session-oriented auth/workspace APIs remain intentionally out of scope for the API-key-first CLI

## OAuth And Application Registration

Upstream surface:

- `authorizeApp`
- `renewToken`
- Google connection entrypoints
- Microsoft connection entrypoints
- OAuth code-exchange logic exists in source but is currently commented out upstream

Status:

- `[x]` First-class CLI now covers `renewToken` and public `getAuthorizationUrlForSSO`
- `[blocked]` `authorizeApp` is intentionally out of scope for the API-key-first CLI because upstream guards it with `UserAuthGuard`
- `[blocked]` Google and Microsoft connection entrypoints remain browser-driven and outside the current API-key-first scope

## Connected Accounts And Channel Integrations

Upstream surface:

- connected account list/remove
- channel sync
- message channel settings
- calendar channel settings
- Google and Microsoft connection flows

Status:

- `[x]` First-class `twenty connected-accounts` now covers connected-account list/get plus explicit `startChannelSync`
- `[x]` First-class `twenty connected-accounts` now covers current upstream `getConnectedImapSmtpCaldavAccount` and `saveImapSmtpCaldavAccount`
- `[x]` First-class connected-account output masks access tokens, refresh tokens, and connection parameters by default
- `[x]` First-class manual IMAP/SMTP/CALDAV reads now mask nested protocol passwords by default
- `[x]` First-class `twenty message-channels` now covers message-channel list/get/update
- `[x]` First-class `twenty calendar-channels` now covers calendar-channel list/get/update
- `[x]` Channel command output now hides `syncCursor` by default
- `[blocked]` Google and Microsoft connection flows remain browser-driven and outside the current API-key-first scope

## Remaining Blocked Areas

The remaining uncovered areas are blocked by upstream auth/transport constraints rather than missing API-key-safe CLI work:

- session-oriented auth, invitations, workspace administration, SSO management, and 2FA flows
- `authorizeApp` and browser-driven Google/Microsoft provider connection flows
- event-stream query subscriptions guarded by `UserAuthGuard`
- admin-panel queue and metrics surfaces guarded by `UserAuthGuard` and `AdminPanelGuard`
- generic file delete, because upstream does not expose a public endpoint

## Primary Evidence

- API overview: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-docs/developers/extend/api.mdx#L22-L49
- OpenAPI discovery endpoints: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/open-api/open-api.controller.ts#L13-L32
- Core REST routes: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/api/rest/core/controllers/rest-api-core.controller.ts#L24-L162
- Core GraphQL operation naming: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/utils/get-resolver-name.util.ts#L12-L50
- Search resolver: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/search/search.resolver.ts#L37-L116
- Metadata schema: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-sdk/src/clients/generated/metadata/schema.graphql#L3078-L3474
- Metadata subscriptions: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-sdk/src/clients/generated/metadata/schema.graphql#L4513-L4515
- Views REST controller: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/metadata-modules/view/controllers/view.controller.ts#L44-L171
- View field REST controller: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/metadata-modules/view-field/controllers/view-field.controller.ts#L32-L109
- View filter REST controller: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/metadata-modules/view-filter/controllers/view-filter.controller.ts#L30-L112
- View filter group REST controller: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/metadata-modules/view-filter-group/controllers/view-filter-group.controller.ts#L31-L111
- View group REST controller: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/metadata-modules/view-group/controllers/view-group.controller.ts#L30-L112
- View sort REST controller: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/metadata-modules/view-sort/controllers/view-sort.controller.ts#L29-L106
- Page layouts REST controller: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/metadata-modules/page-layout/controllers/page-layout.controller.ts#L29-L105
- Page layout tab REST controller: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/metadata-modules/page-layout-tab/controllers/page-layout-tab.controller.ts#L28-L99
- Page layout widget REST controller: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/metadata-modules/page-layout-widget/controllers/page-layout-widget.controller.ts#L29-L104
- API keys resolver and controller: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/api-key/api-key.resolver.ts#L28-L122 and https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/api-key/controllers/api-key.controller.ts#L31-L99
- Webhooks resolver and controller: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/webhook/webhook.resolver.ts#L20-L88 and https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/webhook/controllers/webhook.controller.ts#L1-L97
- Application variable resolver: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/application/application-variable/application-variable.resolver.ts#L1-L44
- Current application install, development, and manifest resolvers: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/application/application-install/application-install.resolver.ts#L1-L70 and https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/application/application-development/application-development.resolver.ts#L1-L181 and https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/application/application-manifest/application-manifest.resolver.ts#L1-L81
- App auth and token renewal: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/auth/auth.resolver.ts#L777-L801
- Connected account object and channel sync: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/modules/connected-account/standard-objects/connected-account.workspace-entity.ts#L10-L24 and https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/modules/connected-account/channel-sync/channel-sync.resolver.ts#L16-L31
- Logic-function rename and resolver: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/database/typeorm/core/migrations/common/1769556947746-renameServerless.ts#L6-L28 and https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/metadata-modules/logic-function/logic-function.resolver.ts#L45-L260
- Event logs and event stream guards: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/event-logs/event-logs.resolver.ts#L1-L48 and https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/subscriptions/event-stream.resolver.ts#L1-L186
- File controller and files-field uploads: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/file/controllers/file.controller.ts#L34-L126 and https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/file/files-field/resolvers/files-field.resolver.ts#L23-L75
- Skill resolver and DTOs: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/metadata-modules/skill/skill.resolver.ts#L18-L69 and https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/metadata-modules/skill/dtos/skill.dto.ts#L9-L53
- Workflow webhook controller: https://github.com/twentyhq/twenty/blob/93de331428b59bb1623c487b391a9339b3f0e079/packages/twenty-server/src/engine/core-modules/workflow/controllers/workflow-trigger.controller.ts#L36-L160
