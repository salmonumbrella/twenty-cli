import { Command } from "commander";
import { GlobalOptions, resolveGlobalOptions } from "./global-options";
import { CliServices, createOutputService, createServices } from "./services";
import { OutputService } from "../output/services/output.service";

export interface CommandContext {
  globalOptions: GlobalOptions;
  services: CliServices;
}

export interface OutputContext {
  globalOptions: GlobalOptions;
  output: OutputService;
}

export function createCommandContext(
  command: Command,
  overrides?: Parameters<typeof resolveGlobalOptions>[1],
): CommandContext {
  const globalOptions = resolveGlobalOptions(command, overrides);
  const services = createServices(globalOptions);

  return {
    globalOptions,
    services,
  };
}

export function createOutputContext(
  command: Command,
  overrides?: Parameters<typeof resolveGlobalOptions>[1],
): OutputContext {
  const globalOptions = resolveGlobalOptions(command, overrides);

  return {
    globalOptions,
    output: createOutputService(globalOptions),
  };
}
