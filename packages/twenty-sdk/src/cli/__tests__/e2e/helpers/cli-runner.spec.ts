import { EventEmitter } from "node:events";
import path from "node:path";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ChildProcessWithoutNullStreams } from "node:child_process";

const spawnSyncMock = vi.fn();
const spawnMock = vi.fn();

vi.mock("node:child_process", () => ({
  spawnSync: spawnSyncMock,
  spawn: spawnMock,
}));

describe("cli runner", () => {
  beforeEach(() => {
    spawnSyncMock.mockReset();
    spawnMock.mockReset();
    delete process.env.TWENTY_TOKEN;
    delete process.env.TWENTY_BASE_URL;
  });

  it("resolves the built CLI artifact path", async () => {
    const { resolveBuiltCliPath } = await import("./cli-runner");

    expect(resolveBuiltCliPath()).toBe(
      path.resolve(__dirname, "../../../../../dist/cli/cli.js"),
    );
  });

  it("runs the built CLI synchronously and returns a shared result shape", async () => {
    spawnSyncMock.mockReturnValue({
      status: 0,
      stdout: '{"ok":true}',
      stderr: "",
    });

    const { runBuiltCli } = await import("./cli-runner");
    const result = runBuiltCli(["api", "list", "people", "--output", "json"], {
      cwd: "/tmp/twenty-cli",
      env: {
        CUSTOM_ENV: "present",
      },
    });

    expect(result).toEqual({
      exitCode: 0,
      stdout: '{"ok":true}',
      stderr: "",
    });
    expect(spawnSyncMock).toHaveBeenCalledWith(
      process.execPath,
      expect.arrayContaining(["api", "list", "people", "--output", "json"]),
      expect.objectContaining({
        cwd: "/tmp/twenty-cli",
        encoding: "utf-8",
        maxBuffer: 20 * 1024 * 1024,
        env: expect.objectContaining({
          CUSTOM_ENV: "present",
        }),
      }),
    );
  });

  it("runs the built CLI asynchronously and returns the same result shape", async () => {
    const child = createMockChildProcess();
    spawnMock.mockReturnValue(child);

    const { runBuiltCliAsync } = await import("./cli-runner");
    const pendingResult = runBuiltCliAsync(["auth", "discover", "https://acme.twenty.com"], {
      timeoutMs: 250,
    });

    child.stdout.emit("data", Buffer.from("hello "));
    child.stderr.emit("data", Buffer.from("warn"));
    child.stdout.emit("data", Buffer.from("world"));
    child.emit("exit", 0);

    await expect(pendingResult).resolves.toEqual({
      exitCode: 0,
      stdout: "hello world",
      stderr: "warn",
    });
  });

  it("filters inherited TWENTY_* env vars unless explicitly retained", async () => {
    process.env.UNRELATED_ENV = "keep-me";
    process.env.TWENTY_TOKEN = "drop-me";
    process.env.TWENTY_BASE_URL = "https://drop.example.com";
    spawnSyncMock.mockReturnValue({
      status: 0,
      stdout: "",
      stderr: "",
    });

    const { runBuiltCli } = await import("./cli-runner");

    runBuiltCli(["auth", "status"], {
      env: {
        TWENTY_PROFILE: "from-options",
      },
    });

    expect(spawnSyncMock.mock.calls[0]?.[2]).toEqual(
      expect.objectContaining({
        env: expect.objectContaining({
          UNRELATED_ENV: "keep-me",
          TWENTY_PROFILE: "from-options",
        }),
      }),
    );
    expect(spawnSyncMock.mock.calls[0]?.[2]?.env).not.toHaveProperty("TWENTY_TOKEN");
    expect(spawnSyncMock.mock.calls[0]?.[2]?.env).not.toHaveProperty("TWENTY_BASE_URL");

    runBuiltCli(["auth", "status"], {
      retainInheritedTwentyEnv: true,
    });

    expect(spawnSyncMock.mock.calls[1]?.[2]).toEqual(
      expect.objectContaining({
        env: expect.objectContaining({
          TWENTY_TOKEN: "drop-me",
          TWENTY_BASE_URL: "https://drop.example.com",
          UNRELATED_ENV: "keep-me",
        }),
      }),
    );
  });
});

function createMockChildProcess(): ChildProcessWithoutNullStreams {
  const child = new EventEmitter() as ChildProcessWithoutNullStreams;
  child.stdout = new EventEmitter() as ChildProcessWithoutNullStreams["stdout"];
  child.stderr = new EventEmitter() as ChildProcessWithoutNullStreams["stderr"];
  child.kill = vi.fn(() => true);

  return child;
}
