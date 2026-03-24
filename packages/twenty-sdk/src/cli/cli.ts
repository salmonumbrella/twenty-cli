#!/usr/bin/env node
import { loadCliEnvironment } from "./utilities/config/services/environment.service";
import { formatError, toExitCode } from "./utilities/errors/error-handler";
import { maybeHandleInlineHelp } from "./help";
import { buildProgram } from "./program";

export async function main(argv: string[] = process.argv): Promise<void> {
  const program = buildProgram();

  try {
    if (await maybeHandleInlineHelp(program, argv.slice(2))) {
      return;
    }

    loadCliEnvironment({ argv, cwd: process.cwd() });
    await program.parseAsync(argv);
  } catch (error) {
    const messages = formatError(error);
    for (const line of messages) {
      // eslint-disable-next-line no-console
      console.error(line);
    }
    process.exitCode = toExitCode(error);
  }
}

if (require.main === module) {
  void main();
}
