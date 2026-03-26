import { spawnSync } from "node:child_process";
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
  const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), "twenty-cli-home-"));

  try {
    const inheritedEnv = Object.fromEntries(
      Object.entries(process.env).filter(
        ([key]) => options.retainInheritedTwentyEnv || !key.startsWith("TWENTY_"),
      ),
    );

    const result = spawnSync(process.execPath, [CLI_PATH, ...args], {
      cwd: homeDir,
      env: {
        ...inheritedEnv,
        HOME: homeDir,
        USERPROFILE: homeDir,
        ...options.env,
      },
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
