import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Command } from "commander";
import { registerCalendarChannelsCommand } from "../calendar-channels.command";
import { ApiService } from "../../../utilities/api/services/api.service";
import { CliError } from "../../../utilities/errors/cli-error";
import { mockConstructor } from "../../../test-utils/mock-constructor";

vi.mock("../../../utilities/api/services/api.service");
vi.mock("../../../utilities/config/services/config.service", () => ({
  ConfigService: vi.fn(function MockConfigService() {
    return {
      getConfig: vi.fn().mockResolvedValue({
        apiUrl: "https://api.twenty.com",
        apiKey: "test-token",
        workspace: "default",
      }),
    };
  }),
}));

describe("calendar-channels command", () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;
  let mockGet: ReturnType<typeof vi.fn>;
  let mockPatch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerCalendarChannelsCommand(program);
    consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});
    mockGet = vi.fn();
    mockPatch = vi.fn();
    vi.mocked(ApiService).mockImplementation(
      mockConstructor(
        () =>
          ({
            get: mockGet,
            post: vi.fn(),
            put: vi.fn(),
            patch: mockPatch,
            delete: vi.fn(),
            request: vi.fn(),
          }) as unknown as ApiService,
      ),
    );
  });

  afterEach(() => {
    consoleSpy.mockRestore();
    vi.clearAllMocks();
  });

  describe("command registration", () => {
    it("registers calendar-channels command with correct name and description", () => {
      const cmd = program.commands.find((candidate) => candidate.name() === "calendar-channels");
      expect(cmd).toBeDefined();
      expect(cmd?.description()).toBe("Inspect and update calendar channel settings");
    });

    it("has required operation argument and optional id argument", () => {
      const cmd = program.commands.find((candidate) => candidate.name() === "calendar-channels");
      const args = cmd?.registeredArguments ?? [];

      expect(args.length).toBe(2);
      expect(args[0].name()).toBe("operation");
      expect(args[0].required).toBe(true);
      expect(args[1].name()).toBe("id");
      expect(args[1].required).toBe(false);
    });

    it("has list and update payload options", () => {
      const cmd = program.commands.find((candidate) => candidate.name() === "calendar-channels");
      const opts = cmd?.options ?? [];

      expect(opts.find((option) => option.long === "--limit")).toBeDefined();
      expect(opts.find((option) => option.long === "--cursor")).toBeDefined();
      expect(opts.find((option) => option.long === "--data")).toBeDefined();
      expect(opts.find((option) => option.long === "--file")).toBeDefined();
      expect(opts.find((option) => option.long === "--set")).toBeDefined();
    });
  });

  describe("list operation", () => {
    it("lists calendar channels", async () => {
      mockGet.mockResolvedValue({
        data: {
          data: {
            calendarChannels: [
              {
                id: "cc-1",
                handle: "owner@example.com",
                visibility: "METADATA",
                isSyncEnabled: true,
                syncCursor: "cursor-value",
              },
            ],
          },
        },
      });

      await program.parseAsync([
        "node",
        "test",
        "calendar-channels",
        "list",
        "--limit",
        "5",
        "--cursor",
        "cursor-1",
        "-o",
        "json",
      ]);

      expect(mockGet).toHaveBeenCalledWith("/rest/calendarChannels", {
        params: { limit: "5", starting_after: "cursor-1" },
      });

      const output = consoleSpy.mock.calls[0][0] as string;
      expect(JSON.parse(output)).toEqual([
        {
          id: "cc-1",
          handle: "owner@example.com",
          visibility: "METADATA",
          isSyncEnabled: true,
          syncCursor: "[hidden]",
        },
      ]);
    });
  });

  describe("get operation", () => {
    it("gets one calendar channel by id", async () => {
      mockGet.mockResolvedValue({
        data: {
          data: {
            calendarChannel: {
              id: "cc-1",
              handle: "owner@example.com",
              contactAutoCreationPolicy: "SENT_EMAILS",
              syncCursor: "cursor-value",
            },
          },
        },
      });

      await program.parseAsync(["node", "test", "calendar-channels", "get", "cc-1", "-o", "json"]);

      expect(mockGet).toHaveBeenCalledWith("/rest/calendarChannels/cc-1", { params: {} });
      const output = consoleSpy.mock.calls[0][0] as string;
      expect(JSON.parse(output)).toEqual({
        id: "cc-1",
        handle: "owner@example.com",
        contactAutoCreationPolicy: "SENT_EMAILS",
        syncCursor: "[hidden]",
      });
    });

    it("throws when the id is missing for get", async () => {
      await expect(
        program.parseAsync(["node", "test", "calendar-channels", "get"]),
      ).rejects.toThrow(CliError);
    });
  });

  describe("update operation", () => {
    it("updates one calendar channel with a JSON payload", async () => {
      mockPatch.mockResolvedValue({
        data: {
          data: {
            updateCalendarChannel: {
              id: "cc-1",
              isSyncEnabled: false,
              visibility: "NOTHING",
            },
          },
        },
      });

      await program.parseAsync([
        "node",
        "test",
        "calendar-channels",
        "update",
        "cc-1",
        "-d",
        '{"isSyncEnabled":false,"visibility":"NOTHING"}',
        "-o",
        "json",
      ]);

      expect(mockPatch).toHaveBeenCalledWith("/rest/calendarChannels/cc-1", {
        isSyncEnabled: false,
        visibility: "NOTHING",
      });

      const output = consoleSpy.mock.calls[0][0] as string;
      expect(JSON.parse(output)).toEqual({
        id: "cc-1",
        isSyncEnabled: false,
        visibility: "NOTHING",
      });
    });

    it("throws when the id is missing for update", async () => {
      await expect(
        program.parseAsync([
          "node",
          "test",
          "calendar-channels",
          "update",
          "-d",
          '{"isSyncEnabled":false}',
        ]),
      ).rejects.toThrow(CliError);
    });
  });

  describe("unknown operations", () => {
    it("throws for unknown operations", async () => {
      await expect(
        program.parseAsync(["node", "test", "calendar-channels", "explode"]),
      ).rejects.toThrow(CliError);
    });
  });
});
