import { spawnSync } from "node:child_process";
import { appendFileSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import {
  buildAdminURL,
  DEFAULT_TEST_DATABASE_URL,
  describeDatabaseURL,
  parseDatabaseName,
  runMigrations,
  runPsqlCommand,
  waitForPostgresReady,
} from "./db-helpers.mjs";

const API_MODULE = "github.com/supperjumpin/supperjumpin/apps/api";
const COVERAGE_SUMMARY_EXCLUSIONS = new Set(["cmd/api", "internal/db"]);

function packageCoverageRows(profilePath) {
  const packages = new Map();
  for (const line of readFileSync(profilePath, "utf8").trim().split("\n").slice(1)) {
    const [range, statementCountText, countText] = line.split(" ");
    const filePath = range.slice(0, range.lastIndexOf(":"));
    const packagePath = filePath.slice(0, filePath.lastIndexOf("/"));
    const packageName = packagePath.startsWith(`${API_MODULE}/`)
      ? packagePath.slice(API_MODULE.length + 1)
      : packagePath;
    const statementCount = Number(statementCountText);
    const count = Number(countText);
    const coverage = packages.get(packageName) ?? { covered: 0, total: 0 };
    coverage.total += statementCount;
    if (count > 0) {
      coverage.covered += statementCount;
    }
    packages.set(packageName, coverage);
  }

  return [...packages.entries()]
    .filter(([packageName]) => !COVERAGE_SUMMARY_EXCLUSIONS.has(packageName))
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([packageName, coverage]) => ({
      packageName,
      percent: coverage.total === 0 ? 0 : (coverage.covered / coverage.total) * 100,
    }));
}

function formatCoverageSummary(rows, totalPercent) {
  const lines = ["Go API coverage by package:"];
  for (const row of rows) {
    const percent = row.percent.toFixed(1);
    const prefix = percent === "0.0" ? "WARNING: " : "";
    lines.push(`${prefix}${row.packageName}: ${percent}%`);
  }
  lines.push(
    totalPercent
      ? `total: (statements) ${totalPercent}%`
      : "total: (statements) unavailable"
  );
  return `${lines.join("\n")}\n`;
}

function filterKnownCoverageNoise(output) {
  const lines = output.split("\n");
  const filtered = lines.filter((line) => {
    const trimmed = line.trim();
    for (const packageName of COVERAGE_SUMMARY_EXCLUSIONS) {
      if (
        trimmed.startsWith(`${API_MODULE}/${packageName}`) &&
        trimmed.includes("coverage: 0.0% of statements")
      ) {
        return false;
      }
    }
    return true;
  });
  return filtered.join("\n");
}

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
  console.log(`Test database: ${describeDatabaseURL(testDatabaseURL)}`);
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
  const migrateResult = runMigrations(testDatabaseURL, { ...env, DATABASE_URL: undefined });
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
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
    env: {
      ...env,
      SUPPERJUMPIN_TEST_DATABASE_URL: testDatabaseURL,
    },
  });

  if (goTestResult.stdout) {
    const stdout = goTestResult.status === 0 && coverageEnabled
      ? filterKnownCoverageNoise(goTestResult.stdout)
      : goTestResult.stdout;
    process.stdout.write(stdout);
  }
  if (goTestResult.stderr) {
    process.stderr.write(goTestResult.stderr);
  }

  if (goTestResult.status === 0 && coverageEnabled) {
    const coverResult = spawnSync("go", ["tool", "cover", "-func=coverage/api.coverprofile"], {
      encoding: "utf8",
      stdio: ["ignore", "pipe", "pipe"],
      env,
    });

    if (coverResult.stderr) {
      process.stderr.write(coverResult.stderr);
    }

    const summaryMatch = `${coverResult.stdout ?? ""}${coverResult.stderr ?? ""}`.match(
      /total:\s+\(statements\)\s+([\d.]+)%/
    );
    const coverageSummary = formatCoverageSummary(
      packageCoverageRows("coverage/api.coverprofile"),
      summaryMatch?.[1]
    );
    process.stdout.write(coverageSummary);
    if (env.GITHUB_STEP_SUMMARY) {
      appendFileSync(env.GITHUB_STEP_SUMMARY, `### Go API coverage\n\n\`\`\`\n${coverageSummary}\`\`\`\n`);
    }

    if (summaryMatch) {
      mkdirSync("coverage", { recursive: true });
      writeFileSync("coverage/go-report.json", JSON.stringify({ total: Number(summaryMatch[1]) }) + "\n");
    }
  }

  process.exit(goTestResult.status ?? 0);
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}
