# Twenty CLI

CLI for Twenty CRM. Manage records, metadata, and API access from your terminal.

## Installation

```bash
# Global install
npm install -g twenty-sdk

# Or run directly with npx
npx twenty-sdk <command>
```

## Quick Start

### 1. Authenticate

Get your API token from Twenty CRM Settings -> APIs & Webhooks.

```bash
twenty auth login --token YOUR_API_TOKEN --base-url https://api.twenty.com
```

### 2. Check Status

```bash
twenty auth status
```

### 3. List Records

```bash
twenty api people list
twenty api companies list --limit 10
```

## Commands

### Authentication

```bash
twenty auth login --token TOKEN --base-url URL    # Configure credentials
twenty auth status                                 # Check authentication status
twenty auth status --show-token                    # Show full token value
twenty auth list                                   # List configured workspaces
twenty auth switch <workspace>                     # Set default workspace
twenty auth logout                                 # Remove current workspace
twenty auth logout --workspace <name>              # Remove specific workspace
twenty auth logout --all                           # Remove all workspaces
```

### Records API

```bash
twenty api <object> <operation> [args]
```

**Operations:**

| Operation | Description | Example |
|-----------|-------------|---------|
| `list` | List records | `twenty api people list --limit 50` |
| `get` | Get single record | `twenty api people get <id>` |
| `create` | Create record | `twenty api people create -d '{"name":{"firstName":"John"}}'` |
| `update` | Update record | `twenty api people update <id> --set jobTitle="Engineer"` |
| `delete` | Soft delete | `twenty api people delete <id> --force` |
| `destroy` | Hard delete | `twenty api people destroy <id> --force` |
| `restore` | Restore deleted | `twenty api people restore <id>` |
| `batch-create` | Create multiple | `twenty api people batch-create -f records.json` |
| `batch-update` | Update multiple | `twenty api people batch-update -f records.json` |
| `batch-delete` | Delete multiple | `twenty api people batch-delete --ids id1,id2 --force` |
| `import` | Import from file | `twenty api people import data.csv` |
| `export` | Export to file | `twenty api people export --format json --output-file out.json` |
| `group-by` | Group records | `twenty api opportunities group-by --field stage` |
| `find-duplicates` | Find duplicates | `twenty api people find-duplicates -d '{"fields":["email"]}'` |
| `merge` | Merge records | `twenty api people merge --source <id> --target <id>` |

**Common Options:**

```bash
--limit <n>          # Limit results
--all                # Fetch all records (paginated)
--filter <expr>      # Filter expression
--include <rels>     # Include relations
--sort <field>       # Sort field
--order <asc|desc>   # Sort direction
--fields <f1,f2>     # Select specific fields
-d, --data <json>    # JSON payload
-f, --file <path>    # JSON/CSV file (use - for stdin)
--set <key=value>    # Set field value (repeatable)
--force              # Skip confirmation
```

### Metadata API

```bash
twenty api-metadata <type> <operation> [args]
```

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

| Option | Description |
|--------|-------------|
| `--limit <n>` | Maximum results (default: 20) |
| `--objects <list>` | Comma-separated object names to include |
| `--exclude <list>` | Comma-separated object names to exclude |

**Examples:**

```bash
twenty search "John Doe"                                      # Search all objects
twenty search "acme" --objects companies,people               # Search specific objects
twenty search "engineer" --limit 50 --exclude notes           # Exclude objects
```

### Webhooks

Manage webhook endpoints.

```bash
twenty webhooks <operation> [id] [options]
```

**Operations:**

| Operation | Description | Example |
|-----------|-------------|---------|
| `list` | List webhooks | `twenty webhooks list` |
| `get` | Get webhook details | `twenty webhooks get <id>` |
| `create` | Create webhook | `twenty webhooks create -d '{"targetUrl":"https://..."}'` |
| `update` | Update webhook | `twenty webhooks update <id> -d '{"targetUrl":"..."}'` |
| `delete` | Delete webhook | `twenty webhooks delete <id>` |

**Options:**

```bash
-d, --data <json>     # JSON payload
-f, --file <path>     # JSON file
--set <key=value>     # Set a field value (repeatable)
```

**Examples:**

```bash
twenty webhooks create --set targetUrl=https://example.com/webhook --set operations=create,update
twenty webhooks update abc123 --set description="Updated webhook"
```

### API Keys

Manage API keys.

```bash
twenty api-keys <operation> [id] [options]
```

**Operations:**

| Operation | Description | Example |
|-----------|-------------|---------|
| `list` | List API keys | `twenty api-keys list` |
| `get` | Get API key details | `twenty api-keys get <id>` |
| `create` | Create API key | `twenty api-keys create --name "CI Key"` |
| `revoke` | Revoke API key | `twenty api-keys revoke <id>` |

**Options:**

```bash
--name <name>        # API key name (required for create)
--expires-at <date>  # Expiration date (ISO format)
```

**Examples:**

```bash
twenty api-keys create --name "Production" --expires-at 2025-12-31
twenty api-keys list -o json
twenty api-keys revoke abc123
```

### Files

Manage file attachments.

```bash
twenty files <operation> [path-or-id] [options]
```

**Operations:**

| Operation | Description | Example |
|-----------|-------------|---------|
| `list` | List all files | `twenty files list` |
| `upload` | Upload file | `twenty files upload ./document.pdf` |
| `download` | Download file | `twenty files download <id> --output-file doc.pdf` |
| `delete` | Delete file | `twenty files delete <id>` |

**Options:**

```bash
--output-file <path>  # Output file path (for download)
```

**Examples:**

```bash
twenty files list
twenty files upload ./report.pdf
twenty files download abc123 --output-file ./downloaded.pdf
twenty files delete abc123
```

### Serverless Functions

Manage and execute serverless functions.

```bash
twenty serverless <operation> [id] [options]
```

**Operations:**

| Operation | Description | Example |
|-----------|-------------|---------|
| `list` | List functions | `twenty serverless list` |
| `get` | Get function details | `twenty serverless get <id>` |
| `create` | Create function | `twenty serverless create --name "myFunc"` |
| `update` | Update function | `twenty serverless update <id> -d '{"name":"..."}'` |
| `delete` | Delete function | `twenty serverless delete <id>` |
| `execute` | Execute function | `twenty serverless execute <id> -d '{"input":"..."}'` |
| `publish` | Publish function | `twenty serverless publish <id>` |
| `source` | Get source code | `twenty serverless source <id>` |

**Options:**

```bash
-d, --data <json>      # JSON payload
-f, --file <path>      # JSON file
--name <name>          # Function name (required for create)
--description <text>   # Function description
```

**Examples:**

```bash
twenty serverless create --name "sendNotification" --description "Send email notification"
twenty serverless execute abc123 -d '{"email":"user@example.com"}'
twenty serverless publish abc123
twenty serverless source abc123 -o json
```

## Output Formats

All commands support `--output` / `-o`:

| Format | Description |
|--------|-------------|
| `text` | Human-readable table (default) |
| `json` | JSON output |
| `csv` | CSV output |

**JMESPath Queries:**

Filter JSON output with `--query`:

```bash
twenty api people list -o json --query '[].name.firstName'
twenty api opportunities list -o json --query "[?stage=='Qualification'].id"
```

## Global Options

| Option | Description |
|--------|-------------|
| `-o, --output <format>` | Output format: text, json, csv |
| `--query <expr>` | JMESPath filter expression |
| `--workspace <name>` | Workspace profile to use |
| `--debug` | Show request/response details |
| `--no-retry` | Disable automatic retry |

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

| Variable | Description |
|----------|-------------|
| `TWENTY_TOKEN` | API token (overrides config) |
| `TWENTY_BASE_URL` | API base URL |
| `TWENTY_PROFILE` | Default workspace name |
| `TWENTY_OUTPUT` | Default output format |
| `TWENTY_QUERY` | Default JMESPath query |
| `TWENTY_DEBUG` | Enable debug mode |
| `TWENTY_NO_RETRY` | Disable retry |

## Rate Limiting

The CLI handles rate limiting automatically with exponential backoff:

- Retries with increasing delays (1s, 2s, 4s) plus jitter
- Honors `Retry-After` header from API
- Up to 3 retries on 429/502/503/504 responses

Disable with `--no-retry` or `TWENTY_NO_RETRY=true`.

## Development

```bash
npm run build       # Build TypeScript
npm run dev         # Watch mode
npm run test        # Run tests
npm run test:watch  # Watch tests
```

## Links

- [Twenty CRM](https://twenty.com)
- [Twenty API Documentation](https://twenty.com/developers/apis)
