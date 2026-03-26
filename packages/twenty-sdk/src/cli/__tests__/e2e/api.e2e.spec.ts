import { describe, it, expect, vi } from "vitest";
import { execFileSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";
import { resolveLiveSmokeConfig } from "./helpers/live-config";

function resolveCliPath(): string {
  return path.resolve(__dirname, "../../../../dist/cli/cli.js");
}

const config = resolveLiveSmokeConfig({ required: false });
const cliPath = resolveCliPath();
const canRun = !!config && fs.existsSync(cliPath);

const describeIf = canRun ? describe : describe.skip;

describeIf("twenty api e2e", () => {
  it("lists people as json", () => {
    const env = {
      ...process.env,
      ...(config?.token ? { TWENTY_TOKEN: config.token } : {}),
      ...(config?.baseUrl ? { TWENTY_BASE_URL: config.baseUrl } : {}),
      ...(config?.profile ? { TWENTY_PROFILE: config.profile } : {}),
    };

    const output = execFileSync(
      "node",
      [cliPath, "api", "list", "people", "--limit", "1", "--output", "json"],
      {
        env,
        encoding: "utf-8",
      },
    );

    const parsed = JSON.parse(output);
    expect(Array.isArray(parsed)).toBe(true);
  });
});

const describeMutations = process.env.TWENTY_E2E_MUTATION === "true" ? describe : describe.skip;

describeMutations("twenty api e2e mutations", () => {
  it("creates and deletes a person", () => {
    if (!config) {
      throw new Error("Missing config");
    }
    runCreateAndDeletePerson({ config, cliPath, runCommand: execFileSync });
  });
});

describe("api mutation cleanup helper", () => {
  it("deletes created records even when later work fails", () => {
    const calls: Array<{ command: string; args: string[] }> = [];
    const runCommand = vi.fn((command: string, args: string[]) => {
      calls.push({ command, args });
      if (args[2] === "create") {
        return JSON.stringify({ id: "person-123" });
      }

      return "";
    });

    expect(() =>
      runCreateAndDeletePerson({
        config: {
          token: "env-token",
          baseUrl: "https://api.example.com",
          profile: "smoke",
        },
        cliPath: "/tmp/twenty",
        runCommand,
        afterCreate: () => {
          throw new Error("later failure");
        },
      }),
    ).toThrow("later failure");

    expect(calls.map(({ args }) => args[2])).toEqual(["create", "delete"]);
    expect(runCommand).toHaveBeenCalledWith(
      "node",
      ["/tmp/twenty", "api", "delete", "people", "person-123", "--yes"],
      expect.objectContaining({
        env: expect.objectContaining({
          TWENTY_TOKEN: "env-token",
          TWENTY_BASE_URL: "https://api.example.com",
          TWENTY_PROFILE: "smoke",
        }),
        encoding: "utf-8",
      }),
    );
  });
});

interface RunCreateAndDeletePersonOptions {
  config: NonNullable<typeof config>;
  cliPath: string;
  runCommand: typeof execFileSync;
  afterCreate?: () => void;
}

function runCreateAndDeletePerson({
  config: liveConfig,
  cliPath: resolvedCliPath,
  runCommand,
  afterCreate,
}: RunCreateAndDeletePersonOptions): void {
  const env = {
    ...process.env,
    ...(liveConfig.token ? { TWENTY_TOKEN: liveConfig.token } : {}),
    ...(liveConfig.baseUrl ? { TWENTY_BASE_URL: liveConfig.baseUrl } : {}),
    ...(liveConfig.profile ? { TWENTY_PROFILE: liveConfig.profile } : {}),
  };

  const createPayload = JSON.stringify({ name: { firstName: "E2E", lastName: "Test" } });
  let createdId: string | undefined;

  try {
    const createdRaw = runCommand(
      "node",
      [resolvedCliPath, "api", "create", "people", "--data", createPayload, "--output", "json"],
      {
        env,
        encoding: "utf-8",
      },
    );
    const created = JSON.parse(createdRaw);
    createdId = created.id;
    expect(created.id).toBeTruthy();
    afterCreate?.();
  } finally {
    if (createdId) {
      runCommand("node", [resolvedCliPath, "api", "delete", "people", createdId, "--yes"], {
        env,
        encoding: "utf-8",
      });
    }
  }
}
