import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Command } from "commander";
import axios from "axios";
import { registerRoutesCommand } from "../routes.command";
import { CliError } from "../../../utilities/errors/cli-error";

const mockLoadConfigFile = vi.fn();

vi.mock("axios");
vi.mock("../../../utilities/config/services/config.service", () => ({
  ConfigService: vi.fn(function MockConfigService() {
    return {
      loadConfigFile: mockLoadConfigFile,
    };
  }),
}));

const defaultConfigFile = {
  defaultWorkspace: "default",
  workspaces: {
    default: {
      apiUrl: "https://api.twenty.com",
      apiKey: "test-token",
    },
  },
};

describe("routes command", () => {
  let program: Command;
  let consoleSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerRoutesCommand(program);
    consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});
    vi.mocked(axios.request).mockReset();
    mockLoadConfigFile.mockResolvedValue(defaultConfigFile);
  });

  afterEach(() => {
    consoleSpy.mockRestore();
    vi.clearAllMocks();
  });

  describe("command registration", () => {
    it("registers routes root command", () => {
      const routesCmd = program.commands.find((candidate) => candidate.name() === "routes");

      expect(routesCmd).toBeDefined();
      expect(routesCmd?.description()).toBe("Invoke public route trigger endpoints");
    });

    it("registers invoke subcommand with required path", () => {
      const routesCmd = program.commands.find((candidate) => candidate.name() === "routes");
      const invokeCmd = routesCmd?.commands.find((candidate) => candidate.name() === "invoke");
      const args = invokeCmd?.registeredArguments ?? [];

      expect(invokeCmd).toBeDefined();
      expect(invokeCmd?.description()).toBe("Invoke a public /s/* route endpoint");
      expect(args.length).toBe(1);
      expect(args[0].name()).toBe("routePath");
      expect(args[0].required).toBe(true);
    });

    it("registers route invocation options and global options", () => {
      const routesCmd = program.commands.find((candidate) => candidate.name() === "routes");
      const invokeCmd = routesCmd?.commands.find((candidate) => candidate.name() === "invoke");
      const options = invokeCmd?.options ?? [];

      expect(options.find((option) => option.long === "--method")).toBeDefined();
      expect(options.find((option) => option.long === "--data")).toBeDefined();
      expect(options.find((option) => option.long === "--file")).toBeDefined();
      expect(options.find((option) => option.long === "--param")).toBeDefined();
      expect(options.find((option) => option.long === "--header")).toBeDefined();
      expect(options.find((option) => option.long === "--output")).toBeDefined();
      expect(options.find((option) => option.long === "--query")).toBeDefined();
      expect(options.find((option) => option.long === "--workspace")).toBeDefined();
    });
  });

  describe("invoke", () => {
    it("invokes a route with GET by default and normalized path", async () => {
      vi.mocked(axios.request).mockResolvedValue({
        data: {
          ok: true,
          source: "route",
        },
      } as never);

      await program.parseAsync([
        "node",
        "test",
        "routes",
        "invoke",
        "contacts/sync",
        "--param",
        "source=cli",
        "--param",
        "dryRun=true",
        "-o",
        "json",
      ]);

      expect(axios.request).toHaveBeenCalledWith({
        method: "get",
        url: "https://api.twenty.com/s/contacts/sync",
        data: undefined,
        params: {
          source: "cli",
          dryRun: "true",
        },
        headers: {
          Authorization: "Bearer test-token",
        },
      });

      expect(JSON.parse(consoleSpy.mock.calls[0][0] as string)).toEqual({
        ok: true,
        source: "route",
      });
    });

    it("invokes a route with a request body and custom headers", async () => {
      vi.mocked(axios.request).mockResolvedValue({
        data: {
          created: true,
        },
      } as never);

      await program.parseAsync([
        "node",
        "test",
        "routes",
        "invoke",
        "/s/hooks/import",
        "--method",
        "post",
        "-d",
        '{"batch":"one"}',
        "--header",
        "x-source=cli",
        "--header",
        "x-trace-id=trace-1",
        "-o",
        "json",
      ]);

      expect(axios.request).toHaveBeenCalledWith({
        method: "post",
        url: "https://api.twenty.com/s/hooks/import",
        data: {
          batch: "one",
        },
        params: undefined,
        headers: {
          Authorization: "Bearer test-token",
          "x-source": "cli",
          "x-trace-id": "trace-1",
        },
      });
    });

    it("omits the authorization header when no api key is configured", async () => {
      mockLoadConfigFile.mockResolvedValue({
        defaultWorkspace: "default",
        workspaces: {
          default: {
            apiUrl: "https://api.twenty.com",
          },
        },
      });
      vi.mocked(axios.request).mockResolvedValue({
        data: {
          ok: true,
        },
      } as never);

      await program.parseAsync(["node", "test", "routes", "invoke", "public/ping", "-o", "json"]);

      expect(axios.request).toHaveBeenCalledWith({
        method: "get",
        url: "https://api.twenty.com/s/public/ping",
        data: undefined,
        params: undefined,
        headers: undefined,
      });
    });

    it("throws when GET is used with a request body", async () => {
      await expect(
        program.parseAsync([
          "node",
          "test",
          "routes",
          "invoke",
          "public/ping",
          "--method",
          "get",
          "-d",
          '{"hello":"world"}',
        ]),
      ).rejects.toThrow(CliError);
    });

    it("throws when the request body is not a json object", async () => {
      await expect(
        program.parseAsync([
          "node",
          "test",
          "routes",
          "invoke",
          "hooks/import",
          "--method",
          "post",
          "-d",
          '["bad"]',
        ]),
      ).rejects.toThrow(CliError);
    });

    it("throws for unsupported methods", async () => {
      await expect(
        program.parseAsync([
          "node",
          "test",
          "routes",
          "invoke",
          "public/ping",
          "--method",
          "options",
        ]),
      ).rejects.toThrow(CliError);
    });
  });
});
