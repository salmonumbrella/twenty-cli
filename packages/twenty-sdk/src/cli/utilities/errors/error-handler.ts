import { AxiosError } from 'axios';
import { CliError } from './cli-error';

export function toExitCode(error: unknown): number {
  if (isCommanderError(error)) {
    const code = String(error.code ?? '');
    if (code.startsWith('commander.help') || code === 'commander.version') {
      return typeof (error as any).exitCode === 'number' ? (error as any).exitCode : 0;
    }
    if (typeof (error as any).exitCode === 'number') {
      return (error as any).exitCode;
    }
  }

  if (error instanceof CliError) {
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

export function formatError(error: unknown): string[] {
  if (isCommanderError(error)) {
    const code = String(error.code ?? '');
    if (code.startsWith('commander.help') || code === 'commander.version') {
      return [];
    }
  }

  if (error instanceof CliError) {
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
      return [`Request failed with status ${status}.`, detail].filter(Boolean) as string[];
    }
    return [`Network error: ${error.message}`];
  }

  if (error instanceof Error) {
    return [error.message];
  }

  return ['Unknown error'];
}

function isCommanderError(error: unknown): error is { message: string; code?: string } {
  return typeof error === 'object' && error !== null && 'code' in error && 'message' in error;
}

function isAxiosError(error: unknown): error is AxiosError {
  return !!error && typeof error === 'object' && (error as AxiosError).isAxiosError === true;
}
