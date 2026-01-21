"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.ImportService = void 0;
const papaparse_1 = __importDefault(require("papaparse"));
const fs_extra_1 = __importDefault(require("fs-extra"));
const path_1 = __importDefault(require("path"));
class ImportService {
    async import(filePath, options) {
        const content = await fs_extra_1.default.readFile(filePath, 'utf-8');
        const ext = path_1.default.extname(filePath).toLowerCase();
        let records = [];
        if (ext === '.csv') {
            const result = papaparse_1.default.parse(content, {
                header: true,
                skipEmptyLines: true,
                transformHeader: (header) => header.trim(),
            });
            records = result.data;
        }
        else if (ext === '.json') {
            const parsed = JSON.parse(content);
            records = Array.isArray(parsed) ? parsed : [parsed];
        }
        else {
            throw new Error(`Unsupported file format: ${ext}`);
        }
        if (options?.dryRun) {
            // eslint-disable-next-line no-console
            console.log(`Would import ${records.length} records`);
            if (records[0]) {
                // eslint-disable-next-line no-console
                console.log('First record:', JSON.stringify(records[0], null, 2));
            }
        }
        return records;
    }
}
exports.ImportService = ImportService;
