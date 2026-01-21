"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.readStdin = readStdin;
exports.readFileOrStdin = readFileOrStdin;
exports.safeJsonParse = safeJsonParse;
exports.readJsonInput = readJsonInput;
const fs_extra_1 = __importDefault(require("fs-extra"));
async function readStdin() {
    const chunks = [];
    return new Promise((resolve, reject) => {
        process.stdin.on('data', (chunk) => chunks.push(Buffer.from(chunk)));
        process.stdin.on('end', () => resolve(Buffer.concat(chunks).toString('utf-8')));
        process.stdin.on('error', (err) => reject(err));
    });
}
async function readFileOrStdin(path) {
    if (path === '-') {
        return readStdin();
    }
    return fs_extra_1.default.readFile(path, 'utf-8');
}
function safeJsonParse(input) {
    return JSON.parse(input);
}
async function readJsonInput(data, filePath) {
    if (data && data.trim() !== '') {
        return safeJsonParse(data);
    }
    if (filePath && filePath.trim() !== '') {
        const content = await readFileOrStdin(filePath.trim());
        if (content.trim() === '') {
            return undefined;
        }
        return safeJsonParse(content);
    }
    return undefined;
}
