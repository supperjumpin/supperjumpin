# apps/api KNOWLEDGE BASE

## OVERVIEW

Go backend API for Supperjumpin. Owns game rules, durable domain state, and the REST/OpenAPI contract consumed by the mobile app.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Entry point / wiring | `cmd/api/main.go` | Env vars: PORT, DATABASE_URL, SUPPERJUMPIN_DEV_AUTH_TOKEN |
| Add API endpoint | `internal/httpapi/server.go` | Closures over ServerConfig; call transport helpers in `store.go` |
| Change DTO / JSON shape | `internal/httpapi/store.go` | DTO structs, transport helpers, error mapping |
| In-memory tests | `internal/httpapi/store.go` | `MemoryStore` implements full `Persistence` interface |
| Production persistence | `internal/httpapi/postgres_store.go` | Raw SQL, transactions, DTO assembly |
| Game rules / domain logic | `internal/game/*.go` | Pure functions, repository interfaces, no HTTP/DB imports |
| DB schema | `db/migrations/*.sql` | 9 numbered migrations; pre-stable: fold changes into existing |
| API contract | `openapi.yaml` | Source of truth for generated TypeScript client |
| sqlc config | `sqlc.yaml` | Configured and generating into `internal/db/` |

## CONVENTIONS

- **Standard library HTTP only**: `net/http` + `http.NewServeMux()` with Go 1.22 path patterns. No Gin, Echo, or Fiber.
- **Auth middleware pattern**: Every protected route calls `signedInProfile(w, r, config)` first. Bearer token from `Authorization` header.
- **Transport helpers** in `store.go` bridge between game snapshots and JSON DTOs. Example: `createGroup()` calls `game.CreateGroup()` then assembles `GroupHomeResponse`.
- **Error mapping**: `mapGameErr()` in `store.go` translates domain errors (`game.ErrInvalidJudgmentScore`) to transport errors (`httpapi.ErrInvalidJudgmentScore`) for HTTP status codes.
- **Transactions**: Multi-step DB operations use `BeginTx` + `defer tx.Rollback()` + `tx.Commit()`.

## ANTI-PATTERNS

- Adding business logic to `server.go` handlers. Keep handlers thin; delegate to transport helpers, which delegate to `internal/game/`.
- Using an ORM or query builder. Raw SQL in `postgres_store.go` is the current pattern.
- Modifying `openapi.yaml` without regenerating the TypeScript client. CI will fail.
- Creating new numbered migration files before DB stability. Fold into existing table-creation migrations.

## NOTES

- `MemoryStore` and `PostgresStore` must stay in sync — both implement `Persistence`. Any new repository method needs both implementations.
- `PostgresStore` uses `database/sql` with `pgx` driver, not `pgxpool` directly.
- `stableID(kind, value)` generates deterministic IDs. Never use UUIDs or auto-increment for domain entities.
