import { ApiService } from "../api/services/api.service";
import { PublicHttpService } from "../api/services/public-http.service";
import { ConfigService } from "../config/services/config.service";
import { MetadataService } from "../metadata/services/metadata.service";
import { RecordsService } from "../records/services/records.service";
import { OutputService } from "../output/services/output.service";
import { QueryService } from "../output/services/query.service";
import { TableService } from "../output/services/table.service";
import { ExportService } from "../file/services/export.service";
import { ImportService } from "../file/services/import.service";
import { McpService } from "../mcp/services/mcp.service";
import { SearchService } from "../search/services/search.service";
import { ApiSearchService } from "../search/services/api-search.service";
import { DbConfigResolverService } from "../db/services/db-config-resolver.service";
import { DbRecordsReadService } from "../db/services/db-records-read.service";
import { DbSearchService } from "../db/services/db-search.service";
import { DbProfileService } from "../db/services/db-profile.service";
import { DbStatusService } from "../db/services/db-status.service";
import { ReadBackendService } from "../readbackend/read-backend.service";
import { ApiRecordsReadService } from "../records/services/api-records-read.service";
import { GlobalOptions } from "./global-options";

export interface CliServices {
  config: ConfigService;
  dbProfiles: DbProfileService;
  dbStatus: DbStatusService;
  api: ApiService;
  publicHttp: PublicHttpService;
  search: SearchService;
  mcp: McpService;
  records: RecordsService;
  metadata: MetadataService;
  output: OutputService;
  importer: ImportService;
  exporter: ExportService;
}

export function createOutputService(globalOptions: GlobalOptions): OutputService {
  return new OutputService(new TableService(), new QueryService(), {
    kind: globalOptions.outputKind,
  });
}

export function createServices(globalOptions: GlobalOptions): CliServices {
  const config = new ConfigService();
  const dbProfiles = new DbProfileService(config);
  const dbConfigResolver = new DbConfigResolverService(dbProfiles);
  const dbStatus = new DbStatusService(dbConfigResolver);
  const api = new ApiService(config, {
    workspace: globalOptions.workspace,
    debug: globalOptions.debug,
    noRetry: globalOptions.noRetry,
  });
  const publicHttp = new PublicHttpService(config, {
    workspace: globalOptions.workspace,
    debug: globalOptions.debug,
    noRetry: globalOptions.noRetry,
  });
  const metadata = new MetadataService(api);
  const apiSearch = new ApiSearchService(api);
  const apiRecordsRead = new ApiRecordsReadService(api);
  const readBackend = new ReadBackendService(
    dbConfigResolver,
    apiSearch,
    apiRecordsRead,
    {
      search: new DbSearchService(metadata),
      records: new DbRecordsReadService(metadata),
    },
    { workspace: globalOptions.workspace },
  );
  const search = new SearchService(api, readBackend);
  const mcp = new McpService(api, config, {
    workspace: globalOptions.workspace,
    debug: globalOptions.debug,
  });
  const records = new RecordsService(api, { readBackend });
  const output = createOutputService(globalOptions);
  const importer = new ImportService();
  const exporter = new ExportService();

  return {
    config,
    dbProfiles,
    dbStatus,
    api,
    publicHttp,
    search,
    mcp,
    records,
    metadata,
    output,
    importer,
    exporter,
  };
}
