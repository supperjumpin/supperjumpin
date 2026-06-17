import { test } from "node:test";
import assert from "node:assert";
import { spawnSync } from "node:child_process";
import { mkdtempSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const scriptPath = new URL("./db-migrate.mjs", import.meta.url).pathname;

function runScript(binDir, env = {}) {
  return spawnSync("node", [scriptPath], {
    encoding: "utf8",
    env: {
      ...process.env,
      SUPPERJUMPIN_MIGRATE_BIN_DIR: binDir,
      ...env,
    },
  });
}

test("exits with clear error when local migrate binary is missing", () => {
  const emptyDir = mkdtempSync(join(tmpdir(), "sj-test-"));
  const result = runScript(emptyDir);

  assert.notStrictEqual(result.status, 0, "Expected non-zero exit code");
  assert.ok(
    result.stderr.includes("run `npm run setup` first") || result.stdout.includes("run `npm run setup` first"),
    `Expected error message to mention 'npm run setup' first. Got stdout: ${result.stdout} stderr: ${result.stderr}`
  );
});

test("invokes local migrate binary with default local database URL when env is unset", () => {
  const binDir = mkdtempSync(join(tmpdir(), "sj-test-"));
  const fakeMigrate = join(binDir, "migrate");
  writeFileSync(
    fakeMigrate,
    "#!/bin/sh\nfor arg in \"$@\"; do echo \"$arg\"; done\n",
    { mode: 0o755 }
  );

  const result = runScript(binDir, { DATABASE_URL: undefined });

  assert.strictEqual(result.status, 0, `Expected exit 0. stderr: ${result.stderr}`);
  const output = result.stdout + result.stderr;
  assert.ok(output.includes("postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable"),
    `Expected default local database URL in output. Got: ${output}`
  );
  assert.ok(output.includes("apps/api/db/migrations"),
    `Expected migrations path in output. Got: ${output}`
  );
  assert.ok(output.includes("up"),
    `Expected 'up' command in output. Got: ${output}`
  );
});

test("ignores SUPPERJUMPIN_DATABASE_URL so shell config cannot retarget migrations", () => {
  const binDir = mkdtempSync(join(tmpdir(), "sj-test-"));
  const fakeMigrate = join(binDir, "migrate");
  writeFileSync(
    fakeMigrate,
    "#!/bin/sh\nfor arg in \"$@\"; do echo \"$arg\"; done\n",
    { mode: 0o755 }
  );

  const customURL = "postgres://user:pass@host:9999/db?sslmode=require";
  const result = runScript(binDir, { SUPPERJUMPIN_DATABASE_URL: customURL });

  assert.strictEqual(result.status, 0, `Expected exit 0. stderr: ${result.stderr}`);
  const output = result.stdout + result.stderr;
  assert.ok(output.includes("Ignoring SUPPERJUMPIN_DATABASE_URL"),
    `Expected warning about ignored SUPPERJUMPIN_DATABASE_URL. Got: ${output}`
  );
  assert.ok(output.includes("postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable"),
    `Expected local default DATABASE_URL in output. Got: ${output}`
  );
  assert.ok(!output.includes("host:9999"),
    `Expected custom SUPPERJUMPIN_DATABASE_URL to be ignored. Got: ${output}`
  );
});

test("ignores ambient DATABASE_URL so shell config cannot retarget migrations", () => {
  const binDir = mkdtempSync(join(tmpdir(), "sj-test-"));
  const fakeMigrate = join(binDir, "migrate");
  writeFileSync(
    fakeMigrate,
    "#!/bin/sh\nfor arg in \"$@\"; do echo \"$arg\"; done\n",
    { mode: 0o755 }
  );

  const result = runScript(binDir, { DATABASE_URL: "postgres://user:pass@staging:9999/db?sslmode=require" });

  assert.strictEqual(result.status, 0, `Expected exit 0. stderr: ${result.stderr}`);
  const output = result.stdout + result.stderr;
  assert.ok(output.includes("Ignoring ambient DATABASE_URL"),
    `Expected warning about ignored DATABASE_URL. Got: ${output}`
  );
  assert.ok(output.includes("postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable"),
    `Expected local default DATABASE_URL in output. Got: ${output}`
  );
  assert.ok(!output.includes("staging:9999"),
    `Expected staging DATABASE_URL to be ignored. Got: ${output}`
  );
});
