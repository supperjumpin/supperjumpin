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
| Prompt catalog HTTP surface | `server.go` (`GET /v1/prompt-catalog`) + `dto.go` (`PromptCatalogResponse`, `PromptPackDTO`, `PromptDTO`) | Public, unauthenticated. Handler must set `actor_type=public` per root AGENTS logging conventions. |
| Round start HTTP surface | `server.go` (`POST /v1/rounds`, `GET /v1/reveal-timeframes`) + `dto.go` (`StartRoundRequest`, `StartRoundResponse`, `RoundDTO`, `RevealTimeframeDTO`, `RevealTimeframesResponse`) | `POST /v1/rounds` is authenticated (bearerAuth) and maps domain errors to 403; `GET /v1/reveal-timeframes` is public. |
| Reveal HTTP surface | `server.go` (`POST /v1/rounds/{roundId}/reveal`) + `dto.go` (`RevealRoundResponse`) | Authenticated (bearerAuth). Calls `game.EvaluateReveal` with injected clock for condition evaluation. Returns round state and `revealed` flag. |
| Stamp catalog HTTP surface | `server.go` (`GET /v1/stamp-catalog`) + `dto.go` (`StampCatalogResponse`, `StampDTO`) | Public, unauthenticated. Returns data-driven stamp catalog (stance = stable identity; label/glyph/copy = tunable data). |
| Reaction HTTP surface | `server.go` (`POST /v1/rounds/{roundId}/jumps/{jumpId}/reactions`) + `dto.go` (`ApplyReactionRequest`, `ApplyReactionResponse`, `ReactionDTO`) | Authenticated (bearerAuth). Calls `game.ApplyReaction` — applies a Stamp to a revealed Jump, one-tap, repeatable. |
| Comment HTTP surface | `server.go` (`POST/GET /v1/rounds/{roundId}/comments`, `POST/GET /v1/rounds/{roundId}/jumps/{jumpId}/comments`) + `dto.go` (`PostCommentRequest`, `PostCommentResponse`, `CommentDTO`, `ListCommentsResponse`) | Authenticated (bearerAuth). Calls `game.PostComment` (round-level or jump-level) and `game.ListComments`. Comments are free-form, post-reveal only, distinct from Stamps. |
| Lore HTTP surface | `server.go` (`GET /v1/communities/{communityId}/lore`) + `dto.go` (`LoreEntryDTO`, `CommunityLoreResponse`) | Authenticated (bearerAuth). Calls `game.DeriveCommunityLore` — pure derivation from Stamp density on revealed Jumps, keyed to moments, never per-Player. |
| Integration tests | `*_test.go` | `httptest` + Postgres-backed fixtures |

## CONVENTIONS

- **Handler closures** over `ServerConfig` — no global state. Each route captures `config`.
- **DTO structs** use camelCase JSON tags. Domain snapshots use PascalCase Go fields.
- **PostgresStore** is the canonical persistence path and supports clock injection in tests (`SetClock`).
- Unit tests should use narrow per-test fakes or mocks when they do not need durable Postgres behavior.

## ANTI-PATTERNS

- Adding game rules to transport helpers. Transport helpers should only marshal/unmarshal and call `game.*` functions.
- Adding HTTP-specific logic (status codes, headers) to `PostgresStore`.
