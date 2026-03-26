import { spawn, spawnSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";

const CLI_PATH = path.resolve(__dirname, "../../../../../dist/cli/cli.js");

export interface TempHomeCliRunOptions {
  env?: NodeJS.ProcessEnv;
  retainInheritedTwentyEnv?: boolean;
}

export interface TempHomeCliRunResult {
  exitCode: number | null;
  stdout: string;
  stderr: string;
}

export function runCliWithTempHome(
  args: string[],
  options: TempHomeCliRunOptions = {},
): TempHomeCliRunResult {
  const homeDir = createTempHomeDir();

  try {
    const result = spawnSync(process.execPath, [CLI_PATH, ...args], {
      cwd: homeDir,
      env: createTempHomeEnv(homeDir, options),
      encoding: "utf-8",
      maxBuffer: 20 * 1024 * 1024,
    });

    if (result.error) {
      throw result.error;
    }

    return {
      exitCode: result.status,
      stdout: result.stdout ?? "",
      stderr: result.stderr ?? "",
    };
  } finally {
    fs.rmSync(homeDir, { recursive: true, force: true });
  }
}

export async function runCliWithTempHomeAsync(
  args: string[],
  options: TempHomeCliRunOptions = {},
): Promise<TempHomeCliRunResult> {
  const homeDir = createTempHomeDir();

  try {
    const child = spawn(process.execPath, [CLI_PATH, ...args], {
      cwd: homeDir,
      env: createTempHomeEnv(homeDir, options),
      stdio: ["ignore", "pipe", "pipe"],
    });

    let stdout = "";
    let stderr = "";

    child.stdout.on("data", (chunk) => {
      stdout += chunk.toString("utf-8");
    });
    child.stderr.on("data", (chunk) => {
      stderr += chunk.toString("utf-8");
    });

    const exitPromise = new Promise<{ exitCode: number | null }>((resolve) => {
      child.on("exit", (exitCode) => {
        resolve({ exitCode });
      });
    });

    const timeoutPromise = new Promise<{ exitCode: number | null }>((resolve) => {
      setTimeout(() => {
        child.kill("SIGKILL");
        resolve({ exitCode: null });
      }, 10000).unref();
    });

    const { exitCode } = await Promise.race([exitPromise, timeoutPromise]);

    return {
      exitCode,
      stdout,
      stderr,
    };
  } finally {
    fs.rmSync(homeDir, { recursive: true, force: true });
  }
}

function createTempHomeDir(): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), "twenty-cli-home-"));
}

function createTempHomeEnv(
  homeDir: string,
  options: TempHomeCliRunOptions,
): NodeJS.ProcessEnv {
  const inheritedEnv = Object.fromEntries(
    Object.entries(process.env).filter(
      ([key]) => options.retainInheritedTwentyEnv || !key.startsWith("TWENTY_"),
    ),
  );

  return {
    ...inheritedEnv,
    HOME: homeDir,
    USERPROFILE: homeDir,
    ...options.env,
  };
}
