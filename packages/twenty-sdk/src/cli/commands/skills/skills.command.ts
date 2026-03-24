import { Command } from "commander";
import { type GraphQLResponse } from "../../utilities/api/graphql-response";
import { CliError } from "../../utilities/errors/cli-error";
import { parseBody } from "../../utilities/shared/body";
import { applyGlobalOptions, resolveGlobalOptions } from "../../utilities/shared/global-options";
import { createServices } from "../../utilities/shared/services";

interface SkillsOptions {
  data?: string;
  file?: string;
  set?: string[];
}

const endpoint = "/graphql";

const SKILL_FIELDS = `
  id
  standardId
  name
  label
  icon
  description
  content
  isCustom
  isActive
  applicationId
  createdAt
  updatedAt
`;

const LIST_SKILLS_QUERY = `query ListSkills {
  skills {
    ${SKILL_FIELDS}
  }
}`;

const GET_SKILL_QUERY = `query GetSkill($id: UUID!) {
  skill(id: $id) {
    ${SKILL_FIELDS}
  }
}`;

const CREATE_SKILL_MUTATION = `mutation CreateSkill($input: CreateSkillInput!) {
  createSkill(input: $input) {
    ${SKILL_FIELDS}
  }
}`;

const UPDATE_SKILL_MUTATION = `mutation UpdateSkill($input: UpdateSkillInput!) {
  updateSkill(input: $input) {
    ${SKILL_FIELDS}
  }
}`;

const DELETE_SKILL_MUTATION = `mutation DeleteSkill($id: UUID!) {
  deleteSkill(id: $id) {
    ${SKILL_FIELDS}
  }
}`;

const ACTIVATE_SKILL_MUTATION = `mutation ActivateSkill($id: UUID!) {
  activateSkill(id: $id) {
    ${SKILL_FIELDS}
  }
}`;

const DEACTIVATE_SKILL_MUTATION = `mutation DeactivateSkill($id: UUID!) {
  deactivateSkill(id: $id) {
    ${SKILL_FIELDS}
  }
}`;

function collect(value: string, previous: string[] = []): string[] {
  return previous.concat([value]);
}

function requireSkillId(id: string | undefined): string {
  if (!id) {
    throw new CliError("Missing skill ID.", "INVALID_ARGUMENTS");
  }

  return id;
}

export function registerSkillsCommand(program: Command): void {
  const cmd = program
    .command("skills")
    .description("Manage workspace AI skills")
    .argument("<operation>", "list, get, create, update, delete, activate, deactivate")
    .argument("[id]", "Skill ID")
    .option("-d, --data <json>", "JSON payload")
    .option("-f, --file <path>", "JSON file")
    .option("--set <key=value>", "Set a field value", collect);

  applyGlobalOptions(cmd);

  cmd.action(
    async (operation: string, id: string | undefined, options: SkillsOptions, command: Command) => {
      const globalOptions = resolveGlobalOptions(command);
      const services = createServices(globalOptions);
      const op = operation.toLowerCase();

      switch (op) {
        case "list": {
          const response = await services.api.post<GraphQLResponse<{ skills?: unknown[] }>>(
            endpoint,
            {
              query: LIST_SKILLS_QUERY,
            },
          );

          await services.output.render(response.data?.data?.skills ?? [], {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "get": {
          const skillId = requireSkillId(id);
          const response = await services.api.post<GraphQLResponse<{ skill?: unknown }>>(endpoint, {
            query: GET_SKILL_QUERY,
            variables: { id: skillId },
          });

          await services.output.render(response.data?.data?.skill, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "create": {
          const payload = await parseBody(options.data, options.file, options.set);
          const response = await services.api.post<GraphQLResponse<{ createSkill?: unknown }>>(
            endpoint,
            {
              query: CREATE_SKILL_MUTATION,
              variables: { input: payload },
            },
          );

          await services.output.render(response.data?.data?.createSkill, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "update": {
          const skillId = requireSkillId(id);
          const payload = await parseBody(options.data, options.file, options.set);
          const response = await services.api.post<GraphQLResponse<{ updateSkill?: unknown }>>(
            endpoint,
            {
              query: UPDATE_SKILL_MUTATION,
              variables: {
                input: {
                  id: skillId,
                  ...payload,
                },
              },
            },
          );

          await services.output.render(response.data?.data?.updateSkill, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "delete": {
          const skillId = requireSkillId(id);
          const response = await services.api.post<GraphQLResponse<{ deleteSkill?: unknown }>>(
            endpoint,
            {
              query: DELETE_SKILL_MUTATION,
              variables: { id: skillId },
            },
          );

          await services.output.render(response.data?.data?.deleteSkill, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "activate": {
          const skillId = requireSkillId(id);
          const response = await services.api.post<GraphQLResponse<{ activateSkill?: unknown }>>(
            endpoint,
            {
              query: ACTIVATE_SKILL_MUTATION,
              variables: { id: skillId },
            },
          );

          await services.output.render(response.data?.data?.activateSkill, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "deactivate": {
          const skillId = requireSkillId(id);
          const response = await services.api.post<GraphQLResponse<{ deactivateSkill?: unknown }>>(
            endpoint,
            {
              query: DEACTIVATE_SKILL_MUTATION,
              variables: { id: skillId },
            },
          );

          await services.output.render(response.data?.data?.deactivateSkill, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        default:
          throw new CliError(`Unknown operation: ${operation}`, "INVALID_ARGUMENTS");
      }
    },
  );
}
