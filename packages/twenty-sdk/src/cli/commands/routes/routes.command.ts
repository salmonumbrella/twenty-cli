import axios from "axios";
import { Command } from "commander";
import { ConfigService } from "../../utilities/config/services/config.service";
import { CliError } from "../../utilities/errors/cli-error";
import { applyGlobalOptions } from "../../utilities/shared/global-options";
import { readJsonInput } from "../../utilities/shared/io";
import { parseKeyValuePairs } from "../../utilities/shared/parse";
import { createOutputContext } from "../../utilities/shared/context";

interface RouteInvokeOptions {
  method?: string;
  data?: string;
  file?: string;
  param?: string[];
  header?: string[];
}

interface PublicConnection {
  apiUrl: string;
  apiKey?: string;
}

type RouteMethod = "delete" | "get" | "patch" | "post" | "put";

export function registerRoutesCommand(program: Command): void {
  const routesCmd = program.command("routes").description("Invoke public route trigger endpoints");

  const invokeCmd = routesCmd
    .command("invoke")
    .description("Invoke a public /s/* route endpoint")
    .argument("<routePath>", "Route path relative to /s")
    .option("--method <method>", "Route method: get, post, put, patch, or delete", "get")
    .option("-d, --data <json>", "JSON payload for non-GET requests")
    .option("-f, --file <path>", "JSON payload file for non-GET requests")
    .option("--param <key=value>", "Query parameter", collect)
    .option("--header <key=value>", "Request header", collect);

  applyGlobalOptions(invokeCmd);

  invokeCmd.action(async (routePath: string, options: RouteInvokeOptions, command: Command) => {
    const { globalOptions, output } = createOutputContext(command);
    const method = normalizeMethod(options.method);
    const connection = await resolvePublicConnection(globalOptions.workspace);
    const payload = await readPayload(method, options);
    const params = normalizeQueryParams(parseKeyValuePairs(options.param));
    const headers = buildHeaders(connection.apiKey, options.header);

    const response = await axios.request({
      method,
      url: buildRouteUrl(connection.apiUrl, routePath),
      data: method === "get" ? undefined : payload,
      params: Object.keys(params).length > 0 ? params : undefined,
      headers,
    });

    await output.render(response.data, {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });
}

function collect(value: string, previous: string[] = []): string[] {
  return previous.concat([value]);
}

function normalizeMethod(rawMethod: string | undefined): RouteMethod {
  const method = rawMethod?.toLowerCase() ?? "get";

  if (!["delete", "get", "patch", "post", "put"].includes(method)) {
    throw new CliError(
      "Unsupported route method. Use GET, POST, PUT, PATCH, or DELETE.",
      "INVALID_ARGUMENTS",
    );
  }

  return method as RouteMethod;
}

async function resolvePublicConnection(workspace: string | undefined): Promise<PublicConnection> {
  const configService = new ConfigService();
  const config = await configService.loadConfigFile();
  const workspaceName =
    workspace ?? process.env.TWENTY_PROFILE ?? config?.defaultWorkspace ?? "default";
  const workspaceConfig = config?.workspaces?.[workspaceName] ?? {};

  return {
    apiUrl: process.env.TWENTY_BASE_URL ?? workspaceConfig.apiUrl ?? "https://api.twenty.com",
    apiKey: process.env.TWENTY_TOKEN ?? workspaceConfig.apiKey ?? undefined,
  };
}

function buildRouteUrl(apiUrl: string, routePath: string): string {
  const trimmedPath = routePath.trim();
  if (!trimmedPath) {
    throw new CliError("Route path must not be empty.", "INVALID_ARGUMENTS");
  }

  const normalizedBaseUrl = apiUrl.replace(/\/+$/, "");
  const normalizedPath = trimmedPath.replace(/^\/+/, "");

  if (normalizedPath === "s" || normalizedPath.startsWith("s/")) {
    return `${normalizedBaseUrl}/${normalizedPath}`;
  }

  return `${normalizedBaseUrl}/s/${normalizedPath}`;
}

function normalizeQueryParams(params: Record<string, string[]>): Record<string, string | string[]> {
  return Object.fromEntries(
    Object.entries(params).map(([key, values]) => [key, values.length === 1 ? values[0] : values]),
  );
}

async function readPayload(
  method: RouteMethod,
  options: RouteInvokeOptions,
): Promise<Record<string, unknown>> {
  if (method === "get") {
    if (options.data || options.file) {
      throw new CliError(
        "GET route invocations do not accept --data or --file. Use --param or a mutating method instead.",
        "INVALID_ARGUMENTS",
      );
    }

    return {};
  }

  const payload = await readJsonInput(options.data, options.file);
  if (payload == null) {
    return {};
  }

  if (typeof payload !== "object" || Array.isArray(payload)) {
    throw new CliError("Route payload must be a JSON object.", "INVALID_ARGUMENTS");
  }

  return payload as Record<string, unknown>;
}

function buildHeaders(
  apiKey: string | undefined,
  rawHeaders: string[] | undefined,
): Record<string, string> | undefined {
  const headers = normalizeStringMap(parseKeyValuePairs(rawHeaders));
  if (apiKey) {
    headers.Authorization = `Bearer ${apiKey}`;
  }

  return Object.keys(headers).length > 0 ? headers : undefined;
}

function normalizeStringMap(entries: Record<string, string[]>): Record<string, string> {
  return Object.fromEntries(
    Object.entries(entries).map(([key, values]) => [key, values[values.length - 1] ?? ""]),
  );
}
