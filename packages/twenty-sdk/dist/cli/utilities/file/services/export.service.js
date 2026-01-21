"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.ExportService = void 0;
const papaparse_1 = __importDefault(require("papaparse"));
const fs_extra_1 = __importDefault(require("fs-extra"));
class ExportService {
    async export(records, options) {
        let content;
        if (options.format === 'csv') {
            content = papaparse_1.default.unparse(records);
        }
        else {
            content = JSON.stringify(records, null, 2);
        }
        if (options.output) {
            await fs_extra_1.default.writeFile(options.output, content);
            // eslint-disable-next-line no-console
            console.error(`Exported ${records.length} records to ${options.output}`);
        }
        else {
            // eslint-disable-next-line no-console
            console.log(content);
        }
    }
}
exports.ExportService = ExportService;
