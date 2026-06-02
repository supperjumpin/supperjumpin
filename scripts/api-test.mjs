import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

/**
 * Determine the test database URL from environment.
 * Uses SUPPERJUMPIN_TEST_DATABASE_URL when set, otherwise falls back
 * to the local Docker Compose Postgres test database.
 */
export function getTestDatabaseURL(env) {
  return (
    env.SUPPERJUMPIN_TEST_DATABASE_URL ??
    "postgres://postgres:postgres@localhost:5432/supperjumpin_test?sslmode=disable"
  );
}

/**
 * Extract the database name from a postgres:// URL.
 */
export function parseDatabaseName(url) {
  const parsed = new URL(url);
  const dbName = parsed.pathname.replace(/^\//, "");
  if (!dbName) {
    throw new Error(`Could not parse database name from URL: ${url}`);
  }
  return dbName;
}

/**
 * Check whether it is safe to destructively reset the given database.
 * Safe when the name ends with _test, or when allowUnsafe is true.
 */
export function isSafeToReset(dbName, allowUnsafe) {
  if (allowUnsafe) {
    return true;
  }
  return dbName.endsWith("_test");
}

/**
 * Build an admin database URL by replacing the database name with "postgres".
 */
export function buildAdminURL(databaseURL) {
  const parsed = new URL(databaseURL);
  parsed.pathname = "/postgres";
  return parsed.toString();
}

/**
 * Run a psql command. Uses docker compose exec when running locally,
 * otherwise expects psql to be available on PATH.
 */
function runPsqlCommand(databaseURL, sql, isLocalDocker) {
  if (isLocalDocker) {
    return spawnSync("docker", ["compose", "exec", "-T", "postgres", "psql", databaseURL, "-c", sql], {
      stdio: "pipe",
    });
  }
  return spawnSync("psql", [databaseURL, "-c", sql], { stdio: "pipe" });
}

/**
 * Wait for local Docker Compose Postgres to be ready.
 */
function waitForPostgresReady(timeoutMs = 30000) {
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
    // Busy-wait 500ms
    const waitStart = Date.now();
    while (Date.now() - waitStart < 500) {
      // no-op
    }
  }
  return false;
}

/**
 * Run migrations against the given database URL.
 */
function runMigrations(databaseURL) {
  const binDir = process.env.SUPPERJUMPIN_MIGRATE_BIN_DIR ?? "bin";
  const migratePath = join(resolve(binDir), "migrate");
  if (!existsSync(migratePath)) {
    console.error(`Local migrate binary not found at ${migratePath}. Run \`npm run setup\` first.`);
    return { status: 1 };
  }
  const migrationsPath = resolve("apps/api/db/migrations");
  return spawnSync(
    migratePath,
    ["-database", databaseURL, "-path", migrationsPath, "up"],
    { stdio: "inherit" }
  );
}

function main() {
  const env = process.env;
  const testDatabaseURL = getTestDatabaseURL(env);
  const dbName = parseDatabaseName(testDatabaseURL);
  const allowUnsafe = env.SUPPERJUMPIN_TEST_ALLOW_UNSAFE_RESET === "1";
  const isLocalDocker = !env.SUPPERJUMPIN_TEST_DATABASE_URL;

  if (!isSafeToReset(dbName, allowUnsafe)) {
    console.error(
      `Refusing to reset database "${dbName}" because it does not end with "_test".`
    );
    console.error(`Set SUPPERJUMPIN_TEST_ALLOW_UNSAFE_RESET=1 to override.`);
    process.exit(1);
  }

  if (isLocalDocker) {
    console.log("Starting local Docker Compose Postgres...");
    const upResult = spawnSync("docker", ["compose", "up", "-d", "postgres"], {
      stdio: "inherit",
    });
    if (upResult.status !== 0) {
      console.error("Failed to start Docker Compose Postgres.");
      process.exit(1);
    }

    console.log("Waiting for Postgres to be ready...");
    if (!waitForPostgresReady()) {
      console.error("Postgres did not become ready in time.");
      process.exit(1);
    }
  }

  const adminURL = buildAdminURL(testDatabaseURL);

  console.log(`Resetting test database "${dbName}"...`);
  const dropResult = runPsqlCommand(adminURL, `DROP DATABASE IF EXISTS "${dbName}";`, isLocalDocker);
  if (dropResult.status !== 0) {
    console.error(`Failed to drop test database "${dbName}".`);
    console.error(dropResult.stderr?.toString() || "");
  }

  const createResult = runPsqlCommand(
    adminURL,
    `CREATE DATABASE "${dbName}";`,
    isLocalDocker
  );
  if (createResult.status !== 0) {
    console.error(`Failed to create test database "${dbName}".`);
    console.error(createResult.stderr?.toString() || "");
    process.exit(1);
  }

  console.log("Applying migrations...");
  const migrateResult = runMigrations(testDatabaseURL);
  if (migrateResult.status !== 0) {
    console.error("Migrations failed.");
    process.exit(1);
  }

  console.log("Running Go tests...");
  const goArgs = ["test", "./apps/api/...", ...process.argv.slice(2)];
  const goTestResult = spawnSync("go", goArgs, {
    stdio: "inherit",
    env: {
      ...env,
      SUPPERJUMPIN_TEST_DATABASE_URL: testDatabaseURL,
    },
  });

  process.exit(goTestResult.status ?? 0);
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}
