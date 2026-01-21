import { ApiService } from '../api/services/api.service';
import { ConfigService } from '../config/services/config.service';
import { MetadataService } from '../metadata/services/metadata.service';
import { RecordsService } from '../records/services/records.service';
import { OutputService } from '../output/services/output.service';
import { QueryService } from '../output/services/query.service';
import { TableService } from '../output/services/table.service';
import { ExportService } from '../file/services/export.service';
import { ImportService } from '../file/services/import.service';
import { GlobalOptions } from './global-options';

export interface CliServices {
  api: ApiService;
  records: RecordsService;
  metadata: MetadataService;
  output: OutputService;
  importer: ImportService;
  exporter: ExportService;
}

export function createServices(globalOptions: GlobalOptions): CliServices {
  const configService = new ConfigService();
  const api = new ApiService(configService, {
    workspace: globalOptions.workspace,
    debug: globalOptions.debug,
    noRetry: globalOptions.noRetry,
  });
  const records = new RecordsService(api);
  const metadata = new MetadataService(api);
  const output = new OutputService(new TableService(), new QueryService());
  const importer = new ImportService();
  const exporter = new ExportService();

  return { api, records, metadata, output, importer, exporter };
}
