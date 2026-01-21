# twenty CLI

CLI for Twenty CRM — TypeScript implementation in `packages/twenty-sdk/src/cli/`

## Development

### Building

```bash
cd packages/twenty-sdk
npm install
npm run build    # Build TypeScript
npm test         # Run Vitest tests
```

### Project Structure

```
packages/twenty-sdk/src/cli/
├── cli.ts                          # Entry point
├── commands/
│   ├── api/                        # twenty api <object> <operation>
│   ├── api-metadata/               # twenty api-metadata <type> <operation>
│   ├── auth/                       # twenty auth <command>
│   └── raw/                        # twenty rest, twenty graphql
└── utilities/
    ├── api/                        # HTTP client with retry
    ├── config/                     # Workspace configuration
    ├── errors/                     # Error handling
    ├── file/                       # Import/export services
    ├── metadata/                   # Metadata operations
    ├── output/                     # Output formatting
    ├── records/                    # CRUD operations
    └── shared/                     # Parsing utilities
```

### Adding Commands

1. Create operation file in appropriate `commands/` subdirectory
2. Export a function that accepts context and options
3. Register with parent command

### Output Formats

All commands support `--output/-o` flag:
- `text` (default): Human-readable table
- `json`: JSON output
- `csv`: CSV output

### Testing

Uses Vitest. Test files are in `__tests__/` directories alongside source.

```bash
npm test                           # Run all tests
npm test -- src/cli/commands/auth  # Run specific tests
```
