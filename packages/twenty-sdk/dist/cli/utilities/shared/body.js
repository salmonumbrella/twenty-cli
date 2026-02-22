"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.parseBody = parseBody;
exports.parseArrayPayload = parseArrayPayload;
const io_1 = require("./io");
const parse_1 = require("./parse");
async function parseBody(data, filePath, sets) {
    const payload = await (0, io_1.readJsonInput)(data, filePath);
    let base = {};
    if (payload != null) {
        if (typeof payload !== 'object' || Array.isArray(payload)) {
            throw new Error('Payload must be a JSON object');
        }
        base = payload;
    }
    const merged = (0, parse_1.mergeSets)(base, sets);
    if (payload == null && (!sets || sets.length === 0)) {
        throw new Error('Missing JSON payload; use --data, --file, or --set');
    }
    return merged;
}
async function parseArrayPayload(data, filePath) {
    const payload = await (0, io_1.readJsonInput)(data, filePath);
    if (payload == null) {
        throw new Error('Missing JSON payload; use --data or --file');
    }
    if (!Array.isArray(payload)) {
        throw new Error('Batch payload must be a JSON array');
    }
    return payload;
}
