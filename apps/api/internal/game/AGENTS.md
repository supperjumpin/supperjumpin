# apps/api/internal/game KNOWLEDGE BASE

## OVERVIEW

Pure domain logic for Supperjumpin. No `net/http`, no `database/sql`, no JSON tags. Repository interfaces are injected; game rules are expressed as standalone functions.

## MODULES

| File | Responsibility |
|------|---------------|
| `identity.go` | Idempotent ensure-Player operation: given opaque (playerID, communityID), ensures both entities exist. Adapter resolves platform actors → internal IDs before calling. |
| `prompts.go` | Prompt/Pack catalog: first-class reusable Prompts grouped into Packs. `ListCatalog` assembles the full catalog (packs with their prompts); `SelectPrompt` looks up by id (returns `ErrPromptNotFound`); `SelectRandomPrompt` picks one from the catalog (returns `ErrNoPromptsAvailable` when empty). Random-pick randomness is injected via a `func(n int) int` picker for testability. Platform-authored global curation (ADR-0039). |
| `round.go` | Round lifecycle start: `ListRevealTimeframes` returns the data-driven reveal-timeframe menu (tunable data, not a hardcoded enum). `StartRound` creates a Round for a Community — validates player/community exist, enforces one-active-Round-per-Community invariant (`ErrRoundAlreadyActive`), resolves prompt (explicit id or random pick via injected `func(n int) int`), computes reveal_by from the timeframe duration, and creates the Round with a deterministic stable ID. The Round is the aggregate root for the Round-centric domain (ADR-0038, ADR-0040). |
| `reveal.go` | Reveal condition evaluation: `EvaluateReveal` checks whether a Round's reveal condition is met (v1 = scheduled time via injected `now`), transitions the Round to "revealed" status if so, and is idempotent. The condition-evaluation seam admits future variants (initiator-triggered, threshold-triggered) without reshaping the Round. |
| `reaction.go` | Reaction / Stamp application: `ListStampCatalog` returns the data-driven Stamp catalog (stance = stable identity, label/glyph/copy = tunable data, never an enum, never in `openapi.yaml` — ADR-0034). `ApplyReaction` applies a Stamp to a revealed Jump — one-tap, repeatable, no rubric. Validates jump exists, round is revealed, stamp exists, player exists. No head-to-head vote or winner is produced. |
| `comment.go` | Free-form Comments channel distinct from Stamps: `PostComment` creates a comment on a revealed Round or Jump (enforces round-revealed, player-exists, jump-exists-if-specified, body-not-empty). `ListComments` lists comments scoped to a round (optionally filtered to a jump). Non-Jumpers can comment; comments are always visible (no sealing). Consistently uses `domainStableID` with nanosecond timestamps for multiple-comment uniqueness. |

## CONVENTIONS

- **Repository-per-flow**: Each file defines its own focused repository interface. Interfaces are small and cohesive.
- **Input/Result structs**: Every operation has explicit `XxxInput` and `XxxResult{Allowed, Created, Err}` structs.
- **Allowed bool**: Authorization failures return `Allowed=false`. HTTP layer maps this to 403.
- **Snapshot pattern**: Read-only views use `XxxSnapshot` structs. Persistence layers assemble these from DB rows.
- **Clock injection**: Time-dependent logic accepts explicit `time.Time` values from callers; adapters own their clocks.
- **Error naming**: Sentinel errors use `ErrXxx`.

## ANTI-PATTERNS

- Importing `net/http`, `database/sql`, or any transport/persistence package. `game/` must remain pure.
- Returning HTTP status codes or JSON shapes from domain functions.
- Using UUIDs or auto-incrementing IDs. `stableID(kind, value)` is the project's ID generation rule.
