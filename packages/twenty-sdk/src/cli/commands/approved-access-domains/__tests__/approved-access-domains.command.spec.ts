import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Command } from "commander";
import { registerApprovedAccessDomainsCommand } from "../approved-access-domains.command";
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

describe("approved-access-domains command", () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;
  let mockPost: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerApprovedAccessDomainsCommand(program);
    consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});
    mockPost = vi.fn();
    vi.mocked(ApiService).mockImplementation(
      mockConstructor(
        () =>
          ({
            post: mockPost,
            get: vi.fn(),
            put: vi.fn(),
            patch: vi.fn(),
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
    it("registers approved-access-domains command with correct name and description", () => {
      const cmd = program.commands.find(
        (candidate) => candidate.name() === "approved-access-domains",
      );
      expect(cmd).toBeDefined();
      expect(cmd?.description()).toBe("Manage approved access domains");
    });

    it("has required operation argument and optional id argument", () => {
      const cmd = program.commands.find(
        (candidate) => candidate.name() === "approved-access-domains",
      );
      const args = cmd?.registeredArguments ?? [];

      expect(args.length).toBe(2);
      expect(args[0].name()).toBe("operation");
      expect(args[0].required).toBe(true);
      expect(args[1].name()).toBe("id");
      expect(args[1].required).toBe(false);
    });

    it("has validate options", () => {
      const cmd = program.commands.find(
        (candidate) => candidate.name() === "approved-access-domains",
      );
      const opts = cmd?.options ?? [];

      expect(opts.find((option) => option.long === "--validation-token")).toBeDefined();
    });
  });

  describe("list operation", () => {
    it("lists approved access domains", async () => {
      mockPost.mockResolvedValue({
        data: {
          data: {
            getApprovedAccessDomains: [
              {
                id: "domain-1",
                domain: "acme.com",
                isValidated: true,
                createdAt: "2026-03-21T00:00:00.000Z",
              },
            ],
          },
        },
      });

      await program.parseAsync(["node", "test", "approved-access-domains", "list", "-o", "json"]);

      expect(mockPost).toHaveBeenCalledWith("/graphql", {
        query: expect.stringContaining("getApprovedAccessDomains"),
      });

      const output = consoleSpy.mock.calls[0][0] as string;
      expect(JSON.parse(output)).toEqual([
        {
          id: "domain-1",
          domain: "acme.com",
          isValidated: true,
          createdAt: "2026-03-21T00:00:00.000Z",
        },
      ]);
    });
  });

  describe("delete operation", () => {
    it("deletes an approved access domain by id", async () => {
      mockPost.mockResolvedValue({
        data: {
          data: {
            deleteApprovedAccessDomain: true,
          },
        },
      });

      await program.parseAsync([
        "node",
        "test",
        "approved-access-domains",
        "delete",
        "domain-1",
        "-o",
        "json",
      ]);

      expect(mockPost).toHaveBeenCalledWith("/graphql", {
        query: expect.stringContaining("deleteApprovedAccessDomain"),
        variables: {
          input: {
            id: "domain-1",
          },
        },
      });

      const output = consoleSpy.mock.calls[0][0] as string;
      expect(JSON.parse(output)).toEqual({
        success: true,
        id: "domain-1",
      });
    });

    it("throws when id is missing for delete", async () => {
      await expect(
        program.parseAsync(["node", "test", "approved-access-domains", "delete"]),
      ).rejects.toThrow(CliError);
    });
  });

  describe("validate operation", () => {
    it("validates an approved access domain by id and token", async () => {
      mockPost.mockResolvedValue({
        data: {
          data: {
            validateApprovedAccessDomain: {
              id: "domain-1",
              domain: "acme.com",
              isValidated: true,
              createdAt: "2026-03-21T00:00:00.000Z",
            },
          },
        },
      });

      await program.parseAsync([
        "node",
        "test",
        "approved-access-domains",
        "validate",
        "domain-1",
        "--validation-token",
        "token-1",
        "-o",
        "json",
      ]);

      expect(mockPost).toHaveBeenCalledWith("/graphql", {
        query: expect.stringContaining("validateApprovedAccessDomain"),
        variables: {
          input: {
            approvedAccessDomainId: "domain-1",
            validationToken: "token-1",
          },
        },
      });

      const output = consoleSpy.mock.calls[0][0] as string;
      expect(JSON.parse(output)).toEqual({
        id: "domain-1",
        domain: "acme.com",
        isValidated: true,
        createdAt: "2026-03-21T00:00:00.000Z",
      });
    });

    it("throws when id or validation token is missing", async () => {
      await expect(
        program.parseAsync(["node", "test", "approved-access-domains", "validate", "domain-1"]),
      ).rejects.toThrow(CliError);
    });
  });

  describe("unknown operations", () => {
    it("throws for unknown operations", async () => {
      await expect(
        program.parseAsync(["node", "test", "approved-access-domains", "explode"]),
      ).rejects.toThrow(CliError);
    });
  });
});
