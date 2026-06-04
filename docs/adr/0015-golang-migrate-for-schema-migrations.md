# golang-migrate for Schema Migrations

We use golang-migrate to run Postgres schema migrations. The bespoke Node.js migration runner previously embedded in `scripts/demo-api.mjs` was hand-rolled, not wired into CI, and introduced a Node.js dependency into a backend concern that is otherwise entirely Go. It is now historical.

golang-migrate is Go-native, provides both a CLI and a library, and integrates cleanly with the existing stack. It runs migrations as a separate CLI step in CI and production deploys, and via `npm run db:migrate` locally. goose was the main alternative considered; both are suitable, but golang-migrate is more widely used and requires less opinion about versioning conventions.

The migration files in `apps/api/db/migrations/` use a numeric prefix naming scheme compatible with golang-migrate.
