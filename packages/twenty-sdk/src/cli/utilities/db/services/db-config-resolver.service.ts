import { DbProfileService } from "./db-profile.service";

export interface ResolvedDbConfig {
  workspace: string;
  mode: "api" | "db";
  source: "env" | "profile" | "none" | "override";
  databaseUrl?: string;
  profileName?: string;
}

export class DbConfigResolverService {
  constructor(private readonly dbProfiles: DbProfileService) {}

  async resolve(options?: { workspace?: string }): Promise<ResolvedDbConfig> {
    const workspace = await this.dbProfiles.resolveWorkspace(options?.workspace);
    const activeProfile = await this.dbProfiles.getActiveProfile(workspace);

    if (process.env.TWENTY_INTERNAL_READ_BACKEND === "api") {
      return {
        workspace,
        mode: "api",
        source: "override",
        profileName: activeProfile?.name,
      };
    }

    const envDatabaseUrl = process.env.TWENTY_DATABASE_URL?.trim();

    if (envDatabaseUrl) {
      return {
        workspace,
        mode: "db",
        source: "env",
        databaseUrl: envDatabaseUrl,
      };
    }

    if (activeProfile?.databaseUrl) {
      return {
        workspace,
        mode: "db",
        source: "profile",
        databaseUrl: activeProfile.databaseUrl,
        profileName: activeProfile.name,
      };
    }

    return {
      workspace,
      mode: "api",
      source: "none",
    };
  }
}
