"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.ApiService = void 0;
const axios_1 = __importDefault(require("axios"));
const axios_retry_1 = __importDefault(require("axios-retry"));
class ApiService {
    constructor(configService, options = {}) {
        this.configService = configService;
        this.options = options;
        this.client = axios_1.default.create();
        if (!options.noRetry) {
            (0, axios_retry_1.default)(this.client, {
                retries: 3,
                retryDelay: (retryCount, error) => {
                    const retryAfter = error.response?.headers?.['retry-after'];
                    if (retryAfter) {
                        const seconds = Number.parseInt(String(retryAfter), 10);
                        if (!Number.isNaN(seconds)) {
                            return seconds * 1000;
                        }
                    }
                    const baseDelay = Math.pow(2, retryCount) * 1000;
                    const jitter = Math.random() * 1000;
                    return baseDelay + jitter;
                },
                retryCondition: (error) => {
                    const status = error.response?.status;
                    return status === 429 || status === 502 || status === 503 || status === 504;
                },
                onRetry: (retryCount, error) => {
                    if (this.options.debug) {
                        // eslint-disable-next-line no-console
                        console.error(`Retry ${retryCount}: ${error.message}`);
                    }
                },
            });
        }
        this.client.interceptors.request.use(async (config) => {
            const resolved = await this.configService.getConfig({
                workspace: this.options.workspace,
            });
            config.baseURL = resolved.apiUrl;
            config.headers = config.headers ?? {};
            config.headers.Authorization = `Bearer ${resolved.apiKey}`;
            if (this.options.debug) {
                const url = `${config.baseURL ?? ''}${config.url ?? ''}`;
                // eslint-disable-next-line no-console
                console.error(`→ ${config.method?.toUpperCase()} ${url}`);
                if (config.data) {
                    const preview = JSON.stringify(config.data).slice(0, 500);
                    // eslint-disable-next-line no-console
                    console.error(`  Body: ${preview}`);
                }
            }
            return config;
        });
        this.client.interceptors.response.use((response) => {
            if (this.options.debug) {
                // eslint-disable-next-line no-console
                console.error(`← ${response.status} ${response.statusText}`);
            }
            return response;
        }, (error) => {
            if (this.options.debug) {
                // eslint-disable-next-line no-console
                console.error(`← ${error.response?.status ?? ''} ${error.message}`);
            }
            throw error;
        });
    }
    async get(url, config) {
        return this.client.get(url, config);
    }
    async post(url, data, config) {
        return this.client.post(url, data, config);
    }
    async patch(url, data, config) {
        return this.client.patch(url, data, config);
    }
    async delete(url, config) {
        return this.client.delete(url, config);
    }
    async request(config) {
        return this.client.request(config);
    }
}
exports.ApiService = ApiService;
