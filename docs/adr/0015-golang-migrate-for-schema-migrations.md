# golang-migrate for Schema Migrations

We will use golang-migrate to run Postgres schema migrations rather than the bespoke Node.js migration runner currently embedded in `scripts/demo-api.mjs`. The existing runner is hand-rolled, not wired into CI, and introduces a Node.js dependency into a backend concern that is otherwise entirely Go.

golang-migrate is Go-native, provides both a CLI and a library, and integrates cleanly with the existing stack. It can run migrations as a separate CLI step in CI and production deploys, or from `main.go` on startup. goose was the main alternative considered; both are suitable, but golang-migrate is more widely used and requires less opinion about versioning conventions.

The existing migration files in `apps/api/db/migrations/` use a numeric prefix naming scheme compatible with golang-migrate.
