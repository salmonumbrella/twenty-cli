"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.TableService = void 0;
class TableService {
    render(data) {
        const records = normalizeRecords(data);
        if (records.length === 0) {
            // eslint-disable-next-line no-console
            console.log('No records found.');
            return;
        }
        if (records.length === 1 && !isRecord(records[0])) {
            // eslint-disable-next-line no-console
            console.log(String(records[0]));
            return;
        }
        const rows = records.map((record) => (isRecord(record) ? record : { value: record }));
        const columns = extractColumns(rows[0]);
        const widths = calculateWidths(columns, rows);
        // eslint-disable-next-line no-console
        console.log(columns.map((col, i) => col.toUpperCase().padEnd(widths[i])).join('  '));
        for (const record of rows) {
            const row = columns.map((col, i) => {
                const value = getValue(record, col);
                const cell = formatValue(value).slice(0, widths[i]);
                return cell.padEnd(widths[i]);
            });
            // eslint-disable-next-line no-console
            console.log(row.join('  '));
        }
    }
}
exports.TableService = TableService;
function normalizeRecords(data) {
    if (Array.isArray(data))
        return data;
    if (data == null)
        return [];
    return [data];
}
function isRecord(value) {
    return typeof value === 'object' && value !== null && !Array.isArray(value);
}
function extractColumns(record) {
    const priority = ['id', 'name', 'email', 'title', 'status', 'createdAt'];
    const keys = Object.keys(record);
    return [
        ...priority.filter((k) => keys.includes(k)),
        ...keys.filter((k) => !priority.includes(k)).sort(),
    ];
}
function calculateWidths(columns, records) {
    return columns.map((column) => {
        const maxCell = records.reduce((max, record) => {
            const value = formatValue(getValue(record, column));
            return Math.max(max, value.length);
        }, column.length);
        return Math.min(Math.max(maxCell, column.length), 60);
    });
}
function getValue(record, path) {
    return path.split('.').reduce((obj, key) => {
        if (obj && typeof obj === 'object' && !Array.isArray(obj)) {
            return obj[key];
        }
        return undefined;
    }, record);
}
function formatValue(value) {
    if (value == null)
        return '';
    if (typeof value === 'string')
        return value;
    if (typeof value === 'number' || typeof value === 'boolean')
        return String(value);
    try {
        return JSON.stringify(value);
    }
    catch {
        return String(value);
    }
}
