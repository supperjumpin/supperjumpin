import { test } from "node:test";
import assert from "node:assert";
import { mkdirSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { tmpdir } from "node:os";
import { parseJestCoveragePct } from "./test-coverage.mjs";

test("parseJestCoveragePct extracts lines pct from valid jest summary", () => {
  const dir = join(tmpdir(), `test-coverage-test-${Date.now()}`);
  mkdirSync(dir, { recursive: true });
  try {
    const summaryPath = join(dir, "coverage-summary.json");
    writeFileSync(
      summaryPath,
      JSON.stringify({
        total: {
          lines: { total: 254, covered: 237, skipped: 0, pct: 93.3 },
          statements: { total: 283, covered: 254, skipped: 0, pct: 89.75 },
        },
      })
    );
    assert.strictEqual(parseJestCoveragePct(summaryPath), 93.3);
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
});

test("parseJestCoveragePct returns null for non-existent file", () => {
  assert.strictEqual(parseJestCoveragePct("/nonexistent/path.json"), null);
});

test("parseJestCoveragePct returns null for invalid JSON", () => {
  const dir = join(tmpdir(), `test-coverage-test-${Date.now()}`);
  mkdirSync(dir, { recursive: true });
  try {
    const summaryPath = join(dir, "bad.json");
    writeFileSync(summaryPath, "not json");
    assert.strictEqual(parseJestCoveragePct(summaryPath), null);
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
});

test("parseJestCoveragePct returns null when lines pct is missing", () => {
  const dir = join(tmpdir(), `test-coverage-test-${Date.now()}`);
  mkdirSync(dir, { recursive: true });
  try {
    const summaryPath = join(dir, "incomplete.json");
    writeFileSync(
      summaryPath,
      JSON.stringify({ total: { statements: { pct: 50 } } })
    );
    assert.strictEqual(parseJestCoveragePct(summaryPath), null);
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
});
