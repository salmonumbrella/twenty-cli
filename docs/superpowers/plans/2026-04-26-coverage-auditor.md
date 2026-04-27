# Coverage Auditor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `twenty coverage compare --upstream <path>` to audit first-class CLI coverage against a local Twenty checkout.

**Architecture:** Keep command registration thin and put comparison logic in `utilities/coverage`. The first implementation reads local upstream source files and this CLI's help/source surface, then emits a stable structured report through the existing output service.

**Tech Stack:** TypeScript, Commander, Vitest, Node `fs`/`path`, existing CLI output and help-json utilities.

---

## File Structure

- Create `packages/twenty-sdk/src/cli/utilities/coverage/coverage-auditor.ts`
  - Extract upstream metadata resources.
  - Parse GraphQL Query/Mutation field names from SDL.
  - Build expected core REST and metadata REST requirements.
  - Compare requirements against CLI help/source coverage.
- Create `packages/twenty-sdk/src/cli/utilities/coverage/__tests__/coverage-auditor.spec.ts`
  - Test upstream extraction and comparison with temp fixtures.
- Create `packages/twenty-sdk/src/cli/commands/coverage/coverage.command.ts`
  - Register `coverage compare`.
  - Use `createOutputContext` and `output.render`.
- Create `packages/twenty-sdk/src/cli/commands/coverage/__tests__/coverage.command.spec.ts`
  - Test command registration and execution.
- Modify `packages/twenty-sdk/src/cli/program.ts`
  - Import and register coverage command.
- Modify `packages/twenty-sdk/src/cli/help/constants.ts`
  - Add coverage examples and read-only operation metadata.
- Modify `packages/twenty-sdk/src/cli/__tests__/help.spec.ts`
  - Add coverage help-json assertions.

## Task 1: Coverage Auditor Core

**Files:**

- Create: `packages/twenty-sdk/src/cli/utilities/coverage/coverage-auditor.ts`
- Test: `packages/twenty-sdk/src/cli/utilities/coverage/__tests__/coverage-auditor.spec.ts`

- [x] **Step 1: Write failing tests**

Create tests that write a temp upstream fixture with:

```ts
const openApiService = `
const metadata = [
  { nameSingular: 'object', namePlural: 'objects' },
  { nameSingular: 'apiKey', namePlural: 'apiKeys' },
  { nameSingular: 'viewFieldGroup', namePlural: 'viewFieldGroups' },
];
`;

const schema = `
type Query {
  currentUser: User!
  getViewFieldGroups(viewId: String!): [ViewFieldGroup!]!
}

type Mutation {
  createApiKey(input: CreateApiKeyInput!): ApiKey!
  deleteViewFieldGroup(input: DeleteViewFieldGroupInput!): ViewFieldGroup!
}
`;
```

Assert:

- Metadata resources normalize to `objects`, `api-keys`, and `view-field-groups`.
- GraphQL operation names are parsed as `currentUser`, `getViewFieldGroups`, `createApiKey`, and `deleteViewFieldGroup`.
- `compareCoverage()` returns `missing_coverage` when `api-keys delete`, `view-field-groups list/get/create/update/delete`, and the GraphQL operations are not first-class covered.
- Raw GraphQL is noted as fallback coverage but does not remove GraphQL first-class missing counts.
- An invalid upstream path throws a clear error mentioning the missing expected file.

Run:

```bash
pnpm --filter ./packages/twenty-sdk test -- src/cli/utilities/coverage/__tests__/coverage-auditor.spec.ts
```

Expected: FAIL because the module does not exist yet.

- [x] **Step 2: Implement minimal auditor**

Implement:

```ts
export interface CoverageCompareOptions {
  upstreamPath: string;
  cliRoot?: string;
}

export interface CoverageReport {
  status: "ok" | "missing_coverage";
  upstreamPath: string;
  summary: Record<string, number>;
  missing: CoverageGap[];
  notes: string[];
  _cli?: { message: string };
}

export async function compareCoverage(options: CoverageCompareOptions): Promise<CoverageReport>;
```

Implementation requirements:

- Use `fs-extra` read helpers.
- Default `cliRoot` to the current package source root.
- Use this local CLI's `buildProgram()` and `buildHelpJson()` to inspect command paths.
- Read every `.ts` file under `src/cli` and scan for GraphQL operation literals.
- Parse GraphQL SDL operation blocks with brace depth, ignoring argument bodies.
- Normalize camel-case metadata resource plurals to kebab-case command names.
- Produce flat missing gap items:

```ts
{
  surface: "metadata-rest",
  name: "apiKeys.delete",
  upstreamPath: "/metadata/apiKeys/{id}",
  suggestedCommand: "twenty api-metadata api-keys delete <id>"
}
```

- [x] **Step 3: Run tests and commit**

Run:

```bash
pnpm --filter ./packages/twenty-sdk test -- src/cli/utilities/coverage/__tests__/coverage-auditor.spec.ts
```

Expected: PASS.

Commit:

```bash
git add packages/twenty-sdk/src/cli/utilities/coverage
git commit -m "feat: add coverage auditor core"
```

## Task 2: Coverage Command

**Files:**

- Create: `packages/twenty-sdk/src/cli/commands/coverage/coverage.command.ts`
- Test: `packages/twenty-sdk/src/cli/commands/coverage/__tests__/coverage.command.spec.ts`
- Modify: `packages/twenty-sdk/src/cli/program.ts`

- [x] **Step 1: Write failing command tests**

Assert:

- Top-level `coverage` command exists.
- `compare` subcommand exists.
- `--upstream <path>` is required.
- The command renders the auditor result through `console.log` when `-o json` is used against a fixture.
- `buildProgram()` includes the `coverage` command.

Run:

```bash
pnpm --filter ./packages/twenty-sdk test -- src/cli/commands/coverage/__tests__/coverage.command.spec.ts
```

Expected: FAIL because the command does not exist.

- [x] **Step 2: Implement command registration**

Register:

```ts
export function registerCoverageCommand(program: Command): void {
  const coverage = program.command("coverage").description("Audit CLI coverage against upstream");
  applyGlobalOptions(coverage);

  registerCommand(
    coverage,
    "compare",
    "Compare CLI coverage against an upstream checkout",
    (command) => {
      command.requiredOption("--upstream <path>", "Path to upstream Twenty checkout");
      applyGlobalOptions(command);
      command.action(async (_options, actionCommand) => {
        const options = actionCommand.opts() as { upstream: string };
        const { globalOptions, output } = createOutputContext(actionCommand);
        const result = await compareCoverage({ upstreamPath: options.upstream });
        await output.render(result, {
          format: globalOptions.output,
          query: globalOptions.query,
          kind: globalOptions.outputKind,
        });
      });
    },
  );
}
```

- [x] **Step 3: Run tests and commit**

Run:

```bash
pnpm --filter ./packages/twenty-sdk test -- src/cli/commands/coverage/__tests__/coverage.command.spec.ts
```

Expected: PASS.

Commit:

```bash
git add packages/twenty-sdk/src/cli/commands/coverage packages/twenty-sdk/src/cli/program.ts
git commit -m "feat: add coverage compare command"
```

## Task 3: Help Contract

**Files:**

- Modify: `packages/twenty-sdk/src/cli/help/constants.ts`
- Modify: `packages/twenty-sdk/src/cli/__tests__/help.spec.ts`

- [x] **Step 1: Write failing help tests**

Assert:

- Root help JSON includes `coverage`.
- `twenty coverage --help-json` exposes `compare` as a read-only operation.
- `twenty coverage compare --help-json` exposes required `upstream`, global `output`, global `query`, and the output contract.

Run:

```bash
pnpm --filter ./packages/twenty-sdk test -- src/cli/__tests__/help.spec.ts
```

Expected: FAIL until metadata is added.

- [x] **Step 2: Add help metadata**

Add examples:

```ts
"twenty coverage": {
  operations: [{ name: "compare", summary: "Compare CLI coverage against upstream", mutates: false }],
  examples: ["twenty coverage compare --upstream /path/to/twenty -o json"],
}
```

- [x] **Step 3: Run tests and commit**

Run:

```bash
pnpm --filter ./packages/twenty-sdk test -- src/cli/__tests__/help.spec.ts
```

Expected: PASS.

Commit:

```bash
git add packages/twenty-sdk/src/cli/help/constants.ts packages/twenty-sdk/src/cli/__tests__/help.spec.ts
git commit -m "test: document coverage help contract"
```

## Task 4: Verification

- [x] Run targeted coverage command against the local Twenty checkout:

```bash
pnpm build
node packages/twenty-sdk/dist/cli/cli.js coverage compare --upstream /path/to/twenty -o json
```

Expected after the later gap-closing commits: JSON report with `status: "ok"` and no missing gaps.

- [x] Run full checks:

```bash
pnpm test
pnpm typecheck
pnpm lint
pnpm build
```

Expected: all pass, except any pre-existing repo hygiene drift if that separate command is run.
