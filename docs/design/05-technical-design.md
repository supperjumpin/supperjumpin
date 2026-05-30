# Technical Design

_Part of the [Supperjumpin Design Package](./README.md). Depends on: [Product/UX Design](./02-product-ux-design.md) (#106), [Backend/Data Architecture](./03-backend-data-architecture.md) (#107). Parent tracker: #66._

## 1. Introduction and Scope

This document translates product decisions from the [Product/UX Design](./02-product-ux-design.md) (#106) and [Backend/Data Architecture](./03-backend-data-architecture.md) (#107) into implementation-ready specifications. It sits between those design documents and the [Implementation Backlog](./06-implementation-backlog.md) — an engineer should be able to read this document and know exactly what to build, what tests to write, and in what order.

### Cross-reference strategy

- **Summarize and reference** where #106 or #107 already contain deep detail (data schemas, wireframes, endpoint request/response shapes). This document provides enough context to understand the specification without reading the source, but the source document is authoritative.
- **Specify fully** where this document introduces new work: the state machine, eligibility decision table, scoring formulas, test specifications, and implementation tracer bullets.

### v1 scope boundary

v1 delivers: public feed, Jump lifecycle (Draft → Performed Jump → Judged Jump), tap-to-select Judging with four factors, Guest Judges, The Open (monthly competition), Evidence upload, Report flow, and Removed Jump.

Deferred to v2: Groups, Seasons, Season Commissioner, Disqualified Jump, formal Dispute tooling, auto-hide on reports, and image scanning. These concepts are reserved in the data model but receive no code paths in v1.

---

## 2. Jump Lifecycle State Machine

### 2.1 State diagram

```
                      ┌─────────────────┐
                      │ Draft (client)  │
                      │ Not persisted   │
                      └────────┬────────┘
                               │ submit Evidence
                               ▼
                      ┌─────────────────────────┐
                 ┌──► │ Performed Jump           │
                 │    │ (Grace Period active)    │
                 │    └────────┬────────────────┘
                 │             │ grace_period_expires_at < NOW()
                 │             ▼
                 │    ┌─────────────────────────┐
                 │    │ Performed Jump           │
       edit ─────┤    │ (Judging Window open)    │
       Caption   │    └────┬───────────┬────────┘
       (grace    │         │           │
       only)     │  first  │           │ Report →
                 │  Judgment│          │ team removal
                 │  received│          │
                 │         ▼           ▼
                 │  ┌────────────┐  ┌─────────────────┐
                 └─ │ Judged Jump│  │ Removed Jump    │
                    │ (Window    │  │ (tombstoned)    │
                    │  remains   │  └─────────────────┘
                    │  open)     │
                    └─────┬──────┘
                          │
                   v2 only│ Season Judging Grace Period closes
                   (zero  │ with zero Season Judgments
                   Season │
                   Jmts)  ▼
                     ┌──────────────────┐
                     │ Unwitnessed Jump │
                     │ (v2 reservation) │
                     └──────────────────┘

           v2 only: Judged Jump ──Dispute resolution──► Disqualified Jump
```

### 2.2 State transition table

| # | Current State | Trigger | Guard | New State | Side Effect |
|---|--------------|---------|-------|-----------|-------------|
| T1 | Draft (client) | Player submits Evidence | Authenticated; Evidence photo + Caption present | Performed Jump (Grace Period) | Insert `jumps` row; set `grace_period_expires_at = NOW() + 10min`; authorize Evidence upload |
| T2 | Performed Jump (Grace Period) | Grace period expires | `grace_period_expires_at < NOW()` | Performed Jump (Judging Window open) | No row update — expiry is computed, not a status change; enable "Judge" button for other Players |
| T3 | Performed Jump (Grace Period) | Performer edits Caption | `player_id == authenticated_player_id`; `grace_period_expires_at > NOW()` | Performed Jump (Grace Period) | Update `jumps.caption`; no status change |
| T4 | Performed Jump (Grace Period) | Performer retracts | `player_id == authenticated_player_id`; `grace_period_expires_at > NOW()` | Removed Jump | Set `jumps.status = 'Removed Jump'`; tombstone Evidence; suppress from feed |
| T5 | Performed Jump (Judging Window open) | First Judgment received | Eligibility rules pass (see §7) | Judged Jump | Set `jumps.status = 'Judged Jump'`; compute running average |
| T6 | Judged Jump | Subsequent Judgment received | Eligibility rules pass (see §7) | Judged Jump | Recompute running average; no status change |
| T7 | Any visible state | Team removes Jump | Admin authority | Removed Jump | Set `jumps.status = 'Removed Jump'`; tombstone Evidence; suppress from all queries; serve tombstone for deep links |
| T8 | Judged Jump | Season closes with zero Season Judgments | v2 only; `season_id` set; no Season-provenance Judgments | Unwitnessed Jump | Set `jumps.status = 'Unwitnessed Jump'`; exclude from Season Standings; public feed unaffected |
| T9 | Judged Jump | Dispute resolution upholds violation | v2 only; formal Dispute process | Disqualified Jump | Set `jumps.status = 'Disqualified Jump'`; exclude from Standings; Jump remains visible |

### 2.3 Persistence rules per state

| State | `jumps.status` | `grace_period_expires_at` | Feed visible | Can receive Judgments | Notes |
|-------|---------------|--------------------------|-------------|----------------------|-------|
| Draft | N/A (not stored) | N/A | No | No | Client-only; never sent to server |
| Performed Jump (Grace) | `'Performed Jump'` | `> NOW()` | Yes (with badge) | No | Other Players see countdown; performer sees Edit |
| Performed Jump (Judging) | `'Performed Jump'` | `≤ NOW()` | Yes | Yes | Computed state — no row update needed |
| Judged Jump | `'Judged Jump'` | `≤ NOW()` | Yes | Yes | Judging Window remains open on public feed |
| Unwitnessed Jump | `'Unwitnessed Jump'` | N/A | Yes | No (Season) / Yes (public) | v2 only; public feed Judging Window unaffected |
| Removed Jump | `'Removed Jump'` | N/A | No (tombstoned) | No | Evidence suppressed; deep links return tombstone |
| Disqualified Jump | `'Disqualified Jump'` | N/A | Yes | No | v2 only; visible but excluded from Standings |

### 2.4 Domain command signatures

```go
// JumpService orchestrates Jump lifecycle transitions.
type JumpService struct {
    JumpWrite    JumpWriteRepo
    JumpRead     JumpReadRepo
    EvidenceRepo EvidenceRepository
    Now          func() time.Time
}

func (s *JumpService) CreatePerformedJump(ctx context.Context, input CreateJumpInput) (JumpSnapshot, error)
func (s *JumpService) EditJumpCaption(ctx context.Context, jumpID, playerID, caption string) error
func (s *JumpService) RetractJump(ctx context.Context, jumpID, playerID string) error
func (s *JumpService) RemoveJump(ctx context.Context, jumpID string) error  // admin only
```

---

## 3. API Contract Specification

This section summarizes the API endpoints. Full request/response schemas are in #107 §4.

### 3.1 Endpoint summary

#### Feed and Discovery

| Method | Path | Purpose | Auth | Domain Command |
|--------|------|---------|------|----------------|
| `GET` | `/v1/feed` | Public chronological feed (paginated, 20 per page) | No | `JumpReadRepo.Feed()` |

#### Jump Lifecycle

| Method | Path | Purpose | Auth | Domain Command |
|--------|------|---------|------|----------------|
| `POST` | `/v1/jumps` | Create a Performed Jump | Yes | `JumpService.CreatePerformedJump()` |
| `GET` | `/v1/jumps/{jumpID}` | Jump detail with scores | No | `JumpReadRepo.Jump()` |
| `PATCH` | `/v1/jumps/{jumpID}` | Edit Caption during Grace Period | Yes (performer) | `JumpService.EditJumpCaption()` |
| `POST` | `/v1/jumps/{jumpID}/retract` | Retract during Grace Period | Yes (performer) | `JumpService.RetractJump()` |

#### Judging

| Method | Path | Purpose | Auth | Domain Command |
|--------|------|---------|------|----------------|
| `POST` | `/v1/jumps/{jumpID}/judgments` | Submit a Judgment (4 factors, 1–4) | No (Guest allowed) | `JudgmentService.SubmitJudgment()` |

#### Evidence

| Method | Path | Purpose | Auth | Domain Command |
|--------|------|---------|------|----------------|
| `POST` | `/v1/jumps/{jumpID}/evidence-authorization` | Get signed upload URL | Yes (performer) | See §6 |
| `POST` | `/v1/jumps/{jumpID}/evidence` | Confirm Evidence uploaded | Yes (performer) | See §6 |

#### The Open

| Method | Path | Purpose | Auth | Domain Command |
|--------|------|---------|------|----------------|
| `GET` | `/v1/opens/{yearMonth}/standings` | Open Standings for a month | No | `OpenReadRepo.OpenStandings()` |

#### Guest Sessions

| Method | Path | Purpose | Auth | Domain Command |
|--------|------|---------|------|----------------|
| `POST` | `/v1/guest-sessions` | Initialize a Guest Judge session | No | `GuestSessionRepo.CreateGuestSession()` |

#### Player Profile

| Method | Path | Purpose | Auth | Domain Command |
|--------|------|---------|------|----------------|
| `GET` | `/v1/players/{playerID}` | Player profile with Jump history | No | `JumpReadRepo.JumpsByPlayer()` |

#### Safety

| Method | Path | Purpose | Auth | Domain Command |
|--------|------|---------|------|----------------|
| `POST` | `/v1/jumps/{jumpID}/reports` | Report a Jump (4 categories + Other) | Yes | `DisputeRepository.InsertReport()` |

### 3.2 Deprecated endpoints (frozen in v1)

All Group, Season, and Invite endpoints are frozen — they survive in the router but receive no new features. See #107 §4.3 for the full list.

### 3.3 Error mapping

| Domain Error | HTTP Status | Client Message |
|-------------|-------------|----------------|
| `ErrJumpNotFound` | 404 | "Jump not found" |
| `ErrJudgingWindowClosed` | 403 | "Judging Window closed" |
| `ErrGracePeriodActive` | 403 | "Judging Window opens in [MM:SS]" |
| `ErrInvalidJudgmentScore` | 400 | "Judgment scores must be between 1 and 4" |
| `ErrAlreadyJudged` | 409 | "You have already entered your Judgment" |
| `ErrSelfJudging` | 403 | "You cannot Judge your own Jump" |
| `ErrGuestCapReached` | 403 | "Guest Judgment cap reached" |
| `ErrJumpRemoved` | 410 | "This Jump has been removed" |

---

## 4. Mobile Flow Specification

This section maps screens to API endpoints and Jump lifecycle states. Full screen inventory, onboarding paths, and visual design are in #106 §1–5.

### 4.1 Navigation structure

```
Feed (root, default screen)
├── Jump Detail (push)
│   ├── Judging Screen (modal)
│   │   └── Judgment Receipt (modal)
│   ├── Report Screen (modal)
│   └── Player Profile (push)
├── Create Jump (modal, via FAB)
│   └── Auth Gate (modal, if unauth)
├── Open Standings (push from header icon)
│   └── Player Profile (push)
└── Player Profile (push from Feed)
```

### 4.2 Screen-to-endpoint mapping

| Screen | Primary API Call | Jump States Visible | Auth Required |
|--------|-----------------|---------------------|---------------|
| Feed | `GET /v1/feed` | Performed Jump, Judged Jump | No |
| Jump Detail | `GET /v1/jumps/{id}` | All except Draft and Removed | No |
| Judging | `POST /v1/jumps/{id}/judgments` | Performed Jump (post-grace), Judged Jump | No (Guest allowed) |
| Create Jump | `POST /v1/jumps` + Evidence flow | N/A (creates new Jump) | Yes |
| Auth Gate | N/A (auth provider SDK) | N/A | N/A (conversion screen) |
| Open Standings | `GET /v1/opens/{month}/standings` | N/A | No |
| Player Profile | `GET /v1/players/{id}` | Performed Jump, Judged Jump | No |
| Report | `POST /v1/jumps/{id}/reports` | Any visible Jump | Yes |

### 4.3 Jump state → UI treatment

| Jump State | Feed Card Treatment | Detail Screen Treatment | Actions Available |
|-----------|-------------------|----------------------|-----------------|
| Performed Jump (Grace) | "Editing" badge for performer; countdown for others | Grace Period banner with countdown; Edit/Retract for performer | Edit Caption, Retract (performer only) |
| Performed Jump (Judging) | Standard card; "Judge →" CTA | "Judge" button active | Judge, Share, Report |
| Judged Jump | Score + Judgment count shown | Score breakdown; "You have entered your Judgment" if judged | Judge (if not yet judged), Share, Report |
| Removed Jump | Not shown | Tombstone page | "Browse Feed" CTA |

### 4.4 Auth gate locations

Auth is required at: creating a Jump, retracting a Jump, reporting a Jump, and viewing own profile. Auth is NOT required for: viewing the feed, viewing Jump details, and submitting Judgments (Guest allowed). Guest Judges hit a soft cap after 5 Judgments (#106 §1.2).

---

## 5. Persistence and Migration Plan

This section summarizes schema changes and migration sequencing. Full SQL DDL is in #107 §3.

### 5.1 Migration sequence

Migrations must run in this order due to foreign key and column dependencies:

| Order | Migration | Purpose | ADR |
|-------|-----------|---------|-----|
| 1 | Rename `stunts` → `jumps` | Table and FK column rename | ADR-0020 |
| 2 | Update `jumps.status` CHECK | Remove `Idea`, `Planned Stunt`; add v1 states | #107 §2.3 |
| 3 | `jumps.group_id` nullable | Jumps are public by default | ADR-0019 |
| 4 | Add `jumps.grace_period_expires_at` | Author Grace Period tracking | #107 §2.3 |
| 5 | `jumps.final_score` → `open_final_score` + `season_final_score` | Multi-score model | ADR-0023 |
| 6 | Rename judgment columns | `difficulty` → `commitment`, `documentation` → `presentation` | ADR-0020 |
| 7 | Update judgment CHECK constraints | 0–10 → 1–4 scale | ADR-0022 |
| 8 | Add judgment columns | `guest_session_id`, `provenance`, `open_month` | ADR-0021, ADR-0023 |
| 9 | Update judgment UNIQUE constraints | Split into partial indexes for Player vs. Guest | #107 §5.4 |
| 10 | Create `guest_sessions` table | Guest Judge session tracking | #107 §5.4 |
| 11 | Create `opens` table | Monthly competition tracking | ADR-0023 |
| 12 | Create `open_standings` table | Player rankings per Open month | ADR-0023 |
| 13 | Create feed index | `jumps_feed_idx` on `created_at DESC` WHERE status in feed-visible states | #107 §4.4 |

### 5.2 Key schema changes summary

| Table | Change | Reference |
|-------|--------|-----------|
| `jumps` | Rename from `stunts`; nullable `group_id`; add `grace_period_expires_at`; split `final_score` | #107 §3.1–3.2 |
| `judgments` | Rename factors; 1–4 scale; add `guest_session_id`, `provenance`, `open_month`; partial unique indexes | #107 §3.2, §5.2–5.4 |
| `guest_sessions` | New table | #107 §3.3 |
| `opens` | New table | #107 §3.3 |
| `open_standings` | New table | #107 §3.3 |
| `groups`, `seasons`, `invites` | Frozen — no v1 changes | #107 §3.4 |

### 5.3 Repository port interfaces

```go
type JumpWriteRepo interface {
    InsertJump(ctx context.Context, jump JumpSnapshot) error
    UpdateJumpStatus(ctx context.Context, jumpID, status string) error
    SetJumpOpenFinalScore(ctx context.Context, jumpID string, score float64) error
}

type JumpReadRepo interface {
    Jump(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
    Feed(ctx context.Context, cursor time.Time, limit int) ([]JumpSnapshot, error)
    JumpsByPlayer(ctx context.Context, playerID string, cursor time.Time, limit int) ([]JumpSnapshot, error)
}

type JudgmentWriteRepo interface {
    InsertJudgment(ctx context.Context, judgment JudgmentSnapshot) error
    JudgmentCountByGuestSession(ctx context.Context, guestSessionID string) (int, error)
}

type JudgmentReadRepo interface {
    JudgmentsForJump(ctx context.Context, jumpID string) ([]JudgmentSnapshot, error)
    RunningAverage(ctx context.Context, jumpID string) (float64, error)
}

type OpenWriteRepo interface {
    UpsertOpen(ctx context.Context, yearMonth string) error
    SetOpenFinalScores(ctx context.Context, yearMonth string) error
}

type OpenReadRepo interface {
    OpenStandings(ctx context.Context, yearMonth string) ([]StandingEntry, error)
    CurrentOpen(ctx context.Context) (OpenSnapshot, error)
}

type GuestSessionRepo interface {
    CreateGuestSession(ctx context.Context, id string) error
    GuestSession(ctx context.Context, id string) (GuestSessionSnapshot, bool, error)
    IncrementJudgmentCount(ctx context.Context, id string) error
    ClaimGuestSession(ctx context.Context, id, playerID string) error
}
```

---

## 6. Evidence Upload Flow

### 6.1 Sequence

```
Player                     Server                      Object Store
  │                          │                             │
  │  1. Select photo         │                             │
  │  2. POST /v1/jumps       │                             │
  │  ──────────────────────► │                             │
  │                          │ 3. Create Jump row          │
  │                          │    (status=Performed Jump)   │
  │  ◄────────────────────── │                             │
  │  { jump_id }             │                             │
  │                          │                             │
  │  4. POST /jumps/{id}/    │                             │
  │     evidence-authorization│                             │
  │  ──────────────────────► │                             │
  │                          │ 5. Validate: owns Jump,     │
  │                          │    Grace Period active       │
  │                          │ 6. Generate signed URL      │
  │  ◄────────────────────── │                             │
  │  { uploadUrl, key }      │                             │
  │                          │                             │
  │  7. PUT photo to         │                             │
  │     signed URL           │                             │
  │  ─────────────────────────────────────────────────────►│
  │                          │                             │
  │  8. POST /jumps/{id}/    │                             │
  │     evidence             │                             │
  │  ──────────────────────► │                             │
  │                          │ 9. Validate: auth exists,   │
  │                          │    not expired, obj exists   │
  │                          │ 10. Create evidences row    │
  │  ◄────────────────────── │                             │
  │  { Jump + Evidence }     │                             │
```

### 6.2 Validation rules

| Rule | Enforcement |
|------|-------------|
| One Evidence per Jump | `evidences.jump_id` is `UNIQUE` |
| Upload authorization required | `evidences.upload_authorization_id` is `NOT NULL` and `UNIQUE` |
| Authorization expiry | `evidence_upload_authorizations.expires_at > NOW()` (15-minute expiry) |
| Player owns the Jump | Domain command validates `jumps.player_id == authenticated_player_id` |
| Content type | Server validates `Content-Type` is `image/*` before issuing authorization |
| Size limit | 10 MB per upload (server-configurable) |

### 6.3 Tombstoning

When a Jump transitions to Removed Jump, the `evidences` row and object storage object are preserved (for potential appeal review) but excluded from all read queries. Deep links return a tombstone page with no Evidence, no performer info, and a "Browse Feed" CTA.

---

## 7. Judging Eligibility Rules

### 7.1 Eligibility decision table

Guards are evaluated in this order. The first failing guard short-circuits the rest.

| Priority | Rule | Condition | Domain Check | Error | HTTP Status |
|----------|------|-----------|-------------|-------|-------------|
| 1 | Jump exists | `jumpID` resolves to a Jump | `JumpReadRepo.Jump()` returns `(snapshot, true, nil)` | `ErrJumpNotFound` | 404 |
| 2 | Jump not Removed | `jumps.status != 'Removed Jump'` | `snapshot.Status != "Removed Jump"` | `ErrJumpRemoved` | 410 |
| 3 | Judging Window open | `jumps.grace_period_expires_at < NOW()` | `snapshot.GracePeriodExpiresAt.Before(s.Now())` | `ErrGracePeriodActive` | 403 |
| 4 | Not own Jump | `jumps.player_id != judge_player_id` | `snapshot.PlayerID != input.PlayerID` | `ErrSelfJudging` | 403 |
| 5 | Not already judged | No existing Judgment for same `(jump_id, player_id)` or `(jump_id, guest_session_id)` | Partial unique index prevents insert; domain pre-check queries | `ErrAlreadyJudged` | 409 |
| 6 | Guest soft cap | `guest_sessions.judgment_count < server_cap` (default 5) | `GuestSessionRepo.GuestSession()` → check `JudgmentCount < cap` | `ErrGuestCapReached` | 403 |
| 7 | Valid scores | Each factor in range 1–4 | `1 ≤ commitment ≤ 4` ∧ same for other 3 factors | `ErrInvalidJudgmentScore` | 400 |

### 7.2 Eligibility check pseudocode

```go
func (s *JudgmentService) checkEligibility(ctx context.Context, input SubmitJudgmentInput) error {
    // 1. Jump exists
    jump, found, err := s.JumpRead.Jump(ctx, input.JumpID)
    if err != nil {
        return err
    }
    if !found {
        return ErrJumpNotFound
    }

    // 2. Jump not Removed
    if jump.Status == "Removed Jump" {
        return ErrJumpRemoved
    }

    // 3. Judging Window open
    if jump.GracePeriodExpiresAt.After(s.Now()) {
        return ErrGracePeriodActive
    }

    // 4. Not own Jump
    if input.PlayerID != "" && jump.PlayerID == input.PlayerID {
        return ErrSelfJudging
    }

    // 5. Not already judged (domain pre-check; UNIQUE constraint is final guard)
    // Handled by INSERT with partial unique index

    // 6. Guest soft cap
    if input.GuestSessionID != "" {
        session, found, err := s.GuestSession.GuestSession(ctx, input.GuestSessionID)
        if err != nil {
            return err
        }
        if found && session.JudgmentCount >= s.GuestCap {
            return ErrGuestCapReached
        }
    }

    // 7. Valid scores
    for _, score := range []int{input.Commitment, input.Transgression, input.Creativity, input.Presentation} {
        if score < 1 || score > 4 {
            return ErrInvalidJudgmentScore
        }
    }

    return nil
}
```

### 7.3 Judgment submission flow

```go
type JudgmentService struct {
    JudgmentWrite JudgmentWriteRepo
    JudgmentRead  JudgmentReadRepo
    JumpRead      JumpReadRepo
    GuestSession  GuestSessionRepo
    Now           func() time.Time
    GuestCap      int
}

func (s *JudgmentService) SubmitJudgment(ctx context.Context, input SubmitJudgmentInput) (JudgmentResult, error)
```

After eligibility passes, `SubmitJudgment`:
1. Determines provenance: if `input.CreatedAt` falls within an active Open month → `provenance = "open"`, else `"public"`.
2. Sets `open_month` to the current `YYYY-MM` if provenance is `"open"`.
3. Inserts the Judgment row.
4. If Guest Judge, increments `guest_sessions.judgment_count`.
5. If first Judgment for this Jump, updates `jumps.status` to `"Judged Jump"`.
6. Returns `JudgmentResult` with the confirmed verdict for the receipt screen.

---

## 8. Scoring Mechanics

### 8.1 Composite score per Judgment

Each Judgment produces a composite score — the mean of its four factors:

```
composite = (commitment + transgression + creativity + presentation) / 4.0
```

Range: 1.0–4.0. This is the atomic unit of scoring; all aggregates are computed from composites.

### 8.2 Running average (live, public feed)

The running average is displayed on Jump cards and Jump detail for any Jump with at least one Judgment:

```
running_average = AVG(composite) across ALL Judgments for the Jump
```

All Judgments contribute regardless of provenance (public, open, season). The running average updates in real time with each new Judgment. Displayed to one decimal place (e.g., "3.2").

Implementation: `JudgmentReadRepo.RunningAverage(ctx, jumpID)` queries the aggregate.

### 8.3 Open Final Score (monthly, at soft-close)

When an Open month soft-closes, the Open Final Score for each Jump is computed from Judgments received before the soft-close timestamp:

```
open_final_score = AVG(composite)
    FROM judgments
    WHERE jump_id = :jump_id
      AND provenance IN ('open', 'public')
      AND created_at < :soft_closed_at
```

Key rules:
- Both `open`-provenance and `public`-provenance Judgments contribute.
- `season`-provenance Judgments are **excluded** (ADR-0021).
- Only Judgments received before the soft-close timestamp count.
- A Jump with zero qualifying Judgments does not receive an Open Final Score.

### 8.4 Season Final Score (v2 reservation)

```
season_final_score = AVG(composite)
    FROM judgments
    WHERE jump_id = :jump_id
      AND provenance = 'season'
      AND season_id = :season_id
```

Only Season-linked Judgments contribute, per ADR-0021. This score is null in v1 (no Seasons).

### 8.5 Provenance assignment

| Context | `provenance` | `open_month` | `season_id` |
|---------|-------------|-------------|-------------|
| Public feed, no active Open | `"public"` | null | null |
| Public feed, active Open month | `"open"` | `"YYYY-MM"` | null |
| Season-linked Jump (v2) | `"season"` | null or `"YYYY-MM"` | season UUID |

The `provenance` is set at Judgment insertion time based on the Jump's context at that moment. It does not change retroactively.

### 8.6 Score display rules

| Score | When Displayed | Where |
|-------|---------------|-------|
| Running average | Immediately after first Judgment | Feed card, Jump detail, share card |
| Open Final Score | After Open month soft-closes | Jump detail, Open Standings |
| Season Final Score | After Season finalizes (v2) | Jump detail, Season Standings (v2) |
| No score | Zero Judgments | Feed card shows "No Judgments yet" |

### 8.7 Score calculator interface

```go
type ScoreCalculator interface {
    RunningAverage(ctx context.Context, jumpID string) (float64, error)
    OpenFinalScore(ctx context.Context, jumpID string, yearMonth string) (float64, error)
}
```

---

## 9. Unwitnessed Jump Behavior

### 9.1 Public feed

On the public feed, the Judging Window is open-ended — it never closes. A Jump with zero Judgments remains a `Performed Jump` indefinitely. It displays "No Judgments yet" with no running average. The `Unwitnessed Jump` status does not apply on the public feed.

### 9.2 Season context (v2)

When a Season's Judging Grace Period closes, any Jump with zero Season-provenance Judgments transitions to `Unwitnessed Jump`. This affects Season Standings but not public feed visibility. The Jump remains visible and Judgable on the public feed with its public running average unaffected.

### 9.3 Unwitnessed Performance Award (v2)

An end-of-Season Award given to a Player whose Season-linked Jump closed as an Unwitnessed Jump. Recognizes commitment without an audience. Does not affect Season Score or Standings.

---

## 10. Group, Visibility, Season, and Standings

### 10.1 v1 vs. v2 boundary

| Concept | v1 Status | v2 Addition |
|---------|-----------|-------------|
| Public feed | Implemented | No changes |
| Jump visibility | Public by default; no Group wall | No changes |
| Guest Judge visibility | Full access to all non-Removed Jumps | No changes |
| The Open | Monthly soft-close, Standings | Awards, weekly checkpoints (experiments) |
| `jumps.group_id` | Always NULL | Set when Jump submitted to Group Season |
| Groups | Frozen schema, no UI | CRUD, membership, invites |
| Seasons | Frozen schema, no UI | Lifecycle (start/close/finalize), Commissioner |
| Season Standings | Not implemented | Computed from `provenance = 'season'` Judgments |
| Group Admin | Not implemented | Override authority over active Season |
| Disqualified Jump | Reserved in CHECK constraint | Formal status via Dispute resolution |
| Dispute tooling | Report button only (4 categories + Other) | Formal Dispute lifecycle with resolution |

### 10.2 Jump-Group association in v1

In v1, `jumps.group_id` is `NULL` for all Jumps. The column is nullable but not removed because v2 will re-introduce Group association. No code path in v1 sets or reads `group_id`.

### 10.3 Visibility rules

- All Performed Jumps and Judged Jumps are visible to everyone (authenticated and unauthenticated).
- Removed Jumps are visible only as tombstones (no content, no performer info).
- Draft Jumps are client-only and never leave the device.
- There is no "private Jump," "friends-only Jump," or "Group-only Jump" in v1.

### 10.4 The Open as v1 competitive context

The Open replaces Seasons as the competitive context in v1. See ADR-0023 and #107 §3.3 for the data model. Key rules:
- The Open is always active. Any Player with at least one Performed Jump and one Judgment in the calendar month competes automatically.
- Soft-close at month-end: Open Final Scores are computed from existing Judgments.
- Open Standings are separate from any future Season Standings.
- The Open is treated as a signal surface, not a retention engine (see #106 §7.6).

---

## 11. Test Specifications

### 11.1 Jump lifecycle state transitions

| ID | Given | When | Then | Test File |
|----|-------|------|------|-----------|
| LT-1 | A Player is authenticated and has selected a photo + Caption | Player calls `CreatePerformedJump` | Jump is inserted with status `"Performed Jump"`; `grace_period_expires_at` is set to `NOW() + 10min` | `internal/game/jump_lifecycle_test.go` |
| LT-2 | A Jump is in Grace Period (`grace_period_expires_at > NOW()`) | Another Player calls `SubmitJudgment` | Returns `ErrGracePeriodActive` | `internal/game/jump_lifecycle_test.go` |
| LT-3 | A Jump's Grace Period has expired (`grace_period_expires_at < NOW()`) | A Player calls `SubmitJudgment` | Judgment is inserted; Jump status transitions to `"Judged Jump"` | `internal/game/jump_lifecycle_test.go` |
| LT-4 | A Jump is in Grace Period and the performer edits the Caption | Performer calls `EditJumpCaption` | Caption is updated; `grace_period_expires_at` unchanged | `internal/game/jump_lifecycle_test.go` |
| LT-5 | A Jump is in Grace Period and the performer retracts | Performer calls `RetractJump` | Jump status becomes `"Removed Jump"`; Evidence tombstoned | `internal/game/jump_lifecycle_test.go` |
| LT-6 | Grace Period has expired and the performer attempts to edit | Performer calls `EditJumpCaption` | Returns `ErrGracePeriodExpired` | `internal/game/jump_lifecycle_test.go` |

### 11.2 Judging eligibility

| ID | Given | When | Then | Test File |
|----|-------|------|------|-----------|
| JE-1 | A Jump exists and the Judge is the performer | Judge calls `SubmitJudgment` | Returns `ErrSelfJudging` | `internal/game/judgment_eligibility_test.go` |
| JE-2 | A Jump is a Removed Jump | Judge calls `SubmitJudgment` | Returns `ErrJumpRemoved` | `internal/game/judgment_eligibility_test.go` |
| JE-3 | A Judge has already Judged a Jump | Judge calls `SubmitJudgment` again | Returns `ErrAlreadyJudged` (UNIQUE constraint violation) | `internal/game/judgment_eligibility_test.go` |
| JE-4 | A Guest Judge has submitted 5 Judgments | Guest calls `SubmitJudgment` for a 6th | Returns `ErrGuestCapReached` | `internal/game/judgment_eligibility_test.go` |
| JE-5 | A Guest Judge has submitted 4 Judgments | Guest calls `SubmitJudgment` for a 5th | Judgment is accepted; `judgment_count` incremented to 5 | `internal/game/judgment_eligibility_test.go` |

### 11.3 Scoring calculations

| ID | Given | When | Then | Test File |
|----|-------|------|------|-----------|
| SC-1 | A Jump has 3 Judgments with composites 2.0, 3.0, 4.0 | Running average is computed | Returns 3.0 (to one decimal) | `internal/game/scoring_test.go` |
| SC-2 | A Jump has 2 public Judgments and 1 season-provenance Judgment (composite 2.0) | Open Final Score is computed | Season Judgment excluded; Open Final Score = AVG(public composites) | `internal/game/scoring_test.go` |
| SC-3 | A Jump has 0 Judgments | Running average is requested | Returns 0.0; UI displays "No Judgments yet" | `internal/game/scoring_test.go` |

### 11.4 Evidence upload

| ID | Given | When | Then | Test File |
|----|-------|------|------|-----------|
| EU-1 | A Jump exists and the performer requests upload authorization | Performer calls evidence authorization endpoint | Signed URL is returned with 15-minute expiry; authorization row persisted | `internal/httpapi/evidence_test.go` |
| EU-2 | An upload authorization has expired | Performer calls evidence confirmation | Returns `ErrAuthorizationExpired` | `internal/httpapi/evidence_test.go` |

### 11.5 Guest Judge session

| ID | Given | When | Then | Test File |
|----|-------|------|------|-----------|
| GJ-1 | A Guest creates a session and submits a Judgment | Guest creates an Account | Guest Judgments are migrated: `player_id` set, `guest_session_id` nulled | `internal/game/guest_session_test.go` |
| GJ-2 | A Guest session exists with 3 Judgments | Guest Judge submits another Judgment | `judgment_count` incremented to 4; Judgment inserted with `guest_session_id` | `internal/game/guest_session_test.go` |

### 11.6 API handler tests

| ID | Given | When | Then | Test File |
|----|-------|------|------|-----------|
| AH-1 | An unauthenticated request to `GET /v1/feed` | Request is sent | Returns 200 with Jump list; no auth required | `internal/httpapi/server_test.go` |
| AH-2 | An unauthenticated request to `POST /v1/jumps` | Request is sent | Returns 401; auth required for Jump creation | `internal/httpapi/server_test.go` |
| AH-3 | A Guest Judge submits a Judgment with all 4 factors | `POST /v1/jumps/{id}/judgments` with `guestSessionId` | Returns 200; Judgment persisted with `provenance` set | `internal/httpapi/server_test.go` |

---

## 12. Implementation Tracer Bullets

Each tracer bullet is a vertical slice that delivers testable value. Slices are ordered by dependency.

### Bullet 1: Schema Migration and Domain Types

**Scope:** Rename `stunts` → `jumps`; update status enum; add `grace_period_expires_at`; rename judgment columns and update CHECK constraints to 1–4; add `guest_sessions`, `opens`, `open_standings` tables; create feed indexes.

**Dependencies:** None (foundation).

**Acceptance criteria:**
- All migrations run without error against a clean database.
- `jumps.status` CHECK allows only: `'Performed Jump'`, `'Judged Jump'`, `'Unwitnessed Jump'`, `'Removed Jump'`, `'Disqualified Jump'`.
- `judgments` CHECK constraints enforce 1–4 for all four factors.
- `guest_sessions`, `opens`, `open_standings` tables exist with correct schemas.

**Key files:** `apps/api/db/migrations/`, `apps/api/internal/game/` (domain type updates).

### Bullet 2: Jump Lifecycle State Machine

**Scope:** Implement `JumpService` with `CreatePerformedJump`, `EditJumpCaption`, `RetractJump`, `RemoveJump`. Implement all state transitions from §2.2. Write domain tests for LT-1 through LT-6.

**Dependencies:** Bullet 1 (schema and domain types).

**Acceptance criteria:**
- All LT test cases pass.
- Grace Period is enforced: no Judgments accepted while active.
- Performer can edit Caption and retract during Grace Period only.
- Admin removal transitions to Removed Jump with Evidence tombstoning.

**Key files:** `internal/game/jump_service.go`, `internal/game/jump_lifecycle_test.go`.

### Bullet 3: Feed Endpoint and Persistence

**Scope:** Implement `GET /v1/feed` with cursor-based pagination and `GET /v1/jumps/{id}` detail. Implement `JumpReadRepo.Feed()` and `JumpReadRepo.Jump()` in PostgresStore. Feed excludes Removed Jumps and returns reverse-chronological order.

**Dependencies:** Bullet 2 (Jump rows must exist to query).

**Acceptance criteria:**
- `GET /v1/feed` returns 20 Jumps per page, ordered by `created_at DESC`.
- Cursor-based pagination works: `?cursor=2026-05-30T12:00:00Z` returns Jumps older than cursor.
- `GET /v1/jumps/{id}` returns Jump detail with `runningAverage`, `judgmentCount`, `gracePeriodExpiresAt`.
- Removed Jumps are excluded from feed; deep links return tombstone.
- AH-1 test passes.

**Key files:** `internal/httpapi/server.go`, `internal/httpapi/postgres_store.go`, `apps/api/openapi.yaml`.

### Bullet 4: Judging Eligibility and Submission

**Scope:** Implement `JudgmentService` with `SubmitJudgment` and `checkEligibility`. Implement all eligibility guards from §7.1. Write tests for JE-1 through JE-5 and SC-1 through SC-3.

**Dependencies:** Bullet 3 (need Jump detail to Judge against).

**Acceptance criteria:**
- All JE and SC test cases pass.
- Self-judging prevented; duplicate Judgment prevented; Grace Period enforced.
- Running average is computed correctly.
- Provenance is assigned based on active Open month.
- AH-2 and AH-3 tests pass.

**Key files:** `internal/game/judgment_service.go`, `internal/game/judgment_eligibility_test.go`, `internal/game/scoring_test.go`.

### Bullet 5: Guest Session Management

**Scope:** Implement `GuestSessionRepo` and `POST /v1/guest-sessions`. Guest-to-Player migration on Account creation. Soft cap enforcement (default 5). Write tests for GJ-1 and GJ-2.

**Dependencies:** Bullet 4 (Judgment submission needs Guest session).

**Acceptance criteria:**
- Guest session created on first app open.
- Guest Judgments tracked with `guest_session_id`; soft cap enforced.
- Guest-to-Player migration sets `player_id` on existing Judgments.
- `ClaimedGuestSession` prevents double-claim.

**Key files:** `internal/game/guest_session.go`, `internal/game/guest_session_test.go`.

### Bullet 6: The Open (Monthly Competition)

**Scope:** Implement `OpenService` with `SoftCloseMonth` and `ComputeStandings`. Implement `GET /v1/opens/{yearMonth}/standings`. Open Final Score computation per §8.3.

**Dependencies:** Bullet 4 (need Judgments and provenance data).

**Acceptance criteria:**
- `SoftCloseMonth` computes Open Final Scores for all Jumps with qualifying Judgments.
- `ComputeStandings` returns Player rankings by aggregate Open Score.
- Standings endpoint returns correct rankings with `open_score`, `judged_jumps`, and `rank`.
- Season-provenance Judgments excluded from Open Final Score.

**Key files:** `internal/game/open_service.go`, `internal/game/open_service_test.go`.

### Bullet 7: Evidence Upload and Safety

**Scope:** Implement Evidence authorization + confirmation flow per §6. Implement Report endpoint (`POST /v1/jumps/{jumpID}/reports`). Admin removal tool. Write tests for EU-1 and EU-2.

**Dependencies:** Bullet 2 (Jump must exist for Evidence; removal requires lifecycle).

**Acceptance criteria:**
- Signed URL generated with 15-minute expiry.
- Evidence confirmation validates authorization, expiry, and object existence.
- Report endpoint accepts 4 categories + Other; persists dispute row.
- Admin removal transitions Jump to Removed Jump with tombstoning.

**Key files:** `internal/httpapi/evidence.go`, `internal/httpapi/evidence_test.go`, `internal/game/dispute.go`.

---

## 13. ADR and CONTEXT.md Updates

These updates should be performed as part of implementation, not in this document.

### 13.1 ADRs requiring updates

| ADR | Update Required | Reason |
|-----|-----------------|--------|
| ADR-0005 (Postgres/sqlc) | Note table rename `stunts` → `jumps`; note new tables `guest_sessions`, `opens`, `open_standings` | Schema evolution |
| ADR-0007 (Direct Evidence Uploads) | Add tombstoning behavior for Removed Jumps; clarify object key format scoped by `jump_id` | Safety model (ADR-0024) |
| ADR-0008 (Stunts Belong to One Group) | Mark as fully superseded by ADR-0019; note `group_id` is now nullable | ADR-0019 |
| ADR-0011 (Season Close and Judging Grace Period) | Clarify that Author Grace Period is separate from Season's Judging Grace Period | Product/UX Design §1.4 |
| ADR-0016 (Client-Side Eligibility Guards) | Add Guest Judge eligibility; add Grace Period countdown guard | ADR-0022, Product/UX Design |

### 13.2 CONTEXT.md sections requiring updates

| Section | Update Required |
|---------|-----------------|
| **Jump** definition | Add Author Grace Period mention; clarify that Draft is client-only |
| **Author Grace Period** | Verify alignment with `grace_period_expires_at` column |
| **Judgment** definition | Note Guest Judge eligibility; note provenance tracking |
| **Guest Judge** definition | Note `guest_sessions` table; note soft cap enforcement |
| **Unwitnessed Jump** definition | Clarify that this is a v2/Season concept only; public feed has open-ended Judging Window |
| **Removed Jump** definition | Confirm tombstoning behavior aligns with ADR-0024 |
| **Open** definition | Note `opens` and `open_standings` tables; note soft-close mechanism |
| **Final Score** definition | Note three score types: running average (live), Open Final Score (monthly), Season Final Score (v2) |

### 13.3 New ADRs to write

| Proposed ADR | Topic | Trigger |
|--------------|-------|---------|
| ADR-0025 | Guest Judge session model | This document §7, #107 §5.4 |
| ADR-0026 | Open monthly competition data model | This document §8, #107 §3.3, ADR-0023 |
| ADR-0027 | Hexagonal architecture formalization | This document §5.3, #107 §8 |

---

## 14. References

| Document | Relationship | Key Sections Referenced |
|----------|-------------|----------------------|
| [Product/UX Design](./02-product-ux-design.md) (#106) | UX flows, screen inventory, onboarding, judging interface, share UX | §1–5, §7 |
| [Backend/Data Architecture](./03-backend-data-architecture.md) (#107) | Data model, API contract, persistence, hexagonal architecture | §2–8 |
| [MVP Roadmap](./04-mvp-roadmap.md) | Scope boundary, metrics, decision gates | §1–2 |
| [Product Vision](./01-product-vision.md) | Design pillars, target Player, growth model | §1–2 |
| ADR-0019 | Jumps are public by default | — |
| ADR-0020 | Rename Stunt → Jump, Documentation → Presentation | — |
| ADR-0021 | Season scoring excludes pre-existing public Judgments | — |
| ADR-0022 | Judgment interaction model (tap-to-select, 1–4 scale) | — |
| ADR-0023 | The Open (platform-run monthly competition) | — |
| ADR-0024 | House Rules, safety, Removed Jump behavior | — |
| ADR-0007 | Direct Evidence Uploads | — |
| ADR-0004 | REST/OpenAPI contract | — |
| ADR-0005 | Postgres/sqlc persistence | — |
| ADR-0002 | Backend owns game rules | — |
| CONTEXT.md | Domain language — all terminology must align | — |

---

_Document status: Complete. Parent tracker: #66. Depends on: Product/UX Design #106, Backend/Data Architecture #107, MVP Roadmap #65. ADRs: 0019, 0020, 0021, 0022, 0023, 0024._
