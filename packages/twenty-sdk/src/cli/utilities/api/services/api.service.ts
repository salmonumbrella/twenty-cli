import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import axiosRetry from 'axios-retry';
import { ConfigService } from '../../config/services/config.service';

export interface ApiServiceOptions {
  workspace?: string;
  debug?: boolean;
  noRetry?: boolean;
}

export class ApiService {
  private client: AxiosInstance;
  private configService: ConfigService;
  private options: ApiServiceOptions;

  constructor(configService: ConfigService, options: ApiServiceOptions = {}) {
    this.configService = configService;
    this.options = options;
    this.client = axios.create();

    if (!options.noRetry) {
      axiosRetry(this.client, {
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

    this.client.interceptors.response.use(
      (response) => {
        if (this.options.debug) {
          // eslint-disable-next-line no-console
          console.error(`← ${response.status} ${response.statusText}`);
        }
        return response;
      },
      (error) => {
        if (this.options.debug) {
          // eslint-disable-next-line no-console
          console.error(`← ${error.response?.status ?? ''} ${error.message}`);
        }
        throw error;
      }
    );
  }

  async get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> {
    return this.client.get<T>(url, config);
  }

  async post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> {
    return this.client.post<T>(url, data, config);
  }

  async patch<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> {
    return this.client.patch<T>(url, data, config);
  }

  async delete<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> {
    return this.client.delete<T>(url, config);
  }

  async request<T = unknown>(config: AxiosRequestConfig): Promise<AxiosResponse<T>> {
    return this.client.request<T>(config);
  }
}
