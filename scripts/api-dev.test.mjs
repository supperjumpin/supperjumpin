import { test } from "node:test";
import assert from "node:assert";
import { mkdtempSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";

const scriptPath = new URL("./api-dev.mjs", import.meta.url).pathname;
const DEFAULT_DATABASE_URL = "postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable";

function makeFakeGo(binDir) {
	const goPath = join(binDir, "go");
	writeFileSync(
		goPath,
		`#!/bin/sh
printf '%s\n' "GO_DATABASE_URL=\${DATABASE_URL}"
printf '%s\n' "GO_SUPPERJUMPIN_DATABASE_URL=\${SUPPERJUMPIN_DATABASE_URL}"
printf '%s\n' "GO_SUPPERJUMPIN_DEV_AUTH_TOKEN=\${SUPPERJUMPIN_DEV_AUTH_TOKEN}"
`,
		{ mode: 0o755 }
	);
}

function runScript(binDir, env = {}) {
	return spawnSync("node", [scriptPath], {
		encoding: "utf8",
		env: {
			...process.env,
			PATH: `${binDir}:${process.env.PATH}`,
			...env,
		},
	});
}

test("passes the local database and default dev token to the API binary", () => {
	const binDir = mkdtempSync(join(tmpdir(), "sj-test-"));
	makeFakeGo(binDir);

	const result = runScript(binDir);

	assert.strictEqual(result.status, 0, `Expected exit 0. stderr: ${result.stderr}`);
	const output = `${result.stdout}${result.stderr}`;
	assert.ok(output.includes("Starting Supperjumpin API on http://localhost:8080"), `Missing startup log. Got: ${output}`);
	assert.ok(output.includes("Demo bearer token: dev-token"), `Missing default token log. Got: ${output}`);
	assert.ok(output.includes("API database: localhost:5432/supperjumpin"), `Missing database log. Got: ${output}`);
	assert.ok(output.includes(`GO_SUPPERJUMPIN_DATABASE_URL=${DEFAULT_DATABASE_URL}`), `Expected local database URL in child env. Got: ${output}`);
	assert.ok(output.includes("GO_SUPPERJUMPIN_DEV_AUTH_TOKEN=dev-token"), `Expected default dev token in child env. Got: ${output}`);
});

test("ignores ambient DATABASE_URL and still starts against local Docker Postgres", () => {
	const binDir = mkdtempSync(join(tmpdir(), "sj-test-"));
	makeFakeGo(binDir);

	const result = runScript(binDir, {
		DATABASE_URL: "postgres://user:pass@staging:5432/supperjumpin?sslmode=require",
	});

	assert.strictEqual(result.status, 0, `Expected exit 0. stderr: ${result.stderr}`);
	const output = `${result.stdout}${result.stderr}`;
	assert.ok(output.includes("Ignoring ambient DATABASE_URL"), `Expected warning about ignored DATABASE_URL. Got: ${output}`);
	assert.ok(output.includes(`GO_SUPPERJUMPIN_DATABASE_URL=${DEFAULT_DATABASE_URL}`), `Expected local database URL in child env. Got: ${output}`);
	assert.ok(!output.includes("staging:5432"), `Expected ambient DATABASE_URL to be ignored. Got: ${output}`);
});
