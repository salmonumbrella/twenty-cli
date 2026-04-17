import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Command } from "commander";
import { registerDbCommand } from "../db.command";
import { createCommandContext } from "../../../utilities/shared/context";

vi.mock("../../../utilities/shared/context", () => ({
  createCommandContext: vi.fn(),
}));

describe("db command", () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;
  let outputRender: ReturnType<typeof vi.fn>;
  let mockStatus: ReturnType<typeof vi.fn>;
  let mockListProfiles: ReturnType<typeof vi.fn>;
  let mockInitProfile: ReturnType<typeof vi.fn>;
  let mockGetProfile: ReturnType<typeof vi.fn>;
  let mockSetActiveProfile: ReturnType<typeof vi.fn>;
  let mockTest: ReturnType<typeof vi.fn>;
  let mockRefreshCreds: ReturnType<typeof vi.fn>;
  let mockRemoveProfile: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    delete process.env.TWENTY_DB_PROFILE;
    delete process.env.TWENTY_DB_PROXY_URL;
    program = new Command();
    program.exitOverride();
    registerDbCommand(program);
    consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});
    outputRender = vi.fn();
    mockStatus = vi.fn();
    mockListProfiles = vi.fn();
    mockInitProfile = vi.fn();
    mockGetProfile = vi.fn();
    mockSetActiveProfile = vi.fn();
    mockTest = vi.fn();
    mockRefreshCreds = vi.fn();
    mockRemoveProfile = vi.fn();

    vi.mocked(createCommandContext).mockImplementation(() => ({
      globalOptions: {
        output: "json",
        query: undefined,
        workspace: undefined,
      },
      services: {
        config: {} as never,
        api: {} as never,
        publicHttp: {} as never,
        search: {} as never,
        mcp: {} as never,
        records: {} as never,
        metadata: {} as never,
        output: {
          render: outputRender,
        } as never,
        importer: {} as never,
        exporter: {} as never,
        dbStatus: {
          getStatus: mockStatus,
        } as never,
        dbProfiles: {
          listProfiles: mockListProfiles,
          initProfile: mockInitProfile,
          getProfile: mockGetProfile,
          setActiveProfile: mockSetActiveProfile,
          test: mockTest,
          refreshCreds: mockRefreshCreds,
          removeProfile: mockRemoveProfile,
        } as never,
      },
    }));
  });

  afterEach(() => {
    consoleSpy.mockRestore();
    delete process.env.TWENTY_DB_PROFILE;
    delete process.env.TWENTY_DB_PROXY_URL;
    vi.clearAllMocks();
  });

  it("registers the db command family", () => {
    const command = program.commands.find((candidate) => candidate.name() === "db");
    const profile = command?.commands.find((candidate) => candidate.name() === "profile");
    const help = command?.helpInformation() ?? "";

    expect(command).toBeDefined();
    expect(command?.description()).toBe("Manage db-first read profiles and diagnostics");
    expect(command?.commands.map((candidate) => candidate.name())).toEqual(
      expect.arrayContaining(["profile", "status"]),
    );
    expect(profile?.commands.map((candidate) => candidate.name())).toEqual(
      expect.arrayContaining(["init", "list", "show", "use", "test", "refresh-creds", "remove"]),
    );
    expect(help).toContain("profile");
    expect(help).toContain("status");
  });

  it("shows db status via the status service", async () => {
    mockStatus.mockResolvedValue({
      workspace: "prod",
      configured: true,
      mode: "db",
      profileName: "readonly",
    });

    await program.parseAsync(["node", "test", "db", "status"]);

    expect(mockStatus).toHaveBeenCalledWith({ workspace: undefined });
    expect(outputRender).toHaveBeenCalledWith(
      {
        workspace: "prod",
        configured: true,
        mode: "db",
        profileName: "readonly",
      },
      {
        format: "json",
        query: undefined,
      },
    );
  });

  it("prefers an explicit db profile name when showing a profile", async () => {
    mockGetProfile.mockResolvedValue({
      name: "explicit",
      workspace: "prod",
      proxyUrl: "http://localhost:4010",
      credentialSource: "manual",
    });

    await program.parseAsync(["node", "test", "db", "profile", "show", "explicit"]);

    expect(mockGetProfile).toHaveBeenCalledWith(undefined, "explicit");
    expect(outputRender).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "explicit",
      }),
      expect.objectContaining({
        format: "json",
      }),
    );
  });

  it("falls back to TWENTY_DB_PROFILE when showing a profile", async () => {
    process.env.TWENTY_DB_PROFILE = "cached";
    mockGetProfile.mockResolvedValue({
      name: "cached",
      workspace: "prod",
      proxyUrl: "http://localhost:4011",
      credentialSource: "manual",
    });

    await program.parseAsync(["node", "test", "db", "profile", "show"]);

    expect(mockGetProfile).toHaveBeenCalledWith(undefined, "cached");
  });

  it("falls back to the status profile when showing a profile", async () => {
    mockStatus.mockResolvedValue({
      workspace: "prod",
      configured: true,
      mode: "db",
      profileName: "status-profile",
    });
    mockGetProfile.mockResolvedValue({
      name: "status-profile",
      workspace: "prod",
      proxyUrl: "http://localhost:4012",
      credentialSource: "manual",
    });

    await program.parseAsync(["node", "test", "db", "profile", "show"]);

    expect(mockStatus).toHaveBeenCalledWith({ workspace: undefined });
    expect(mockGetProfile).toHaveBeenCalledWith(undefined, "status-profile");
  });

  it("fails clearly when no profile can be resolved for show", async () => {
    mockStatus.mockResolvedValue({
      workspace: "prod",
      configured: false,
      mode: "api",
    });

    await expect(program.parseAsync(["node", "test", "db", "profile", "show"])).rejects.toThrow(
      "No DB profile selected.",
    );
  });

  it("initializes a db profile from a proxy URL", async () => {
    mockInitProfile.mockResolvedValue({
      name: "readonly",
      workspace: "default",
      proxyUrl: "http://localhost:4010",
      credentialSource: "manual",
      notes: "seeded",
    });

    await program.parseAsync([
      "node",
      "test",
      "db",
      "profile",
      "init",
      "readonly",
      "--proxy-url",
      "http://localhost:4010",
      "--notes",
      "seeded",
    ]);

    expect(mockInitProfile).toHaveBeenCalledWith({
      workspace: undefined,
      name: "readonly",
      proxyUrl: "http://localhost:4010",
      notes: "seeded",
    });
    expect(outputRender).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "readonly",
        proxyUrl: "http://localhost:4010",
        credentialSource: "manual",
        notes: "seeded",
      }),
      expect.objectContaining({
        format: "json",
      }),
    );
  });

  it("uses TWENTY_DB_PROXY_URL when initializing a db profile", async () => {
    process.env.TWENTY_DB_PROXY_URL = "http://localhost:4999";
    mockInitProfile.mockResolvedValue({
      name: "cached",
      workspace: "default",
      proxyUrl: "http://localhost:4999",
      credentialSource: "manual",
    });

    await program.parseAsync(["node", "test", "db", "profile", "init", "cached"]);

    expect(mockInitProfile).toHaveBeenCalledWith({
      workspace: undefined,
      name: "cached",
      proxyUrl: "http://localhost:4999",
      notes: undefined,
    });
  });
});
