# Project Memory

## ?? PRD #1: First Playable Group Stunt Loop
**Status**: OPEN | **Author**: Ben Turney

A single-player-per-group social game where players perform absurd food-location stunts, share evidence, judge each other's stunts on 4 axes (Difficulty, Transgression, Creativity, Documentation), and compete for season standings. Uses Supabase Auth + Postgres + Expo mobile app + Go REST API. MVP excludes progression systems, public feed, auto-scoring, video evidence, push notifications.

---

### Key Mechanisms
- Groups multi-member with admin/commissioner roles  
- Season phases: submission ? judging grace period ? finalized  
- Stunt lifecycle: Idea ? Planned ? Performed (requires evidence + caption)  
- Judging: one judgment per judge, editable during window, locked after close  
- Standings based on season-linked judged stunts only; off-season excluded

### Current Focus
**Objective**: Implement Peer Judging Loop (T-A #19)  
**Active Issue**: #19  
**Status**: In-progress (Designing schema and API)

## ??? Architecture & Decisions
- **Dual-Track Approach**: Track A (Engine/Foundation) and Track B (User Experience) are running in parallel. Track B depends on Track A's underlying logic.
- **Identity**: Many-to-One mapping (Auth $\rightarrow$ Account $\rightarrow$ Player) implemented in PostgresStore.
- **Stunt Lifecycle**: Idea $\rightarrow$ Planned $\rightarrow$ Performed (gated by Evidence).

## ?? Hurdles & Gotchas
- **Concurrency**: Ensure judgments are idempotent (one per player per stunt).
- **Eligibility**: Strictly enforce that only Group members (excluding the performer) can judge.

## ?? Working Hypotheses
- **Tracer Bullet**: By implementing a basic scoring API first, we can verify the end-to-end loop before adding the complex gesture-driven UI of #20.

## ?? PRD #1 Highlights
- **Auth**: Supabase (email magic + social), separate from in-game Player ID  
- **Stunts**: Season-linked by default when active; off-season for casual play
- **Judging Window**: Enforces one judgment/judge/stunt, edits allowed until close  
- **Scoring**: Final Score = aggregation of judgments; Season Score accumulates non-disqualified stunts  
- **Disputes**: Commissioner resolves disputes; removal reserved for serious violations

## ?? Hand-off Notes
- **Next Steps**: Implement judgments table, add SubmitJudgment to Store, and expose via HTTP.
- **Warning**: Ensure the state machine guard for 'Performed Stunt' is checked before allowing a judgment.
