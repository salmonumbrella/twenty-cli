import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { Command } from "commander";
import { registerApiMetadataCommand } from "../api-metadata.command";
import { CliError } from "../../../utilities/errors/cli-error";

vi.mock("../../../utilities/shared/services", () => ({
  createServices: vi.fn(),
}));

import { createServices } from "../../../utilities/shared/services";

function createMockServices() {
  return {
    metadata: {
      getObject: vi.fn().mockResolvedValue({
        id: "obj-1",
        nameSingular: "person",
        namePlural: "people",
      }),
      listViews: vi.fn().mockResolvedValue([{ id: "view-1", name: "All People" }]),
      listViewSorts: vi.fn().mockResolvedValue([{ id: "sort-1", viewId: "view-1" }]),
      listPageLayouts: vi.fn().mockResolvedValue([{ id: "layout-1" }]),
      listPageLayoutTabs: vi.fn().mockResolvedValue([{ id: "tab-1" }]),
      listPageLayoutWidgets: vi.fn().mockResolvedValue([{ id: "widget-1" }]),
      listCommandMenuItems: vi.fn().mockResolvedValue([{ id: "cmd-1" }]),
      listFrontComponents: vi.fn().mockResolvedValue([{ id: "front-1" }]),
      listNavigationMenuItems: vi.fn().mockResolvedValue([{ id: "nav-1" }]),
    },
    output: {
      render: vi.fn(),
    },
  };
}

describe("api-metadata command", () => {
  let program: Command;
  let mockServices: ReturnType<typeof createMockServices>;

  beforeEach(() => {
    program = new Command();
    program.exitOverride();
    registerApiMetadataCommand(program);
    mockServices = createMockServices();
    vi.mocked(createServices).mockReturnValue(mockServices as any);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("command registration", () => {
    it("registers api-metadata command with the expected options", () => {
      const command = program.commands.find((cmd) => cmd.name() === "api-metadata");

      expect(command).toBeDefined();
      expect(command?.description()).toBe("Schema operations");

      const options = command?.options ?? [];
      expect(options.find((opt) => opt.long === "--object")).toBeDefined();
      expect(options.find((opt) => opt.long === "--view")).toBeDefined();
      expect(options.find((opt) => opt.long === "--page-layout")).toBeDefined();
      expect(options.find((opt) => opt.long === "--page-layout-tab")).toBeDefined();
      expect(options.find((opt) => opt.long === "--page-layout-type")?.description).toContain(
        "requires --object",
      );
    });
  });

  describe("routing", () => {
    it("routes views list through object resolution", async () => {
      await program.parseAsync([
        "node",
        "test",
        "api-metadata",
        "views",
        "list",
        "--object",
        "person",
        "-o",
        "json",
      ]);

      expect(mockServices.metadata.getObject).toHaveBeenCalledWith("person");
      expect(mockServices.metadata.listViews).toHaveBeenCalledWith({
        objectMetadataId: "obj-1",
      });
      expect(mockServices.output.render).toHaveBeenCalledWith(
        [{ id: "view-1", name: "All People" }],
        { format: "json", query: undefined },
      );
    });

    it("routes view-sorts list with --view", async () => {
      await program.parseAsync([
        "node",
        "test",
        "api-metadata",
        "view-sorts",
        "list",
        "--view",
        "view-1",
        "-o",
        "json",
      ]);

      expect(mockServices.metadata.listViewSorts).toHaveBeenCalledWith({
        viewId: "view-1",
      });
      expect(mockServices.output.render).toHaveBeenCalled();
    });

    it("routes page-layout-tabs list with --page-layout", async () => {
      await program.parseAsync([
        "node",
        "test",
        "api-metadata",
        "page-layout-tabs",
        "list",
        "--page-layout",
        "layout-1",
        "-o",
        "json",
      ]);

      expect(mockServices.metadata.listPageLayoutTabs).toHaveBeenCalledWith({
        pageLayoutId: "layout-1",
      });
    });

    it("routes page-layout-widgets list with --page-layout-tab", async () => {
      await program.parseAsync([
        "node",
        "test",
        "api-metadata",
        "page-layout-widgets",
        "list",
        "--page-layout-tab",
        "tab-1",
        "-o",
        "json",
      ]);

      expect(mockServices.metadata.listPageLayoutWidgets).toHaveBeenCalledWith({
        pageLayoutTabId: "tab-1",
      });
    });

    it("routes command-menu-items list", async () => {
      await program.parseAsync([
        "node",
        "test",
        "api-metadata",
        "command-menu-items",
        "list",
        "-o",
        "json",
      ]);

      expect(mockServices.metadata.listCommandMenuItems).toHaveBeenCalledWith();
      expect(mockServices.output.render).toHaveBeenCalled();
    });

    it("routes front-components list", async () => {
      await program.parseAsync([
        "node",
        "test",
        "api-metadata",
        "front-components",
        "list",
        "-o",
        "json",
      ]);

      expect(mockServices.metadata.listFrontComponents).toHaveBeenCalledWith();
      expect(mockServices.output.render).toHaveBeenCalled();
    });

    it("routes navigation-menu-items list", async () => {
      await program.parseAsync([
        "node",
        "test",
        "api-metadata",
        "navigation-menu-items",
        "list",
        "-o",
        "json",
      ]);

      expect(mockServices.metadata.listNavigationMenuItems).toHaveBeenCalledWith();
      expect(mockServices.output.render).toHaveBeenCalled();
    });
  });

  describe("validation", () => {
    it("rejects page-layout-type without object filter", async () => {
      await expect(
        program.parseAsync([
          "node",
          "test",
          "api-metadata",
          "page-layouts",
          "list",
          "--page-layout-type",
          "RECORD_PAGE",
          "-o",
          "json",
        ]),
      ).rejects.toThrow(CliError);
    });
  });
});
