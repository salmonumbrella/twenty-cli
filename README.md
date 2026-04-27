# Twenty CLI

CLI access to Twenty CRM. Use it to manage records, metadata, auth profiles, workflows, applications, files, and admin surfaces from a terminal or automation agent.

## Agent Mode

Use `--agent-mode` when an automation agentor script will consume the output. Agent mode forces JSON, defaults to compact light fields, still applies `--query` before output projection, and can be expanded with `--full` when canonical field names matter.

<!-- GENERATED:INSTALL_AND_AGENT_CONTRACT:START -->

## Installation

Install a standalone release archive, use the Homebrew formula if your account has tap access, or build from source.

```bash
# Latest macOS ARM64 archive; use linux_amd64, linux_arm64, or darwin_amd64 as needed
gh release download --repo salmonumbrella/twenty-cli --pattern 'twenty_*_darwin_arm64.tar.gz'
tar -xzf twenty_*_darwin_arm64.tar.gz
mkdir -p ~/.local/bin
install -m 0755 twenty ~/.local/bin/twenty

# Homebrew formula, updated by tagged releases; tap access required
brew install salmonumbrella/tap/twenty-cli
```

Tagged releases publish standalone `twenty` archives for macOS and Linux and update the maintained Homebrew formula. When `NPM_TOKEN` is configured, releases also publish the scoped npm package.

```bash
# Build from source
pnpm install
pnpm build
node packages/twenty-sdk/dist/cli/cli.js --help
```

## Agents

The CLI ships with agent-mode output plus a curated root help contract and machine-readable help for automation.

```bash
twenty --help
twenty --help-json
twenty --hj
twenty roles --help-json
twenty routes invoke --hj
twenty auth list --help-json
```

- Use `--agent-mode`, `--ai`, or `TWENTY_AGENT=true` to force JSON output and default to light fields
- Add `--full` when an agent or script needs canonical field names instead of compact light keys
- Prefer twenty CMD --help-json before executing mutations
- Stable JSON fields: path, args, options, operations, capabilities, exit_codes, output_contract

### Environment Loading

.env then .env.local then the explicit env file; existing environment variables win.

### Output Guarantees

- no-flag output is compact JSON
- --query runs before light projection and output formatting
- --light/--li renders compact short-key JSON fields
- --full renders canonical JSON field names
- --agent-mode forces JSON and behaves like --li unless --full is present
- jsonl renders one compact JSON record per line
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

## Quick Start

Configure a workspace profile with a Twenty API key:

```bash
twenty auth login --token "$TWENTY_TOKEN" --base-url https://api.twenty.com
twenty auth status
twenty auth workspace
```

For self-hosted Twenty, pass your instance URL as `--base-url`. You can keep
multiple profiles and switch between them:

```bash
twenty auth login --workspace staging --token "$STAGING_TOKEN" --base-url https://crm.example.com
twenty auth list
twenty auth switch staging
```

Before a mutation, inspect the exact command contract:

```bash
twenty api create --help-json
twenty roles upsert-object-permissions --help-json
```

Read and write records:

```bash
twenty search "acme" --objects companies,people
twenty api list people --limit 50
twenty api get companies <company-id>
twenty api create notes --data '{"title":"Follow up"}'
twenty api update people <person-id> --set jobTitle="Engineer"
```

Fetch Twenty's generated OpenAPI schemas:

```bash
twenty openapi core
twenty openapi metadata --output-file metadata-openapi.json
```

## How It Is Organized

Twenty has schema-per-workspace APIs, so the object and field names you create
in your workspace become your REST and GraphQL surface. This CLI mirrors that
model:

| Area             | Commands                                                                                    | Use For                                                                                                                |
| ---------------- | ------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| Workspace access | `auth`, `api-keys`, `approved-access-domains`                                               | Configure profiles, inspect the active workspace, manage API keys and access domains.                                  |
| Records          | `api`, `search`                                                                             | CRUD, imports, exports, duplicate detection, merges, full-text search, and grouping for standard or custom objects.    |
| Metadata         | `api-metadata`, `schema`, `openapi`                                                         | Manage objects, fields, views, layouts, cached discovery schemas, and REST OpenAPI discovery.                          |
| Admin surfaces   | `roles`, `public-domains`, `emailing-domains`, `postgres-proxy`, `dashboards`, `event-logs` | Manage workspace roles, domains, proxy credentials, dashboards, and event log queries.                                 |
| Applications     | `applications`, `application-registrations`, `marketplace-apps`                             | Sync app manifests, create development apps, generate app tokens, inspect registrations, and install marketplace apps. |
| Automation       | `workflows`, `routes`, `route-triggers`, `serverless`, `skills`                             | Invoke workflow webhooks, call route triggers, manage serverless functions, and manage AI skills.                      |
| Integrations     | `webhooks`, `connected-accounts`, `message-channels`, `calendar-channels`, `files`          | Manage webhook endpoints, channel state, manual IMAP/SMTP/CALDAV accounts, and file upload/download flows.             |
| Raw access       | `raw`, `graphql`, `mcp`                                                                     | Use escape-hatch REST/GraphQL calls or discover and execute Twenty MCP tools.                                          |
| DB-first reads   | `db`                                                                                        | Configure optional direct database profiles for supported self-hosted read paths.                                      |

Use `twenty <command> --help` for human-readable help and
`twenty <command> --help-json` for the contract that agents and scripts should
read.

## Records

Record commands use the object plural name, including custom objects:

```bash
twenty api list people --limit 25 -o text
twenty api get opportunities <opportunity-id> --include company
twenty api create companies --data '{"name":"Acme"}'
twenty api update people <person-id> --set city="Vancouver"
twenty api delete notes <note-id> --yes
twenty api import people ./people.csv --dry-run
twenty api export companies --format csv --output-file companies.csv
twenty api group-by opportunities --field stage
twenty api find-duplicates people --ids <person-id>
```

Batch and destructive operations support `--ids`, `--filter`, `--data`,
`--file`, and `--yes` depending on the operation. Check the command contract
before running a broad mutation:

```bash
twenty api batch-update --help-json
twenty api destroy --help-json
```

For self-hosted deployments, supported read paths can use direct database reads
when `TWENTY_DATABASE_URL` or an active `twenty db profile` is configured.
Mutations always stay on the official Twenty API.

## Metadata And Schema

Use metadata commands when changing the CRM data model or UI metadata:

```bash
twenty api-metadata objects list
twenty api-metadata objects get person
twenty api-metadata fields list --object person
twenty api-metadata views list --object person
twenty api-metadata page-layouts list --object person --page-layout-type RECORD_PAGE
```

The metadata families are:

```text
objects, fields, command-menu-items, front-components, navigation-menu-items,
views, view-fields, view-filters, view-filter-groups, view-groups, view-sorts,
page-layouts, page-layout-tabs, page-layout-widgets
```

Use `schema` for cached discovery schema diagnostics, and `openapi` when you
need the generated core or metadata OpenAPI documents.

## Automation And Apps

Twenty automation commands cover workflow webhooks, route triggers, serverless
functions, applications, registrations, marketplace apps, MCP, and skills:

```bash
twenty workflows invoke-webhook <workflow-id> --workspace-id <workspace-id>
twenty workflows run <workflow-version-id> --data '{"source":"cli"}'
twenty routes invoke public/ping
twenty route-triggers list
twenty serverless list
twenty serverless logs <serverless-function-id> --max-events 1 -o jsonl
twenty applications sync --manifest-file ./manifest.json
twenty applications create-development com.example.app --name "Example App"
twenty application-registrations tarball-url <application-registration-id>
twenty marketplace-apps list
twenty mcp catalog -o json
twenty mcp schema find_companies
twenty mcp exec find_companies --data '{"query":"Acme"}'
twenty skills list
```

Some upstream Twenty GraphQL surfaces require a user-authenticated bearer token
instead of a workspace API key. When a command hits one of those surfaces, the
CLI fails explicitly rather than silently returning partial data.

## Output And Configuration

Global options are available on most commands:

| Option                                  | Purpose                                                              |
| --------------------------------------- | -------------------------------------------------------------------- |
| `-o, --output <json\|jsonl\|csv\|text>` | Choose output format.                                                |
| `--query <expr>`                        | Apply a JMESPath query before formatting.                            |
| `--workspace <name>`                    | Select a saved workspace profile.                                    |
| `--env-file <path>`                     | Load an explicit environment file after `.env` and `.env.local`.     |
| `--debug`                               | Print request and response details.                                  |
| `--no-retry`                            | Disable retry/backoff for transient failures and rate limits.        |
| `--light`, `--li`                       | Emit compact short-key JSON.                                         |
| `--full`                                | Emit canonical field names.                                          |
| `--agent-mode`, `--ai`                  | Force JSON output and use light payloads unless `--full` is present. |

Configuration is stored in `~/.twenty/config.json`:

```json
{
  "defaultWorkspace": "default",
  "workspaces": {
    "default": {
      "apiKey": "...",
      "apiUrl": "https://api.twenty.com"
    }
  }
}
```

Environment variables can override saved configuration:

| Variable              | Purpose                                              |
| --------------------- | ---------------------------------------------------- |
| `TWENTY_TOKEN`        | API token.                                           |
| `TWENTY_BASE_URL`     | API base URL.                                        |
| `TWENTY_PROFILE`      | Default workspace profile.                           |
| `TWENTY_DB_PROFILE`   | Default DB profile.                                  |
| `TWENTY_DATABASE_URL` | Direct database URL for supported self-hosted reads. |
| `TWENTY_OUTPUT`       | Default output format.                               |
| `TWENTY_AGENT`        | Enable agent mode.                                   |
| `TWENTY_QUERY`        | Default JMESPath output filter.                      |
| `TWENTY_ENV_FILE`     | Default explicit env file path.                      |
| `TWENTY_DEBUG`        | Enable debug output.                                 |
| `TWENTY_NO_RETRY`     | Disable retries.                                     |

## Raw API Access

Use raw commands when the dedicated command surface does not cover a request:

```bash
twenty raw rest GET /health
twenty raw graphql query --document 'query { currentWorkspace { id displayName } }'
twenty graphql currentUser --selection 'id email'
twenty graphql schema --output-file schema.json
```

Prefer dedicated commands for stable automation. Raw commands are intentionally
thin wrappers around the active workspace API.

## Development

```bash
pnpm install
pnpm setup
pnpm build
pnpm lint
pnpm format:check
pnpm typecheck
pnpm test
pnpm test:e2e
```

Refresh generated README snippets after changing root help text:

```bash
pnpm readme:generate
```

Run the full repository verification used by CI:

```bash
pnpm verify:ci
```

## Links

- [Twenty CLI source](https://github.com/salmonumbrella/twenty-cli)
- [Twenty CRM](https://github.com/twentyhq/twenty)
- [Twenty API documentation](https://docs.twenty.com/developers/extend/api)
