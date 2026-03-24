import fs from "fs-extra";
import FormData from "form-data";
import path from "path";
import { Command } from "commander";
import { type GraphQLResponse } from "../../utilities/api/graphql-response";
import { CliError } from "../../utilities/errors/cli-error";
import { applyGlobalOptions, resolveGlobalOptions } from "../../utilities/shared/global-options";
import { createServices } from "../../utilities/shared/services";

interface FilesOptions {
  outputFile?: string;
  target?: string;
  folder?: string;
  token?: string;
  universalIdentifier?: string;
  workspaceId?: string;
  applicationId?: string;
  applicationUniversalIdentifier?: string;
  fileFolder?: string;
  filePath?: string;
  fieldMetadataId?: string;
  fieldMetadataUniversalIdentifier?: string;
}

const endpoint = "/metadata";
const SUPPORTED_DOWNLOAD_FOLDERS = [
  "core-picture",
  "files-field",
  "workflow",
  "agent-chat",
  "app-tarball",
] as const;
const UPLOAD_TARGETS = [
  "ai-chat",
  "workflow",
  "field",
  "workspace-logo",
  "profile-picture",
  "app-tarball",
  "application-file",
] as const;
const APPLICATION_FILE_FOLDERS = {
  "built-logic-function": "BuiltLogicFunction",
  "built-front-component": "BuiltFrontComponent",
  "public-asset": "PublicAsset",
  source: "Source",
  dependencies: "Dependencies",
} as const;

type UploadTarget = (typeof UPLOAD_TARGETS)[number];

function normalizeUploadTarget(target?: string): UploadTarget {
  const normalized = target?.toLowerCase();

  if (!normalized) {
    throw new CliError(
      `Missing --target option. Expected one of: ${UPLOAD_TARGETS.join(", ")}.`,
      "INVALID_ARGUMENTS",
    );
  }

  if (!UPLOAD_TARGETS.includes(normalized as UploadTarget)) {
    throw new CliError(
      `Unsupported --target "${target}". Expected one of: ${UPLOAD_TARGETS.join(", ")}.`,
      "INVALID_ARGUMENTS",
    );
  }

  return normalized as UploadTarget;
}

function normalizeDownloadFolder(folder?: string): (typeof SUPPORTED_DOWNLOAD_FOLDERS)[number] {
  const normalized = folder?.toLowerCase();

  if (!normalized) {
    throw new CliError(
      "Download by file ID requires --folder and --token. You can also pass a full signed URL instead of an ID.",
      "INVALID_ARGUMENTS",
    );
  }

  if (
    !SUPPORTED_DOWNLOAD_FOLDERS.includes(normalized as (typeof SUPPORTED_DOWNLOAD_FOLDERS)[number])
  ) {
    throw new CliError(
      `Unsupported --folder "${folder}". Expected one of: ${SUPPORTED_DOWNLOAD_FOLDERS.join(", ")}.`,
      "INVALID_ARGUMENTS",
    );
  }

  return normalized as (typeof SUPPORTED_DOWNLOAD_FOLDERS)[number];
}

function normalizeApplicationFileFolder(
  folder?: string,
): (typeof APPLICATION_FILE_FOLDERS)[keyof typeof APPLICATION_FILE_FOLDERS] {
  const normalized = folder?.toLowerCase();

  if (!normalized) {
    throw new CliError(
      `Application file uploads require --file-folder. Expected one of: ${Object.keys(APPLICATION_FILE_FOLDERS).join(", ")}.`,
      "INVALID_ARGUMENTS",
    );
  }

  const graphqlFolder =
    APPLICATION_FILE_FOLDERS[normalized as keyof typeof APPLICATION_FILE_FOLDERS];

  if (!graphqlFolder) {
    throw new CliError(
      `Unsupported --file-folder "${folder}". Expected one of: ${Object.keys(APPLICATION_FILE_FOLDERS).join(", ")}.`,
      "INVALID_ARGUMENTS",
    );
  }

  return graphqlFolder;
}

function buildGraphqlUploadForm(
  query: string,
  variables: Record<string, unknown>,
  filePath: string,
): FormData {
  const form = new FormData();

  form.append(
    "operations",
    JSON.stringify({
      query,
      variables: {
        ...variables,
        file: null,
      },
    }),
  );
  form.append("map", JSON.stringify({ 0: ["variables.file"] }));
  form.append("0", fs.createReadStream(filePath), path.basename(filePath));

  return form;
}

function inferOutputPath(reference: string): string {
  let pathname = reference;

  try {
    pathname = new URL(reference).pathname;
  } catch {
    pathname = reference;
  }

  const cleanPath = pathname.split("?")[0] ?? pathname;
  const fileName = path.basename(cleanPath);

  return fileName || "download";
}

function buildDownloadUrl(pathOrId: string, options: FilesOptions): string {
  if (
    pathOrId.startsWith("http://") ||
    pathOrId.startsWith("https://") ||
    pathOrId.startsWith("/file/") ||
    pathOrId.startsWith("/public-assets/")
  ) {
    return pathOrId;
  }

  const folder = normalizeDownloadFolder(options.folder);

  if (!options.token) {
    throw new CliError(
      "Download by file ID requires --token. You can also pass a full signed URL instead of an ID.",
      "INVALID_ARGUMENTS",
    );
  }

  return `/file/${folder}/${pathOrId}?token=${encodeURIComponent(options.token)}`;
}

function buildPublicAssetUrl(assetPath: string, options: FilesOptions): string {
  if (!options.workspaceId) {
    throw new CliError("Missing --workspace-id option.", "INVALID_ARGUMENTS");
  }

  if (!options.applicationId) {
    throw new CliError("Missing --application-id option.", "INVALID_ARGUMENTS");
  }

  const normalizedPath = assetPath.replace(/^\/+/, "");

  return `/public-assets/${options.workspaceId}/${options.applicationId}/${normalizedPath}`;
}

function buildUploadMutation(
  target: UploadTarget,
  options: FilesOptions,
): {
  query: string;
  variables: Record<string, unknown>;
  resultKey: string;
} {
  switch (target) {
    case "ai-chat":
      return {
        query: `mutation UploadAIChatFile($file: Upload!) { uploadAIChatFile(file: $file) { id path size createdAt url } }`,
        variables: {},
        resultKey: "uploadAIChatFile",
      };
    case "workflow":
      return {
        query: `mutation UploadWorkflowFile($file: Upload!) { uploadWorkflowFile(file: $file) { id path size createdAt url } }`,
        variables: {},
        resultKey: "uploadWorkflowFile",
      };
    case "workspace-logo":
      return {
        query: `mutation UploadWorkspaceLogo($file: Upload!) { uploadWorkspaceLogo(file: $file) { url } }`,
        variables: {},
        resultKey: "uploadWorkspaceLogo",
      };
    case "profile-picture":
      return {
        query: `mutation UploadWorkspaceMemberProfilePicture($file: Upload!) { uploadWorkspaceMemberProfilePicture(file: $file) { url } }`,
        variables: {},
        resultKey: "uploadWorkspaceMemberProfilePicture",
      };
    case "app-tarball":
      return {
        query: `mutation UploadAppTarball($file: Upload!, $universalIdentifier: String) {
          uploadAppTarball(file: $file, universalIdentifier: $universalIdentifier) {
            id
            universalIdentifier
            name
          }
        }`,
        variables: {
          universalIdentifier: options.universalIdentifier ?? null,
        },
        resultKey: "uploadAppTarball",
      };
    case "application-file": {
      if (!options.applicationUniversalIdentifier) {
        throw new CliError(
          "Application file uploads require --application-universal-identifier.",
          "INVALID_ARGUMENTS",
        );
      }

      if (!options.filePath) {
        throw new CliError("Application file uploads require --file-path.", "INVALID_ARGUMENTS");
      }

      return {
        query: `mutation UploadApplicationFile(
          $file: Upload!
          $applicationUniversalIdentifier: String!
          $fileFolder: FileFolder!
          $filePath: String!
        ) {
          uploadApplicationFile(
            file: $file
            applicationUniversalIdentifier: $applicationUniversalIdentifier
            fileFolder: $fileFolder
            filePath: $filePath
          ) {
            id
            path
            size
            createdAt
            url
          }
        }`,
        variables: {
          applicationUniversalIdentifier: options.applicationUniversalIdentifier,
          fileFolder: normalizeApplicationFileFolder(options.fileFolder),
          filePath: options.filePath,
        },
        resultKey: "uploadApplicationFile",
      };
    }
    case "field": {
      if (options.fieldMetadataId && options.fieldMetadataUniversalIdentifier) {
        throw new CliError(
          "Provide only one of --field-metadata-id or --field-metadata-universal-identifier.",
          "INVALID_ARGUMENTS",
        );
      }

      if (options.fieldMetadataId) {
        return {
          query: `mutation UploadFilesFieldFile($file: Upload!, $fieldMetadataId: String!) { uploadFilesFieldFile(file: $file, fieldMetadataId: $fieldMetadataId) { id path size createdAt url } }`,
          variables: { fieldMetadataId: options.fieldMetadataId },
          resultKey: "uploadFilesFieldFile",
        };
      }

      if (options.fieldMetadataUniversalIdentifier) {
        return {
          query: `mutation UploadFilesFieldFileByUniversalIdentifier($file: Upload!, $fieldMetadataUniversalIdentifier: String!) { uploadFilesFieldFileByUniversalIdentifier(file: $file, fieldMetadataUniversalIdentifier: $fieldMetadataUniversalIdentifier) { id path size createdAt url } }`,
          variables: {
            fieldMetadataUniversalIdentifier: options.fieldMetadataUniversalIdentifier,
          },
          resultKey: "uploadFilesFieldFileByUniversalIdentifier",
        };
      }

      throw new CliError(
        "Field uploads require --field-metadata-id or --field-metadata-universal-identifier.",
        "INVALID_ARGUMENTS",
      );
    }
  }
}

export function registerFilesCommand(program: Command): void {
  const cmd = program
    .command("files")
    .description("Upload and download files through verified Twenty file APIs")
    .argument("<operation>", "upload, download, public-asset")
    .argument("[path-or-id]", "Local file path, signed URL, file ID, or public asset path")
    .option("--output-file <path>", "Output file path (download/public-asset)")
    .option(
      "--target <target>",
      "Upload target: ai-chat, workflow, field, workspace-logo, profile-picture, app-tarball, application-file",
    )
    .option(
      "--folder <folder>",
      "Signed file folder: core-picture, files-field, workflow, agent-chat, app-tarball",
    )
    .option("--token <token>", "Signed file token for /file downloads")
    .option("--universal-identifier <id>", "Optional universal identifier for app tarball uploads")
    .option("--workspace-id <id>", "Workspace ID for public asset downloads")
    .option("--application-id <id>", "Application ID for public asset downloads")
    .option(
      "--application-universal-identifier <id>",
      "Application universal identifier for application-file uploads",
    )
    .option(
      "--file-folder <folder>",
      "Application file folder: built-logic-function, built-front-component, public-asset, source, dependencies",
    )
    .option("--file-path <path>", "Remote application file path for application-file uploads")
    .option("--field-metadata-id <id>", "Field metadata ID for field uploads")
    .option(
      "--field-metadata-universal-identifier <id>",
      "Field metadata universal identifier for field uploads",
    );

  applyGlobalOptions(cmd);

  cmd.action(
    async (
      operation: string,
      pathOrId: string | undefined,
      options: FilesOptions,
      command: Command,
    ) => {
      const globalOptions = resolveGlobalOptions(command);
      const services = createServices(globalOptions);
      const op = operation.toLowerCase();

      switch (op) {
        case "upload": {
          if (!pathOrId) throw new CliError("Missing file path.", "INVALID_ARGUMENTS");
          if (!(await fs.pathExists(pathOrId))) {
            throw new CliError(`File not found: ${pathOrId}`, "INVALID_ARGUMENTS");
          }

          const target = normalizeUploadTarget(options.target);
          const { query, variables, resultKey } = buildUploadMutation(target, options);
          const form = buildGraphqlUploadForm(query, variables, pathOrId);
          const response = await services.api.post<GraphQLResponse<Record<string, unknown>>>(
            endpoint,
            form,
            {
              headers: form.getHeaders(),
            },
          );

          await services.output.render(response.data?.data?.[resultKey], {
            format: globalOptions.output,
            query: globalOptions.query,
          });
          break;
        }
        case "download": {
          if (!pathOrId) {
            throw new CliError("Missing file ID or signed URL.", "INVALID_ARGUMENTS");
          }

          const downloadUrl = buildDownloadUrl(pathOrId, options);
          const outputPath = options.outputFile || inferOutputPath(downloadUrl);
          const response = await services.api.get<ArrayBuffer>(downloadUrl, {
            responseType: "arraybuffer",
          });

          await fs.writeFile(outputPath, Buffer.from(response.data));
          // eslint-disable-next-line no-console
          console.log(`Downloaded to ${outputPath}`);
          break;
        }
        case "public-asset": {
          if (!pathOrId) {
            throw new CliError("Missing public asset path.", "INVALID_ARGUMENTS");
          }

          const assetUrl = buildPublicAssetUrl(pathOrId, options);
          const outputPath = options.outputFile || inferOutputPath(pathOrId);
          const response = await services.api.get<ArrayBuffer>(assetUrl, {
            responseType: "arraybuffer",
          });

          await fs.writeFile(outputPath, Buffer.from(response.data));
          // eslint-disable-next-line no-console
          console.log(`Downloaded to ${outputPath}`);
          break;
        }
        case "list":
          throw new CliError(
            "Current Twenty does not expose a public generic file list API. Use domain-specific uploads and signed or public download URLs instead.",
            "INVALID_ARGUMENTS",
          );
        case "delete":
          throw new CliError(
            "Current Twenty does not expose a public generic file delete API.",
            "INVALID_ARGUMENTS",
          );
        default:
          throw new CliError(`Unknown operation: ${operation}`, "INVALID_ARGUMENTS");
      }
    },
  );
}
