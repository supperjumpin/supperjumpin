# Round lifecycle: commit, submit, reveal-by-condition, react, recap

The **Round** is the aggregate root (ADR-0032). Its lifecycle has two distinct participation acts and a condition-driven reveal.

## The lifecycle

1. **Commit** — a **Player** becomes a **Jumper** by committing to the **Round**'s **Prompt** (the "I'm In" beat). This is a distinct act *before* any **Jump** exists.
2. **Submit** — a **Jumper** privately submits a sealed **Jump** (**Evidence** + **Caption**). It stays sealed until reveal.
3. **Reveal** — the **Round**'s reveal condition fires; all sealed **Jumps** flip to revealed simultaneously.
4. **React** — **Reactions** (**Stamps**) and **Comments** accrue on the revealed **Jumps**.
5. **Recap** — a **Recap** is produced; standout bits become **Lore** (ADR-0033/0034).

## Two distinct acts, not one

Committing and submitting are **separate states**: `committed → submitted`. A **Jumper** who commits but never submits is a **Ghost Jumper** — a first-class, distinguishable state — not the same as a **Player** who didn't play.

Why two acts (the simpler one-act "submit = join" was rejected):
- The visible "I'm In" count creates anticipation and mild social pressure that makes a *synchronized* reveal feel like an event, not a deadline.
- The commit-without-submit gap is a **comedy surface** — the **Recap** can affectionately roast the ghosters. This anchors the "affectionate failure" register in domain state rather than only in a **Stamp** label.
- It is what makes a future threshold-based reveal (below) possible at all — you can't trigger on "all committed Jumpers submitted" without knowing who committed.

The cost is one extra state transition, which is cheap.

## Reveal is a condition, not a hardcoded timer

The **Round** holds a **reveal condition** it evaluates, rather than a bare timestamp check. For v1 the only implemented condition is **scheduled-time**: the initiator sets the **Round**'s reveal time at start, choosing from a curated **menu of timeframes** (e.g. "in 24 hours," "this Friday 8pm," "in 3 days") rather than a fixed platform default or free-entry. The set of menu options is tunable data (edge); that the initiator picks a per-**Round** reveal time at start is spine. Modeling reveal as a condition is the deliberate escape hatch: **organizer-triggered** ("/reveal now") and **threshold-triggered** ("all committed **Jumpers** have submitted") become additional condition variants later, without reshaping the **Round**.

Time-scheduled was chosen as the v1 spine because a known reveal moment is what turns the reveal into a shared, anticipated event (the synchronized-reveal beat is the heart of the pivot); organizer-only would make it fragile/human-dependent, and threshold-only removes the fixed-moment anticipation.

## Edge (tunable, not contract)

Cadence (weekly vs. other), the specific reveal time, and submission rules are tunable data/config, not contract shape.

## Status

accepted
