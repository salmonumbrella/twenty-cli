"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.CliError = void 0;
exports.errorWithCause = errorWithCause;
class CliError extends Error {
    constructor(message, code, suggestion) {
        super(message);
        this.code = code;
        this.suggestion = suggestion;
    }
}
exports.CliError = CliError;
function errorWithCause(message, code, suggestion, cause) {
    const err = new CliError(message, code, suggestion);
    if (cause) {
        err.cause = cause;
    }
    return err;
}
