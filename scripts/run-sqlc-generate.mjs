import { existsSync } from "node:fs";
import { join, resolve } from "node:path";
import { spawnSync } from "node:child_process";

const binDir = process.env.SUPPERJUMPIN_SQLC_BIN_DIR ?? "bin";
const sqlcPath = join(resolve(binDir), "sqlc");

if (!existsSync(sqlcPath)) {
  console.error(`Local sqlc binary not found at ${sqlcPath}. run \`npm run setup\` first.`);
  process.exit(1);
}

const result = spawnSync(sqlcPath, ["generate"], {
  cwd: "apps/api",
  stdio: "inherit",
});

process.exit(result.status ?? 0);
