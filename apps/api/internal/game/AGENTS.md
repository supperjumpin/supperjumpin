# apps/api/internal/game KNOWLEDGE BASE

## OVERVIEW

Pure domain logic for Supperjumpin. No `net/http`, no `database/sql`, no JSON tags. Repository interfaces are injected; game rules are expressed as standalone functions.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Group/invite rules | `group.go` | `CreateGroup`, `CreateInvite`, `AcceptInvite`, `GroupHome`, `ListGroups` |
| Jump lifecycle | `jump_planning.go` | `CreateIdea` → `CreatePlannedJump` → `CreatePerformedJump` |
| Evidence rules | `evidence.go` | `AuthorizeEvidenceUpload`, `SubmitEvidence` |
| Judgment scoring | `judgment.go` | `SubmitJudgment` (upsert, self-judge guard, score validation) |
| Season lifecycle | `season.go` | `StartSeason`, `CloseSeasonSubmissions`, `FinalizeSeason`, `AutoFinalizeSeason`, `Standings` |
| Dispute handling | `dispute.go` | `CreateDispute`, `ResolveDispute` (override/appeal flow) |
| Service wrapper | `game.go` | Thin `Service` struct with `Now` clock injection |
| Domain unit tests | `*_test.go` | Hand-rolled `mock*Repo` structs, co-located |

## CONVENTIONS

- **Repository-per-flow**: Each file defines its own focused repository interface (e.g., `JudgmentRepository`). Interfaces are small and cohesive.
- **Input/Result structs**: Every operation has explicit `XxxInput` and `XxxResult{Allowed, Created, Err}` structs.
- **Allowed bool**: Authorization failures return `Allowed=false`. HTTP layer maps this to 403.
- **Snapshot pattern**: Read-only views use `XxxSnapshot` structs. Persistence layers assemble these from DB rows.
- **Clock injection**: Time-dependent logic accepts `func() time.Time` (e.g., `Service.Now`, `PostgresStore` clock).
- **Error naming**: Sentinel errors use `ErrXxx` (e.g., `ErrInvalidJudgmentScore`, `ErrJumpNotFound`).

## ANTI-PATTERNS

- Importing `net/http`, `database/sql`, or any transport/persistence package. `game/` must remain pure.
- Returning HTTP status codes or JSON shapes from domain functions.
- Using UUIDs or auto-incrementing IDs. `stableID(kind, value)` is the project's ID generation rule.

## NOTES

- **Jump status machine**: `Idea` → `Planned Jump` → `Performed Jump` → `Judged Jump` / `Unjudged Jump` / `Disqualified Jump`.
- **Season status machine**: `Active` → `Judging Grace Period` → `Finalized`. Auto-finalization is triggered on read operations (`GroupHomeForGroup`, `GroupHomeForSeason`).
- **Grace periods**: Performed jumps have a 10-minute `GracePeriodExpiresAt` window where the performer can edit/retract.
- **Dispute hierarchy**: Commissioner resolves open disputes on season-linked jumps; Group Admin can override any resolution; off-season disputes resolved by Group Admin directly.
