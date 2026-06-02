import { test } from "node:test";
import assert from "node:assert";
import { mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";

const scriptPath = new URL("./db-lifecycle.mjs", import.meta.url).pathname;

function makeFakeDocker(binDir, logPath) {
  const dockerPath = join(binDir, "docker");
  writeFileSync(
    dockerPath,
    `#!/bin/sh
printf '%s\n' "$@" > "${logPath}"
`,
    { mode: 0o755 }
  );
}

test("down mode stops local postgres without deleting data", () => {
  const binDir = mkdtempSync(join(tmpdir(), "sj-test-"));
  const logPath = join(binDir, "docker-args.log");
  makeFakeDocker(binDir, logPath);

  const result = spawnSync("node", [scriptPath, "down"], {
    encoding: "utf8",
    env: {
      ...process.env,
      PATH: `${binDir}:${process.env.PATH}`,
    },
  });

  assert.strictEqual(result.status, 0, `Expected exit 0. stderr: ${result.stderr}`);
  assert.strictEqual(
    readFileSync(logPath, "utf8"),
    "compose\nstop\npostgres\n",
    "Expected down mode to run 'docker compose stop postgres'"
  );
});

test("up mode starts local postgres in detached mode", () => {
  const binDir = mkdtempSync(join(tmpdir(), "sj-test-"));
  const logPath = join(binDir, "docker-args.log");
  makeFakeDocker(binDir, logPath);

  const result = spawnSync("node", [scriptPath, "up"], {
    encoding: "utf8",
    env: {
      ...process.env,
      PATH: `${binDir}:${process.env.PATH}`,
    },
  });

  assert.strictEqual(result.status, 0, `Expected exit 0. stderr: ${result.stderr}`);
  assert.strictEqual(
    readFileSync(logPath, "utf8"),
    "compose\nup\n-d\npostgres\n",
    "Expected up mode to run 'docker compose up -d postgres'"
  );
});
