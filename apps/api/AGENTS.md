# apps/api Guide

Go backend API for Supperjumpin. Owns domain logic, durable state, and the REST/OpenAPI contract.

## Where To Look

| Task | Location | Notes |
|------|----------|-------|
| Entry point / wiring | `cmd/api/main.go` | Env vars: PORT, SUPPERJUMPIN_DATABASE_URL, SUPPERJUMPIN_ADAPTER_TOKEN, SUPPERJUMPIN_LOG_FORMAT, SUPPERJUMPIN_LOG_LEVEL |
| Add API endpoint | `internal/httpapi/server.go` | Closures over ServerConfig |
| Change DTO / JSON shape | `internal/httpapi/dto.go` | camelCase JSON tags |
| Postgres-backed tests | `mage test` | Canonical test path against Postgres; API prep is built into the root target |
| Persistence adapter | `internal/httpapi/postgres_store.go` | sqlc-generated queries via `db.Queries` |
| External identity resolution | `internal/httpapi/external_identity.go` | Adapter-owned mapping from platform actors to (player, community) |
| Game rules / domain logic | `internal/game/*.go` | Pure functions, repository interfaces, no HTTP/DB imports |
| DB schema | `db/migrations/*.sql` | Communities, players, external_identity, prompt_packs, prompts, reveal_timeframes, rounds, commits, jumps, jump_evidence, stamps, reactions, comments. Pre-stable: fold schema changes into existing migration files |
| API contract | `openapi.yaml` | Source of truth for the HTTP contract |
| sqlc config & generation | `db/queries/*.sql` → `sqlc.yaml` → `internal/db/` | Source `.sql` files; generated Go in `internal/db/` |

`internal/game/AGENTS.md` owns domain-flow specifics. `internal/httpapi/AGENTS.md` owns route/DTO/logging specifics.

## Core Rules

- **Standard library HTTP only**: `net/http` + `http.NewServeMux()` with Go 1.22 path patterns. No Gin, Echo, or Fiber.
- **Auth middleware pattern**: Every protected route calls `signedInProfile(w, r, config)` first. Bearer token from `Authorization` plus `X-Adapter-Actor: discord:<guildID>:<userID>`. MVP development uses `SUPPERJUMPIN_ADAPTER_TOKEN` for local-first adapter auth.
- **Error mapping**: Domain errors are mapped to HTTP status codes in transport helpers.
- **sqlc for queries**: All repository interface methods delegate to `s.queries.*` (generated `*db.Queries`). Add/modify a query → edit its `.sql` file in `db/queries/`, run `mage generate:sqlc`.
- **Transactions**: Multi-step DB operations use `BeginTx` + `defer tx.Rollback()` + `tx.Commit()` with `qtx := s.queries.WithTx(tx)`.

## Logging

The API uses `log/slog` structured JSON logs on a **unit-of-work** model. The middleware emits one final log line per request using accumulated metadata and the highest severity raised during the request.

### Route handler checklist

1. **Set operation metadata** by calling `setRequestOperation(r, "<METHOD /path>", "<snake_case_operation>")` at the top of the handler.
2. **Add domain entity IDs** with `AddRequestLogField(r.Context(), ...)` when known from path params or DB results.
3. **Set actor identity** via `signedInProfile()` (authenticated), `optionalProfile()` (optional auth), or manually add `AddRequestLogField(r.Context(), "actor_type", "public")` for unauthenticated endpoints.
4. **Call `recordHTTPError(r, status, code, err)`** on every error return.

### Approved log fields

| Field | Meaning | Set by |
|-------|---------|--------|
| `actor_type` | `"player"`, `"guest"`, or `"public"` | `signedInProfile` / handler |
| `player_id` | Stable player ID | `signedInProfile` |
| `route` | HTTP method + path pattern | `setRequestOperation` |
| `operation` | Short snake_case action name | `setRequestOperation` |
| `outcome` | `"success"`, `"client_error"`, `"forbidden"`, `"not_found"`, `"conflict"`, `"server_error"` | `recordHTTPError` / middleware |
| `error_code` | Machine-readable error | `recordHTTPError` |
| `internal_error` | Redacted internal error text (5xx only) | `recordHTTPError` |
| `stack` | Full goroutine stack trace (5xx / panic only) | `recordHTTPError` / middleware |

The middleware automatically adds `request_id`, `method`, `path`, `status`, `response_bytes`, `duration_ms`, and `user_agent`.

### Forbidden log fields

These must **never** appear in log output:
- Bearer tokens, session tokens, or any auth material
- Email addresses, player display names, or personally identifiable text
- Raw request bodies or response bodies
- SQL query text or query parameters

### Logger-free domain

`internal/game/` must remain logger-free. Domain functions return structured errors that the transport layer maps to log fields.

## Avoid

- Adding business logic to `server.go` handlers. Keep handlers thin; delegate to transport helpers, which delegate to `internal/game/`.
- Using an ORM or query builder. sqlc-generated queries is the current pattern.
- Modifying `openapi.yaml` without keeping the handlers and DTOs aligned with it.
- Creating new numbered migration files before DB stability. Fold into existing table-creation migrations.

## Notes

- `PostgresStore` uses `database/sql` with `pgx` driver, not `pgxpool` directly.
- `stableID(kind, value)` generates deterministic IDs. Never use UUIDs or auto-increment for domain entities.
- `internal/db/` is the generated sqlc package. It must never import `internal/httpapi` or `internal/game`.
