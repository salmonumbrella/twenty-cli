import path from "node:path";
import { describe, expect, it } from "vitest";
import { BuiltCliRunError, runNodeScriptAsync } from "./process-runner";

describe("process runner", () => {
  it("kills timed out child processes and preserves output context", async () => {
    const fixturePath = path.resolve(__dirname, "./fixtures/hanging-child.js");

    const error = await runNodeScriptAsync(fixturePath, [], {
      timeoutMs: 250,
      throwOnNonZeroExit: true,
    }).catch((caughtError) => caughtError);

    expect(error).toBeInstanceOf(BuiltCliRunError);
    expect(error.result).toEqual({
      exitCode: null,
      stdout: "fixture-stdout\n",
      stderr: "fixture-stderr\n",
    });
    expect(error.message).toMatch(
      /Built CLI exited with code null[\s\S]*stderr:\nfixture-stderr[\s\S]*stdout:\nfixture-stdout/,
    );
  });
});
