"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.OutputService = void 0;
const papaparse_1 = __importDefault(require("papaparse"));
class OutputService {
    constructor(table, queryService) {
        this.table = table;
        this.queryService = queryService;
    }
    async render(data, options) {
        let result = data;
        if (options.query) {
            result = this.queryService.apply(result, options.query);
        }
        const format = options.format ?? 'text';
        switch (format) {
            case 'json':
                // eslint-disable-next-line no-console
                console.log(JSON.stringify(result, null, 2));
                break;
            case 'csv':
                // eslint-disable-next-line no-console
                console.log(this.formatCsv(result));
                break;
            case 'text':
            default:
                this.table.render(result);
                break;
        }
    }
    formatCsv(data) {
        const records = Array.isArray(data) ? data : [data];
        return papaparse_1.default.unparse(records);
    }
}
exports.OutputService = OutputService;
