"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.createServices = createServices;
const api_service_1 = require("../api/services/api.service");
const config_service_1 = require("../config/services/config.service");
const metadata_service_1 = require("../metadata/services/metadata.service");
const records_service_1 = require("../records/services/records.service");
const output_service_1 = require("../output/services/output.service");
const query_service_1 = require("../output/services/query.service");
const table_service_1 = require("../output/services/table.service");
const export_service_1 = require("../file/services/export.service");
const import_service_1 = require("../file/services/import.service");
function createServices(globalOptions) {
    const configService = new config_service_1.ConfigService();
    const api = new api_service_1.ApiService(configService, {
        workspace: globalOptions.workspace,
        debug: globalOptions.debug,
        noRetry: globalOptions.noRetry,
    });
    const records = new records_service_1.RecordsService(api);
    const metadata = new metadata_service_1.MetadataService(api);
    const output = new output_service_1.OutputService(new table_service_1.TableService(), new query_service_1.QueryService());
    const importer = new import_service_1.ImportService();
    const exporter = new export_service_1.ExportService();
    return { api, records, metadata, output, importer, exporter };
}
