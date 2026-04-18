import { DbConfigResolverService } from "./db-config-resolver.service";

export interface DbStatusSummary {
  workspace: string;
  configured: boolean;
  mode: "api" | "db";
  source: "env" | "profile" | "none" | "override";
  profileName?: string;
}

export class DbStatusService {
  constructor(private readonly dbConfigResolver: DbConfigResolverService) {}

  async getStatus(options?: { workspace?: string }): Promise<DbStatusSummary> {
    const resolved = await this.dbConfigResolver.resolve(options);

    return {
      workspace: resolved.workspace,
      configured: resolved.mode === "db",
      mode: resolved.mode,
      source: resolved.source,
      profileName: resolved.profileName,
    };
  }
}
