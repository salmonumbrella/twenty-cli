import { Command } from "commander";
import { CliError } from "../../utilities/errors/cli-error";
import { applyGlobalOptions } from "../../utilities/shared/global-options";
import { createCommandContext } from "../../utilities/shared/context";
import { registerCommand } from "../../utilities/shared/register-command";

interface DbProfileInitOptions {
  databaseUrl?: string;
  notes?: string;
}

function resolveProfileName(
  explicitName: string | undefined,
  envName: string | undefined,
  statusName: string | undefined,
): string | undefined {
  return explicitName || envName || statusName;
}

export function registerDbCommand(program: Command): void {
  const db = program.command("db").description("Manage db-first read profiles and diagnostics");
  applyGlobalOptions(db);

  const statusCmd = db.command("status").description("Show db-first read diagnostics");
  applyGlobalOptions(statusCmd);
  statusCmd.action(async (_options: unknown, command: Command) => {
    const { globalOptions, services } = createCommandContext(command);
    const status = await services.dbStatus.getStatus({ workspace: globalOptions.workspace });

    await services.output.render(status, {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });

  const profileCmd = db.command("profile").description("cached db profiles");
  applyGlobalOptions(profileCmd);

  registerCommand(profileCmd, "init", "Initialize a db profile", (command) => {
    command.argument("<name>", "Profile name");
    command.option("--database-url <url>", "Database URL");
    command.option("--notes <text>", "Profile notes");
    applyGlobalOptions(command);
    command.action(async (name: string, options: DbProfileInitOptions, actionCommand: Command) => {
      const { globalOptions, services } = createCommandContext(actionCommand);
      const databaseUrl = options.databaseUrl ?? process.env.TWENTY_DATABASE_URL;

      if (!databaseUrl) {
        throw new CliError(
          "Missing database URL.",
          "INVALID_ARGUMENTS",
          "Provide --database-url or set TWENTY_DATABASE_URL.",
        );
      }

      const profile = await services.dbProfiles.initProfile({
        workspace: globalOptions.workspace,
        name,
        databaseUrl,
        notes: options.notes,
      });

      await services.output.render(profile, {
        format: globalOptions.output,
        query: globalOptions.query,
      });
    });
  });

  registerCommand(profileCmd, "list", "List db profiles", (command) => {
    applyGlobalOptions(command);
    command.action(async (_options: unknown, actionCommand: Command) => {
      const { globalOptions, services } = createCommandContext(actionCommand);
      const profiles = await services.dbProfiles.listProfiles(globalOptions.workspace);

      await services.output.render(profiles, {
        format: globalOptions.output,
        query: globalOptions.query,
      });
    });
  });

  registerCommand(profileCmd, "show", "Show a db profile", (command) => {
    command.argument("[name]", "Profile name");
    applyGlobalOptions(command);
    command.action(async (name: string | undefined, _options: unknown, actionCommand: Command) => {
      const { globalOptions, services } = createCommandContext(actionCommand);
      const envName = process.env.TWENTY_DB_PROFILE;
      const resolvedStatus =
        name || envName
          ? undefined
          : await services.dbStatus.getStatus({ workspace: globalOptions.workspace });
      const resolvedName = resolveProfileName(name, envName, resolvedStatus?.profileName);

      if (!resolvedName) {
        throw new CliError(
          "No DB profile selected.",
          "INVALID_ARGUMENTS",
          'Use "twenty db profile use <name>" or set TWENTY_DB_PROFILE.',
        );
      }

      const profile = await services.dbProfiles.getProfile(globalOptions.workspace, resolvedName);

      await services.output.render(profile, {
        format: globalOptions.output,
        query: globalOptions.query,
      });
    });
  });

  registerCommand(profileCmd, "use", "Set the active db profile", (command) => {
    command.argument("<name>", "Profile name");
    applyGlobalOptions(command);
    command.action(async (name: string, _options: unknown, actionCommand: Command) => {
      const { globalOptions, services } = createCommandContext(actionCommand);
      const profile = await services.dbProfiles.setActiveProfile(globalOptions.workspace, name);

      await services.output.render(profile, {
        format: globalOptions.output,
        query: globalOptions.query,
      });
    });
  });

  registerCommand(profileCmd, "test", "Test a db profile", (command) => {
    command.argument("[name]", "Profile name");
    applyGlobalOptions(command);
    command.action(async (name: string | undefined, _options: unknown, actionCommand: Command) => {
      const { globalOptions, services } = createCommandContext(actionCommand);
      const resolved = await resolveDbProfileSelection(services, globalOptions.workspace, name);
      const profile = await services.dbProfiles.test(globalOptions.workspace, resolved);

      await services.output.render(profile, {
        format: globalOptions.output,
        query: globalOptions.query,
      });
    });
  });

  registerCommand(profileCmd, "refresh-creds", "Refresh db credentials", (command) => {
    command.argument("[name]", "Profile name");
    applyGlobalOptions(command);
    command.action(async (name: string | undefined, _options: unknown, actionCommand: Command) => {
      const { globalOptions, services } = createCommandContext(actionCommand);
      const resolved = await resolveDbProfileSelection(services, globalOptions.workspace, name);
      const profile = await services.dbProfiles.refreshCreds(globalOptions.workspace, resolved);

      await services.output.render(profile, {
        format: globalOptions.output,
        query: globalOptions.query,
      });
    });
  });

  registerCommand(profileCmd, "remove", "Remove a db profile", (command) => {
    command.argument("<name>", "Profile name");
    applyGlobalOptions(command);
    command.action(async (name: string, _options: unknown, actionCommand: Command) => {
      const { globalOptions, services } = createCommandContext(actionCommand);
      await services.dbProfiles.removeProfile(globalOptions.workspace, name);

      await services.output.render(
        {
          workspace: globalOptions.workspace,
          name,
          removed: true,
        },
        {
          format: globalOptions.output,
          query: globalOptions.query,
        },
      );
    });
  });
}

async function resolveDbProfileSelection(
  services: ReturnType<typeof createCommandContext>["services"],
  workspace: string | undefined,
  explicitName: string | undefined,
): Promise<string> {
  if (explicitName) {
    return explicitName;
  }

  const envName = process.env.TWENTY_DB_PROFILE;
  if (envName) {
    return envName;
  }

  const status = await services.dbStatus.getStatus({ workspace });
  if (status.profileName) {
    return status.profileName;
  }

  throw new CliError(
    "No DB profile selected.",
    "INVALID_ARGUMENTS",
    'Use "twenty db profile use <name>" or set TWENTY_DB_PROFILE.',
  );
}
