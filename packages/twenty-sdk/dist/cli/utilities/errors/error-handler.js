"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.toExitCode = toExitCode;
exports.formatError = formatError;
const cli_error_1 = require("./cli-error");
function toExitCode(error) {
    if (isCommanderError(error)) {
        const code = String(error.code ?? '');
        if (code.startsWith('commander.help') || code === 'commander.version') {
            return typeof error.exitCode === 'number' ? error.exitCode : 0;
        }
        if (typeof error.exitCode === 'number') {
            return error.exitCode;
        }
    }
    if (error instanceof cli_error_1.CliError) {
        switch (error.code) {
            case 'INVALID_ARGUMENTS':
                return 2;
            case 'AUTH':
                return 3;
            case 'NETWORK':
                return 4;
            case 'RATE_LIMIT':
                return 5;
            default:
                return 1;
        }
    }
    if (isCommanderError(error)) {
        return 2;
    }
    if (isAxiosError(error)) {
        const status = error.response?.status;
        if (status === 401 || status === 403) {
            return 3;
        }
        if (status === 429) {
            return 5;
        }
        if (!status) {
            return 4;
        }
        return 1;
    }
    return 1;
}
function formatError(error) {
    if (isCommanderError(error)) {
        const code = String(error.code ?? '');
        if (code.startsWith('commander.help') || code === 'commander.version') {
            return [];
        }
    }
    if (error instanceof cli_error_1.CliError) {
        const lines = [error.message];
        if (error.suggestion) {
            lines.push(`Suggestion: ${error.suggestion}`);
        }
        return lines;
    }
    if (isCommanderError(error)) {
        return [error.message];
    }
    if (isAxiosError(error)) {
        const status = error.response?.status;
        if (status) {
            const detail = typeof error.response?.data === 'string'
                ? error.response?.data
                : JSON.stringify(error.response?.data ?? {}, null, 2);
            return [`Request failed with status ${status}.`, detail].filter(Boolean);
        }
        return [`Network error: ${error.message}`];
    }
    if (error instanceof Error) {
        return [error.message];
    }
    return ['Unknown error'];
}
function isCommanderError(error) {
    return typeof error === 'object' && error !== null && 'code' in error && 'message' in error;
}
function isAxiosError(error) {
    return !!error && typeof error === 'object' && error.isAxiosError === true;
}
