import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { join, resolve } from "node:path";

export const DEFAULT_DEVELOPMENT_DATABASE_URL =
  "postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable";

export const DEFAULT_TEST_DATABASE_URL =
  "postgres://postgres:postgres@localhost:5432/supperjumpin_test?sslmode=disable";

export function parseDatabaseName(url) {
  const parsed = new URL(url);
  const dbName = parsed.pathname.replace(/^\//, "");
  if (!dbName) {
    throw new Error(`Could not parse database name from URL: ${url}`);
  }
  return dbName;
}

export function buildAdminURL(databaseURL) {
  const parsed = new URL(databaseURL);
  parsed.pathname = "/postgres";
  return parsed.toString();
}

export function runPsqlCommand(databaseURL, sql, isLocalDocker) {
  if (isLocalDocker) {
    return spawnSync("docker", ["compose", "exec", "-T", "postgres", "psql", databaseURL, "-c", sql], {
      stdio: "pipe",
    });
  }

  return spawnSync("psql", [databaseURL, "-c", sql], { stdio: "pipe" });
}

export function waitForPostgresReady(timeoutMs = 30000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    const result = spawnSync(
      "docker",
      ["compose", "exec", "-T", "postgres", "pg_isready", "-U", "postgres"],
      { stdio: "pipe" }
    );
    if (result.status === 0) {
      return true;
    }

    const waitStart = Date.now();
    while (Date.now() - waitStart < 500) {
      // no-op
    }
  }

  return false;
}

export function runMigrations(databaseURL, env = process.env) {
  const binDir = env.SUPPERJUMPIN_MIGRATE_BIN_DIR ?? "bin";
  const migratePath = join(resolve(binDir), "migrate");
  if (!existsSync(migratePath)) {
    console.error(`Local migrate binary not found at ${migratePath}. Run \`npm run setup\` first.`);
    return { status: 1 };
  }

  return spawnSync(
    migratePath,
    ["-database", databaseURL, "-path", "apps/api/db/migrations", "up"],
    { stdio: "inherit" }
  );
}
