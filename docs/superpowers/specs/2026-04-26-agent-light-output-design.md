# Agent-Light Output Design

Date: 2026-04-26
Status: Approved for local implementation

## Goal

Make Twenty CLI default to compact, agent-friendly JSON and replace the stale `agent` output format with a Chatwoot-style runtime contract: `--agent-mode` is behavior, `--light` is compact schema, and `--full` is the explicit canonical escape hatch.

## Runtime Contract

- No-flag commands render compact flat JSON by default.
- Valid output formats are `json`, `jsonl`, `csv`, and `text`.
- `agent` is removed from `--output` and `TWENTY_OUTPUT`; using it is an invalid argument.
- `--light` / `--li` renders compact short-key JSON.
- `--full` renders canonical full JSON with Twenty field names.
- `--agent-mode` / `--ai` implies JSON output, quiet/non-interactive behavior where available, and effective light output unless `--full` is explicit.
- `--light` and `--full` conflict.
- Destructive commands keep requiring `--yes`; agent mode does not auto-confirm mutations in this change.

## Alias Contract

Command aliases should optimize for agent desire paths without creating an unbounded alias matrix.

- Resource families get short aliases such as `records` -> `r`, `metadata` -> `md`, `schema` -> `sc`, `coverage` -> `cov`, and `api-metadata` -> `amd`.
- Common operations get stable aliases: `list` -> `ls`, `get` -> `g`, `create` -> `mk`, `update` -> `up`, `delete` -> `rm`, `destroy` -> `des`, `restore` -> `rs`, `batch-*` -> `b*`, `group-by` -> `gb`, `find-duplicates` -> `fd`, and `merge` -> `mg`.
- Cache-backed dynamic resources use kebab-case command names while preserving the schema name internally for API calls.
- Alias collisions are test failures.

## Light JSON Contract

Light output is a projection over canonical JSON, not a separate command implementation.

- The canonical payload remains the source of truth.
- Light mode recursively rewrites object keys through a central alias registry.
- One canonical field maps to one compact key.
- One compact key maps to one canonical field.
- Reserved structural keys are protected: `ok`, `err`, `items`, `is`, `meta`, `m`, `hint`, `h`, `bc`.
- Unknown keys pass through unchanged.
- Queries run before light projection so users can query canonical fields with `--query` and still get short-key output.

## Documentation

The curated root help and generated README snippets must not advertise `agent` as an output format. They should describe default compact JSON, `--li`, `--full`, and `--agent-mode`.

## Verification

Required verification:

- Unit tests for global option resolution.
- Unit tests for output projection.
- Help JSON tests proving `agent` is gone and new flags appear.
- Command alias tests for static and dynamic commands.
- `pnpm typecheck`
- `pnpm lint`
- `pnpm build`
- `pnpm test`
- Read-only built CLI smoke against a private Twenty instance.
