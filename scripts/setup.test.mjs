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
