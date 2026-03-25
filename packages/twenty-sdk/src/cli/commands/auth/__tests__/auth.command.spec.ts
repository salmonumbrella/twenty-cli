import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { Command } from "commander";
import { registerAuthCommand } from "../auth.command";
import { ApiService } from "../../../utilities/api/services/api.service";
import {
  ConfigService,
  WorkspaceInfo,
  ResolvedConfig,
} from "../../../utilities/config/services/config.service";
import { CliError } from "../../../utilities/errors/cli-error";
import { mockConstructor } from "../../../test-utils/mock-constructor";
import { loadCliEnvironment } from "../../../utilities/config/services/environment.service";

vi.mock("../../../utilities/config/services/config.service");
vi.mock("../../../utilities/api/services/api.service");
vi.mock("../../../utilities/config/services/environment.service", () => ({
  loadCliEnvironment: vi.fn(),
  resolveEnvFileFromArgv: vi.fn(),
}));

describe("auth commands", () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;
  let mockPost: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerAuthCommand(program);
    consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});
    mockPost = vi.fn();
    vi.mocked(loadCliEnvironment).mockReset();
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

  describe("auth list", () => {
    it("shows message when no workspaces configured", async () => {
      vi.mocked(ConfigService.prototype.listWorkspaces).mockResolvedValue([]);

      await program.parseAsync(["node", "test", "auth", "list"]);

      expect(consoleSpy).toHaveBeenCalledWith(
        'No workspaces configured. Use "twenty auth login" to add a workspace.',
      );
    });

    it("lists configured workspaces with default marker", async () => {
      const workspaces: WorkspaceInfo[] = [
        { name: "production", isDefault: true, apiUrl: "https://api.twenty.com" },
        { name: "staging", isDefault: false, apiUrl: "https://staging.twenty.com" },
      ];
      vi.mocked(ConfigService.prototype.listWorkspaces).mockResolvedValue(workspaces);

      await program.parseAsync(["node", "test", "auth", "list", "-o", "json"]);

      expect(ConfigService.prototype.listWorkspaces).toHaveBeenCalled();
      // JSON output will contain the formatted workspace data
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual([
        { name: "production", default: "Y", apiUrl: "https://api.twenty.com" },
        { name: "staging", default: "", apiUrl: "https://staging.twenty.com" },
      ]);
    });

    it("loads env handling once through shared output context", async () => {
      vi.mocked(ConfigService.prototype.listWorkspaces).mockResolvedValue([]);

      await program.parseAsync([
        "node",
        "test",
        "auth",
        "list",
        "--env-file",
        ".env.test",
      ]);

      expect(loadCliEnvironment).toHaveBeenCalledTimes(1);
      expect(loadCliEnvironment).toHaveBeenCalledWith({
        argv: process.argv,
        cwd: process.cwd(),
        explicitEnvFile: ".env.test",
      });
    });
  });

  describe("auth switch", () => {
    it("switches default workspace successfully", async () => {
      vi.mocked(ConfigService.prototype.setDefaultWorkspace).mockResolvedValue(undefined);

      await program.parseAsync(["node", "test", "auth", "switch", "production"]);

      expect(ConfigService.prototype.setDefaultWorkspace).toHaveBeenCalledWith("production");
      expect(consoleSpy).toHaveBeenCalledWith('Switched to workspace "production".');
    });
  });

  describe("auth login", () => {
    it("saves workspace with token and default URL/workspace", async () => {
      vi.mocked(ConfigService.prototype.saveWorkspace).mockResolvedValue(undefined);

      await program.parseAsync(["node", "test", "auth", "login", "--token", "my-api-token"]);

      expect(ConfigService.prototype.saveWorkspace).toHaveBeenCalledWith("default", {
        apiKey: "my-api-token",
        apiUrl: "https://api.twenty.com",
      });
      expect(consoleSpy).toHaveBeenCalledWith('Workspace "default" configured.');
      expect(consoleSpy).toHaveBeenCalledWith("API URL: https://api.twenty.com");
    });

    it("saves workspace with custom base-url and workspace name", async () => {
      vi.mocked(ConfigService.prototype.saveWorkspace).mockResolvedValue(undefined);

      await program.parseAsync([
        "node",
        "test",
        "auth",
        "login",
        "--token",
        "custom-token",
        "--base-url",
        "https://custom.twenty.com",
        "--workspace",
        "production",
      ]);

      expect(ConfigService.prototype.saveWorkspace).toHaveBeenCalledWith("production", {
        apiKey: "custom-token",
        apiUrl: "https://custom.twenty.com",
      });
      expect(consoleSpy).toHaveBeenCalledWith('Workspace "production" configured.');
      expect(consoleSpy).toHaveBeenCalledWith("API URL: https://custom.twenty.com");
    });
  });

  describe("auth logout", () => {
    it("removes specified workspace", async () => {
      vi.mocked(ConfigService.prototype.removeWorkspace).mockResolvedValue(undefined);

      await program.parseAsync(["node", "test", "auth", "logout", "--workspace", "staging"]);

      expect(ConfigService.prototype.removeWorkspace).toHaveBeenCalledWith("staging");
      expect(consoleSpy).toHaveBeenCalledWith('Workspace "staging" removed.');
    });

    it("removes all workspaces with --all flag", async () => {
      const workspaces: WorkspaceInfo[] = [
        { name: "production", isDefault: true, apiUrl: "https://api.twenty.com" },
        { name: "staging", isDefault: false, apiUrl: "https://staging.twenty.com" },
      ];
      vi.mocked(ConfigService.prototype.listWorkspaces).mockResolvedValue(workspaces);
      vi.mocked(ConfigService.prototype.removeWorkspace).mockResolvedValue(undefined);

      await program.parseAsync(["node", "test", "auth", "logout", "--all"]);

      expect(ConfigService.prototype.listWorkspaces).toHaveBeenCalled();
      expect(ConfigService.prototype.removeWorkspace).toHaveBeenCalledWith("production");
      expect(ConfigService.prototype.removeWorkspace).toHaveBeenCalledWith("staging");
      expect(consoleSpy).toHaveBeenCalledWith("All workspaces removed.");
    });

    it("removes default workspace when no option specified", async () => {
      const config: ResolvedConfig = {
        apiUrl: "https://api.twenty.com",
        apiKey: "test-token",
        workspace: "current-workspace",
      };
      vi.mocked(ConfigService.prototype.getConfig).mockResolvedValue(config);
      vi.mocked(ConfigService.prototype.removeWorkspace).mockResolvedValue(undefined);

      await program.parseAsync(["node", "test", "auth", "logout"]);

      expect(ConfigService.prototype.getConfig).toHaveBeenCalled();
      expect(ConfigService.prototype.removeWorkspace).toHaveBeenCalledWith("current-workspace");
      expect(consoleSpy).toHaveBeenCalledWith('Workspace "current-workspace" removed.');
    });
  });

  describe("auth status", () => {
    it("shows authenticated status with masked token", async () => {
      const config: ResolvedConfig = {
        apiUrl: "https://api.twenty.com",
        apiKey: "abcd1234efgh5678",
        workspace: "production",
      };
      vi.mocked(ConfigService.prototype.getConfig).mockResolvedValue(config);

      await program.parseAsync(["node", "test", "auth", "status", "-o", "json"]);

      expect(ConfigService.prototype.getConfig).toHaveBeenCalledWith({
        workspace: undefined,
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual({
        authenticated: true,
        workspace: "production",
        apiUrl: "https://api.twenty.com",
        apiKey: "abcd****5678",
      });
    });

    it("uses the selected workspace profile when provided", async () => {
      const config: ResolvedConfig = {
        apiUrl: "https://smoke.example.com",
        apiKey: "abcd1234efgh5678",
        workspace: "smoke",
      };
      vi.mocked(ConfigService.prototype.getConfig).mockResolvedValue(config);

      await program.parseAsync([
        "node",
        "test",
        "auth",
        "status",
        "--workspace",
        "smoke",
        "-o",
        "json",
      ]);

      expect(ConfigService.prototype.getConfig).toHaveBeenCalledWith({
        workspace: "smoke",
      });
      const output = consoleSpy.mock.calls[0][0] as string;
      expect(JSON.parse(output)).toMatchObject({
        authenticated: true,
        workspace: "smoke",
        apiUrl: "https://smoke.example.com",
      });
    });

    it("shows full token when --show-token used", async () => {
      const config: ResolvedConfig = {
        apiUrl: "https://api.twenty.com",
        apiKey: "abcd1234efgh5678",
        workspace: "production",
      };
      vi.mocked(ConfigService.prototype.getConfig).mockResolvedValue(config);

      await program.parseAsync(["node", "test", "auth", "status", "--show-token", "-o", "json"]);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual({
        authenticated: true,
        workspace: "production",
        apiUrl: "https://api.twenty.com",
        apiKey: "abcd1234efgh5678",
      });
    });

    it("shows unauthenticated status when no config exists", async () => {
      const authError = new CliError(
        "Missing API token.",
        "AUTH",
        "Set TWENTY_TOKEN or configure ~/.twenty/config.json with an apiKey.",
      );
      vi.mocked(ConfigService.prototype.getConfig).mockRejectedValue(authError);

      await program.parseAsync(["node", "test", "auth", "status", "-o", "json"]);

      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed).toEqual({
        authenticated: false,
        error: "Missing API token.",
      });
    });

    it("masks short tokens properly", async () => {
      const config: ResolvedConfig = {
        apiUrl: "https://api.twenty.com",
        apiKey: "short",
        workspace: "production",
      };
      vi.mocked(ConfigService.prototype.getConfig).mockResolvedValue(config);

      await program.parseAsync(["node", "test", "auth", "status", "-o", "json"]);

      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.apiKey).toBe("****");
    });

    it("loads env handling once through shared output context", async () => {
      const config: ResolvedConfig = {
        apiUrl: "https://api.twenty.com",
        apiKey: "short",
        workspace: "production",
      };
      vi.mocked(ConfigService.prototype.getConfig).mockResolvedValue(config);

      await program.parseAsync([
        "node",
        "test",
        "auth",
        "status",
        "--env-file",
        ".env.test",
        "-o",
        "json",
      ]);

      expect(loadCliEnvironment).toHaveBeenCalledTimes(1);
      expect(loadCliEnvironment).toHaveBeenCalledWith({
        argv: process.argv,
        cwd: process.cwd(),
        explicitEnvFile: ".env.test",
      });
    });
  });

  describe("auth workspace", () => {
    it("loads the current workspace from the API", async () => {
      mockPost.mockResolvedValue({
        data: {
          data: {
            currentWorkspace: {
              id: "ws-1",
              displayName: "Acme",
              activationStatus: "ACTIVE",
              workspaceUrls: {
                subdomainUrl: "https://acme.twenty.com",
                customUrl: null,
              },
            },
          },
        },
      });

      await program.parseAsync(["node", "test", "auth", "workspace", "-o", "json"]);

      expect(mockPost).toHaveBeenCalledWith("/metadata", {
        query: expect.stringContaining("currentWorkspace"),
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.displayName).toBe("Acme");
      expect(parsed.workspaceUrls.subdomainUrl).toBe("https://acme.twenty.com");
    });
  });

  describe("auth discover", () => {
    it("loads public workspace auth settings for an origin", async () => {
      mockPost.mockResolvedValue({
        data: {
          data: {
            getPublicWorkspaceDataByDomain: {
              id: "ws-1",
              displayName: "Acme",
              workspaceUrls: {
                subdomainUrl: "https://acme.twenty.com",
                customUrl: "https://crm.acme.com",
              },
              authProviders: {
                google: true,
                magicLink: false,
                password: true,
                microsoft: false,
                sso: [],
              },
              authBypassProviders: {
                google: false,
                password: false,
                microsoft: false,
              },
            },
          },
        },
      });

      await program.parseAsync([
        "node",
        "test",
        "auth",
        "discover",
        "https://acme.twenty.com",
        "-o",
        "json",
      ]);

      expect(mockPost).toHaveBeenCalledWith("/metadata", {
        query: expect.stringContaining("getPublicWorkspaceDataByDomain"),
        variables: { origin: "https://acme.twenty.com" },
      });
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls[0][0] as string;
      const parsed = JSON.parse(output);
      expect(parsed.displayName).toBe("Acme");
      expect(parsed.authProviders.google).toBe(true);
      expect(parsed.workspaceUrls.customUrl).toBe("https://crm.acme.com");
    });
  });

  describe("auth renew-token", () => {
    it("renews an app token through the core auth GraphQL mutation", async () => {
      mockPost.mockResolvedValue({
        data: {
          data: {
            renewToken: {
              tokens: {
                accessToken: "access-token",
                refreshToken: "refresh-token",
              },
            },
          },
        },
      });

      await program.parseAsync([
        "node",
        "test",
        "auth",
        "renew-token",
        "--app-token",
        "refresh-token",
        "-o",
        "json",
      ]);

      expect(mockPost).toHaveBeenCalledWith("/graphql", {
        query: expect.stringContaining("renewToken"),
        variables: { appToken: "refresh-token" },
      });

      const output = consoleSpy.mock.calls[0][0] as string;
      expect(JSON.parse(output)).toEqual({
        tokens: {
          accessToken: "access-token",
          refreshToken: "refresh-token",
        },
      });
    });
  });

  describe("auth sso-url", () => {
    it("gets an SSO authorization URL for an identity provider", async () => {
      mockPost.mockResolvedValue({
        data: {
          data: {
            getAuthorizationUrlForSSO: {
              authorizationURL: "https://idp.example.com/login",
              type: "OIDC",
              id: "idp-1",
            },
          },
        },
      });

      await program.parseAsync([
        "node",
        "test",
        "auth",
        "sso-url",
        "idp-1",
        "--workspace-invite-hash",
        "invite-123",
        "-o",
        "json",
      ]);

      expect(mockPost).toHaveBeenCalledWith("/graphql", {
        query: expect.stringContaining("getAuthorizationUrlForSSO"),
        variables: {
          input: {
            identityProviderId: "idp-1",
            workspaceInviteHash: "invite-123",
          },
        },
      });

      const output = consoleSpy.mock.calls[0][0] as string;
      expect(JSON.parse(output)).toEqual({
        authorizationURL: "https://idp.example.com/login",
        type: "OIDC",
        id: "idp-1",
      });
    });

    it("omits invite hash when not provided", async () => {
      mockPost.mockResolvedValue({
        data: {
          data: {
            getAuthorizationUrlForSSO: {
              authorizationURL: "https://idp.example.com/login",
              type: "OIDC",
              id: "idp-1",
            },
          },
        },
      });

      await program.parseAsync(["node", "test", "auth", "sso-url", "idp-1", "-o", "json"]);

      expect(mockPost).toHaveBeenCalledWith("/graphql", {
        query: expect.stringContaining("getAuthorizationUrlForSSO"),
        variables: {
          input: {
            identityProviderId: "idp-1",
          },
        },
      });
    });
  });
});
