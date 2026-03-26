import { Command } from "commander";
import { CliError } from "../../utilities/errors/cli-error";
import { applyGlobalOptions, resolveGlobalOptions } from "../../utilities/shared/global-options";
import { readFileOrStdin, safeJsonParse } from "../../utilities/shared/io";
import { createServices } from "../../utilities/shared/services";

export function registerMcpCommand(program: Command): void {
  const cmd = program.command("mcp").description("Interact with the official Twenty MCP server");

  const statusCmd = cmd
    .command("status")
    .description("Show MCP availability for the active workspace");
  applyGlobalOptions(statusCmd);
  statusCmd.action(async (_options, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const result = await services.mcp.status();

    await services.output.render(result, {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });

  const catalogCmd = cmd.command("catalog").description("Show the official MCP tool catalog");
  applyGlobalOptions(catalogCmd);
  catalogCmd.action(async (_options, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const result = await services.mcp.callTool("get_tool_catalog", {});

    await services.output.render(result, {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });

  const learnCmd = cmd
    .command("learn")
    .description("Get guidance for one or more MCP tools")
    .argument("<toolNames...>", "Tool names to learn");
  applyGlobalOptions(learnCmd);
  learnCmd.action(async (toolNames: string[], _options, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const result = await services.mcp.callTool("learn_tools", { toolNames });

    await services.output.render(result, {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });

  const callCmd = cmd
    .command("call")
    .description("Call an official MCP tool by name")
    .argument("<tool>", "Official MCP tool name")
    .option("--data <json>", "Tool arguments as inline JSON")
    .option("--file <path>", "Path to a JSON file containing tool arguments (use - for stdin)");
  applyGlobalOptions(callCmd);
  callCmd.action(async (tool: string, options: CallOptions, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const argumentsObject = await readMcpCallArguments(options);
    const result = await services.mcp.callTool(tool, argumentsObject);

    await services.output.render(result, {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });

  const loadSkillsCmd = cmd.command("load-skills").description("Load official MCP skills");
  loadSkillsCmd.argument("<skillNames...>", "Skill names to load");
  applyGlobalOptions(loadSkillsCmd);
  loadSkillsCmd.action(async (skillNames: string[], _options, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const result = await services.mcp.callTool("load_skills", { skillNames });

    await services.output.render(result, {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });

  const helpCenterCmd = cmd
    .command("help-center")
    .description("Search the official MCP help center");
  helpCenterCmd.argument("<query>", "Help center query");
  applyGlobalOptions(helpCenterCmd);
  helpCenterCmd.action(async (query: string, _options, command: Command) => {
    const globalOptions = resolveGlobalOptions(command);
    const services = createServices(globalOptions);
    const result = await services.mcp.callTool("search_help_center", { query });

    await services.output.render(result, {
      format: globalOptions.output,
      query: globalOptions.query,
    });
  });
}

interface CallOptions {
  data?: string;
  file?: string;
}

async function readMcpCallArguments(options: CallOptions): Promise<Record<string, unknown>> {
  const sources = [options.data, options.file].filter((value) => value != null);
  if (sources.length > 1) {
    throw new CliError(
      "provide MCP call arguments via --data or --file, not multiple sources",
      "INVALID_ARGUMENTS",
    );
  }

  if (sources.length === 0) {
    return {};
  }

  if (options.file) {
    let content: string;
    try {
      content = await readFileOrStdin(options.file);
    } catch {
      throw new CliError(
        `Unable to read MCP call arguments file: ${options.file}`,
        "INVALID_ARGUMENTS",
      );
    }

    if (content.trim() === "") {
      return {};
    }

    return parseMcpCallArguments(content);
  }

  return parseMcpCallArguments(options.data);
}

function parseMcpCallArguments(rawJson: string | undefined): Record<string, unknown> {
  if (rawJson == null || rawJson.trim() === "") {
    return {};
  }

  let payload: unknown;
  try {
    payload = safeJsonParse(rawJson);
  } catch {
    throw new CliError("MCP call arguments must be valid JSON.", "INVALID_ARGUMENTS");
  }

  if (payload === null || typeof payload !== "object" || Array.isArray(payload)) {
    throw new CliError("MCP call arguments must be a JSON object.", "INVALID_ARGUMENTS");
  }

  return payload as Record<string, unknown>;
}
