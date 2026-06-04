import { test } from "node:test";
import assert from "node:assert";

import {
  getTestDatabaseURL,
  isSafeToReset,
} from "./api-test.mjs";
import { buildAdminURL, parseDatabaseName } from "./db-helpers.mjs";

test("parseDatabaseName extracts db name from postgres URL", () => {
  assert.strictEqual(
    parseDatabaseName("postgres://user:pass@localhost:5432/supperjumpin_test?sslmode=disable"),
    "supperjumpin_test"
  );
});

test("parseDatabaseName extracts db name from URL without query params", () => {
  assert.strictEqual(
    parseDatabaseName("postgres://user:pass@localhost:5432/mydb"),
    "mydb"
  );
});

test("isSafeToReset allows db names ending with _test", () => {
  assert.strictEqual(isSafeToReset("supperjumpin_test", false), true);
  assert.strictEqual(isSafeToReset("myapp_test", false), true);
});

test("isSafeToReset refuses non-test db names", () => {
  assert.strictEqual(isSafeToReset("supperjumpin", false), false);
  assert.strictEqual(isSafeToReset("production", false), false);
});

test("isSafeToReset allows override with allowUnsafe", () => {
  assert.strictEqual(isSafeToReset("supperjumpin", true), true);
  assert.strictEqual(isSafeToReset("production", true), true);
});

test("getTestDatabaseURL uses SUPPERJUMPIN_TEST_DATABASE_URL when set", () => {
  const env = {
    SUPPERJUMPIN_TEST_DATABASE_URL: "postgres://u:p@host:5432/db_test",
  };
  assert.strictEqual(getTestDatabaseURL(env), "postgres://u:p@host:5432/db_test");
});

test("getTestDatabaseURL falls back to local Docker Compose URL", () => {
  const env = {};
  assert.strictEqual(
    getTestDatabaseURL(env),
    "postgres://postgres:postgres@localhost:5432/supperjumpin_test?sslmode=disable"
  );
});

test("buildAdminURL replaces db name with postgres", () => {
  assert.strictEqual(
    buildAdminURL("postgres://user:pass@localhost:5432/supperjumpin_test?sslmode=disable"),
    "postgres://user:pass@localhost:5432/postgres?sslmode=disable"
  );
});

test("buildAdminURL works without query params", () => {
  assert.strictEqual(
    buildAdminURL("postgres://u:p@host:5432/mydb"),
    "postgres://u:p@host:5432/postgres"
  );
});
