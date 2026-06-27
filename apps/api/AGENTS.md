# apps/api KNOWLEDGE BASE

## OVERVIEW

Go backend API for Supperjumpin. Owns domain logic, durable state, and the REST/OpenAPI contract.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Entry point / wiring | `cmd/api/main.go` | Env vars: PORT, SUPPERJUMPIN_DATABASE_URL, SUPPERJUMPIN_DEV_AUTH_TOKEN, SUPPERJUMPIN_LOG_FORMAT, SUPPERJUMPIN_LOG_LEVEL |
| Add API endpoint | `internal/httpapi/server.go` | Closures over ServerConfig |
| Change DTO / JSON shape | `internal/httpapi/dto.go` | DTO structs with camelCase JSON tags |
| Postgres-backed tests | `npm run api:test` | Canonical test path against Postgres |
| Persistence adapter | `internal/httpapi/postgres_store.go` | sqlc-generated queries via `db.Queries` |
| Game rules / domain logic | `internal/game/*.go` | Pure functions, repository interfaces, no HTTP/DB imports |
| DB schema | `db/migrations/*.sql` | Pre-stable: fold schema changes into existing migration files |
| API contract | `openapi.yaml` | Source of truth for generated TypeScript client |
| sqlc config & generation | `db/queries/*.sql` → `sqlc.yaml` → `internal/db/` | Source `.sql` files; generated Go in `internal/db/` |

## CONVENTIONS

- **Standard library HTTP only**: `net/http` + `http.NewServeMux()` with Go 1.22 path patterns. No Gin, Echo, or Fiber.
- **Auth middleware pattern**: Every protected route calls `signedInProfile(w, r, config)` first. Bearer token from `Authorization` header. MVP development uses `SUPPERJUMPIN_DEV_AUTH_TOKEN` for local-first auth.
- **Error mapping**: Domain errors are mapped to HTTP status codes in transport helpers.
- **sqlc for queries**: All repository interface methods delegate to `s.queries.*` (generated `*db.Queries`). Add/modify a query → edit its `.sql` file in `db/queries/`, run `npm run generate:sqlc`.
- **Transactions**: Multi-step DB operations use `BeginTx` + `defer tx.Rollback()` + `tx.Commit()` with `qtx := s.queries.WithTx(tx)`.

## LOGGING CONVENTIONS

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

## ANTI-PATTERNS

- Adding business logic to `server.go` handlers. Keep handlers thin; delegate to transport helpers, which delegate to `internal/game/`.
- Using an ORM or query builder. sqlc-generated queries is the current pattern.
- Modifying `openapi.yaml` without regenerating the TypeScript client. CI will fail.
- Creating new numbered migration files before DB stability. Fold into existing table-creation migrations.

## NOTES

- `PostgresStore` uses `database/sql` with `pgx` driver, not `pgxpool` directly.
- `stableID(kind, value)` generates deterministic IDs. Never use UUIDs or auto-increment for domain entities.
- `internal/db/` is the generated sqlc package. It must never import `internal/httpapi` or `internal/game`.
