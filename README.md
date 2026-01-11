# 📇 Twenty CLI — CRM in your terminal.

Twenty CRM in your terminal. Manage people, companies, opportunities, tasks, notes, webhooks, and custom objects.

## Features

- **Authentication** - authenticate once with API token and base URL
- **People** - create, update, delete, import/export contacts
- **Companies** - manage company records with all fields
- **Opportunities** - track deals with stages, amounts, and close dates
- **Tasks** - create and manage tasks with assignees and due dates
- **Notes** - attach notes to any record
- **Webhooks** - configure webhook endpoints
- **Custom Objects** - work with any Twenty object via generic records API
- **Multiple profiles** - manage multiple Twenty workspaces
- **Import/Export** - bulk operations with CSV and JSON files

## Installation

### Homebrew

```bash
brew install salmonumbrella/tap/twenty-cli
```

## Quick Start

### 1. Authenticate

Get your API token from Twenty CRM Settings -> APIs & Webhooks.

```bash
twenty auth login --token YOUR_API_TOKEN --base-url https://api.twenty.com
```

### 2. Test Authentication

```bash
twenty auth status
```

## Configuration

### Profile Selection

Specify the profile using either a flag or environment variable:

```bash
# Via flag
twenty people list --profile work

# Via environment
export TWENTY_PROFILE=work
twenty people list
```

### Environment Variables

- `TWENTY_TOKEN` - API token (bypasses keychain)
- `TWENTY_BASE_URL` - API base URL
- `TWENTY_PROFILE` - Default profile name to use
- `TWENTY_OUTPUT` - Output format: `text` (default), `json`, or `csv`
- `TWENTY_QUERY` - JQ-style query for JSON output
- `TWENTY_DEBUG` - Enable debug mode (`true`/`false`)
- `TWENTY_NO_COLOR` or `NO_COLOR` - Disable colors
- `TWENTY_NO_RETRY` - Disable automatic retry (`true`/`false`)

## Security

### Credential Storage

Credentials are stored securely in your system's keychain:
- **macOS**: Keychain Access
- **Linux**: Secret Service (GNOME Keyring, KWallet) or encrypted file fallback
- **Windows**: Credential Manager

For headless environments, use the file backend:

```bash
export TWENTY_KEYRING_BACKEND=file
export TWENTY_KEYRING_PASSWORD="your-secure-password"
```

## Rate Limiting

The Twenty API enforces rate limits. The CLI automatically handles rate limiting with:

- **Exponential backoff** - Retries with increasing delays (1s, 2s, 4s) plus jitter
- **Retry-After header respect** - Honors the API's suggested retry timing
- **Maximum retry attempts** - Up to 3 retries on 429 (Too Many Requests) responses
- **Retryable errors** - HTTP 429, 502, 503, 504

To disable automatic retry:

```bash
twenty --no-retry people list
```

## Commands

### Authentication

```bash
twenty auth login --token TOKEN --base-url URL    # Authenticate with token
twenty auth status                                 # Check authentication status
twenty auth status --show-token                    # Show full token value
twenty auth logout                                 # Remove credentials
twenty auth list                                   # List configured profiles
twenty auth switch <profile>                       # Set default profile
```

### People

```bash
twenty people list                                 # List people (default 20)
twenty people list --limit 50                      # Custom limit
twenty people list --all                           # Fetch all records
twenty people list --email john@example.com        # Filter by email
twenty people list --name John                     # Filter by name
twenty people get <id>                             # Get person by ID
twenty people get <id> --include company           # Include related company
twenty people create --first-name John --last-name Doe --email john@example.com
twenty people update <id> --job-title "Engineer"
twenty people delete <id> --force
twenty people upsert --email john@example.com --first-name John --job-title "CTO"
twenty people batch-create --file people.json
twenty people batch-delete --file ids.json --force
twenty people import people.csv                    # Import from CSV/JSON
twenty people export --format json --output people.json
```

### Companies

```bash
twenty companies list
twenty companies list --all
twenty companies get <id>
twenty companies create --name "Acme Corp" --domain "acme.com"
twenty companies create --name "TechCo" --employees 100 --revenue 1000000
twenty companies update <id> --employees 150
twenty companies delete <id> --force
```

### Opportunities

```bash
twenty opportunities list
twenty opportunities list --all
twenty opportunities get <id>
twenty opportunities create --name "Enterprise Deal" --amount 50000 --stage "Qualification"
twenty opportunities create --name "Big Deal" --amount 100000 --close-date 2024-06-30 --probability 75
twenty opportunities update <id> --stage "Negotiation"
twenty opportunities delete <id> --force
```

### Tasks

```bash
twenty tasks list
twenty tasks list --all
twenty tasks get <id>
twenty tasks create --title "Follow up with client"
twenty tasks create --title "Send proposal" --body "Include pricing" --due-at "2024-06-15T10:00:00Z"
twenty tasks update <id> --status "DONE"
twenty tasks delete <id> --force
```

### Notes

```bash
twenty notes list
twenty notes list --all
twenty notes get <id>
twenty notes create --title "Meeting notes" --body "Discussed Q2 roadmap..."
twenty notes update <id> --body "Updated content"
twenty notes delete <id> --force
```

### Webhooks

```bash
twenty webhooks list
twenty webhooks create --url "https://example.com/webhook" --operation "person.created"
twenty webhooks create --url "https://example.com/webhook" --operation "*.updated" --description "Track all updates"
twenty webhooks delete <id> --force
```

### Generic Records (Any Object)

```bash
twenty records list people --all
twenty records get companies <id>
twenty records create opportunities --data '{"name":"Big Deal"}'
twenty records update tasks <id> --set status="DONE"
twenty records delete people <id> --force
twenty records destroy people <id> --force         # Hard delete
twenty records restore people <id>
twenty records batch-create people --file people.json
twenty records batch-update people --file people.json
twenty records batch-delete people --ids id1,id2 --force
twenty records group-by opportunities --param groupBy=stage
twenty records find-duplicates people --data '{"fields":["email"]}'
twenty records merge people --data '{"sourceId":"...","targetId":"..."}'
twenty records export people --output people.json
twenty records import people people.json
```

### Metadata (Objects & Fields)

```bash
twenty objects list                                # List all object types
twenty objects get person                          # Get object details with fields
twenty objects create --data '{"nameSingular":"widget","namePlural":"widgets"}'
twenty fields list                                 # List all fields
twenty fields get <field-id>
twenty fields create --data '{"objectMetadataId":"...","name":"rating","type":"NUMBER"}'
```

### Raw API Access

```bash
twenty rest get /rest/people
twenty rest post /rest/people --data '{"name":{"firstName":"Ada","lastName":"Lovelace"}}'
twenty graphql query --query 'query { people(first: 5) { edges { node { id } } } }'
twenty graphql mutate --query 'mutation { createPerson(data:{...}) { id } }'
twenty graphql schema --endpoint graphql
```

### Configuration

```bash
twenty config show
twenty config set base_url https://api.twenty.com
twenty config set keyring_backend keychain
```

## Output Formats

### Text

Human-readable tables with colors and formatting:

```bash
$ twenty people list
ID                                    FIRST NAME    LAST NAME    EMAIL
550e8400-e29b-41d4-a716-446655440000  John          Doe          john@example.com
550e8400-e29b-41d4-a716-446655440001  Jane          Smith        jane@example.com
```

### JSON

Machine-readable output:

```bash
$ twenty people list -o json
[
  {"id": "550e8400...", "name": {"firstName": "John", "lastName": "Doe"}, ...},
  {"id": "550e8400...", "name": {"firstName": "Jane", "lastName": "Smith"}, ...}
]
```

Data goes to stdout, errors and progress to stderr for clean piping.

### CSV

```bash
twenty people list -o csv > people.csv
```

## Examples

### Create a person with company association

```bash
# First, create a company
twenty companies create --name "Acme Corp" --domain "acme.com"

# Then create a person linked to the company
twenty people create \
  --first-name John \
  --last-name Doe \
  --email john@acme.com \
  --company-id <companyId>
```

### Track an opportunity through stages

```bash
# Create the opportunity
twenty opportunities create \
  --name "Enterprise Deal" \
  --amount 50000 \
  --stage "Qualification" \
  --company-id <companyId>

# Update as it progresses
twenty opportunities update <id> --stage "Proposal"
twenty opportunities update <id> --stage "Negotiation" --probability 75
twenty opportunities update <id> --stage "Closed Won"
```

### Bulk import contacts

```bash
# From JSON file
twenty people batch-create --file contacts.json

# From CSV file
twenty people import contacts.csv

# Preview without importing
twenty people import contacts.csv --dry-run
```

### Switch between workspaces

```bash
# Check production workspace
twenty people list --profile production

# Check staging workspace
twenty people list --profile staging

# Or set default
export TWENTY_PROFILE=production
twenty people list
```

### JQ Filtering

Filter JSON output with JQ expressions:

```bash
# Get only email addresses
twenty people list -o json --query '.[].emails.primaryEmail'

# Filter by job title
twenty people list -o json --query '.[] | select(.jobTitle == "Engineer")'

# Extract opportunity IDs by stage
twenty opportunities list -o json --query '.[] | select(.stage == "Qualification") | .id'
```

## Global Flags

All commands support these flags:

- `--output <format>`, `-o` - Output format: `text`, `json`, or `csv` (default: text)
- `--profile <name>`, `-p` - Profile to use (overrides TWENTY_PROFILE)
- `--base-url <url>` - Custom API base URL
- `--config <path>` - Config file path (default: `$HOME/.twenty.yaml`)
- `--debug` - Enable debug output (shows API requests/responses)
- `--no-color` - Disable colored output
- `--no-retry` - Disable automatic retry on rate limiting
- `--query <expr>` - JQ filter expression for JSON output
- `--help` - Show help for any command
- `--version` - Show version information

## Shell Completions

Generate shell completions for your preferred shell:

### Bash

```bash
# macOS (Homebrew):
twenty completion bash > $(brew --prefix)/etc/bash_completion.d/twenty

# Linux:
twenty completion bash > /etc/bash_completion.d/twenty

# Or source directly in current session:
source <(twenty completion bash)
```

### Zsh

```zsh
# Save to fpath:
twenty completion zsh > "${fpath[1]}/_twenty"

# Or add to .zshrc for auto-loading:
echo 'source <(twenty completion zsh)' >> ~/.zshrc
```

### Fish

```fish
twenty completion fish > ~/.config/fish/completions/twenty.fish
```

### PowerShell

```powershell
# Load for current session:
twenty completion powershell | Out-String | Invoke-Expression

# Or add to profile for persistence:
twenty completion powershell >> $PROFILE
```

## Development

### Prerequisites

- Go 1.21+
- Make

### Building

```bash
make build      # Build binary
make install    # Install to GOPATH/bin
make test       # Run tests
make lint       # Run linters
make fmt        # Format code
make tidy       # Tidy dependencies
```

### Project Structure

```
cmd/twenty/           # Entry point
internal/
  api/                # API client (REST + GraphQL)
  auth/               # Authentication handling
  cmd/                # Cobra commands
  config/             # Configuration handling
  outfmt/             # Output formatting
  secrets/            # Secure credential storage
  types/              # Shared types
```

## License

MIT

## Links

- [Twenty CRM](https://twenty.com)
- [Twenty API Documentation](https://twenty.com/developers/apis)
