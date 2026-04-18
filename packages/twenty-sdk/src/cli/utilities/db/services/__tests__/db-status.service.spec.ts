import { afterEach, describe, expect, it, vi } from "vitest";
import { DbConfigResolverService } from "../db-config-resolver.service";
import { DbStatusService } from "../db-status.service";

describe("DbStatusService", () => {
  afterEach(() => {
    delete process.env.TWENTY_INTERNAL_READ_BACKEND;
    delete process.env.TWENTY_DATABASE_URL;
    vi.clearAllMocks();
  });

  it("reports env db configuration as the active db-first source", async () => {
    const resolver = new DbConfigResolverService({
      resolveWorkspace: vi.fn().mockResolvedValue("prod"),
      getActiveProfile: vi.fn().mockResolvedValue(undefined),
    } as never);

    process.env.TWENTY_DATABASE_URL = "postgresql://db.example.com:5432/twenty";
    const service = new DbStatusService(resolver);

    await expect(service.getStatus({ workspace: "prod" })).resolves.toEqual({
      workspace: "prod",
      configured: true,
      mode: "db",
      source: "env",
    });
  });

  it("reports the active saved profile when db-first reads come from profile config", async () => {
    const service = new DbStatusService(
      new DbConfigResolverService({
        resolveWorkspace: vi.fn().mockResolvedValue("prod"),
        getActiveProfile: vi.fn().mockResolvedValue({
          name: "staging",
          workspace: "prod",
          databaseUrl: "postgresql://db.example.com:5432/twenty",
        }),
      } as never),
    );

    await expect(service.getStatus({ workspace: "prod" })).resolves.toEqual({
      workspace: "prod",
      configured: true,
      mode: "db",
      source: "profile",
      profileName: "staging",
    });
  });

  it("reports api mode when the internal backend override disables db-first reads", async () => {
    process.env.TWENTY_INTERNAL_READ_BACKEND = "api";
    process.env.TWENTY_DATABASE_URL = "postgresql://db.example.com:5432/twenty";

    const service = new DbStatusService(
      new DbConfigResolverService({
        resolveWorkspace: vi.fn().mockResolvedValue("prod"),
        getActiveProfile: vi.fn().mockResolvedValue({
          name: "staging",
          workspace: "prod",
          databaseUrl: "postgresql://db.example.com:5432/twenty",
        }),
      } as never),
    );

    await expect(service.getStatus({ workspace: "prod" })).resolves.toEqual({
      workspace: "prod",
      configured: false,
      mode: "api",
      source: "override",
      profileName: "staging",
    });
  });

  it("reports api mode when no db configuration exists", async () => {
    const service = new DbStatusService(
      new DbConfigResolverService({
        resolveWorkspace: vi.fn().mockResolvedValue("prod"),
        getActiveProfile: vi.fn().mockResolvedValue(undefined),
      } as never),
    );

    await expect(service.getStatus({ workspace: "prod" })).resolves.toEqual({
      workspace: "prod",
      configured: false,
      mode: "api",
      source: "none",
    });
  });

  it("uses the resolver workspace even when no workspace option is passed", async () => {
    const resolveWorkspace = vi.fn().mockResolvedValue("resolved-workspace");
    const service = new DbStatusService(
      new DbConfigResolverService({
        resolveWorkspace,
        getActiveProfile: vi.fn().mockResolvedValue({
          name: "readonly",
          workspace: "resolved-workspace",
          databaseUrl: "postgresql://db.example.com:5432/twenty",
        }),
      } as never),
    );

    await expect(service.getStatus()).resolves.toEqual({
      workspace: "resolved-workspace",
      configured: true,
      mode: "db",
      source: "profile",
      profileName: "readonly",
    });
    expect(resolveWorkspace).toHaveBeenCalledWith(undefined);
  });
});
