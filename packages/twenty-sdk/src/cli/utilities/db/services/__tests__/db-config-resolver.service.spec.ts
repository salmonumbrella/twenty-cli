import { afterEach, describe, expect, it, vi } from "vitest";
import { DbConfigResolverService } from "../db-config-resolver.service";

describe("DbConfigResolverService", () => {
  afterEach(() => {
    delete process.env.TWENTY_INTERNAL_READ_BACKEND;
    delete process.env.TWENTY_DATABASE_URL;
    vi.clearAllMocks();
  });

  it("prefers the internal api override over any db configuration while preserving profile context", async () => {
    process.env.TWENTY_INTERNAL_READ_BACKEND = "api";
    process.env.TWENTY_DATABASE_URL = "postgresql://env-user:env-pass@db.example.com:5432/twenty";

    const dbProfiles = {
      resolveWorkspace: vi.fn().mockResolvedValue("prod"),
      getActiveProfile: vi.fn().mockResolvedValue({
        name: "staging",
        workspace: "prod",
        databaseUrl: "postgresql://profile-user:profile-pass@db.example.com:5432/twenty_staging",
      }),
    };

    const resolver = new DbConfigResolverService(dbProfiles as never);

    await expect(resolver.resolve({ workspace: "prod" })).resolves.toEqual({
      workspace: "prod",
      mode: "api",
      source: "override",
      profileName: "staging",
    });
    expect(dbProfiles.getActiveProfile).toHaveBeenCalledWith("prod");
  });

  it("prefers TWENTY_DATABASE_URL over the saved active db profile", async () => {
    process.env.TWENTY_DATABASE_URL = "postgresql://env-user:env-pass@db.example.com:5432/twenty";

    const resolver = new DbConfigResolverService({
      resolveWorkspace: vi.fn().mockResolvedValue("prod"),
      getActiveProfile: vi.fn().mockResolvedValue({
        name: "staging",
        workspace: "prod",
        databaseUrl: "postgresql://profile-user:profile-pass@db.example.com:5432/twenty_staging",
      }),
    } as never);

    await expect(resolver.resolve({ workspace: "prod" })).resolves.toEqual({
      workspace: "prod",
      mode: "db",
      source: "env",
      databaseUrl: "postgresql://env-user:env-pass@db.example.com:5432/twenty",
    });
  });

  it("uses the saved active db profile when no env database url is set", async () => {
    const resolver = new DbConfigResolverService({
      resolveWorkspace: vi.fn().mockResolvedValue("prod"),
      getActiveProfile: vi.fn().mockResolvedValue({
        name: "staging",
        workspace: "prod",
        databaseUrl: "postgresql://profile-user:profile-pass@db.example.com:5432/twenty_staging",
      }),
    } as never);

    await expect(resolver.resolve({ workspace: "prod" })).resolves.toEqual({
      workspace: "prod",
      mode: "db",
      source: "profile",
      databaseUrl: "postgresql://profile-user:profile-pass@db.example.com:5432/twenty_staging",
      profileName: "staging",
    });
  });

  it("falls back to api mode when no db configuration exists", async () => {
    const resolver = new DbConfigResolverService({
      resolveWorkspace: vi.fn().mockResolvedValue("prod"),
      getActiveProfile: vi.fn().mockResolvedValue(undefined),
    } as never);

    await expect(resolver.resolve({ workspace: "prod" })).resolves.toEqual({
      workspace: "prod",
      mode: "api",
      source: "none",
    });
  });
});
