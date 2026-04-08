# Twenty CLI

CLI for Twenty CRM. Manage records, metadata, and API access from your terminal.

<!-- GENERATED:INSTALL_AND_AGENT_CONTRACT:START -->

## Installation

Tagged releases publish standalone `twenty` archives for macOS and Linux, update the maintained Homebrew formula, and can publish the scoped npm package `@salmonumbrella/twenty-cli` when `NPM_TOKEN` is configured.

```bash
# Build from source
pnpm install
pnpm build
node packages/twenty-sdk/dist/cli/cli.js --help
```

## Agent Discovery

The CLI ships with a curated root help contract plus machine-readable help output for agents and automation.

```bash
twenty --help
twenty --help-json
twenty --hj
twenty roles --help-json
twenty routes invoke --hj
twenty auth list --help-json
```

- Prefer twenty CMD --help-json before executing mutations
- Stable JSON fields: path, args, options, operations, capabilities, exit_codes, output_contract

### Environment Loading

.env then .env.local then the explicit env file; existing environment variables win.

### Output Guarantees

- --query runs before output formatting
- json renders pretty-printed 2-space JSON
- jsonl renders one compact JSON record per line
- agent renders stable envelopes for arrays, objects, and scalar data
- csv wraps singleton values and JSON-encodes nested objects/arrays
- text renders best-effort tables

### Exit Codes

```text
0  Success, help output, or version output
1  General or unexpected error
2  Invalid arguments or command usage error
3  Authentication or permission error
4  Network error or request failed before a response
5  Rate limited (429)
```

<!-- GENERATED:INSTALL_AND_AGENT_CONTRACT:END -->

### OpenAPI

Fetch the published REST OpenAPI schemas without falling back to raw REST.

```bash
twenty openapi core
twenty openapi metadata --output-file metadata-openapi.json
```

This wraps Twenty's public `/rest/open-api/core` and `/rest/open-api/metadata` endpoints directly.

### MCP

Discover and execute logical Twenty tools through the unified MCP endpoint.
Examples:

- twenty mcp status
- twenty mcp catalog -o json
- twenty mcp schema find_companies
- twenty mcp exec find_companies --data '{"query":"Acme"}'
- twenty mcp skills workflow-building
- twenty mcp search "MCP setup"

`twenty mcp schema` wraps `learn_tools`, and `twenty mcp exec` wraps `execute_tool`.

## Commands

### Authentication

```bash
twenty auth login --token TOKEN --base-url URL    # Configure credentials
twenty auth status                                 # Check authentication status
twenty auth status --show-token                    # Show full token value
twenty auth workspace                              # Query current workspace from the API
twenty auth discover <origin>                      # Inspect public workspace auth providers
twenty auth renew-token --app-token TOKEN          # Exchange an app refresh token
twenty auth sso-url <identity-provider-id>         # Get SSO authorization URL
twenty auth list                                   # List configured workspaces
twenty auth switch <workspace>                     # Set default workspace
twenty auth logout                                 # Remove current workspace
twenty auth logout --workspace <name>              # Remove specific workspace
twenty auth logout --all                           # Remove all workspaces
twenty approved-access-domains list                # List approved access domains
```

`twenty auth renew-token` maps the public `renewToken` GraphQL mutation.
`twenty auth sso-url` maps the public `getAuthorizationUrlForSSO` GraphQL mutation.

### Approved Access Domains

Manage workspace approved access domains and validation flow.

```bash
twenty approved-access-domains <operation> [id] [options]
```

**Operations:**

| Operation  | Description                      | Example                                                                 |
| ---------- | -------------------------------- | ----------------------------------------------------------------------- |
| `list`     | List approved access domains     | `twenty approved-access-domains list`                                   |
| `delete`   | Delete an approved access domain | `twenty approved-access-domains delete <id>`                            |
| `validate` | Validate a domain with token     | `twenty approved-access-domains validate <id> --validation-token TOKEN` |

**Options:**

```bash
--validation-token <token>         # Validation token for validate
```

**Examples:**

```bash
twenty approved-access-domains list
twenty approved-access-domains delete <approved-access-domain-id>
twenty approved-access-domains validate <approved-access-domain-id> --validation-token <validation-token>
```

`createApprovedAccessDomain` remains upstream user-session-oriented because it requires `@AuthUser()`, so it is not exposed as a first-class API-key command here.

### Public Domains

Manage public domains used for workspace exposure and DNS validation.

```bash
twenty public-domains <operation> [options]
```

**Operations:**

| Operation       | Description                        | Example                                                        |
| --------------- | ---------------------------------- | -------------------------------------------------------------- |
| `list`          | List configured public domains     | `twenty public-domains list`                                   |
| `create`        | Create a public domain             | `twenty public-domains create --domain app.example.com`        |
| `delete`        | Delete a public domain             | `twenty public-domains delete --domain app.example.com`        |
| `check-records` | Refresh and inspect DNS validation | `twenty public-domains check-records --domain app.example.com` |

**Options:**

```bash
--domain <domain>                    # Public domain name
```

**Examples:**

```bash
twenty public-domains list
twenty public-domains create --domain app.example.com
twenty public-domains check-records --domain app.example.com -o json
twenty public-domains delete --domain app.example.com
```

This wraps the current upstream GraphQL admin surface for `findManyPublicDomains`, `createPublicDomain`, `deletePublicDomain`, and `checkPublicDomainValidRecords`.

### Emailing Domains

Manage transactional email sender domains and verification state.

```bash
twenty emailing-domains <operation> [id] [options]
```

**Operations:**

| Operation | Description                     | Example                                                    |
| --------- | ------------------------------- | ---------------------------------------------------------- |
| `list`    | List emailing domains           | `twenty emailing-domains list`                             |
| `create`  | Create an emailing domain       | `twenty emailing-domains create --domain mail.example.com` |
| `verify`  | Verify an emailing domain by ID | `twenty emailing-domains verify <id>`                      |
| `delete`  | Delete an emailing domain by ID | `twenty emailing-domains delete <id>`                      |

**Options:**

```bash
--domain <domain>                    # Emailing domain name for create
--driver <driver>                    # Emailing driver (default: AWS_SES)
```

**Examples:**

```bash
twenty emailing-domains list
twenty emailing-domains create --domain mail.example.com
twenty emailing-domains create --domain mail.example.com --driver AWS_SES -o json
twenty emailing-domains verify <emailing-domain-id>
twenty emailing-domains delete <emailing-domain-id>
```

The current upstream enum only exposes `AWS_SES`, so the CLI defaults `--driver` to that value while still validating it explicitly.

### Postgres Proxy

Manage the workspace Postgres proxy credential lifecycle.

```bash
twenty postgres-proxy <operation> [options]
```

**Operations:**

| Operation | Description                           | Example                                        |
| --------- | ------------------------------------- | ---------------------------------------------- |
| `get`     | Get current proxy credentials         | `twenty postgres-proxy get`                    |
| `enable`  | Enable proxy access and mint creds    | `twenty postgres-proxy enable --show-password` |
| `disable` | Disable proxy access and revoke creds | `twenty postgres-proxy disable`                |

**Options:**

```bash
--show-password                    # Show the returned Postgres proxy password
```

**Examples:**

```bash
twenty postgres-proxy get
twenty postgres-proxy get --show-password -o json
twenty postgres-proxy enable --show-password
twenty postgres-proxy disable
```

Upstream returns plaintext credentials for all three operations, so the CLI masks `password` by default unless `--show-password` is set explicitly.

### Roles

Manage workspace role definitions.

```bash
twenty roles <operation> [id] [options]
```

**Operations:**

| Operation                   | Description                     | Example                                               |
| --------------------------- | ------------------------------- | ----------------------------------------------------- |
| `list`                      | List roles                      | `twenty roles list`                                   |
| `get`                       | Get one role by ID              | `twenty roles get <id> --include-targets`             |
| `create`                    | Create a role                   | `twenty roles create -d '{"label":"Support"}'`        |
| `update`                    | Update a role                   | `twenty roles update <id> --set description=...`      |
| `delete`                    | Delete a role                   | `twenty roles delete <id>`                            |
| `upsert-permission-flags`   | Replace role setting flags      | `twenty roles upsert-permission-flags -d '{...}'`     |
| `upsert-object-permissions` | Replace role object permissions | `twenty roles upsert-object-permissions -d '{...}'`   |
| `upsert-field-permissions`  | Replace role field permissions  | `twenty roles upsert-field-permissions -d '{...}'`    |
| `assign-agent`              | Assign a role to an agent       | `twenty roles assign-agent <agent-id> --role-id <id>` |
| `remove-agent`              | Remove an agent role assignment | `twenty roles remove-agent <agent-id>`                |

**Options:**

```bash
-d, --data <json>                  # Inline JSON payload for create/update
-f, --file <path>                  # JSON file payload for create/update
--set <key=value>                  # Set one field value for create/update
--role-id <id>                     # Role ID for assign-agent
--include-targets                  # Include assigned workspace members, agents, and API keys on list/get
--include-permissions              # Include nested permission flags and object/field permissions on list/get
```

**Examples:**

```bash
twenty roles list
twenty roles get <role-id> --include-targets --include-permissions
twenty roles create -d '{"label":"Support","description":"Support role"}'
twenty roles update <role-id> --set description="Updated support role"
twenty roles delete <role-id>
twenty roles upsert-permission-flags -d '{"roleId":"<role-id>","permissionFlagKeys":["WORKSPACE"]}'
twenty roles upsert-object-permissions -d '{"roleId":"<role-id>","objectPermissions":[{"objectMetadataId":"<object-metadata-id>","canReadObjectRecords":true}]}'
twenty roles upsert-field-permissions -d '{"roleId":"<role-id>","fieldPermissions":[{"objectMetadataId":"<object-metadata-id>","fieldMetadataId":"<field-metadata-id>","canReadFieldValue":true}]}'
twenty roles assign-agent <agent-id> --role-id <role-id>
twenty roles remove-agent <agent-id>
```

`assign-agent` and `remove-agent` depend on Twenty's upstream AI feature flag being enabled for the workspace.

### Dashboards

Duplicate an existing dashboard through Twenty's REST dashboard controller.

```bash
twenty dashboards duplicate <dashboard-id> [options]
```

**Examples:**

```bash
twenty dashboards duplicate <dashboard-id>
twenty dashboards duplicate <dashboard-id> -o json
```

This wraps `POST /rest/dashboards/:id/duplicate` and returns the duplicated dashboard payload, including its new `id`, `pageLayoutId`, and `position`.

### Skills

Manage workspace AI skills.

```bash
twenty skills <operation> [id] [options]
```

**Operations:**

| Operation    | Description           | Example                                                  |
| ------------ | --------------------- | -------------------------------------------------------- |
| `list`       | List skills           | `twenty skills list`                                     |
| `get`        | Get one skill by ID   | `twenty skills get <id>`                                 |
| `create`     | Create a custom skill | `twenty skills create -d '{"name":"workflow-building"}'` |
| `update`     | Update a custom skill | `twenty skills update <id> -d '{"label":"Workflow AI"}'` |
| `delete`     | Delete a custom skill | `twenty skills delete <id>`                              |
| `activate`   | Activate a skill      | `twenty skills activate <id>`                            |
| `deactivate` | Deactivate a skill    | `twenty skills deactivate <id>`                          |

**Options:**

```bash
-d, --data <json>                  # Inline JSON payload for create/update
-f, --file <path>                  # JSON file payload for create/update
--set <key=value>                  # Set one field value for create/update
```

**Examples:**

```bash
twenty skills list
twenty skills get <skill-id> -o json
twenty skills create -d '{"name":"workflow-building","label":"Workflow Building","content":"# Steps"}'
twenty skills update <skill-id> --set label="Workflow Design"
twenty skills activate <skill-id>
twenty skills deactivate <skill-id>
twenty skills delete <skill-id>
```

The upstream skills resolver is guarded by workspace auth plus the AI settings permission, so your API key or workspace member must be allowed to manage AI features.

### Workflows

Invoke public workflow webhooks and control workflow runs.

```bash
twenty workflows <command> [options]
```

**Commands:**

| Command          | Description                               | Example                                                                       |
| ---------------- | ----------------------------------------- | ----------------------------------------------------------------------------- |
| `invoke-webhook` | Invoke a public workflow webhook endpoint | `twenty workflows invoke-webhook <workflow-id> --workspace-id <workspace-id>` |
| `activate`       | Activate a workflow version               | `twenty workflows activate <workflow-version-id>`                             |
| `deactivate`     | Deactivate a workflow version             | `twenty workflows deactivate <workflow-version-id>`                           |
| `run`            | Run a workflow version                    | `twenty workflows run <workflow-version-id> -d '{"source":"cli"}'`            |
| `stop-run`       | Stop an existing workflow run             | `twenty workflows stop-run <workflow-run-id>`                                 |

**Webhook Options:**

```bash
--workspace-id <id>               # Explicit workspace ID for the public webhook path
--method <get|post>               # Invoke with GET or POST (default: post)
-d, --data <json>                 # JSON body for POST requests
-f, --file <path>                 # JSON file body for POST requests
--param <key=value>               # Query parameter (repeatable)
```

**Run Options:**

```bash
--workflow-run-id <id>            # Continue an existing workflow run
-d, --data <json>                 # JSON payload for run input
-f, --file <path>                 # JSON payload file for run input
```

**Examples:**

```bash
twenty workflows invoke-webhook <workflow-id> --workspace-id <workspace-id>
twenty workflows invoke-webhook <workflow-id> --workspace-id <workspace-id> --method get --param source=cli
twenty workflows invoke-webhook <workflow-id> -d '{"contactId":"<person-id>"}'
twenty workflows activate <workflow-version-id>
twenty workflows run <workflow-version-id> -d '{"source":"cli"}'
twenty workflows run <workflow-version-id> --workflow-run-id <workflow-run-id> -f ./payload.json
twenty workflows stop-run <workflow-run-id>
```

When `--workspace-id` is omitted, the CLI uses the authenticated `currentWorkspace` query to discover it automatically. If you want to use the public webhook path without auth at all, pass `--workspace-id` explicitly.

`activate`, `deactivate`, `run`, and `stop-run` call the private workflow control GraphQL mutations. Upstream guards these with workspace auth, user auth, and the workflows settings permission, so these commands require a user-authenticated bearer token rather than a workspace API key.

### Routes

Invoke public `/s/*` route trigger endpoints.

```bash
twenty routes invoke <route-path> [options]
```

**Options:**

```bash
--method <get|post|put|patch|delete>  # HTTP method (default: get)
-d, --data <json>                     # JSON body for non-GET requests
-f, --file <path>                     # JSON file body for non-GET requests
--param <key=value>                   # Query parameter (repeatable)
--header <key=value>                  # Request header (repeatable)
```

**Examples:**

```bash
twenty routes invoke public/ping
twenty routes invoke /s/hooks/import --method post -d '{"batch":"one"}'
twenty routes invoke contacts/sync --param dryRun=true --header x-source=cli
```

The CLI normalizes relative paths to `/s/*`, so `public/ping` and `/s/public/ping` both resolve correctly. If the underlying route trigger requires auth, the configured API token is sent automatically as a Bearer token; otherwise the command also works against fully public routes with no token configured.

### Records API

```bash
twenty api <object> <operation> [args]
```

**Operations:**

| Operation         | Description       | Example                                                                            |
| ----------------- | ----------------- | ---------------------------------------------------------------------------------- |
| `list`            | List records      | `twenty api people list --limit 50`                                                |
| `get`             | Get single record | `twenty api people get <id>`                                                       |
| `create`          | Create record     | `twenty api people create -d '{"name":{"firstName":"John"}}'`                      |
| `update`          | Update record     | `twenty api people update <id> --set jobTitle="Engineer"`                          |
| `delete`          | Soft delete       | `twenty api people delete <id> --force`                                            |
| `destroy`         | Hard delete       | `twenty api people destroy <id> --force`                                           |
| `restore`         | Restore deleted   | `twenty api people restore <id>`                                                   |
| `batch-create`    | Create multiple   | `twenty api people batch-create -f records.json`                                   |
| `batch-update`    | Update multiple   | `twenty api people batch-update --filter 'status[eq]:TODO' -d '{"status":"DONE"}'` |
| `batch-delete`    | Delete multiple   | `twenty api people batch-delete --ids id1,id2 --force`                             |
| `import`          | Import from file  | `twenty api people import data.csv`                                                |
| `export`          | Export to file    | `twenty api people export --format json --output-file out.json`                    |
| `group-by`        | Group records     | `twenty api opportunities group-by --field stage`                                  |
| `find-duplicates` | Find duplicates   | `twenty api people find-duplicates -d '{"ids":["person_123"]}'`                    |
| `merge`           | Merge records     | `twenty api people merge --source <id> --target <id>`                              |

`batch-update` now supports both current upstream collection updates and the older per-record array mode. Use `--filter` or `--ids` with a shared object payload or `--set` for a true collection `PATCH /rest/{object}` update; keep passing an array or CSV file when you want the legacy per-record fan-out behavior.

`restore` and `destroy` also support collection-level flows when no single record ID is provided. Use `--filter` or `--ids` with `twenty api <object> restore` or `twenty api <object> destroy --force` to hit the current upstream many-record restore and destroy routes.

`find-duplicates` now follows the current upstream REST contract as well: use `--ids` for record-based duplicate lookup or `--data` / `--file` with a raw payload like `{"data":[{...}]}`. `group-by` now serializes the upstream `group_by` query format correctly, and the `--field` shorthand expands to the required JSON-array grouping contract.

`list` and `export` no longer advertise or accept `--fields`. Current upstream REST find-many derives field selection from `depth`, not a `fields` query parameter, so the CLI now fails fast instead of silently sending a no-op.

**Common Options:**

```bash
--limit <n>          # Limit results
--all                # Fetch all records (paginated)
--filter <expr>      # Filter expression
--include <rels>     # Include relations
--sort <field>       # Sort field
--order <asc|desc>   # Sort direction
-d, --data <json>    # JSON payload
-f, --file <path>    # JSON/CSV file (use - for stdin)
--set <key=value>    # Set field value (repeatable)
--force              # Skip confirmation
```

### Metadata API

```bash
twenty api-metadata <type> <operation> [args]
```

Supported types:
`objects`, `fields`, `command-menu-items`, `front-components`, `navigation-menu-items`, `views`, `view-fields`, `view-filters`, `view-filter-groups`, `view-groups`, `view-sorts`, `page-layouts`, `page-layout-tabs`, `page-layout-widgets`

**Objects:**

```bash
twenty api-metadata objects list                              # List all object types
twenty api-metadata objects get person                        # Get object with fields
twenty api-metadata objects create -d '{"nameSingular":"widget","namePlural":"widgets"}'
twenty api-metadata objects update <id> -d '{"labelSingular":"Widget"}'
twenty api-metadata objects delete <id>                       # Delete custom object
```

**Fields:**

```bash
twenty api-metadata fields list                               # List all fields
twenty api-metadata fields list --object person               # Fields for object
twenty api-metadata fields get <field-id>                     # Get field details
twenty api-metadata fields create -d '{"objectMetadataId":"...","name":"rating","type":"NUMBER"}'
```

**Views and layout metadata:**

```bash
twenty api-metadata views list --object person
twenty api-metadata view-fields list --view <view-id>
twenty api-metadata view-filters create -d '{"viewId":"...","fieldMetadataId":"...","operand":"ACTIVE"}'
twenty api-metadata view-filter-groups update <id> -d '{"logicalOperator":"AND"}'
twenty api-metadata view-groups delete <id>
twenty api-metadata view-sorts list --view <view-id>
twenty api-metadata page-layouts list --object person --page-layout-type RECORD_PAGE
twenty api-metadata page-layout-tabs list --page-layout <layout-id>
twenty api-metadata page-layout-widgets list --page-layout-tab <tab-id>
```

`--page-layout-type` requires `--object`, because the upstream endpoint only applies that filter when an object metadata ID is present.

**UI extension metadata:**

```bash
twenty api-metadata command-menu-items list
twenty api-metadata command-menu-items update <id> -d '{"label":"Open widget"}'
twenty api-metadata front-components list
twenty api-metadata front-components update <id> -d '{"name":"Widget panel"}'
twenty api-metadata navigation-menu-items list
twenty api-metadata navigation-menu-items update <id> -d '{"name":"Accounts"}'
```

`front-components update` and `navigation-menu-items update` accept the inner update payload directly; the CLI wraps it into the upstream GraphQL `update` input shape for you.

### Raw API Access

**REST:**

```bash
twenty rest get /rest/people
twenty rest post /rest/people -d '{"name":{"firstName":"Ada","lastName":"Lovelace"}}'
twenty rest patch /rest/people/<id> -d '{"jobTitle":"Engineer"}'
twenty rest delete /rest/people/<id>
```

**GraphQL:**

```bash
twenty graphql query --query 'query { people(first: 5) { edges { node { id } } } }'
twenty graphql mutate --query 'mutation { createPerson(data:{...}) { id } }'
twenty graphql schema                                         # Fetch introspection schema
twenty graphql schema --output-file schema.json               # Save to file
```

### Search

Full-text search across records.

```bash
twenty search <query> [options]
```

**Options:**

| Option                | Description                                                  |
| --------------------- | ------------------------------------------------------------ |
| `--limit <n>`         | Maximum results (default: 20)                                |
| `--objects <list>`    | Comma-separated object names to include (singular or plural) |
| `--exclude <list>`    | Comma-separated object names to exclude (singular or plural) |
| `--cursor <cursor>`   | Pagination cursor for the next page                          |
| `--include-page-info` | Include top-level `pageInfo` in output                       |
| `--filter <json>`     | JSON filter object                                           |
| `--filter-file`       | Load JSON filter from a file or stdin                        |

**Examples:**

```bash
twenty search "John Doe"                                      # Search all objects
twenty search "acme" --objects companies,people               # Search specific objects
twenty search "engineer" --limit 50 --exclude notes           # Exclude objects
twenty search "john" --cursor eyJvZmZzZXQiOjIwfQ==            # Fetch the next page
twenty search "john" --include-page-info -o json             # Return data plus pageInfo
twenty search "john" --filter '{"city":{"eq":"Vancouver"}}'   # Apply metadata filter JSON
twenty search "john" --filter-file ./search-filter.json       # Read filter JSON from a file
```

### Webhooks

Manage webhook endpoints.

```bash
twenty webhooks <operation> [id] [options]
```

**Operations:**

| Operation | Description         | Example                                                   |
| --------- | ------------------- | --------------------------------------------------------- |
| `list`    | List webhooks       | `twenty webhooks list`                                    |
| `get`     | Get webhook details | `twenty webhooks get <id>`                                |
| `create`  | Create webhook      | `twenty webhooks create -d '{"targetUrl":"https://..."}'` |
| `update`  | Update webhook      | `twenty webhooks update <id> -d '{"targetUrl":"..."}'`    |
| `delete`  | Delete webhook      | `twenty webhooks delete <id>`                             |

**Options:**

```bash
-d, --data <json>     # JSON payload
-f, --file <path>     # JSON file
--set <key=value>     # Set a field value (repeatable)
```

**Examples:**

```bash
twenty webhooks create --set targetUrl=https://example.com/webhook --set operations=create,update
twenty webhooks update <webhook-id> --set description="Updated webhook"
```

### Route Triggers

Manage saved route-trigger definitions for serverless functions.

```bash
twenty route-triggers <operation> [id] [options]
```

**Operations:**

| Operation | Description                 | Example                                                   |
| --------- | --------------------------- | --------------------------------------------------------- |
| `list`    | List route triggers         | `twenty route-triggers list`                              |
| `get`     | Get one route trigger by ID | `twenty route-triggers get <id>`                          |
| `create`  | Create a route trigger      | `twenty route-triggers create -d '{"path":"/hello",...}'` |
| `update`  | Update a route trigger      | `twenty route-triggers update <id> -d '{"path":"/v2"}'`   |
| `delete`  | Delete a route trigger      | `twenty route-triggers delete <id>`                       |

**Options:**

```bash
-d, --data <json>     # JSON payload
-f, --file <path>     # JSON file
--set <key=value>     # Set a field value (repeatable)
```

**Examples:**

```bash
twenty route-triggers list
twenty route-triggers get <route-trigger-id> -o json
twenty route-triggers create -d '{"path":"/hello","isAuthRequired":true,"httpMethod":"GET","serverlessFunctionId":"<serverless-function-id>"}'
twenty route-triggers update <route-trigger-id> -d '{"path":"/hello-v2","isAuthRequired":false,"httpMethod":"POST"}'
twenty route-triggers delete <route-trigger-id>
```

`route-triggers` manages the metadata definitions. Invoking a public route still happens through the runtime `/s/*` endpoint.

### API Keys

Manage API keys.

```bash
twenty api-keys <operation> [id] [options]
```

**Operations:**

| Operation | Description         | Example                                  |
| --------- | ------------------- | ---------------------------------------- |
| `list`    | List API keys       | `twenty api-keys list`                   |
| `get`     | Get API key details | `twenty api-keys get <id>`               |
| `create`  | Create API key      | `twenty api-keys create --name "CI Key"` |
| `revoke`  | Revoke API key      | `twenty api-keys revoke <id>`            |

**Options:**

```bash
--name <name>        # API key name (required for create)
--expires-at <date>  # Expiration date (ISO format)
```

**Examples:**

```bash
twenty api-keys create --name "Production" --expires-at 2025-12-31
twenty api-keys list -o json
twenty api-keys revoke <api-key-id>
```

### Applications

Manage workspace applications and application variables.

```bash
twenty applications <operation> [target] [options]
```

**Operations:**

| Operation            | Description                                       | Example                                                                       |
| -------------------- | ------------------------------------------------- | ----------------------------------------------------------------------------- |
| `list`               | List installed applications                       | `twenty applications list`                                                    |
| `get`                | Get one application by ID                         | `twenty applications get <id>`                                                |
| `sync`               | Install or update an application manifest         | `twenty applications sync --manifest-file app.json`                           |
| `create-development` | Create or resolve a local development application | `twenty applications create-development com.example.app --name "Example App"` |
| `generate-token`     | Generate an application access/refresh token pair | `twenty applications generate-token <id>`                                     |
| `uninstall`          | Uninstall by universal identifier                 | `twenty applications uninstall com.example.my-app`                            |
| `update-variable`    | Update one application variable by app ID         | `twenty applications update-variable <id> --key K ...`                        |

**Options:**

```bash
--manifest <json>            # Inline manifest JSON for sync
--manifest-file <path>       # Manifest JSON file for sync
--package-json <json>        # Legacy package.json JSON for older syncApplication schemas
--package-json-file <path>   # Legacy package.json file for older syncApplication schemas
--yarn-lock-file <path>      # Legacy yarn.lock file for older syncApplication schemas
--name <name>                # Display name for create-development
--key <key>                  # Variable key for update-variable
--value <value>              # Variable value for update-variable
```

**Examples:**

```bash
twenty applications list
twenty applications get <application-id> -o json
twenty applications sync --manifest-file ./manifest.json
twenty applications create-development com.acme.calendar-sync --name "Calendar Sync"
twenty applications generate-token <application-id>
twenty applications uninstall com.acme.calendar-sync
twenty applications update-variable <application-id> --key API_TOKEN --value "$TOKEN"
```

`sync` now follows current upstream Twenty behavior on `/metadata` and only requires the application manifest. The legacy `--package-json*` and `--yarn-lock-file` flags are kept only as a compatibility fallback for older workspaces that still expose the pre-migration sync mutation.

### Application Registrations

Manage distributable application registrations, their variables, and ownership.

```bash
twenty application-registrations <operation> [target] [options]
```

**Operations:**

| Operation            | Description                                          | Example                                                                                       |
| -------------------- | ---------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| `list`               | List application registrations                       | `twenty application-registrations list`                                                       |
| `get`                | Get one application registration by ID               | `twenty application-registrations get <id>`                                                   |
| `stats`              | Get install stats for one registration               | `twenty application-registrations stats <id>`                                                 |
| `tarball-url`        | Get a signed tarball download URL                    | `twenty application-registrations tarball-url <id>`                                           |
| `list-variables`     | List variables for one registration                  | `twenty application-registrations list-variables <id>`                                        |
| `create`             | Create a registration from JSON input                | `twenty application-registrations create -d '{"name":"Widget"}'`                              |
| `update`             | Update a registration with JSON input                | `twenty application-registrations update <id> -d '{"websiteUrl":"https://..."}'`              |
| `delete`             | Delete a registration by ID                          | `twenty application-registrations delete <id>`                                                |
| `create-variable`    | Create a registration variable                       | `twenty application-registrations create-variable -d '{"applicationRegistrationId":"..."}'`   |
| `update-variable`    | Update one registration variable by ID               | `twenty application-registrations update-variable <id> -d '{"description":"..."}'`            |
| `delete-variable`    | Delete one registration variable by ID               | `twenty application-registrations delete-variable <id>`                                       |
| `rotate-secret`      | Rotate the OAuth client secret for a registration    | `twenty application-registrations rotate-secret <id>`                                         |
| `transfer-ownership` | Transfer registration ownership to another workspace | `twenty application-registrations transfer-ownership <id> --target-workspace-subdomain other` |

**Options:**

```bash
-d, --data <json>                         # Inline JSON payload for create/update operations
-f, --file <path>                         # JSON payload file
--set <key=value>                         # Set nested payload values without rewriting full JSON
--target-workspace-subdomain <subdomain>  # Transfer destination for transfer-ownership
```

**Examples:**

```bash
twenty application-registrations list
twenty application-registrations get <application-registration-id> -o json
twenty application-registrations stats <application-registration-id>
twenty application-registrations tarball-url <application-registration-id>
twenty application-registrations list-variables <application-registration-id>
twenty application-registrations create -d '{"name":"Widget App","websiteUrl":"https://example.com"}'
twenty application-registrations rotate-secret <application-registration-id>
```

### Marketplace Apps

Inspect marketplace apps and install them by universal identifier.

```bash
twenty marketplace-apps <operation> [target] [options]
```

**Operations:**

| Operation | Description                              | Example                                              |
| --------- | ---------------------------------------- | ---------------------------------------------------- |
| `list`    | List marketplace apps                    | `twenty marketplace-apps list`                       |
| `get`     | Get one marketplace app by identifier    | `twenty marketplace-apps get com.example.widget`     |
| `install` | Install a marketplace app into workspace | `twenty marketplace-apps install com.example.widget` |

**Options:**

```bash
--version <version>   # Optional version pin for install
```

**Examples:**

```bash
twenty marketplace-apps list
twenty marketplace-apps get com.example.widget -o json
twenty marketplace-apps install com.example.widget --version 1.2.0
```

### Connected Accounts

Inspect connected accounts and trigger channel sync without dumping raw credentials by default.

```bash
twenty connected-accounts <operation> [id] [options]
```

**Operations:**

| Operation               | Description                                        | Example                                                                             |
| ----------------------- | -------------------------------------------------- | ----------------------------------------------------------------------------------- |
| `list`                  | List connected accounts                            | `twenty connected-accounts list`                                                    |
| `get`                   | Get one connected account by ID                    | `twenty connected-accounts get <id>`                                                |
| `sync`                  | Start channel sync for an account                  | `twenty connected-accounts sync <id>`                                               |
| `get-imap-smtp-caldav`  | Get one manual IMAP/SMTP/CALDAV account by ID      | `twenty connected-accounts get-imap-smtp-caldav <id>`                               |
| `save-imap-smtp-caldav` | Create or update a manual IMAP/SMTP/CALDAV account | `twenty connected-accounts save-imap-smtp-caldav --account-owner-id ... --file ...` |

**Options:**

```bash
--limit <number>              # Limit list results
--cursor <cursor>             # Pagination cursor
--show-secrets                # Show access tokens and connection parameters
--account-owner-id <id>       # Workspace member ID for save-imap-smtp-caldav
--handle <value>              # Email/account handle for save-imap-smtp-caldav
-d, --data <json>             # Inline JSON connectionParameters payload
-f, --file <path>             # JSON file with connectionParameters payload
--set <key=value>             # Nested connectionParameters field, e.g. IMAP.host=...
```

**Examples:**

```bash
twenty connected-accounts list
twenty connected-accounts list --limit 10 -o json
twenty connected-accounts get <connected-account-id>
twenty connected-accounts sync <connected-account-id>
twenty connected-accounts get-imap-smtp-caldav <connected-account-id>
twenty connected-accounts save-imap-smtp-caldav --account-owner-id <workspace-member-id> --handle mailbox@example.com --file ./manual-account.json
twenty connected-accounts save-imap-smtp-caldav --account-owner-id <workspace-member-id> --handle mailbox@example.com --set IMAP.host=imap.example.com --set IMAP.port=993 --set IMAP.password=secret
```

For `get-imap-smtp-caldav`, the CLI masks nested protocol passwords by default and only reveals them with `--show-secrets`. On older workspaces that do not expose the upstream GraphQL fields yet, the command fails explicitly instead of returning `undefined`.

### Message Channels

Inspect and update message channel settings. Internal sync cursors are hidden in output.

```bash
twenty message-channels <operation> [id] [options]
```

**Operations:**

| Operation | Description                   | Example                                   |
| --------- | ----------------------------- | ----------------------------------------- |
| `list`    | List message channels         | `twenty message-channels list`            |
| `get`     | Get one message channel by ID | `twenty message-channels get <id>`        |
| `update`  | Update one message channel    | `twenty message-channels update <id> ...` |

**Options:**

```bash
--limit <number>      # Limit list results
--cursor <cursor>     # Pagination cursor
-d, --data <json>     # Inline JSON payload for update
-f, --file <path>     # JSON file payload for update
--set <key=value>     # Set one field value for update
```

**Examples:**

```bash
twenty message-channels list
twenty message-channels get <message-channel-id> -o json
twenty message-channels update <message-channel-id> --set isSyncEnabled=false
```

### Calendar Channels

Inspect and update calendar channel settings. Internal sync cursors are hidden in output.

```bash
twenty calendar-channels <operation> [id] [options]
```

**Operations:**

| Operation | Description                    | Example                                    |
| --------- | ------------------------------ | ------------------------------------------ |
| `list`    | List calendar channels         | `twenty calendar-channels list`            |
| `get`     | Get one calendar channel by ID | `twenty calendar-channels get <id>`        |
| `update`  | Update one calendar channel    | `twenty calendar-channels update <id> ...` |

**Options:**

```bash
--limit <number>      # Limit list results
--cursor <cursor>     # Pagination cursor
-d, --data <json>     # Inline JSON payload for update
-f, --file <path>     # JSON file payload for update
--set <key=value>     # Set one field value for update
```

**Examples:**

```bash
twenty calendar-channels list
twenty calendar-channels get <calendar-channel-id> -o json
twenty calendar-channels update <calendar-channel-id> --set visibility=NOTHING
```

### Files

Upload and download files through verified Twenty file APIs.

```bash
twenty files <operation> [path-or-id] [options]
```

**Operations:**

| Operation      | Description                 | Example                                                                            |
| -------------- | --------------------------- | ---------------------------------------------------------------------------------- |
| `upload`       | Upload to a file target     | `twenty files upload ./document.pdf --target workflow`                             |
| `download`     | Download from signed URL/id | `twenty files download <signed-url>`                                               |
| `public-asset` | Download a public app asset | `twenty files public-asset images/logo.svg --workspace-id ws --application-id app` |

**Options:**

```bash
--output-file <path>                        # Output file path for downloads
--target <target>                           # Upload target: ai-chat, workflow, field, workspace-logo, profile-picture, app-tarball, application-file
--folder <folder>                           # Signed file folder: core-picture, files-field, workflow, agent-chat, app-tarball
--token <token>                             # Signed file token for /file downloads
--universal-identifier <id>                 # Optional universal identifier for app-tarball uploads
--workspace-id <id>                         # Workspace ID for public asset downloads
--application-id <id>                       # Application ID for public asset downloads
--application-universal-identifier <id>     # Application universal identifier for application-file uploads
--file-folder <folder>                      # built-logic-function, built-front-component, public-asset, source, dependencies
--file-path <path>                          # Remote application file path for application-file uploads
--field-metadata-id <id>                    # Field metadata ID for field uploads
--field-metadata-universal-identifier <id>  # Field metadata universal identifier for field uploads
```

**Examples:**

```bash
twenty files upload ./report.pdf --target workflow
twenty files upload ./avatar.png --target profile-picture
twenty files upload ./contract.pdf --target field --field-metadata-id <field-metadata-id>
twenty files upload ./app.tar.gz --target app-tarball --universal-identifier com.example.widget
twenty files upload ./logo.svg --target application-file --application-universal-identifier com.example.widget --file-folder public-asset --file-path images/logo.svg
twenty files download "https://api.twenty.com/file/files-field/file-123?token=..."
twenty files download file-123 --folder files-field --token "$SIGNED_FILE_TOKEN"
twenty files public-asset images/logo.svg --workspace-id <workspace-id> --application-id <application-id>
```

### Event Logs

Query enterprise event logs through the current `/metadata` GraphQL surface.

```bash
twenty event-logs <operation> [options]
```

**Operations:**

| Operation | Description               | Example                                                  |
| --------- | ------------------------- | -------------------------------------------------------- |
| `list`    | Query one event log table | `twenty event-logs list --table workspace-event -o json` |

**Options:**

```bash
--table <table>             # workspace-event, pageview, object-event, usage-event
--first <count>             # Page size (default: 100)
--after <cursor>            # Pagination cursor
--event-type <type>         # Filter by event type
--user-workspace-id <id>    # Filter by user workspace ID
--record-id <id>            # Filter by record ID
--object-metadata-id <id>   # Filter by object metadata ID
--start <date>              # ISO-8601 start timestamp
--end <date>                # ISO-8601 end timestamp
--include-page-info         # Include totalCount and pageInfo in the rendered result
```

**Examples:**

```bash
twenty event-logs list --table workspace-event -o json
twenty event-logs list --table pageview --include-page-info -o json
twenty event-logs list --table object-event --object-metadata-id <object-metadata-id> --event-type record.updated
```

### Serverless Functions

Manage current Twenty serverless functions through the stable `serverless` command name. The CLI prefers the current upstream `ServerlessFunction*` GraphQL surface and automatically falls back to older `LogicFunction*` schemas when your workspace is behind upstream.

```bash
twenty serverless <operation> [id] [options]
```

**Operations:**

| Operation      | Description                         | Example                                                             |
| -------------- | ----------------------------------- | ------------------------------------------------------------------- |
| `list`         | List serverless functions           | `twenty serverless list`                                            |
| `get`          | Get serverless function details     | `twenty serverless get <id>`                                        |
| `create`       | Create a serverless function        | `twenty serverless create --name "myFunc"`                          |
| `update`       | Update a serverless function        | `twenty serverless update <id> -d '{"name":"..."}'`                 |
| `delete`       | Delete a serverless function        | `twenty serverless delete <id>`                                     |
| `publish`      | Publish a serverless function       | `twenty serverless publish <id>`                                    |
| `execute`      | Execute a serverless function       | `twenty serverless execute <id> -d '{"input":"..."}'`               |
| `packages`     | Show available packages             | `twenty serverless packages <id>`                                   |
| `source`       | Get serverless function source code | `twenty serverless source <id>`                                     |
| `logs`         | Stream logic-function logs          | `twenty serverless logs <id> --max-events 1 -o jsonl`               |
| `create-layer` | Create a reusable serverless layer  | `twenty serverless create-layer --package-json ... --yarn-lock ...` |

**Options:**

```bash
-d, --data <json>      # JSON payload
-f, --file <path>      # JSON file
--set <key=value>      # Set a field value (repeatable)
--name <name>          # Function name (required for create)
--description <text>   # Function description
--timeout-seconds <n>  # Function timeout in seconds
--universal-identifier <id>           # Function universal identifier filter for logs
--application-id <id>                 # Application ID filter for logs
--application-universal-identifier    # Application universal identifier filter for logs
--max-events <n>        # Stop logs after N payloads
--wait-seconds <n>      # Stop logs after N seconds
--package-json <json>  # Layer package.json JSON for create-layer
--package-json-file    # Layer package.json file for create-layer
--yarn-lock <text>     # Layer yarn.lock content for create-layer
--yarn-lock-file       # Layer yarn.lock file for create-layer
```

**Examples:**

```bash
twenty serverless create --name "sendNotification" --description "Send email notification" --timeout-seconds 30
twenty serverless execute <serverless-function-id> -d '{"email":"recipient@example.com"}'
twenty serverless packages <serverless-function-id>
twenty serverless source <serverless-function-id> -o json
twenty serverless logs <serverless-function-id> --max-events 1 -o jsonl
twenty serverless create-layer --package-json '{"dependencies":{"zod":"^3.25.0"}}' --yarn-lock 'zod@^3.25.0:'
twenty serverless publish <serverless-function-id>
```

`create-layer` is only available on current upstream serverless schemas. On older workspaces that still expose the legacy `LogicFunction` surface, the CLI fails explicitly with a compatibility message instead of attempting a broken mutation.

## Output Formats

All commands support `--output` / `-o`:

| Format  | Description                    |
| ------- | ------------------------------ |
| `text`  | Human-readable table (default) |
| `json`  | Pretty-printed JSON output     |
| `jsonl` | Newline-delimited JSON output  |
| `agent` | Stable agent envelope output   |
| `csv`   | CSV output                     |

**JMESPath Queries:**

Filter JSON output with `--query`:

```bash
twenty api people list -o json --query '[].name.firstName'
twenty api opportunities list -o json --query "[?stage=='Qualification'].id"
```

## Global Options

| Option                  | Description                                  |
| ----------------------- | -------------------------------------------- |
| `-o, --output <format>` | Output format: text, json, jsonl, agent, csv |
| `--query <expr>`        | JMESPath filter expression                   |
| `--workspace <name>`    | Workspace profile to use                     |
| `--env-file <path>`     | Load env vars from a file                    |
| `--debug`               | Show request/response details                |
| `--no-retry`            | Disable automatic retry                      |

## Configuration

Configuration is stored in `~/.twenty/config.json`:

```json
{
  "defaultWorkspace": "default",
  "workspaces": {
    "default": {
      "apiKey": "...",
      "apiUrl": "https://api.twenty.com"
    },
    "staging": {
      "apiKey": "...",
      "apiUrl": "https://staging.twenty.com"
    }
  }
}
```

### Environment Variables

`.env` and `.env.local` are loaded automatically from the current working directory. You can also pass `--env-file <path>` or set `TWENTY_ENV_FILE`.

| Variable          | Description                  |
| ----------------- | ---------------------------- |
| `TWENTY_TOKEN`    | API token (overrides config) |
| `TWENTY_BASE_URL` | API base URL                 |
| `TWENTY_PROFILE`  | Default workspace name       |
| `TWENTY_OUTPUT`   | Default output format        |
| `TWENTY_QUERY`    | Default JMESPath query       |
| `TWENTY_DEBUG`    | Enable debug mode            |
| `TWENTY_NO_RETRY` | Disable retry                |

## Rate Limiting

The CLI handles rate limiting automatically with exponential backoff:

- Retries with increasing delays (1s, 2s, 4s) plus jitter
- Honors `Retry-After` header from API
- Up to 3 retries on 429/502/503/504 responses

Disable with `--no-retry` or `TWENTY_NO_RETRY=true`.

## Development

```bash
pnpm install        # Install workspace dependencies
pnpm setup          # Install prek hooks
pnpm build          # Build TypeScript
pnpm lint           # Run oxlint
pnpm format:check   # Verify formatting
pnpm typecheck      # Run TypeScript checks
pnpm test           # Run unit tests
pnpm test:e2e       # Run e2e tests
```

## Links

- [Twenty CRM](https://twenty.com)
- [Twenty API Documentation](https://twenty.com/developers/apis)
