import { test } from "node:test";
import assert from "node:assert";
import { spawnSync } from "node:child_process";
import { mkdtempSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const scriptPath = new URL("./run-sqlc-generate.mjs", import.meta.url).pathname;

function runScript(binDir) {
  return spawnSync("node", [scriptPath], {
    encoding: "utf8",
    env: {
      ...process.env,
      SUPPERJUMPIN_SQLC_BIN_DIR: binDir,
    },
  });
}

test("exits with clear error when local sqlc binary is missing", () => {
  const emptyDir = mkdtempSync(join(tmpdir(), "sj-test-"));
  const result = runScript(emptyDir);

  assert.notStrictEqual(result.status, 0, "Expected non-zero exit code");
  assert.ok(
    result.stderr.includes("run `npm run setup` first") || result.stdout.includes("run `npm run setup` first"),
    `Expected error message to mention 'npm run setup' first. Got stdout: ${result.stdout} stderr: ${result.stderr}`
  );
});

test("invokes local sqlc binary when present", () => {
  const binDir = mkdtempSync(join(tmpdir(), "sj-test-"));
  // Create a fake sqlc binary that prints a sentinel and exits 0
  const fakeSqlc = join(binDir, "sqlc");
  writeFileSync(fakeSqlc, "#!/bin/sh\necho 'fake-sqlc-ran'\n", { mode: 0o755 });

  const result = runScript(binDir);

  assert.strictEqual(result.status, 0, `Expected exit 0. stderr: ${result.stderr}`);
  assert.ok(result.stdout.includes("fake-sqlc-ran") || result.stderr.includes("fake-sqlc-ran"),
    `Expected fake sqlc to run. Got stdout: ${result.stdout} stderr: ${result.stderr}`
  );
});
