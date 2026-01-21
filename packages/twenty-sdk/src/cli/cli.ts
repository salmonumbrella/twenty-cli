#!/usr/bin/env node
import { Command } from 'commander';
import { registerApiCommand } from './commands/api/api.command';
import { registerApiMetadataCommand } from './commands/api-metadata/api-metadata.command';
import { registerRestCommand } from './commands/raw/rest.command';
import { registerGraphqlCommand } from './commands/raw/graphql.command';
import { registerAuthCommand } from './commands/auth/auth.command';
import { registerSearchCommand } from './commands/search/search.command';
import { registerWebhooksCommand } from './commands/webhooks/webhooks.command';
import { registerApiKeysCommand } from './commands/api-keys/api-keys.command';
import { registerFilesCommand } from './commands/files/files.command';
import { formatError, toExitCode } from './utilities/errors/error-handler';

async function main(): Promise<void> {
  const program = new Command();
  program.name('twenty');
  program.description('Twenty CLI (TypeScript port)');
  program.exitOverride();

  registerApiCommand(program);
  registerApiMetadataCommand(program);
  registerRestCommand(program);
  registerGraphqlCommand(program);
  registerAuthCommand(program);
  registerSearchCommand(program);
  registerWebhooksCommand(program);
  registerApiKeysCommand(program);
  registerFilesCommand(program);

  try {
    await program.parseAsync(process.argv);
  } catch (error) {
    const messages = formatError(error);
    for (const line of messages) {
      // eslint-disable-next-line no-console
      console.error(line);
    }
    process.exitCode = toExitCode(error);
  }
}

main();
