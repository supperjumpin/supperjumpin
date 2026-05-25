import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { spawn, spawnSync } from "node:child_process";

const skipDatabaseSetup = process.argv.includes("--no-db");
const dockerDatabaseURL = "postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable";
const databaseURL = skipDatabaseSetup ? process.env.DATABASE_URL ?? dockerDatabaseURL : dockerDatabaseURL;
const devToken = process.env.SUPPERJUMPIN_DEV_AUTH_TOKEN ?? "dev-token";

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    encoding: "utf8",
    stdio: options.input ? ["pipe", "inherit", "inherit"] : "inherit",
    input: options.input,
  });

  if (result.status !== 0) {
    throw new Error(`${command} ${args.join(" ")} failed`);
  }
}

function capture(command, args) {
  return spawnSync(command, args, {
    encoding: "utf8",
  });
}

function dockerCompose(args, options = {}) {
  return run("docker", ["compose", ...args], options);
}

function assertDockerAvailable() {
  const docker = capture("docker", ["--version"]);
  if (docker.status !== 0) {
    throw new Error("Docker is required for npm run demo:api. Install Docker Desktop/Engine with Compose, or set DATABASE_URL and use npm run api:dev.");
  }

  const compose = capture("docker", ["compose", "version"]);
  if (compose.status !== 0) {
    throw new Error("Docker Compose is required for npm run demo:api. Install Docker Desktop/Engine with Compose, or set DATABASE_URL and use npm run api:dev.");
  }
}

function waitForPostgres() {
  for (let attempt = 1; attempt <= 80; attempt += 1) {
    const result = capture("docker", [
      "compose",
      "exec",
      "-T",
      "postgres",
      "pg_isready",
      "-U",
      "postgres",
      "-d",
      "supperjumpin",
    ]);

    if (result.status === 0) {
      return;
    }

    Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 500);
  }

  throw new Error("Postgres did not become ready in time");
}

function applyMigrations() {
  const migrationsDirectory = join("apps", "api", "db", "migrations");
  const migrations = readdirSync(migrationsDirectory)
    .filter((file) => file.endsWith(".sql"))
    .sort();

  const bootstrap = "CREATE TABLE IF NOT EXISTS schema_migrations (filename TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now());";
  dockerCompose(["exec", "-T", "postgres", "psql", "-v", "ON_ERROR_STOP=1", "-U", "postgres", "-d", "supperjumpin"], {
    input: bootstrap,
  });

  for (const filename of migrations) {
    const applied = capture("docker", [
      "compose",
      "exec",
      "-T",
      "postgres",
      "psql",
      "-tAc",
      `SELECT 1 FROM schema_migrations WHERE filename = '${filename.replaceAll("'", "''")}'`,
      "-U",
      "postgres",
      "-d",
      "supperjumpin",
    ]);

    if (applied.stdout.trim() === "1") {
      continue;
    }

    const sql = readFileSync(join(migrationsDirectory, filename), "utf8");
    const escapedFilename = filename.replaceAll("'", "''");
    const migration = `BEGIN;\n${sql}\nINSERT INTO schema_migrations (filename) VALUES ('${escapedFilename}');\nCOMMIT;`;
    console.log(`Applying ${filename}`);
    dockerCompose(["exec", "-T", "postgres", "psql", "-v", "ON_ERROR_STOP=1", "-U", "postgres", "-d", "supperjumpin"], {
      input: migration,
    });
  }
}

try {
  if (!skipDatabaseSetup) {
    assertDockerAvailable();
    console.log("Starting Postgres with Docker Compose");
    dockerCompose(["up", "-d", "postgres"]);
    waitForPostgres();
    applyMigrations();
  }

  console.log("Starting Supperjumpin API on http://localhost:8080");
  console.log(`Demo bearer token: ${devToken}`);

  const api = spawn("go", ["run", "./apps/api/cmd/api"], {
    env: {
      ...process.env,
      DATABASE_URL: databaseURL,
      SUPPERJUMPIN_DEV_AUTH_TOKEN: devToken,
    },
    stdio: "inherit",
  });

  api.on("exit", (code) => {
    process.exit(code ?? 0);
  });
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error));
  process.exit(1);
}
