import { test } from "node:test";
import assert from "node:assert";
import { mkdirSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { tmpdir } from "node:os";
import { pct, deltaIcon, loadReport } from "./coverage-diff.mjs";

test("pct formats a number as a percentage string", () => {
  assert.strictEqual(pct(93.3), "93.3%");
  assert.strictEqual(pct(100), "100.0%");
  assert.strictEqual(pct(0), "0.0%");
});

test("pct returns em-dash for null or undefined", () => {
  assert.strictEqual(pct(null), "—");
  assert.strictEqual(pct(undefined), "—");
});

test("deltaIcon returns check for positive delta > 0.5", () => {
  assert.ok(deltaIcon(1.0).includes("✅"));
});

test("deltaIcon returns fire for negative delta < -0.5", () => {
  assert.ok(deltaIcon(-1.0).includes("🔻"));
});

test("deltaIcon returns empty for small deltas", () => {
  assert.strictEqual(deltaIcon(0.3), "");
  assert.strictEqual(deltaIcon(-0.3), "");
  assert.strictEqual(deltaIcon(0), "");
});

test("deltaIcon returns empty for null delta", () => {
  assert.strictEqual(deltaIcon(null), "");
});

test("loadReport returns parsed JSON when file exists", () => {
  const dir = join(tmpdir(), `coverage-diff-test-${Date.now()}`);
  mkdirSync(dir, { recursive: true });
  try {
    writeFileSync(join(dir, "mobile-report.json"), JSON.stringify({ total: 93.3 }));
    const report = loadReport(dir, "mobile");
    assert.deepStrictEqual(report, { total: 93.3 });
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
});

test("loadReport returns null when file does not exist", () => {
  const dir = join(tmpdir(), `coverage-diff-test-nonexistent-${Date.now()}`);
  const report = loadReport(dir, "mobile");
  assert.strictEqual(report, null);
});
