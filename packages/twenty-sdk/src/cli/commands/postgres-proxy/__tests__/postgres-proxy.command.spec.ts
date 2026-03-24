import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Command } from "commander";
import { registerPostgresProxyCommand } from "../postgres-proxy.command";
import { ApiService } from "../../../utilities/api/services/api.service";
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

describe("postgres-proxy command", () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;
  let mockPost: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerPostgresProxyCommand(program);
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

  it("registers the postgres-proxy command", () => {
    const command = program.commands.find((candidate) => candidate.name() === "postgres-proxy");

    expect(command).toBeDefined();
    expect(command?.description()).toBe("Manage Postgres proxy credentials");
  });

  it("gets Postgres credentials and masks the password by default", async () => {
    mockPost.mockResolvedValue({
      data: {
        data: {
          getPostgresCredentials: {
            id: "pg-1",
            user: "workspace_user",
            password: "super-secret",
            workspaceId: "ws-1",
          },
        },
      },
    });

    await program.parseAsync(["node", "test", "postgres-proxy", "get", "-o", "json"]);

    expect(mockPost).toHaveBeenCalledWith(
      "/graphql",
      expect.objectContaining({
        query: expect.stringContaining("getPostgresCredentials"),
      }),
    );
    const output = consoleSpy.mock.calls[0][0] as string;
    expect(JSON.parse(output)).toEqual({
      id: "pg-1",
      user: "workspace_user",
      password: "[hidden]",
      workspaceId: "ws-1",
    });
  });

  it("shows the Postgres password when requested", async () => {
    mockPost.mockResolvedValue({
      data: {
        data: {
          getPostgresCredentials: {
            id: "pg-1",
            user: "workspace_user",
            password: "super-secret",
            workspaceId: "ws-1",
          },
        },
      },
    });

    await program.parseAsync([
      "node",
      "test",
      "postgres-proxy",
      "get",
      "--show-password",
      "-o",
      "json",
    ]);

    const output = consoleSpy.mock.calls[0][0] as string;
    expect(JSON.parse(output)).toEqual({
      id: "pg-1",
      user: "workspace_user",
      password: "super-secret",
      workspaceId: "ws-1",
    });
  });

  it("enables the Postgres proxy", async () => {
    mockPost.mockResolvedValue({
      data: {
        data: {
          enablePostgresProxy: {
            id: "pg-1",
            user: "workspace_user",
            password: "super-secret",
            workspaceId: "ws-1",
          },
        },
      },
    });

    await program.parseAsync(["node", "test", "postgres-proxy", "enable", "-o", "json"]);

    expect(mockPost).toHaveBeenCalledWith(
      "/graphql",
      expect.objectContaining({
        query: expect.stringContaining("enablePostgresProxy"),
      }),
    );
    const output = consoleSpy.mock.calls[0][0] as string;
    const parsed = JSON.parse(output);
    expect(parsed.user).toBe("workspace_user");
    expect(parsed.password).toBe("[hidden]");
  });

  it("disables the Postgres proxy", async () => {
    mockPost.mockResolvedValue({
      data: {
        data: {
          disablePostgresProxy: {
            id: "pg-1",
            user: "workspace_user",
            password: "super-secret",
            workspaceId: "ws-1",
          },
        },
      },
    });

    await program.parseAsync(["node", "test", "postgres-proxy", "disable", "-o", "json"]);

    expect(mockPost).toHaveBeenCalledWith(
      "/graphql",
      expect.objectContaining({
        query: expect.stringContaining("disablePostgresProxy"),
      }),
    );
    const output = consoleSpy.mock.calls[0][0] as string;
    const parsed = JSON.parse(output);
    expect(parsed.password).toBe("[hidden]");
  });

  it("fails clearly when the workspace schema does not expose Postgres proxy operations", async () => {
    mockPost.mockResolvedValue({
      data: {
        errors: [
          {
            message: 'Cannot query field "getPostgresCredentials" on type "Query".',
          },
        ],
      },
    });

    await expect(
      program.parseAsync(["node", "test", "postgres-proxy", "get", "-o", "json"]),
    ).rejects.toThrow(
      "Postgres proxy is not available on this workspace because it does not expose getPostgresCredentials.",
    );
  });
});
