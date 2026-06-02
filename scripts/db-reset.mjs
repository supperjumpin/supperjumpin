import { fileURLToPath } from "node:url";

import {
  buildAdminURL,
  DEFAULT_DEVELOPMENT_DATABASE_URL,
  parseDatabaseName,
  runMigrations,
  runPsqlCommand,
  waitForPostgresReady,
} from "./db-helpers.mjs";
import { runLifecycle } from "./db-lifecycle.mjs";

function main() {
  const databaseURL = DEFAULT_DEVELOPMENT_DATABASE_URL;
  const dbName = parseDatabaseName(databaseURL);

  console.log("Starting local Docker Compose Postgres...");
  const upResult = runLifecycle("up");
  if (upResult.status !== 0) {
    console.error("Failed to start Docker Compose Postgres.");
    process.exit(1);
  }

  console.log("Waiting for Postgres to be ready...");
  if (!waitForPostgresReady()) {
    console.error("Postgres did not become ready in time.");
    process.exit(1);
  }

  const adminURL = buildAdminURL(databaseURL);

  console.log(`Resetting development database \"${dbName}\"...`);
  const dropResult = runPsqlCommand(adminURL, `DROP DATABASE IF EXISTS "${dbName}";`, true);
  if (dropResult.status !== 0) {
    console.error(`Failed to drop development database "${dbName}".`);
    console.error(dropResult.stderr?.toString() || "");
    process.exit(1);
  }

  const createResult = runPsqlCommand(adminURL, `CREATE DATABASE "${dbName}";`, true);
  if (createResult.status !== 0) {
    console.error(`Failed to create development database "${dbName}".`);
    console.error(createResult.stderr?.toString() || "");
    process.exit(1);
  }

  console.log("Applying migrations...");
  const migrateResult = runMigrations(databaseURL);
  if (migrateResult.status !== 0) {
    console.error("Migrations failed.");
    process.exit(1);
  }
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}
