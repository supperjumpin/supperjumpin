# apps/api/internal/game KNOWLEDGE BASE

## OVERVIEW

Pure domain logic for Supperjumpin. No `net/http`, no `database/sql`, no JSON tags. Repository interfaces are injected; game rules are expressed as standalone functions.

## MODULES

| File | Responsibility |
|------|---------------|
| `identity.go` | Idempotent ensure-Player operation: given opaque (playerID, communityID), ensures both entities exist. Adapter resolves platform actors → internal IDs before calling. |
| `prompts.go` | Prompt/Pack catalog: first-class reusable Prompts grouped into Packs. `ListCatalog` assembles the full catalog (packs with their prompts); `SelectPrompt` looks up by id (returns `ErrPromptNotFound`); `SelectRandomPrompt` picks one from the catalog (returns `ErrNoPromptsAvailable` when empty). Random-pick randomness is injected via a `func(n int) int` picker for testability. Platform-authored global curation (ADR-0039). |

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
