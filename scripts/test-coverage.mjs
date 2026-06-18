import { appendFileSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

function runApiClientCoverage(env) {
  const result = spawnSync(
    "npm",
    ["--workspace", "@supperjumpin/api-client", "run", "test:coverage"],
    {
      encoding: "utf8",
      env,
    }
  );

  if (result.stdout) {
    process.stdout.write(result.stdout);
  }
  if (result.stderr) {
    process.stderr.write(result.stderr);
  }

  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }

  const output = `${result.stdout ?? ""}${result.stderr ?? ""}`;
  const match = output.match(/all files \|\s+([\d.]+)\s+\|/);
  const line = match
    ? `Node workspace coverage: ${match[1]}% line coverage.`
    : "Node workspace coverage completed.";

  const summaryPath = env.GITHUB_STEP_SUMMARY;
  if (summaryPath) {
    appendFileSync(summaryPath, `### Node workspace coverage\n${line}\n`);
  }

  if (match) {
    mkdirSync("coverage", { recursive: true });
    writeFileSync(
      "coverage/node-report.json",
      JSON.stringify({ total: Number(match[1]) }, null, 2) + "\n"
    );
  }
}

export function parseJestCoveragePct(jestSummaryPath) {
  try {
    const summary = JSON.parse(readFileSync(jestSummaryPath, "utf8"));
    return summary?.total?.lines?.pct ?? null;
  } catch {
    return null;
  }
}

function runMobileCoverage(env) {
  const result = spawnSync(
    "npm",
    ["--workspace", "@supperjumpin/mobile", "run", "test:coverage"],
    {
      encoding: "utf8",
      env,
    }
  );

  if (result.stdout) {
    process.stdout.write(result.stdout);
  }
  if (result.stderr) {
    process.stderr.write(result.stderr);
  }

  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }

  const pct = parseJestCoveragePct("coverage/jest/coverage-summary.json");

  const line =
    pct != null
      ? `Mobile coverage: ${pct}% line coverage.`
      : "Mobile coverage completed.";

  const summaryPath = env.GITHUB_STEP_SUMMARY;
  if (summaryPath) {
    appendFileSync(summaryPath, `### Mobile coverage\n${line}\n`);
  }

  if (pct != null) {
    mkdirSync("coverage", { recursive: true });
    writeFileSync(
      "coverage/mobile-report.json",
      JSON.stringify({ total: Number(pct) }, null, 2) + "\n"
    );
  }
}

function main() {
  const env = process.env;
  runApiClientCoverage(env);
  runMobileCoverage(env);
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}
