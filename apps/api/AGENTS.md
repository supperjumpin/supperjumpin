# apps/api KNOWLEDGE BASE

## OVERVIEW

Go backend API for Supperjumpin. Owns game rules, durable domain state, and the REST/OpenAPI contract consumed by the mobile app.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Entry point / wiring | `cmd/api/main.go` | Env vars: PORT, SUPPERJUMPIN_DATABASE_URL, SUPPERJUMPIN_DEV_AUTH_TOKEN, SUPPERJUMPIN_LOG_FORMAT, SUPPERJUMPIN_LOG_LEVEL. Most npm scripts use local Docker Postgres by default and pass `SUPPERJUMPIN_DATABASE_URL` explicitly; `npm run db:migrate` is local-only. |
| Add API endpoint | `internal/httpapi/server.go` | Closures over ServerConfig; call transport helpers in `store.go` |
| Change DTO / JSON shape | `internal/httpapi/dto.go` | DTO structs with camelCase JSON tags |
| Postgres-backed tests | `npm run api:test` / `npm run api:test:coverage` | Canonical test path against Postgres; see root AGENTS.md |
| Persistence adapter | `internal/httpapi/postgres_store*.go` | sqlc-generated queries via `db.Queries`; per-repository files |
| Game rules / domain logic | `internal/game/*.go` | Pure functions, repository interfaces, no HTTP/DB imports |
| DB schema | `db/migrations/*.sql` | Pre-stable: fold schema changes into existing migration files |
| API contract | `openapi.yaml` | Source of truth for generated TypeScript client |
| sqlc config & generation | `db/queries/*.sql` → `sqlc.yaml` → `internal/db/` | Source `.sql` files; generated Go in `internal/db/` |

## CONVENTIONS

- **Standard library HTTP only**: `net/http` + `http.NewServeMux()` with Go 1.22 path patterns. No Gin, Echo, or Fiber.
- **Auth middleware pattern**: Every protected route calls `signedInProfile(w, r, config)` first. Bearer token from `Authorization` header. MVP development uses `SUPPERJUMPIN_DEV_AUTH_TOKEN` for local-first auth; hosted auth will be additive when introduced.
- **Transport helpers** in `store.go` bridge between game snapshots and JSON DTOs from `dto.go`. Example: `createPerformedJump()` calls `game.CreatePerformedJump()` then assembles a `Jump` response.
- **Error mapping**: `mapGameErr()` in `store.go` translates domain errors (`game.ErrInvalidJudgmentScore`) to transport errors (`httpapi.ErrInvalidJudgmentScore`) for HTTP status codes.
- **sqlc for queries**: All repository interface methods in `postgres_store_*.go` delegate to `s.queries.*` (generated `*db.Queries`). Add/modify a query → edit its `.sql` file in `db/queries/`, run `npm run generate:sqlc`.
- **Transactions**: Multi-step DB operations use `BeginTx` + `defer tx.Rollback()` + `tx.Commit()` with `qtx := s.queries.WithTx(tx)` for sqlc-generated queries inside the transaction.
- **Complex read queries**: Multi-table DTO assembly queries for public read paths may stay as hand-written raw SQL when sqlc would make the read model harder to follow.

## ANTI-PATTERNS

- Adding business logic to `server.go` handlers. Keep handlers thin; delegate to transport helpers, which delegate to `internal/game/`.
- Using an ORM or query builder. sqlc-generated queries in `postgres_store*.go` is the current pattern.
- Modifying `openapi.yaml` without regenerating the TypeScript client. CI will fail.
- Creating new numbered migration files before DB stability. Fold into existing table-creation migrations.

## NOTES

- `PostgresStore` is the durable persistence adapter. Any new repository method needs a Postgres implementation, and unit tests should use narrow fakes instead of a shared in-memory store.
- `PostgresStore` uses `database/sql` with `pgx` driver, not `pgxpool` directly.
- `stableID(kind, value)` generates deterministic IDs. Never use UUIDs or auto-increment for domain entities.
- `internal/db/` is the generated sqlc package. It must never import `internal/httpapi` or `internal/game` — CI enforces this with a freshness check.
- `npm run api:test:coverage` writes the complete `coverage/api.coverprofile`; the human package summary excludes only `cmd/api` and generated `internal/db` while keeping total coverage visible.
