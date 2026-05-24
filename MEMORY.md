# Project Memory

## ?? PRD #1: First Playable Group Stunt Loop

**Status**: OPEN, ready-for-agent  
**Author**: Ben Turney  
**Scope**: MVP proving the core social game loop without drifting into generic social feed or progression systems.

### Core User Stories
- **Auth & Player Identity**: Email magic links and social login via Supabase Auth; separate external auth from in-game Account/Player identity.
- **Group Formation**: Create/join Groups, admin roles, invites, multi-group membership.
- **Season Lifecycle**: Start Season (any member ? Commissioner), phases: Submission Window ? Judging Grace Period ? Finalized. Admin emergency overrides with visible audit trails.
- **Stunt Lifecycle**: Ideas ? Planned Stunts (Source, Destination, Food) ? Performed Stunts (Evidence + Caption). Defaults to Season-linked during active Season; Off-Season available for casual play.
- **Judging System**: Group members ? performer can judge; one Judgment per Judge per Stunt with 4 axes (Difficulty, Transgression, Creativity, Documentation); editable during Judging Window, locked after close.
- **Scoring & Standings**: Final Score from Judgments; Season Score accumulates non-disqualified Judged Stunts; only Season-linked, Judged Stunts affect Standings.
- **Disputes & Moderation**: Dispute resolution (Commissioner/Admin), disqualification (visible) vs removal (serious violations).

### Implementation Architecture
- **Monorepo**: Expo React Native mobile app + Go backend API + generated TypeScript client (	s-codegen)
- **API**: REST with OpenAPI contract (no GraphQL; explicit game commands dominate MVP)
- **Storage**: Postgres with sqlc-generated Go data access; direct media uploads to object storage via backend authorization
- **Auth**: Supabase for external auth, internal Account/Player records in Go backend

### Testing Priorities
- Domain/service tests for Season/Stunt/Judging lifecycle
- Authorization tests (member, admin, commissioner permissions)
- Evidence upload flow (backend authorize + direct upload)
- Mobile smoke-level flows after Expo scaffold

### Out of Scope (MVP)
- ML scoring, public feed, restaurant discovery, video evidence, push notifications, full progression mechanics (Missions/Bounties/Levels/currency).

### Key Rules & Invariants
1. At most one active or closing Season per Group
2. Player belongs to many Groups via Group Memberships
3. Stunt belongs to exactly one Group for MVP
4. Off-Season Stunts excluded from Standings
5. Self-judging forbidden (performer ? judge)
6. One Judgment per Judge per Stunt; editable until Judging Window close
7. Unjudged Stunts (=1 submitted but window closed) don't award Score
8. Disqualified = visible + excluded from Standings; Removal = serious violations only

### Deferred Decisions
- Go router/framework, migrations (sqlc?), OpenAPI tooling, local Supabase workflow, object storage provider, deployment target, mobile test strategy.

---

## ?? Current Focus
- **Objective**: Implement Peer Judging Loop (T-A #19)
- **Active Issue**: #19
- **Status**: In-progress (Designing schema and API)

### Architecture & Decisions
- Dual-Track Approach: Track A (Engine/Foundation) and Track B (User Experience) parallel.
- Identity: Many-to-One mapping (Auth ? Account ? Player) in PostgresStore.
- Stunt Lifecycle: Idea ? Planned ? Performed (gated by Evidence).

### Hurdles & Gotchas
- Concurrency: Judgment idempotency (one per player per stunt).
- Eligibility: Only Group members ? performer can judge.

### Working Hypotheses
- Tracer Bullet: Basic scoring API first to verify end-to-end loop before complex UI (#20).

### Next Steps
- Implement judgments table, add SubmitJudgment to Store, expose via HTTP.
- Verify state machine guard for 'Performed Stunt' before judgment.
