import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { Command } from "commander";
import { registerApiKeysCommand } from "../api-keys.command";
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

describe("api-keys command", () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;
  let mockPost: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerApiKeysCommand(program);
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
    it("registers api-keys command with correct name and description", () => {
      const apiKeysCmd = program.commands.find((cmd) => cmd.name() === "api-keys");
      expect(apiKeysCmd).toBeDefined();
      expect(apiKeysCmd?.description()).toBe("Manage API keys");
    });

    it("has required operation argument", () => {
      const apiKeysCmd = program.commands.find((cmd) => cmd.name() === "api-keys");
      const args = apiKeysCmd?.registeredArguments ?? [];
      expect(args.length).toBe(2);
      expect(args[0].name()).toBe("operation");
      expect(args[0].required).toBe(true);
    });

    it("has optional id argument", () => {
      const apiKeysCmd = program.commands.find((cmd) => cmd.name() === "api-keys");
      const args = apiKeysCmd?.registeredArguments ?? [];
      expect(args[1].name()).toBe("id");
      expect(args[1].required).toBe(false);
    });

    it("has --name option", () => {
      const apiKeysCmd = program.commands.find((cmd) => cmd.name() === "api-keys");
      const opts = apiKeysCmd?.options ?? [];
      const nameOpt = opts.find((o) => o.long === "--name");
      expect(nameOpt).toBeDefined();
    });

    it("has --expires-at option", () => {
      const apiKeysCmd = program.commands.find((cmd) => cmd.name() === "api-keys");
      const opts = apiKeysCmd?.options ?? [];
      const expiresAtOpt = opts.find((o) => o.long === "--expires-at");
      expect(expiresAtOpt).toBeDefined();
    });

    it("has global options applied", () => {
      const apiKeysCmd = program.commands.find((cmd) => cmd.name() === "api-keys");
      const opts = apiKeysCmd?.options ?? [];
      const outputOpt = opts.find((o) => o.long === "--output");
      const queryOpt = opts.find((o) => o.long === "--query");
      const workspaceOpt = opts.find((o) => o.long === "--workspace");
      expect(outputOpt).toBeDefined();
      expect(queryOpt).toBeDefined();
      expect(workspaceOpt).toBeDefined();
    });
  });

  describe("list operation", () => {
    it("lists API keys", async () => {
      const apiKeys = [
        {
          id: "key-1",
          name: "Test Key 1",
          expiresAt: "2025-01-01",
          revokedAt: null,
          createdAt: "2024-01-01",
        },
        {
          id: "key-2",
          name: "Test Key 2",
          expiresAt: "2025-06-01",
          revokedAt: null,
          createdAt: "2024-01-02",
        },
      ];
      mockPost.mockResolvedValue({ data: { data: { apiKeys } } });

      await program.parseAsync(["node", "test", "api-keys", "list", "-o", "json"]);

      expect(mockPost).toHaveBeenCalledWith("/metadata", {
        query: expect.stringContaining("apiKeys"),
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toHaveLength(2);
      expect(parsed[0].id).toBe("key-1");
    });

    it("handles empty API keys list", async () => {
      mockPost.mockResolvedValue({ data: { data: { apiKeys: [] } } });

      await program.parseAsync(["node", "test", "api-keys", "list", "-o", "json"]);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual([]);
    });

    it("handles null API keys response", async () => {
      mockPost.mockResolvedValue({ data: { data: { apiKeys: null } } });

      await program.parseAsync(["node", "test", "api-keys", "list", "-o", "json"]);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual([]);
    });
  });

  describe("get operation", () => {
    it("gets a single API key by ID", async () => {
      const apiKey = {
        id: "key-1",
        name: "Test Key",
        expiresAt: "2025-01-01",
        revokedAt: null,
        createdAt: "2024-01-01",
        updatedAt: "2024-01-02",
      };
      mockPost.mockResolvedValue({ data: { data: { apiKey } } });

      await program.parseAsync(["node", "test", "api-keys", "get", "key-1", "-o", "json"]);

      expect(mockPost).toHaveBeenCalledWith("/metadata", {
        query: expect.stringContaining("apiKey(input: { id: $id })"),
        variables: { id: "key-1" },
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.id).toBe("key-1");
      expect(parsed.name).toBe("Test Key");
    });

    it("throws error when ID is missing", async () => {
      await expect(program.parseAsync(["node", "test", "api-keys", "get"])).rejects.toThrow(
        CliError,
      );
    });
  });

  describe("create operation", () => {
    it("creates an API key with required metadata", async () => {
      const newApiKey = { id: "key-new", name: "New API Key", expiresAt: "2025-12-31T00:00:00Z" };
      mockPost.mockResolvedValueOnce({ data: { data: { createApiKey: newApiKey } } });
      mockPost.mockResolvedValueOnce({
        data: { data: { generateApiKeyToken: { token: "secret-token-123" } } },
      });

      await program.parseAsync([
        "node",
        "test",
        "api-keys",
        "create",
        "--name",
        "New API Key",
        "--expires-at",
        "2025-12-31T00:00:00Z",
        "--role-id",
        "role-1",
        "-o",
        "json",
      ]);

      expect(mockPost).toHaveBeenNthCalledWith(1, "/metadata", {
        query: expect.stringContaining("createApiKey"),
        variables: {
          input: {
            name: "New API Key",
            expiresAt: "2025-12-31T00:00:00Z",
            roleId: "role-1",
          },
        },
      });
      expect(mockPost).toHaveBeenNthCalledWith(2, "/metadata", {
        query: expect.stringContaining("generateApiKeyToken"),
        variables: { id: "key-new", expiresAt: "2025-12-31T00:00:00Z" },
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.id).toBe("key-new");
      expect(parsed.token).toBe("secret-token-123");
    });

    it("throws error when --name is missing", async () => {
      await expect(program.parseAsync(["node", "test", "api-keys", "create"])).rejects.toThrow(
        CliError,
      );
    });

    it("throws error when --expires-at is missing", async () => {
      await expect(
        program.parseAsync(["node", "test", "api-keys", "create", "--name", "Failed Key"]),
      ).rejects.toThrow(CliError);
    });

    it("throws error when --role-id is missing", async () => {
      await expect(
        program.parseAsync([
          "node",
          "test",
          "api-keys",
          "create",
          "--name",
          "Failed Key",
          "--expires-at",
          "2025-12-31T00:00:00Z",
        ]),
      ).rejects.toThrow(CliError);
    });

    it("throws error when createApiKey returns null", async () => {
      mockPost.mockResolvedValueOnce({ data: { data: { createApiKey: null } } });

      await expect(
        program.parseAsync([
          "node",
          "test",
          "api-keys",
          "create",
          "--name",
          "Failed Key",
          "--expires-at",
          "2025-12-31T00:00:00Z",
          "--role-id",
          "role-1",
        ]),
      ).rejects.toThrow(CliError);
    });
  });

  describe("update operation", () => {
    it("updates an API key by ID", async () => {
      const apiKey = { id: "key-1", name: "Updated Key", expiresAt: "2026-01-01T00:00:00Z" };
      mockPost.mockResolvedValue({ data: { data: { updateApiKey: apiKey } } });

      await program.parseAsync([
        "node",
        "test",
        "api-keys",
        "update",
        "key-1",
        "--name",
        "Updated Key",
        "--expires-at",
        "2026-01-01T00:00:00Z",
        "-o",
        "json",
      ]);

      expect(mockPost).toHaveBeenCalledWith("/metadata", {
        query: expect.stringContaining("updateApiKey"),
        variables: {
          input: {
            id: "key-1",
            name: "Updated Key",
            expiresAt: "2026-01-01T00:00:00Z",
            revokedAt: undefined,
          },
        },
      });
    });
  });

  describe("revoke operation", () => {
    it("revokes an API key by ID", async () => {
      mockPost.mockResolvedValue({ data: { data: { revokeApiKey: { id: "key-1" } } } });

      await program.parseAsync(["node", "test", "api-keys", "revoke", "key-1"]);

      expect(mockPost).toHaveBeenCalledWith("/metadata", {
        query: expect.stringContaining("revokeApiKey"),
        variables: { id: "key-1" },
      });
      expect(consoleSpy).toHaveBeenCalledWith("API key key-1 revoked.");
    });

    it("throws error when ID is missing", async () => {
      await expect(program.parseAsync(["node", "test", "api-keys", "revoke"])).rejects.toThrow(
        CliError,
      );
    });
  });

  describe("assign-role operation", () => {
    it("assigns a role to an API key", async () => {
      mockPost.mockResolvedValue({ data: { data: { assignRoleToApiKey: true } } });

      await program.parseAsync([
        "node",
        "test",
        "api-keys",
        "assign-role",
        "key-1",
        "--role-id",
        "role-2",
        "-o",
        "json",
      ]);

      expect(mockPost).toHaveBeenCalledWith("/metadata", {
        query: expect.stringContaining("assignRoleToApiKey"),
        variables: { id: "key-1", roleId: "role-2" },
      });

      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual({
        apiKeyId: "key-1",
        roleId: "role-2",
        assigned: true,
      });
    });
  });

  describe("error handling", () => {
    it("requires operation argument", async () => {
      await expect(program.parseAsync(["node", "test", "api-keys"])).rejects.toThrow();
    });

    it("throws error for unknown operation", async () => {
      await expect(program.parseAsync(["node", "test", "api-keys", "unknown"])).rejects.toThrow(
        CliError,
      );
    });

    it("handles case insensitive operations", async () => {
      const apiKeys = [
        { id: "key-1", name: "Test", expiresAt: null, revokedAt: null, createdAt: "" },
      ];
      mockPost.mockResolvedValue({ data: { data: { apiKeys } } });

      await program.parseAsync(["node", "test", "api-keys", "LIST", "-o", "json"]);

      expect(mockPost).toHaveBeenCalled();
    });
  });
});
