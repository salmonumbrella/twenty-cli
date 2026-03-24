import { Command } from "commander";
import { type GraphQLResponse } from "../../utilities/api/graphql-response";
import { CliError } from "../../utilities/errors/cli-error";
import { parseBody } from "../../utilities/shared/body";
import { applyGlobalOptions, resolveGlobalOptions } from "../../utilities/shared/global-options";
import { createServices } from "../../utilities/shared/services";

interface RolesOptions {
  data?: string;
  file?: string;
  set?: string[];
  roleId?: string;
  includeTargets?: boolean;
  includePermissions?: boolean;
}

const ROLE_FIELDS = `
  id
  label
  description
  icon
  isEditable
  canUpdateAllSettings
  canAccessAllTools
  canReadAllObjectRecords
  canUpdateAllObjectRecords
  canSoftDeleteAllObjectRecords
  canDestroyAllObjectRecords
  canBeAssignedToUsers
  canBeAssignedToAgents
  canBeAssignedToApiKeys
`;

const ROLE_TARGET_FIELDS = `
  workspaceMembers {
    id
    name {
      firstName
      lastName
    }
    userEmail
    avatarUrl
  }
  agents {
    id
    name
    label
    icon
    description
    roleId
    isCustom
    createdAt
    updatedAt
  }
  apiKeys {
    id
    name
    expiresAt
    revokedAt
  }
`;

const ROLE_PERMISSION_FIELDS = `
  permissionFlags {
    id
    roleId
    flag
  }
  objectPermissions {
    objectMetadataId
    canReadObjectRecords
    canUpdateObjectRecords
    canSoftDeleteObjectRecords
    canDestroyObjectRecords
    restrictedFields
  }
  fieldPermissions {
    id
    roleId
    objectMetadataId
    fieldMetadataId
    canReadFieldValue
    canUpdateFieldValue
  }
`;

const CREATE_ROLE_MUTATION = `mutation CreateOneRole($createRoleInput: CreateRoleInput!) {
  createOneRole(createRoleInput: $createRoleInput) {
    ${ROLE_FIELDS}
  }
}`;

const UPDATE_ROLE_MUTATION = `mutation UpdateOneRole($updateRoleInput: UpdateRoleInput!) {
  updateOneRole(updateRoleInput: $updateRoleInput) {
    ${ROLE_FIELDS}
  }
}`;

const DELETE_ROLE_MUTATION = `mutation DeleteOneRole($roleId: UUID!) {
  deleteOneRole(roleId: $roleId)
}`;

const UPSERT_PERMISSION_FLAGS_MUTATION = `mutation UpsertPermissionFlags($upsertPermissionFlagsInput: UpsertPermissionFlagsInput!) {
  upsertPermissionFlags(upsertPermissionFlagsInput: $upsertPermissionFlagsInput) {
    id
    roleId
    flag
  }
}`;

const UPSERT_OBJECT_PERMISSIONS_MUTATION = `mutation UpsertObjectPermissions($upsertObjectPermissionsInput: UpsertObjectPermissionsInput!) {
  upsertObjectPermissions(upsertObjectPermissionsInput: $upsertObjectPermissionsInput) {
    objectMetadataId
    canReadObjectRecords
    canUpdateObjectRecords
    canSoftDeleteObjectRecords
    canDestroyObjectRecords
    restrictedFields
  }
}`;

const UPSERT_FIELD_PERMISSIONS_MUTATION = `mutation UpsertFieldPermissions($upsertFieldPermissionsInput: UpsertFieldPermissionsInput!) {
  upsertFieldPermissions(upsertFieldPermissionsInput: $upsertFieldPermissionsInput) {
    id
    roleId
    objectMetadataId
    fieldMetadataId
    canReadFieldValue
    canUpdateFieldValue
  }
}`;

const ASSIGN_ROLE_TO_AGENT_MUTATION = `mutation AssignRoleToAgent($agentId: UUID!, $roleId: UUID!) {
  assignRoleToAgent(agentId: $agentId, roleId: $roleId)
}`;

const REMOVE_ROLE_FROM_AGENT_MUTATION = `mutation RemoveRoleFromAgent($agentId: UUID!) {
  removeRoleFromAgent(agentId: $agentId)
}`;

function collect(value: string, previous: string[] = []): string[] {
  return previous.concat([value]);
}

function buildRoleSelection(options: RolesOptions): string {
  return [
    ROLE_FIELDS,
    options.includeTargets ? ROLE_TARGET_FIELDS : "",
    options.includePermissions ? ROLE_PERMISSION_FIELDS : "",
  ]
    .filter(Boolean)
    .join("\n");
}

function buildGetRolesQuery(options: RolesOptions): string {
  return `query GetRoles {
  getRoles {
    ${buildRoleSelection(options)}
  }
}`;
}

function findRoleById(roles: unknown[], id: string): unknown {
  return roles.find((role) => {
    if (typeof role !== "object" || role === null) {
      return false;
    }

    const candidate = role as { id?: unknown };

    return candidate.id === id;
  });
}

export function registerRolesCommand(program: Command): void {
  const cmd = program
    .command("roles")
    .description("Manage workspace roles")
    .argument(
      "<operation>",
      "list, get, create, update, delete, upsert-permission-flags, upsert-object-permissions, upsert-field-permissions, assign-agent, remove-agent",
    )
    .argument("[id]", "Role ID")
    .option("-d, --data <json>", "JSON payload")
    .option("-f, --file <path>", "JSON file")
    .option("--set <key=value>", "Set a field value", collect)
    .option("--role-id <id>", "Role ID for assignment operations")
    .option(
      "--include-targets",
      "Include workspace members, agents, and API keys assigned to each role",
    )
    .option(
      "--include-permissions",
      "Include permission flag, object permission, and field permission details",
    );

  applyGlobalOptions(cmd);

  cmd.action(
    async (operation: string, id: string | undefined, options: RolesOptions, command: Command) => {
      const globalOptions = resolveGlobalOptions(command);
      const services = createServices(globalOptions);
      const op = operation.toLowerCase();

      switch (op) {
        case "list": {
          const response = await services.api.post<GraphQLResponse<{ getRoles?: unknown[] }>>(
            "/graphql",
            {
              query: buildGetRolesQuery(options),
            },
          );

          await services.output.render(response.data?.data?.getRoles ?? [], {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "get": {
          if (!id) {
            throw new CliError("Missing role ID.", "INVALID_ARGUMENTS");
          }

          const response = await services.api.post<GraphQLResponse<{ getRoles?: unknown[] }>>(
            "/graphql",
            {
              query: buildGetRolesQuery(options),
            },
          );

          const role = findRoleById(response.data?.data?.getRoles ?? [], id);

          if (!role) {
            throw new CliError(`Role ${id} not found.`, "NOT_FOUND");
          }

          await services.output.render(role, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "create": {
          const payload = await parseBody(options.data, options.file, options.set);
          const response = await services.api.post<GraphQLResponse<{ createOneRole?: unknown }>>(
            "/graphql",
            {
              query: CREATE_ROLE_MUTATION,
              variables: {
                createRoleInput: payload,
              },
            },
          );

          await services.output.render(response.data?.data?.createOneRole, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "update": {
          if (!id) {
            throw new CliError("Missing role ID.", "INVALID_ARGUMENTS");
          }

          const payload = await parseBody(options.data, options.file, options.set);
          const response = await services.api.post<GraphQLResponse<{ updateOneRole?: unknown }>>(
            "/graphql",
            {
              query: UPDATE_ROLE_MUTATION,
              variables: {
                updateRoleInput: {
                  id,
                  update: payload,
                },
              },
            },
          );

          await services.output.render(response.data?.data?.updateOneRole, {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "delete": {
          if (!id) {
            throw new CliError("Missing role ID.", "INVALID_ARGUMENTS");
          }

          const response = await services.api.post<GraphQLResponse<{ deleteOneRole?: string }>>(
            "/graphql",
            {
              query: DELETE_ROLE_MUTATION,
              variables: {
                roleId: id,
              },
            },
          );

          await services.output.render(
            {
              id: response.data?.data?.deleteOneRole ?? id,
            },
            {
              format: globalOptions.output,
              query: globalOptions.query,
            },
          );
          break;
        }
        case "upsert-permission-flags": {
          const payload = await parseBody(options.data, options.file, options.set);
          const response = await services.api.post<
            GraphQLResponse<{ upsertPermissionFlags?: unknown[] }>
          >("/graphql", {
            query: UPSERT_PERMISSION_FLAGS_MUTATION,
            variables: {
              upsertPermissionFlagsInput: payload,
            },
          });

          await services.output.render(response.data?.data?.upsertPermissionFlags ?? [], {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "upsert-object-permissions": {
          const payload = await parseBody(options.data, options.file, options.set);
          const response = await services.api.post<
            GraphQLResponse<{ upsertObjectPermissions?: unknown[] }>
          >("/graphql", {
            query: UPSERT_OBJECT_PERMISSIONS_MUTATION,
            variables: {
              upsertObjectPermissionsInput: payload,
            },
          });

          await services.output.render(response.data?.data?.upsertObjectPermissions ?? [], {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "upsert-field-permissions": {
          const payload = await parseBody(options.data, options.file, options.set);
          const response = await services.api.post<
            GraphQLResponse<{ upsertFieldPermissions?: unknown[] }>
          >("/graphql", {
            query: UPSERT_FIELD_PERMISSIONS_MUTATION,
            variables: {
              upsertFieldPermissionsInput: payload,
            },
          });

          await services.output.render(response.data?.data?.upsertFieldPermissions ?? [], {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "assign-agent": {
          if (!id) {
            throw new CliError("Missing agent ID.", "INVALID_ARGUMENTS");
          }
          if (!options.roleId) {
            throw new CliError("Missing --role-id option.", "INVALID_ARGUMENTS");
          }

          const response = await services.api.post<
            GraphQLResponse<{ assignRoleToAgent?: boolean }>
          >("/graphql", {
            query: ASSIGN_ROLE_TO_AGENT_MUTATION,
            variables: {
              agentId: id,
              roleId: options.roleId,
            },
          });

          await services.output.render(
            {
              agentId: id,
              roleId: options.roleId,
              assigned: response.data?.data?.assignRoleToAgent ?? false,
            },
            {
              format: globalOptions.output,
              query: globalOptions.query,
            },
          );
          break;
        }
        case "remove-agent": {
          if (!id) {
            throw new CliError("Missing agent ID.", "INVALID_ARGUMENTS");
          }

          const response = await services.api.post<
            GraphQLResponse<{ removeRoleFromAgent?: boolean }>
          >("/graphql", {
            query: REMOVE_ROLE_FROM_AGENT_MUTATION,
            variables: {
              agentId: id,
            },
          });

          await services.output.render(
            {
              agentId: id,
              removed: response.data?.data?.removeRoleFromAgent ?? false,
            },
            {
              format: globalOptions.output,
              query: globalOptions.query,
            },
          );
          break;
        }
        default:
          throw new CliError(`Unknown operation: ${operation}`, "INVALID_ARGUMENTS");
      }
    },
  );
}
