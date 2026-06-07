import { existsSync } from "node:fs";
import { join, resolve } from "node:path";
import { spawnSync } from "node:child_process";

const binDir = process.env.SUPPERJUMPIN_MIGRATE_BIN_DIR ?? "bin";
const migratePath = join(resolve(binDir), "migrate");
const defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable";
const databaseURL = process.env.SUPPERJUMPIN_DATABASE_URL ?? defaultDatabaseURL;

if (process.env.DATABASE_URL && !process.env.SUPPERJUMPIN_DATABASE_URL) {
  console.log("Ignoring ambient DATABASE_URL; set SUPPERJUMPIN_DATABASE_URL to target a non-local database.");
}

if (!existsSync(migratePath)) {
  console.error(`Local migrate binary not found at ${migratePath}. run \`npm run setup\` first.`);
  process.exit(1);
}

const result = spawnSync(migratePath, [
  "-database", databaseURL,
  "-path", "apps/api/db/migrations",
  "up",
], { stdio: "inherit" });

process.exit(result.status ?? 0);
