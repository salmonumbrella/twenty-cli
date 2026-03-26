import { spawn, spawnSync } from "node:child_process";
import path from "node:path";

const DEFAULT_TIMEOUT_MS = 10_000;
const MAX_BUFFER = 20 * 1024 * 1024;

export interface BuiltCliRunOptions {
  cwd?: string;
  env?: NodeJS.ProcessEnv;
  retainInheritedTwentyEnv?: boolean;
  timeoutMs?: number;
  throwOnNonZeroExit?: boolean;
}

export interface BuiltCliRunResult {
  exitCode: number | null;
  stdout: string;
  stderr: string;
}

export class BuiltCliRunError extends Error {
  constructor(
    readonly args: string[],
    readonly result: BuiltCliRunResult,
  ) {
    super(formatRunFailure(args, result));
    this.name = "BuiltCliRunError";
  }
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

  const cliResult = {
    exitCode: result.status,
    stdout: result.stdout ?? "",
    stderr: result.stderr ?? "",
  };

  return options.throwOnNonZeroExit ? assertSuccessfulBuiltCliRun(args, cliResult) : cliResult;
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
    let exitCode: number | null = null;
    let timedOut = false;

    const timeout = setTimeout(() => {
      timedOut = true;
      child.kill("SIGKILL");
    }, options.timeoutMs ?? DEFAULT_TIMEOUT_MS);
    timeout.unref();

    child.once("error", (error) => {
      clearTimeout(timeout);
      reject(error);
    });

    child.once("exit", (code) => {
      exitCode = code;
    });

    child.once("close", (closeCode) => {
      clearTimeout(timeout);
      const cliResult = {
        exitCode: closeCode ?? (timedOut ? null : exitCode),
        stdout,
        stderr,
      };

      if (options.throwOnNonZeroExit) {
        try {
          resolve(assertSuccessfulBuiltCliRun(args, cliResult));
        } catch (error) {
          reject(error);
        }
        return;
      }

      resolve(cliResult);
    });
  });
}

export function assertSuccessfulBuiltCliRun(
  args: string[],
  result: BuiltCliRunResult,
): BuiltCliRunResult {
  if (result.exitCode === 0) {
    return result;
  }

  throw new BuiltCliRunError(args, result);
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

function formatRunFailure(args: string[], result: BuiltCliRunResult): string {
  return [
    `Built CLI exited with code ${result.exitCode ?? "null"}`,
    `args: ${args.join(" ")}`,
    `stderr:\n${result.stderr || "(empty)"}`,
    `stdout:\n${result.stdout || "(empty)"}`,
  ].join("\n");
}
