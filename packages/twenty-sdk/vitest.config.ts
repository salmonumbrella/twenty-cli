import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    exclude: ["src/cli/__tests__/e2e/**"],
  },
});
