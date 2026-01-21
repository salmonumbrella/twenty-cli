import { describe, it, expect, vi } from 'vitest';
import { runFindDuplicatesOperation } from '../find-duplicates.operation';
import { CliError } from '../../../../utilities/errors/cli-error';

describe('runFindDuplicatesOperation', () => {
  it('throws helpful error when no payload provided', async () => {
    const ctx = {
      object: 'people',
      options: {},
      globalOptions: { output: 'json' },
      services: {
        records: { findDuplicates: vi.fn() },
        output: { render: vi.fn() },
      },
    };

    await expect(runFindDuplicatesOperation(ctx as any)).rejects.toThrow(CliError);

    // Verify the error has a helpful suggestion with example
    try {
      await runFindDuplicatesOperation(ctx as any);
    } catch (err) {
      expect(err).toBeInstanceOf(CliError);
      const cliErr = err as CliError;
      expect(cliErr.code).toBe('INVALID_ARGUMENTS');
      expect(cliErr.suggestion).toMatch(/--fields/);
      expect(cliErr.suggestion).toMatch(/email,phone/);  // example included
    }
  });

  it('parses comma-separated fields correctly', async () => {
    const mockResponse = { data: { duplicates: [] } };
    const ctx = {
      object: 'people',
      options: { fields: 'email, phone, name' },
      globalOptions: { output: 'json' },
      services: {
        records: {
          findDuplicates: vi.fn().mockResolvedValue(mockResponse),
        },
        output: { render: vi.fn() },
      },
    };

    await runFindDuplicatesOperation(ctx as any);

    expect(ctx.services.records.findDuplicates).toHaveBeenCalledWith(
      'people',
      { fields: ['email', 'phone', 'name'] }
    );
  });
});
