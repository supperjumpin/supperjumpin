# apps/api/internal/httpapi KNOWLEDGE BASE

## OVERVIEW

HTTP transport layer for the Go API. Handles routing, auth, JSON DTO conversion, and the Postgres-backed persistence implementation.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add/modify route | `server.go` | `mux.HandleFunc` closures; ~18 routes in one file |
| Change JSON request/response shape | `dto.go` | camelCase JSON tags |
| Fix game error → HTTP status mapping | `store.go` (mapGameErr) | e.g., `ErrInvalidJudgmentScore` → 400 |
| Add transport helper | `store.go` | Wraps game function + assembles DTO response |
| Shared transport helpers | `helpers.go` | ID generation, DTO/snapshot mapping, small package helpers |
| Production persistence | `postgres_store.go` (PostgresStore) | Raw SQL implementing same `Persistence` interface |
| Integration tests | `me_test.go`, `public_read_test.go`, `guest_judgment_test.go` | `httptest` + Postgres-backed fixtures |

## CONVENTIONS

- **Handler closures** over `ServerConfig` — no global state. Each route captures `config`.
- **DTO structs** use camelCase JSON tags. Domain snapshots use PascalCase Go fields.
- **PostgresStore** is the canonical persistence path and supports clock injection in tests (`SetClock`).
- Unit tests should use narrow per-test fakes or mocks when they do not need durable Postgres behavior.

## ANTI-PATTERNS

- Adding game rules to transport helpers. Transport helpers should only marshal/unmarshal and call `game.*` functions.
- Adding HTTP-specific logic (status codes, headers) to `PostgresStore`.

## NOTES

- `server.go` contains all routing in one function with inline closures. As the API grows, consider extracting route registration into a separate function or file.
- `guest_judgment_test.go` covers the Guest Judge session and unauthenticated Judgment flow.
- Groups, Seasons, Invites, and Disputes were removed per ADR-0019. `SeasonSnapshot` and `Season()` are retained only for judgment-window checks on season-linked jumps (legacy data).
