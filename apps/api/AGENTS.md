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

## LOGGING CONVENTIONS

The API uses `log/slog` structured JSON logs on a **unit-of-work** model. A unit of work is an externally meaningful API/domain operation (creating a Performed Jump, submitting a Judgment, loading the Feed), not an individual SQL query. The middleware emits one final log line per request using accumulated metadata and the highest severity raised during the request.

### Route handler checklist

Every new route handler must:

1. **Set operation metadata** by calling `setRequestOperation(r, "<METHOD /path>", "<snake_case_operation>")` at the top of the handler.
2. **Add domain entity IDs** with `AddRequestLogField(r.Context(), "jump_id", id)` (or `judgment_id`, `open_year`, `open_month`) when they are known from path params or DB results.
3. **Set actor identity** via `signedInProfile()` (authenticated), `optionalProfile()` (optional auth), or manually add `AddRequestLogField(r.Context(), "actor_type", "public")` for unauthenticated endpoints.
4. **Call `recordHTTPError(r, status, code, err)`** on every error return. This sets `outcome` and `error_code`, and raises severity to `ERROR` on 5xx.

### Approved log fields (safe)

These are the only field names the log accumulator accepts. Any other field name is silently dropped.

| Field | Meaning | Set by |
|-------|---------|--------|
| `actor_type` | `"player"`, `"guest"`, or `"public"` | `signedInProfile` / handler |
| `player_id` | Stable player ID | `signedInProfile` |
| `route` | HTTP method + path pattern (e.g. `GET /v1/me`) | `setRequestOperation` |
| `operation` | Short snake_case action name | `setRequestOperation` |
| `jump_id` | Performed Jump stable ID | Handler |
| `judgment_id` | Judgment stable ID | Handler |
| `open_year` | Open year from path | Handler |
| `open_month` | Open month from path | Handler |
| `outcome` | `"success"`, `"client_error"`, `"forbidden"`, `"not_found"`, `"conflict"`, `"server_error"` | `recordHTTPError` / middleware |
| `error_code` | Machine-readable error (e.g. `"not_authenticated"`, `"invalid_request"`) | `recordHTTPError` |
| `internal_error` | Redacted internal error text (5xx only) | `recordHTTPError` |
| `stack` | Full goroutine stack trace (5xx / panic only) | `recordHTTPError` / middleware |

The middleware automatically adds `request_id`, `method`, `path`, `status`, `response_bytes`, `duration_ms`, and `user_agent` to every log line.

### Forbidden log fields

These must **never** appear in log output, even indirectly:

- Bearer tokens, session tokens, or any auth material
- Email addresses, player display names, or personally identifiable text
- Guest session IDs
- Media object keys or evidence URLs
- Jump captions, Source, Destination, Food text, or any user-authored content
- Raw request bodies or response bodies
- SQL query text or query parameters
- Any content field from the domain layer

### Logger-free domain

`internal/game/` must remain logger-free. It must never import `log/slog` or manipulate request log context. Domain functions return structured errors that the transport layer maps to log fields.

Routine repository and SQL-query logging are out of scope. `PostgresStore` and `internal/db/` must not emit per-query logs unless a future issue deliberately adds targeted persistence diagnostics.

### Log format

Logs are JSON by default (`SUPPERJUMPIN_LOG_FORMAT=json`), use snake_case field names, and write to stderr. Use `SUPPERJUMPIN_LOG_FORMAT=text` for human-readable output in local dev. Level defaults to `info`; set `SUPPERJUMPIN_LOG_LEVEL=debug` for verbose output.

### Request ID

Every request gets a canonical UUID v4 request ID. If the caller sends a valid UUID in `X-Request-ID`, it is preserved; otherwise a new UUID is generated with `crypto/rand`. The request ID is attached to the request context (accessible via `RequestIDFromContext(ctx)`), returned in the `X-Request-ID` response header, and included in every unit-of-work log as `request_id`.

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
