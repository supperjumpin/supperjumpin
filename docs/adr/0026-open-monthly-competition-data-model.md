# Open Monthly Competition Data Model

⚠️ **Superseded by ADR-0035 (competition apparatus deleted).** Preserved as historical record.

## Context

The Open (ADR-0023) is a platform-run monthly competition that provides competitive payoff for Players who have not joined a Group. It requires data model support for tracking months, computing Final Scores, and ranking Players.

## Decision

The `opens` table tracks monthly competition periods with `year_month`, in `YYYY-MM` format and unique, `soft_closed_at`, null until closed, and `created_at`. The `open_standings` table stores Player rankings per Open month: `player_id`, `year_month`, `open_score`, the aggregate of all Open Final Scores for that Player in that month, `judged_jumps` count, and `rank`.

The `jumps` table gains an `open_final_score` column, nullable and set at soft-close. Open Final Score is computed from Judgments with `provenance IN ('open', 'public')` received before the soft-close timestamp. Season-provenance Judgments are excluded per ADR-0021.

## Rationale

A dedicated `opens` table provides a clean anchor for soft-close operations and future extensions such as weekly checkpoints and Awards. Separating `open_standings` from `season_standings` keeps the two competitive contexts independent. The `open_final_score` column on `jumps` avoids recomputing scores on every read.

## Consequences

The soft-close operation must atomically update `opens.soft_closed_at`, compute all Open Final Scores, and populate `open_standings`. A Jump may have both `open_final_score` and `season_final_score` independently. The running average, live, is always computed from all Judgments regardless of provenance.

## References

- ADR-0023
- Technical Design §8
- Backend/Data Architecture (#107) §3.3
