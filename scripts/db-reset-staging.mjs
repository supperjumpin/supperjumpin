import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

import { runMigrations } from "./db-helpers.mjs";

const resetSQL = `
DROP TABLE IF EXISTS
    open_standings,
    judgments,
    guest_sessions,
    evidences,
    evidence_upload_authorizations,
    jumps,
    players,
    auth_identities,
    accounts,
    schema_migrations
CASCADE;
`;

export function isProbablyRemoteSupabaseURL(databaseURL) {
  try {
    const parsed = new URL(databaseURL);
    const host = parsed.hostname.toLowerCase();
    return parsed.protocol.startsWith("postgres") && host.endsWith("supabase.com");
  } catch {
    return false;
  }
}

export function runStagingReset(env = process.env) {
  const databaseURL = env.SUPPERJUMPIN_DATABASE_URL;
  if (!databaseURL) {
    console.error("SUPPERJUMPIN_DATABASE_URL is required.");
    return 1;
  }

  if (!isProbablyRemoteSupabaseURL(databaseURL)) {
    console.error("Refusing staging reset because SUPPERJUMPIN_DATABASE_URL does not look like a Supabase URL.");
    return 1;
  }

  if (env.SUPPERJUMPIN_RESET_STAGING !== "1") {
    console.error("Refusing destructive staging reset. Set SUPPERJUMPIN_RESET_STAGING=1 to confirm.");
    return 1;
  }

  console.log("Dropping Supperjumpin staging app tables...");
  const dropResult = spawnSync("psql", [databaseURL, "-v", "ON_ERROR_STOP=1", "-c", resetSQL], {
    stdio: "inherit",
    env,
  });
  if (dropResult.status !== 0) {
    console.error("Failed to drop Supperjumpin staging app tables.");
    return dropResult.status ?? 1;
  }

  console.log("Reapplying migrations...");
  const migrateResult = runMigrations(databaseURL, env);
  if (migrateResult.status !== 0) {
    console.error("Migrations failed.");
    return migrateResult.status ?? 1;
  }

  return 0;
}

function main() {
  process.exit(runStagingReset());
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}
