import { test } from "node:test";
import assert from "node:assert";
import { mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";

import { isProbablyRemoteSupabaseURL } from "./db-reset-staging.mjs";

const scriptPath = new URL("./db-reset-staging.mjs", import.meta.url).pathname;

function makeFakeCommand(binDir, name, logPath) {
  writeFileSync(
    join(binDir, name),
    `#!/bin/sh
printf '%s\n' "${name}" >> "${logPath}"
for arg in "$@"; do printf '%s\n' "$arg" >> "${logPath}"; done
printf '%s\n' '--' >> "${logPath}"
`,
    { mode: 0o755 }
  );
}

test("isProbablyRemoteSupabaseURL only allows Supabase postgres URLs", () => {
  assert.strictEqual(
    isProbablyRemoteSupabaseURL("postgresql://postgres.ref:pass@aws-1-us-west-2.pooler.supabase.com:5432/postgres?sslmode=require"),
    true
  );
  assert.strictEqual(
    isProbablyRemoteSupabaseURL("postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable"),
    false
  );
  assert.strictEqual(isProbablyRemoteSupabaseURL("https://example.com"), false);
});

test("refuses to reset without explicit confirmation", () => {
  const result = spawnSync("node", [scriptPath], {
    encoding: "utf8",
    env: {
      ...process.env,
      SUPPERJUMPIN_DATABASE_URL: "postgresql://postgres.ref:pass@aws-1-us-west-2.pooler.supabase.com:5432/postgres?sslmode=require",
      SUPPERJUMPIN_RESET_STAGING: undefined,
    },
  });

  assert.notStrictEqual(result.status, 0, "Expected non-zero exit code");
  assert.ok(
    `${result.stdout}${result.stderr}`.includes("SUPPERJUMPIN_RESET_STAGING=1"),
    `Expected confirmation error. Got stdout: ${result.stdout} stderr: ${result.stderr}`
  );
});

test("refuses to reset non-Supabase database URLs", () => {
  const result = spawnSync("node", [scriptPath], {
    encoding: "utf8",
    env: {
      ...process.env,
      SUPPERJUMPIN_DATABASE_URL: "postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable",
      SUPPERJUMPIN_RESET_STAGING: "1",
    },
  });

  assert.notStrictEqual(result.status, 0, "Expected non-zero exit code");
  assert.ok(
    `${result.stdout}${result.stderr}`.includes("does not look like a Supabase URL"),
    `Expected Supabase URL error. Got stdout: ${result.stdout} stderr: ${result.stderr}`
  );
});

test("drops known app tables and reapplies migrations", () => {
  const binDir = mkdtempSync(join(tmpdir(), "sj-test-"));
  const logPath = join(binDir, "commands.log");
  makeFakeCommand(binDir, "psql", logPath);
  makeFakeCommand(binDir, "migrate", logPath);

  const databaseURL = "postgresql://postgres.ref:pass@aws-1-us-west-2.pooler.supabase.com:5432/postgres?sslmode=require";
  const result = spawnSync("node", [scriptPath], {
    encoding: "utf8",
    env: {
      ...process.env,
      PATH: `${binDir}:${process.env.PATH}`,
      SUPPERJUMPIN_MIGRATE_BIN_DIR: binDir,
      SUPPERJUMPIN_DATABASE_URL: databaseURL,
      SUPPERJUMPIN_RESET_STAGING: "1",
    },
  });

  assert.strictEqual(result.status, 0, `Expected exit 0. stderr: ${result.stderr}`);
  const log = readFileSync(logPath, "utf8");
  assert.ok(log.includes(`psql\n${databaseURL}\n-v\nON_ERROR_STOP=1\n-c\n`), `Expected psql reset command. Got: ${log}`);
  assert.ok(log.includes("DROP TABLE IF EXISTS"), `Expected drop SQL. Got: ${log}`);
  assert.ok(log.includes("accounts"), `Expected accounts table in drop SQL. Got: ${log}`);
  assert.ok(log.includes("schema_migrations"), `Expected schema_migrations in drop SQL. Got: ${log}`);
  assert.ok(log.includes(`migrate\n-database\n${databaseURL}\n-path\napps/api/db/migrations\nup\n--\n`), `Expected migration reapply. Got: ${log}`);
});
