# Twenty MCP CLI Design

Date: 2026-03-24

## Goal

Add a first-class `twenty mcp` namespace that talks to the official Twenty MCP server over the existing workspace API-key configuration, exposing the full official high-level MCP surface for diagnostics, discovery, learning, execution, skills, and help-center access.

## Context

The current CLI already has strong direct API coverage through REST, GraphQL, and first-class wrappers for records, metadata, roles, skills, workflows, files, and integrations. What it does not have is an MCP client layer.

The official Twenty MCP server is built into `twenty-server` and exposed at `POST /mcp`. It is not a separate package to install. The official workflow is based on these high-level tools:

- `get_tool_catalog`
- `learn_tools`
- `execute_tool`
- `load_skills`
- `search_help_center`

A locally configured production workspace proved the endpoint exists in production, but currently returns `403 AI feature is not enabled for this workspace`. That means the CLI needs explicit MCP diagnostics, not just tool execution.

## Scope

This design covers only the first implementation slice:

- official Twenty MCP only
- remote HTTP MCP only
- API-key auth only, reusing the existing workspace config in `~/.twenty/config.json`
- first-class `twenty mcp` commands for all official high-level MCP tools
- MCP diagnostics and error handling
- tests and help/contract updates for the new command surface

## Non-Goals

This design does not include:

- OAuth or dynamic client registration for MCP
- stdio/community MCP server support
- compatibility aliases for third-party Twenty MCP tool names
- reimplementing existing CRUD or metadata operations as MCP-first wrappers
- natural-language planner behavior inside the CLI

## Design Principles

- Keep MCP isolated from the existing direct API command families.
- Reuse current workspace selection and API-key loading instead of inventing a second credential store.
- Expose the official MCP workflow explicitly rather than hiding it behind magic behavior.
- Preserve the CLI's current output/query/debug conventions.
- Make workspace gating and auth failures easy to diagnose.

## Architecture

Add a dedicated MCP client layer under the CLI utilities and a new top-level `mcp` command namespace in the CLI program.

The command layer will stay thin. It should translate CLI arguments into calls on an MCP service/client, then hand structured results to the shared output layer. All JSON-RPC transport details, request envelopes, response parsing, and MCP-specific error classification should live in the new MCP utility layer.

The new MCP client should:

- resolve `apiUrl` and `apiKey` from the existing `ConfigService`
- target `POST ${apiUrl}/mcp`
- send JSON-RPC 2.0 requests
- run `initialize` before invoking high-level tools
- support `tools/list` and `tools/call`
- provide typed helpers around the official tool names
- classify transport, auth, and workspace-gating failures into normal CLI errors

This keeps MCP concerns isolated and matches the structure already used by the local Notion CLI.

## Command Surface

Add a new top-level namespace:

- `twenty mcp status`
- `twenty mcp catalog`
- `twenty mcp learn <tool...>`
- `twenty mcp call <tool> [args-json]`
- `twenty mcp load-skills <name...>`
- `twenty mcp help-center <query>`

### `twenty mcp status`

Purpose:

- verify MCP connectivity for the active workspace
- surface endpoint URL, auth mode, protocol version, server info, and availability state

Behavior:

- load workspace config
- send `initialize`
- return a structured status object
- if MCP is unavailable because AI is disabled, return that as a first-class diagnosis rather than generic noise

`status` is a diagnostic command. Unlike the mutating or discovery commands, it should not fail with `AUTH` for expected reachable MCP states. If the endpoint is reached and returns a meaningful MCP/auth/workspace-gating response, `status` should render a structured object and exit `0`.

Expected structured shape:

```json
{
  "endpoint": "https://workspace.example.com/mcp",
  "authMode": "api-key",
  "reachable": true,
  "available": false,
  "state": "ai_feature_disabled",
  "protocolVersion": "2024-11-05",
  "serverInfo": {
    "name": "Twenty MCP Server",
    "version": "0.1.0"
  },
  "message": "AI feature is not enabled for this workspace"
}
```

Expected `state` values:

- `ok`
- `ai_feature_disabled`
- `unauthorized`
- `forbidden`

If the command cannot reach the endpoint at all, it should still fail as `NETWORK`.

### `twenty mcp catalog`

Purpose:

- wrap the official `get_tool_catalog`

Behavior:

- initialize MCP
- call `tools/call` for `get_tool_catalog`
- print the returned catalog in standard CLI formats

### `twenty mcp learn <tool...>`

Purpose:

- wrap the official `learn_tools`

Behavior:

- require one or more exact tool names
- initialize MCP
- call `learn_tools`
- print schemas/descriptions as structured output

### `twenty mcp call <tool> [args-json]`

Purpose:

- provide a raw escape hatch for official MCP tools

Behavior:

- accept positional inline JSON or `--args` / `--args-file`
- initialize MCP
- call the exact MCP tool requested through protocol-level `tools/call`
- parse JSON text results when possible
- fall back to plain text output if the result is not valid JSON

This command is the generic way to call `execute_tool` directly when the user already knows the tool name and arguments they want.

Important contract:

- `call` accepts official MCP tool names like `get_tool_catalog`, `learn_tools`, `execute_tool`, `load_skills`, and `search_help_center`
- `call` does not pre-validate the tool name against the catalog; it forwards the exact name and arguments to the server and lets the server decide
- `call` targets MCP tools, not raw JSON-RPC method names
- positional `[args-json]` is mutually exclusive with `--args` and `--args-file`
- `--args` is mutually exclusive with `--args-file`
- if no arguments are provided, `call` should send an empty object as the MCP tool arguments payload

Protocol envelope sent by `call`:

```json
{
  "jsonrpc": "2.0",
  "id": "<generated-id>",
  "method": "tools/call",
  "params": {
    "name": "<tool-name>",
    "arguments": { "...": "..." }
  }
}
```

Example:

```bash
twenty mcp call execute_tool --args '{"toolName":"find_companies","arguments":{"filter":{"name":{"ilike":"%Acme%"}}}}'
```

That command should send:

```json
{
  "jsonrpc": "2.0",
  "id": "<generated-id>",
  "method": "tools/call",
  "params": {
    "name": "execute_tool",
    "arguments": {
      "toolName": "find_companies",
      "arguments": {
        "filter": {
          "name": {
            "ilike": "%Acme%"
          }
        }
      }
    }
  }
}
```

### `twenty mcp load-skills <name...>`

Purpose:

- wrap the official `load_skills`

Behavior:

- initialize MCP
- call `load_skills`
- print loaded skill content/metadata through the normal output layer

### `twenty mcp help-center <query>`

Purpose:

- wrap the official `search_help_center`

Behavior:

- initialize MCP
- call `search_help_center`
- print structured help-center search results

## Data Flow

For all `twenty mcp` commands:

1. Commander parses arguments and global options.
2. Existing config/environment services resolve the active workspace profile.
3. A new MCP service/client is created with the resolved `apiUrl` and `apiKey`.
4. The client sends `initialize` to `POST /mcp`.
5. The client either:
   - calls `tools/list`, or
   - calls `tools/call` with the relevant high-level MCP tool.
6. MCP result content is normalized into plain objects, arrays, or strings.
7. The shared output service renders the normalized result.

The JSON-RPC layer should remain mostly hidden from normal users, but `--debug` should expose the envelopes for troubleshooting.

## Auth Model

V1 is API-key-only.

The CLI should reuse:

- `TWENTY_BASE_URL`
- `TWENTY_TOKEN`
- `TWENTY_PROFILE`
- `~/.twenty/config.json`

There should be no separate `twenty mcp login` in v1.

If a workspace has no API key configured, MCP commands should fail the same way other authenticated commands fail today, with a normal `AUTH` suggestion path.

## Error Handling

MCP failures should be translated into the CLI's existing error/exit-code model.

Expected mapping:

- auth failure or permission failure -> `AUTH`
- workspace MCP disabled / AI feature disabled -> `AUTH`
- no HTTP response / transport failure -> `NETWORK`
- rate limit -> `RATE_LIMIT`
- malformed tool args or command misuse -> `INVALID_ARGUMENTS`

Important special case:

- `twenty mcp status` is diagnostic and should return exit `0` with structured `state` output for reachable expected states like `ai_feature_disabled`, `unauthorized`, or `forbidden`
- specifically, HTTP/JSON-RPC auth responses that indicate 401/403-style access denial should map to `state=unauthorized` or `state=forbidden` and still exit `0` when returned by `status`, because the endpoint was reached successfully
- non-status MCP commands should still classify `AI feature is not enabled for this workspace` as `AUTH` and suggest using a workspace with MCP/AI enabled

The generic `call` command should still preserve enough context in error output to debug MCP-level failures without requiring `--debug`.

## Output Contract

All new commands should use the shared output service so they inherit:

- `text`
- `json`
- `jsonl`
- `agent`
- `csv`
- `--query`

Normalization rules:

- structured JSON from MCP should remain structured
- `catalog`, `learn`, `load-skills`, and `help-center` should favor object/array output
- `call` should attempt JSON parsing first, then return text if parsing fails
- `status` should always return a structured object

`--debug` should include:

- endpoint URL
- tool name being called
- JSON-RPC request envelope
- JSON-RPC response envelope

## Testing Strategy

Implementation should follow TDD.

Required tests:

### MCP client tests

- request envelope generation for `initialize`, `tools/list`, and `tools/call`
- response parsing for successful JSON-RPC results
- parsing of text-wrapped JSON tool outputs
- error classification for:
  - 401/403 auth failures
  - workspace AI disabled
  - transport/network failure
  - rate limiting

### Command tests

- `mcp status`
- `mcp catalog`
- `mcp learn`
- `mcp call`
- `mcp load-skills`
- `mcp help-center`

Each command test should verify:

- argument parsing
- interaction with the MCP service/client
- structured output behavior
- invalid-argument handling where relevant

### Help/contract tests

- root help text includes the new `mcp` namespace
- command-specific help contracts include the new MCP commands
- `--help-json` remains accurate

## File Structure Direction

Expected new areas:

- a new MCP command module under `packages/twenty-sdk/src/cli/commands/mcp/`
- a new MCP utility/service area under `packages/twenty-sdk/src/cli/utilities/mcp/`

The design intentionally keeps:

- commander wiring in command files
- transport/protocol logic in utility/service files
- error translation close to the MCP client/service

## Rollout Order

Recommended execution order:

1. add MCP utility/client and failing tests
2. add `mcp status`
3. add `mcp catalog`
4. add `mcp learn`
5. add raw `mcp call`
6. add `mcp load-skills`
7. add `mcp help-center`
8. update help docs and contracts
9. run verification and end-to-end mocked checks

## Risks And Mitigations

### Risk: Official MCP payloads vary by workspace

Mitigation:

- keep result normalization defensive
- test against representative mocked envelopes
- avoid over-constraining catalog/learn payload shapes too early

### Risk: Workspace MCP feature is disabled

Mitigation:

- make `status` a first-class diagnostic path
- classify the feature-disabled response cleanly

### Risk: `call` returns inconsistent text vs JSON

Mitigation:

- parse JSON when possible
- preserve plain text when not
- document the fallback behavior

## Success Criteria

The design is successful when:

- a user can run `twenty mcp status` and immediately understand whether MCP is usable for the active workspace
- a user can discover official MCP tools with `catalog`
- a user can inspect tool schemas with `learn`
- a user can execute any official high-level MCP tool with `call`
- a user can access official `load_skills` and `search_help_center` directly
- the new commands follow existing CLI output/help/error conventions
- no direct API command surfaces are regressed or entangled with MCP transport logic
