import { existsSync } from "node:fs";
import { join, resolve } from "node:path";
import { spawnSync } from "node:child_process";

import { DEFAULT_DEVELOPMENT_DATABASE_URL } from "./db-helpers.mjs";

const binDir = process.env.SUPPERJUMPIN_MIGRATE_BIN_DIR ?? "bin";
const migratePath = join(resolve(binDir), "migrate");
const databaseURL = DEFAULT_DEVELOPMENT_DATABASE_URL;

if (process.env.DATABASE_URL) {
  console.log("Ignoring ambient DATABASE_URL; db:migrate always targets the local Docker database for now.");
}

if (process.env.SUPPERJUMPIN_DATABASE_URL) {
  console.log("Ignoring SUPPERJUMPIN_DATABASE_URL; db:migrate always targets the local Docker database for now.");
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
