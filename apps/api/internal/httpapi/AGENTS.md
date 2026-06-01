# apps/api/internal/httpapi KNOWLEDGE BASE

## OVERVIEW

HTTP transport layer for the Go API. Handles routing, auth, JSON DTO conversion, and both persistence implementations (`MemoryStore` + `PostgresStore`).

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add/modify route | `server.go` | `mux.HandleFunc` closures; ~18 routes in one file |
| Change JSON request/response shape | `store.go` (DTO structs) | camelCase `json:"groupId"` tags |
| Fix game error → HTTP status mapping | `store.go` (mapGameErr) | e.g., `ErrInvalidJudgmentScore` → 400 |
| Add transport helper | `store.go` | Wraps game function + assembles DTO response |
| In-memory test double | `store.go` (MemoryStore) | Full `Persistence` implementation using maps |
| Production persistence | `postgres_store.go` (PostgresStore) | Raw SQL implementing same `Persistence` interface |
| Integration tests | `groups_test.go`, `me_test.go` | `httptest` + `MemoryStore`; comprehensive lifecycle tests |

## CONVENTIONS

- **Handler closures** over `ServerConfig` — no global state. Each route captures `config`.
- **DTO structs** use camelCase JSON tags. Domain snapshots use PascalCase Go fields.
- **MemoryStore** is map-backed with clock injection (`NewMemoryStoreWithClock`). Used for fast, isolated tests.
- **PostgresStore** mirrors MemoryStore's behavior exactly using raw SQL. Any logic discrepancy is a bug.

## ANTI-PATTERNS

- Adding game rules to transport helpers. Transport helpers should only marshal/unmarshal and call `game.*` functions.
- Adding HTTP-specific logic (status codes, headers) to `MemoryStore` or `PostgresStore`.
- Changing `MemoryStore` behavior without updating `PostgresStore` (or vice versa).

## NOTES

- `groups_test.go` is the largest file in the project (~2300 lines). It covers the full Group/Jump/Season/Judgment/Dispute lifecycle. Consider splitting into focused test files (e.g., `group_season_test.go`) if it grows further.
- `server.go` contains all routing in one function with ~20 inline closures. As the API grows, consider extracting route registration into a separate function or file.
- `guest_judgment_test.go` covers the Guest Judge session and unauthenticated Judgment flow.
