# apps/api/internal/httpapi KNOWLEDGE BASE

## OVERVIEW

HTTP transport layer for the Go API. Handles routing, auth, JSON DTO conversion, and the Postgres-backed persistence implementation.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add/modify route | `server.go` | `mux.HandleFunc` closures |
| Change JSON request/response shape | `dto.go` | camelCase JSON tags |
| Shared transport helpers | `helpers.go` | ID generation, DTO/snapshot mapping |
| Request logging | `logging.go` | Request ID context, response capture, panic recovery, completion logs |
| Production persistence | `postgres_store.go` (PostgresStore) | sqlc-generated queries implementing per-flow interfaces |
| External identity resolution | `external_identity.go` | Adapter-owned mapping from platform actors → (player_id, community_id) via game.EnsurePlayer |
| Integration tests | `*_test.go` | `httptest` + Postgres-backed fixtures |

## CONVENTIONS

- **Handler closures** over `ServerConfig` — no global state. Each route captures `config`.
- **DTO structs** use camelCase JSON tags. Domain snapshots use PascalCase Go fields.
- **PostgresStore** is the canonical persistence path and supports clock injection in tests (`SetClock`).
- Unit tests should use narrow per-test fakes or mocks when they do not need durable Postgres behavior.

## ANTI-PATTERNS

- Adding game rules to transport helpers. Transport helpers should only marshal/unmarshal and call `game.*` functions.
- Adding HTTP-specific logic (status codes, headers) to `PostgresStore`.
