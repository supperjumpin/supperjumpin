# The Open: Platform-Run Monthly Competition

## Context

Jumps are public by default (ADR-0019). Players on the public feed accumulate Judgments and a running average score, but there is no bounded competition period — no Finals Scores, no Standings, no Awards — unless a Player joins a Group and participates in a Season. This creates a gap: Players who have not formed or joined a Group have no competitive payoff moment.

The original Season model requires a Group and a Season Commissioner. Extending Season to cover a global, platform-run competition would create a degenerate Season with no Commissioner, no House Rules authority, and no Group — violating the assumptions baked into Season Commissioner (ADR-0009), Group Admin Season Override (ADR-0010), and Season Close and Judging Grace Period (ADR-0011).

## Decision

The platform runs a monthly competition called **the Open**, distinct from Group Seasons.

- The Open is platform-run and requires no Season Commissioner or Group membership.
- It runs on a fixed monthly calendar cadence and is always active — there is no cold-start coordination problem.
- The Open soft-closes at month-end: Final Scores are computed from whatever Judgments exist at that moment. There is no explicit Submission Window or Judging Grace Period.
- Any Player may compete in the Open by having at least one Performed Jump with at least one Judgment in the calendar month.
- Open Standings and Awards are separate from Season Standings and Awards. A Jump may earn both an Open Final Score and a Season Final Score independently.
- "Season" retains its existing definition: a bounded competition period within a Group, started by a Season Commissioner. The Open is not a Season.

## Rationale

The Open solves the payoff gap for Players who have not joined a Group without contaminating the Season model. Expanding "Season" to cover both would have silently broken multiple ADRs that assume a Commissioner owns the Season lifecycle. Giving the platform-run competition its own name — the Open, as in the U.S. Open or British Open — signals that an institution runs it, not a Player, and that entry is unrestricted.

The soft-close model (snapshot at month-end, no explicit phases) was chosen over the full Submission Window + Judging Grace Period model because the Open has no Commissioner to manage the close. At MVP user scale, a full phase model adds complexity without meaningfully improving competitive integrity.

The monthly cadence was chosen over weekly because the expected play rate for an active Player is approximately two Jumps per month. Weekly windows would make most Players feel perpetually late; monthly windows give a Jump enough time to accumulate meaningful Judgments before Final Scores lock.

## Consequences

- The data model must track Open membership per Judgment: each Judgment must record whether it was submitted within a given Open month, parallel to the Season Judgment provenance tracking in ADR-0021.
- A Jump may display up to three scores: a public running average (all Judgments), an Open Final Score (Judgments in the Open month), and a Season Final Score (Season-linked Judgments). The UI must make these distinctions legible without overwhelming the Player.
- "Open Season" is the natural community shorthand and is acceptable as informal language; the canonical term is "the Open."
