import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { OutputService } from '../output.service';
import { QueryService } from '../query.service';
import { TableService } from '../table.service';

describe('OutputService', () => {
  let outputService: OutputService;
  let consoleSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    const queryService = new QueryService();
    const tableService = new TableService();
    outputService = new OutputService(tableService, queryService);
    consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
  });

  afterEach(() => {
    consoleSpy.mockRestore();
  });

  describe('CSV output with nested objects', () => {
    it('serializes nested objects to JSON strings', async () => {
      const data = [
        {
          id: '1',
          name: { firstName: 'John', lastName: 'Doe' },
          emails: { primaryEmail: 'john@example.com', additionalEmails: [] },
        },
      ];

      await outputService.render(data, { format: 'csv' });

      const output = consoleSpy.mock.calls[0][0];
      expect(output).toContain('id,name,emails');
      // CSV escapes inner quotes by doubling them: "" inside quoted field
      expect(output).toContain('"{""firstName"":""John"",""lastName"":""Doe""}"');
      expect(output).not.toContain('[object Object]');
    });

    it('handles arrays in CSV output', async () => {
      const data = [{ id: '1', tags: ['a', 'b', 'c'] }];

      await outputService.render(data, { format: 'csv' });

      const output = consoleSpy.mock.calls[0][0];
      // CSV escapes inner quotes by doubling them: "" inside quoted field
      expect(output).toContain('"[""a"",""b"",""c""]"');
      expect(output).not.toContain('[object Object]');
    });

    it('handles null and primitive values correctly', async () => {
      const data = [{ id: '1', name: 'Test', count: 42, active: true, deleted: null }];

      await outputService.render(data, { format: 'csv' });

      const output = consoleSpy.mock.calls[0][0];
      expect(output).toContain('Test');
      expect(output).toContain('42');
      expect(output).toContain('true');
    });
  });
});
