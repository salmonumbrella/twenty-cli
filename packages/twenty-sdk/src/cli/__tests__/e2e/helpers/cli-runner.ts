import { spawn, spawnSync } from "node:child_process";
import path from "node:path";

const DEFAULT_TIMEOUT_MS = 10_000;
const MAX_BUFFER = 20 * 1024 * 1024;

export interface BuiltCliRunOptions {
  cwd?: string;
  env?: NodeJS.ProcessEnv;
  retainInheritedTwentyEnv?: boolean;
  timeoutMs?: number;
}

export interface BuiltCliRunResult {
  exitCode: number | null;
  stdout: string;
  stderr: string;
}

export function resolveBuiltCliPath(): string {
  return path.resolve(__dirname, "../../../../../dist/cli/cli.js");
}

export function runBuiltCli(
  args: string[],
  options: BuiltCliRunOptions = {},
): BuiltCliRunResult {
  const result = spawnSync(process.execPath, [resolveBuiltCliPath(), ...args], {
    cwd: options.cwd,
    env: composeCliEnv(options),
    encoding: "utf-8",
    maxBuffer: MAX_BUFFER,
  });

  if (result.error) {
    throw result.error;
  }

  return {
    exitCode: result.status,
    stdout: result.stdout ?? "",
    stderr: result.stderr ?? "",
  };
}

export async function runBuiltCliAsync(
  args: string[],
  options: BuiltCliRunOptions = {},
): Promise<BuiltCliRunResult> {
  const child = spawn(process.execPath, [resolveBuiltCliPath(), ...args], {
    cwd: options.cwd,
    env: composeCliEnv(options),
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

  return await new Promise<BuiltCliRunResult>((resolve, reject) => {
    const timeout = setTimeout(() => {
      child.kill("SIGKILL");
      resolve({
        exitCode: null,
        stdout,
        stderr,
      });
    }, options.timeoutMs ?? DEFAULT_TIMEOUT_MS);
    timeout.unref();

    child.once("error", (error) => {
      clearTimeout(timeout);
      reject(error);
    });

    child.once("exit", (exitCode) => {
      clearTimeout(timeout);
      resolve({
        exitCode,
        stdout,
        stderr,
      });
    });
  });
}

function composeCliEnv(options: BuiltCliRunOptions): NodeJS.ProcessEnv {
  const inheritedEnv = Object.fromEntries(
    Object.entries(process.env).filter(
      ([key]) => options.retainInheritedTwentyEnv || !key.startsWith("TWENTY_"),
    ),
  );

  return {
    ...inheritedEnv,
    ...options.env,
  };
}
