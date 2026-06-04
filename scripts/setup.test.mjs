import { test } from "node:test";
import assert from "node:assert";
import { parseSemver } from "./setup.mjs";

test("parseSemver extracts major, minor, patch from x.y.z", () => {
  const result = parseSemver("1.26.3");
  assert.deepStrictEqual(result, { major: 1, minor: 26, patch: 3 });
});

test("parseSemver strips leading v", () => {
  const result = parseSemver("v24.16.0");
  assert.deepStrictEqual(result, { major: 24, minor: 16, patch: 0 });
});

import { checkVersionRange } from "./setup.mjs";

test("checkVersionRange returns true for version within range", () => {
  assert.strictEqual(checkVersionRange("24.16.0", ">=24.16.0 <25"), true);
});

test("checkVersionRange returns false for version below minimum", () => {
  assert.strictEqual(checkVersionRange("24.15.9", ">=24.16.0 <25"), false);
});

test("checkVersionRange returns false for version at maximum boundary", () => {
  assert.strictEqual(checkVersionRange("25.0.0", ">=24.16.0 <25"), false);
});

import { formatFailureMessage } from "./setup.mjs";

test("formatFailureMessage includes tool name, current version, and requirement", () => {
  const msg = formatFailureMessage("Node", "24.15.0", ">=24.16.0 <25");
  assert.ok(msg.includes("Node"));
  assert.ok(msg.includes("24.15.0"));
  assert.ok(msg.includes(">=24.16.0 <25"));
});

import { checkExactVersion, formatExactFailureMessage, extractSqlcVersion, extractMigrateVersion } from "./setup.mjs";

test("checkExactVersion returns true for exact match", () => {
  assert.strictEqual(checkExactVersion("1.31.1", "1.31.1"), true);
});

test("checkExactVersion returns false for mismatch", () => {
  assert.strictEqual(checkExactVersion("1.30.0", "1.31.1"), false);
});

test("formatExactFailureMessage includes tool name, current version, and expected version", () => {
  const msg = formatExactFailureMessage("sqlc", "1.30.0", "1.31.1");
  assert.ok(msg.includes("sqlc"));
  assert.ok(msg.includes("1.30.0"));
  assert.ok(msg.includes("1.31.1"));
});

test("extractSqlcVersion strips leading v", () => {
  assert.strictEqual(extractSqlcVersion("v1.31.1\n"), "1.31.1");
});

test("extractSqlcVersion handles plain semver", () => {
  assert.strictEqual(extractSqlcVersion("1.31.1"), "1.31.1");
});

test("extractMigrateVersion strips leading v", () => {
  assert.strictEqual(extractMigrateVersion("v4.19.1\n"), "4.19.1");
});

test("extractMigrateVersion handles plain semver", () => {
  assert.strictEqual(extractMigrateVersion("4.19.1"), "4.19.1");
});

import { checkDockerRunning } from "./setup.mjs";

test("checkDockerRunning returns ok when docker info succeeds", () => {
  const mockCapture = () => ({ status: 0 });
  const result = checkDockerRunning(mockCapture);
  assert.strictEqual(result.ok, true);
  assert.ok(result.message.includes("running"));
});

test("checkDockerRunning returns not ok when docker info fails", () => {
  const mockCapture = () => ({ status: 1 });
  const result = checkDockerRunning(mockCapture);
  assert.strictEqual(result.ok, false);
  assert.ok(result.message.includes("not running"));
});
