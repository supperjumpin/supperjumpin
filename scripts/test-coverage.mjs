import { appendFileSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

function main() {
  const env = process.env;
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
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}
