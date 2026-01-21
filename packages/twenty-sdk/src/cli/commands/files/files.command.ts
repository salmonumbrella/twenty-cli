import { Command } from 'commander';
import { applyGlobalOptions, resolveGlobalOptions } from '../../utilities/shared/global-options';
import { createServices } from '../../utilities/shared/services';
import { CliError } from '../../utilities/errors/cli-error';
import fs from 'fs-extra';
import path from 'path';
import FormData from 'form-data';

export function registerFilesCommand(program: Command): void {
  const cmd = program
    .command('files')
    .description('Manage file attachments')
    .argument('<operation>', 'upload, download, delete')
    .argument('[path-or-id]', 'File path (upload) or file ID (download/delete)')
    .option('--output-file <path>', 'Output file path (download)');

  applyGlobalOptions(cmd);

  cmd.action(async (operation: string, pathOrId: string | undefined, options: FilesOptions, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const op = operation.toLowerCase();

    switch (op) {
      case 'upload': {
        if (!pathOrId) throw new CliError('Missing file path.', 'INVALID_ARGUMENTS');
        const filePath = pathOrId;
        if (!await fs.pathExists(filePath)) {
          throw new CliError(`File not found: ${filePath}`, 'INVALID_ARGUMENTS');
        }
        const form = new FormData();
        form.append('file', fs.createReadStream(filePath), path.basename(filePath));
        const response = await services.api.post('/files', form, {
          headers: form.getHeaders(),
        });
        await services.output.render(response.data, { format: globalOptions.output, query: globalOptions.query });
        break;
      }
      case 'download': {
        if (!pathOrId) throw new CliError('Missing file ID or URL.', 'INVALID_ARGUMENTS');
        const outputPath = options.outputFile || path.basename(pathOrId);
        const response = await services.api.get<ArrayBuffer>(`/files/${pathOrId}`, { responseType: 'arraybuffer' });
        await fs.writeFile(outputPath, Buffer.from(response.data));
        // eslint-disable-next-line no-console
        console.log(`Downloaded to ${outputPath}`);
        break;
      }
      case 'delete': {
        if (!pathOrId) throw new CliError('Missing file ID.', 'INVALID_ARGUMENTS');
        await services.api.post('/graphql', {
          query: `mutation($id: String!) { deleteFile(fileId: $id) }`,
          variables: { id: pathOrId },
        });
        // eslint-disable-next-line no-console
        console.log(`File ${pathOrId} deleted.`);
        break;
      }
      default:
        throw new CliError(`Unknown operation: ${operation}`, 'INVALID_ARGUMENTS');
    }
  });
}

interface FilesOptions {
  outputFile?: string;
}
