import { spawnSync } from "node:child_process";
import { appendFileSync, mkdirSync } from "node:fs";
import { fileURLToPath } from "node:url";

import {
  buildAdminURL,
  DEFAULT_TEST_DATABASE_URL,
  parseDatabaseName,
  runMigrations,
  runPsqlCommand,
  waitForPostgresReady,
} from "./db-helpers.mjs";

/**
 * Determine the test database URL from environment.
 * Uses SUPPERJUMPIN_TEST_DATABASE_URL when set, otherwise falls back
 * to the local Docker Compose Postgres test database.
 */
export function getTestDatabaseURL(env) {
  return env.SUPPERJUMPIN_TEST_DATABASE_URL ?? DEFAULT_TEST_DATABASE_URL;
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

function main() {
  const env = process.env;
  const testDatabaseURL = getTestDatabaseURL(env);
  const dbName = parseDatabaseName(testDatabaseURL);
  const allowUnsafe = env.SUPPERJUMPIN_TEST_ALLOW_UNSAFE_RESET === "1";
  const isLocalDocker = !env.SUPPERJUMPIN_TEST_DATABASE_URL;
  const coverageEnabled = process.argv.includes("--coverage");
  const extraGoArgs = process.argv.slice(2).filter((arg) => arg !== "--coverage");

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
  const migrateResult = runMigrations(testDatabaseURL, env);
  if (migrateResult.status !== 0) {
    console.error("Migrations failed.");
    process.exit(1);
  }

  console.log("Running Go tests...");
  const goArgs = ["test"];
  if (coverageEnabled) {
    mkdirSync("coverage", { recursive: true });
    goArgs.push("-covermode=atomic", "-coverprofile=coverage/api.coverprofile");
  }
  goArgs.push("./apps/api/...", ...extraGoArgs);
  const goTestResult = spawnSync("go", goArgs, {
    stdio: "inherit",
    env: {
      ...env,
      SUPPERJUMPIN_TEST_DATABASE_URL: testDatabaseURL,
    },
  });

  if (goTestResult.status === 0 && coverageEnabled) {
    const coverResult = spawnSync("go", ["tool", "cover", "-func=coverage/api.coverprofile"], {
      encoding: "utf8",
      stdio: ["ignore", "pipe", "pipe"],
      env,
    });

    if (coverResult.stdout) {
      process.stdout.write(coverResult.stdout);
    }
    if (coverResult.stderr) {
      process.stderr.write(coverResult.stderr);
    }

    const summaryMatch = `${coverResult.stdout ?? ""}${coverResult.stderr ?? ""}`.match(
      /total:\s+\(statements\)\s+([\d.]+)%/
    );
    if (env.GITHUB_STEP_SUMMARY) {
      const line = summaryMatch
        ? `Go API coverage: ${summaryMatch[1]}% statement coverage.`
        : "Go API coverage completed.";
      appendFileSync(env.GITHUB_STEP_SUMMARY, `### Go API coverage\n${line}\n`);
    }
  }

  process.exit(goTestResult.status ?? 0);
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}
