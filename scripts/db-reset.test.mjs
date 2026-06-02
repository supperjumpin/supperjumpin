import { test } from "node:test";
import assert from "node:assert";
import { mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";

const scriptPath = new URL("./db-reset.mjs", import.meta.url).pathname;

function makeFakeDocker(binDir, logPath) {
  const dockerPath = join(binDir, "docker");
  writeFileSync(
    dockerPath,
    `#!/bin/sh
printf '%s\n' "$@" >> "${logPath}"
printf '%s\n' '--' >> "${logPath}"
`,
    { mode: 0o755 }
  );
}

function makeFakeMigrate(binDir, logPath) {
  const migratePath = join(binDir, "migrate");
  writeFileSync(
    migratePath,
    `#!/bin/sh
printf '%s\n' "$@" >> "${logPath}"
printf '%s\n' '--' >> "${logPath}"
`,
    { mode: 0o755 }
  );
}

test("reset recreates the local development database and reapplies migrations", () => {
  const dockerBinDir = mkdtempSync(join(tmpdir(), "sj-test-"));
  const migrateBinDir = mkdtempSync(join(tmpdir(), "sj-test-"));
  const dockerLogPath = join(dockerBinDir, "docker.log");
  const migrateLogPath = join(migrateBinDir, "migrate.log");

  makeFakeDocker(dockerBinDir, dockerLogPath);
  makeFakeMigrate(migrateBinDir, migrateLogPath);

  const result = spawnSync("node", [scriptPath], {
    encoding: "utf8",
    env: {
      ...process.env,
      PATH: `${dockerBinDir}:${process.env.PATH}`,
      SUPPERJUMPIN_MIGRATE_BIN_DIR: migrateBinDir,
    },
  });

  assert.strictEqual(result.status, 0, `Expected exit 0. stderr: ${result.stderr}`);

  const dockerLog = readFileSync(dockerLogPath, "utf8");
  assert.ok(dockerLog.includes("compose\nup\n-d\npostgres\n--\n"), `Expected docker compose up. Got: ${dockerLog}`);
  assert.ok(
    dockerLog.includes("compose\nexec\n-T\npostgres\npg_isready\n-U\npostgres\n--\n"),
    `Expected pg_isready probe. Got: ${dockerLog}`
  );
  assert.ok(
    dockerLog.includes(
      "compose\nexec\n-T\npostgres\npsql\npostgres://postgres:postgres@localhost:5432/postgres?sslmode=disable\n-c\nDROP DATABASE IF EXISTS \"supperjumpin\";\n--\n"
    ),
    `Expected DROP DATABASE command. Got: ${dockerLog}`
  );
  assert.ok(
    dockerLog.includes(
      "compose\nexec\n-T\npostgres\npsql\npostgres://postgres:postgres@localhost:5432/postgres?sslmode=disable\n-c\nCREATE DATABASE \"supperjumpin\";\n--\n"
    ),
    `Expected CREATE DATABASE command. Got: ${dockerLog}`
  );

  const migrateLog = readFileSync(migrateLogPath, "utf8");
  assert.ok(
    migrateLog.includes("-database\npostgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable\n"),
    `Expected migrate to target the development database. Got: ${migrateLog}`
  );
  assert.ok(
    migrateLog.includes("-path\napps/api/db/migrations\nup\n--\n"),
    `Expected migrate to run local migrations. Got: ${migrateLog}`
  );
});
