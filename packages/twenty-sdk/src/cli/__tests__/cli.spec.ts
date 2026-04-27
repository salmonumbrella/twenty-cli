import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Command } from "commander";
import { main } from "../cli";
import { loadCliEnvironment } from "../utilities/config/services/environment.service";
import { buildProgram } from "../program";
import { maybeHandleInlineHelp } from "../help";

vi.mock("../utilities/config/services/environment.service", () => ({
  loadCliEnvironment: vi.fn(),
}));

vi.mock("../program", () => ({
  buildProgram: vi.fn(),
}));

vi.mock("../help", () => ({
  maybeHandleInlineHelp: vi.fn(),
}));

describe("CLI entrypoint", () => {
  let events: string[];
  let program: Command;

  beforeEach(() => {
    events = [];
    program = new Command();
    program.parseAsync = vi.fn(async () => {
      events.push("parse");
      return program;
    });

    vi.mocked(loadCliEnvironment).mockImplementation(() => {
      events.push("load-env");
      return { loadedFiles: [] };
    });
    vi.mocked(buildProgram).mockImplementation(() => {
      events.push("build-program");
      return program;
    });
    vi.mocked(maybeHandleInlineHelp).mockResolvedValue(false);
  });

  afterEach(() => {
    vi.clearAllMocks();
    process.exitCode = undefined;
  });

  it("loads environment files before building dynamic commands", async () => {
    const argv = ["node", "twenty", "--env-file", ".env.production", "records"];

    await main(argv);

    expect(events).toEqual(["load-env", "build-program", "parse"]);
    expect(loadCliEnvironment).toHaveBeenCalledWith({
      argv,
      cwd: process.cwd(),
    });
    expect(maybeHandleInlineHelp).toHaveBeenCalledWith(program, [
      "--env-file",
      ".env.production",
      "records",
    ]);
  });
});
