import { test } from "node:test";
import assert from "node:assert";
import { TOOL_VERSIONS } from "./tool-versions.mjs";

test("TOOL_VERSIONS defines exact version for sqlc", () => {
  assert.strictEqual(TOOL_VERSIONS.sqlc, "1.31.1");
});

test("TOOL_VERSIONS defines exact version for golang-migrate", () => {
  assert.strictEqual(TOOL_VERSIONS["golang-migrate"], "4.19.1");
});

test("TOOL_VERSIONS defines exact version for postgres", () => {
  assert.strictEqual(TOOL_VERSIONS.postgres, "16");
});

test("TOOL_VERSIONS defines range for node", () => {
  assert.strictEqual(TOOL_VERSIONS.node, ">=24.16.0 <25");
});

test("TOOL_VERSIONS defines range for go", () => {
  assert.strictEqual(TOOL_VERSIONS.go, ">=1.26.3 <1.27");
});
