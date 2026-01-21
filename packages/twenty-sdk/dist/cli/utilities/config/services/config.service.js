"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.ConfigService = void 0;
const os_1 = __importDefault(require("os"));
const path_1 = __importDefault(require("path"));
const fs_extra_1 = __importDefault(require("fs-extra"));
const cli_error_1 = require("../../errors/cli-error");
class ConfigService {
    constructor(configPath) {
        this.configPath = configPath ?? path_1.default.join(os_1.default.homedir(), '.twenty', 'config.json');
    }
    async loadConfigFile() {
        try {
            const exists = await fs_extra_1.default.pathExists(this.configPath);
            if (!exists)
                return null;
            const content = await fs_extra_1.default.readFile(this.configPath, 'utf-8');
            return JSON.parse(content);
        }
        catch (error) {
            throw new cli_error_1.CliError(`Failed to read config at ${this.configPath}`, 'INVALID_ARGUMENTS', 'Check the config file format or remove the file to recreate it.');
        }
    }
    async getConfig(overrides) {
        const fileConfig = await this.loadConfigFile();
        const workspace = overrides?.workspace
            ?? process.env.TWENTY_PROFILE
            ?? fileConfig?.defaultWorkspace
            ?? 'default';
        const workspaceConfig = fileConfig?.workspaces?.[workspace] ?? {};
        const apiUrl = overrides?.apiUrl
            ?? process.env.TWENTY_BASE_URL
            ?? workspaceConfig.apiUrl
            ?? 'https://api.twenty.com';
        const apiKey = overrides?.apiKey
            ?? process.env.TWENTY_TOKEN
            ?? workspaceConfig.apiKey
            ?? '';
        if (!apiKey) {
            throw new cli_error_1.CliError('Missing API token.', 'AUTH', 'Set TWENTY_TOKEN or configure ~/.twenty/config.json with an apiKey.');
        }
        return {
            apiUrl,
            apiKey,
            workspace,
        };
    }
}
exports.ConfigService = ConfigService;
