import { CliError } from "../../errors/cli-error";
import { DbProfileService } from "./db-profile.service";

export interface DbStatusSummary {
  workspace: string;
  configured: boolean;
  mode: "api" | "db";
  profileName?: string;
}

export class DbStatusService {
  constructor(private readonly dbProfiles: DbProfileService) {}

  async getStatus(options?: { workspace?: string }): Promise<DbStatusSummary> {
    const workspace = await this.dbProfiles.resolveWorkspace(options?.workspace);

    try {
      const profile = await this.dbProfiles.getActiveProfile(workspace);

      if (!profile) {
        return {
          workspace,
          configured: false,
          mode: "api",
        };
      }

      return {
        workspace,
        configured: true,
        mode: "db",
        profileName: profile.name,
      };
    } catch (error) {
      if (this.isMissingWorkspace(error, workspace)) {
        return {
          workspace,
          configured: false,
          mode: "api",
        };
      }

      throw error;
    }
  }

  private isMissingWorkspace(error: unknown, workspace: string): boolean {
    return (
      error instanceof CliError &&
      error.code === "INVALID_ARGUMENTS" &&
      error.message === `Workspace '${workspace}' does not exist`
    );
  }
}
