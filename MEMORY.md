# Project Memory

## 🟢 Current Focus
- **Objective**: Implement Peer Judging Loop (T-A #19)
- **Active Issue**: #19
- **Status**: In-progress (Designing schema and API)

## 🏗️ Architecture & Decisions
- **Dual-Track Approach**: Track A (Engine/Foundation) and Track B (User Experience) are running in parallel. Track B depends on Track A's underlying logic.
- **Identity**: Many-to-One mapping (Auth $\rightarrow$ Account $\rightarrow$ Player) implemented in `PostgresStore`.
- **Stunt Lifecycle**: Idea $\rightarrow$ Planned $\rightarrow$ Performed (gated by Evidence).

## ⚠️ Hurdles & Gotchas
- **Concurrency**: Ensure judgments are idempotent (one per player per stunt).
- **Eligibility**: Strictly enforce that only Group members (excluding the performer) can judge.

## 💡 Working Hypotheses
- **Tracer Bullet**: By implementing a basic scoring API first, we can verify the end-to-end loop before adding the complex gesture-driven UI of #20.

## 📡 Hand-off Notes
- **Next Steps**: Implement `judgments` table, add `SubmitJudgment` to Store, and expose via HTTP.
- **Warning**: Ensure the state machine guard for 'Performed Stunt' is checked before allowing a judgment.
