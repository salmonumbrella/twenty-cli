"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.capitalize = capitalize;
exports.singularize = singularize;
exports.parsePrimitive = parsePrimitive;
exports.applySet = applySet;
exports.mergeSets = mergeSets;
exports.parseKeyValuePairs = parseKeyValuePairs;
exports.splitOnce = splitOnce;
exports.chunkArray = chunkArray;
exports.parseBooleanEnv = parseBooleanEnv;
const io_1 = require("./io");
function capitalize(value) {
    if (!value)
        return value;
    return value.charAt(0).toUpperCase() + value.slice(1);
}
const irregularSingulars = {
    people: 'person',
    men: 'man',
    women: 'woman',
    children: 'child',
};
function singularize(value) {
    const lower = value.toLowerCase();
    if (irregularSingulars[lower]) {
        return irregularSingulars[lower];
    }
    if (lower.endsWith('ies') && lower.length > 3) {
        return value.slice(0, -3) + 'y';
    }
    if (lower.endsWith('ses') && lower.length > 3) {
        return value.slice(0, -2);
    }
    if (lower.endsWith('s') && !lower.endsWith('ss')) {
        return value.slice(0, -1);
    }
    return value;
}
function parsePrimitive(value) {
    const trimmed = value.trim();
    if (trimmed === '') {
        return '';
    }
    if (trimmed === 'true')
        return true;
    if (trimmed === 'false')
        return false;
    if (trimmed === 'null')
        return null;
    if (!Number.isNaN(Number(trimmed)) && trimmed !== '') {
        return Number(trimmed);
    }
    if (trimmed.startsWith('{') || trimmed.startsWith('[') || trimmed.startsWith('"')) {
        try {
            return (0, io_1.safeJsonParse)(trimmed);
        }
        catch {
            return trimmed;
        }
    }
    return trimmed;
}
function applySet(target, expr) {
    const [rawPath, rawValue] = splitOnce(expr, '=');
    if (!rawPath) {
        throw new Error(`Invalid set expression ${JSON.stringify(expr)} (expected key=value)`);
    }
    const path = rawPath.trim();
    if (!path) {
        throw new Error(`Invalid set expression ${JSON.stringify(expr)} (empty key)`);
    }
    const parts = path.split('.');
    let current = target;
    for (let i = 0; i < parts.length; i += 1) {
        const part = parts[i];
        if (!part) {
            throw new Error(`Invalid set expression ${JSON.stringify(expr)} (empty path segment)`);
        }
        if (i === parts.length - 1) {
            current[part] = parsePrimitive(rawValue ?? '');
            return;
        }
        const next = current[part];
        if (next == null) {
            current[part] = {};
            current = current[part];
            continue;
        }
        if (typeof next !== 'object' || Array.isArray(next)) {
            throw new Error(`Set path ${JSON.stringify(path)} conflicts with non-object value`);
        }
        current = next;
    }
}
function mergeSets(base, sets) {
    const result = { ...base };
    if (!sets)
        return result;
    for (const expr of sets) {
        applySet(result, expr);
    }
    return result;
}
function parseKeyValuePairs(pairs) {
    const out = {};
    if (!pairs)
        return out;
    for (const pair of pairs) {
        const [key, value] = splitOnce(pair, '=');
        if (!key) {
            throw new Error(`Invalid param ${JSON.stringify(pair)} (expected key=value)`);
        }
        if (!out[key]) {
            out[key] = [];
        }
        out[key].push(value ?? '');
    }
    return out;
}
function splitOnce(input, delimiter) {
    const index = input.indexOf(delimiter);
    if (index === -1) {
        return [input, ''];
    }
    return [input.slice(0, index), input.slice(index + delimiter.length)];
}
function chunkArray(items, size) {
    const chunks = [];
    for (let i = 0; i < items.length; i += size) {
        chunks.push(items.slice(i, i + size));
    }
    return chunks;
}
function parseBooleanEnv(value) {
    if (value == null)
        return undefined;
    const normalized = value.toLowerCase();
    if (normalized === 'true' || normalized === '1' || normalized === 'yes')
        return true;
    if (normalized === 'false' || normalized === '0' || normalized === 'no')
        return false;
    return undefined;
}
