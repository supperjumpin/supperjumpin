# Season Scoring Excludes Pre-Existing Public Judgments

⚠️ **Moot per ADR-0035 (scoring/Judgments deleted).** Preserved as historical record.

## Context

Jumps are public by default (ADR-0019) and accumulate Judgments on the public feed over an open-ended Judging Window. A Player may submit any existing Performed Jump to a Season during its Submission Window (no age restriction). This creates a question: when a pre-existing Jump enters a Season, do the Judgments it already received on the public feed count toward its Season Final Score?

## Decision

Only Judgments submitted while a Jump is Season-linked count toward that Season's Final Score. Pre-existing public Judgments contribute to the public running average only and are excluded from Season scoring.

## Rationale

If pre-existing Judgments carried into Season scoring, a Player could farm a high-average Jump on the public feed over weeks before submitting it to a Season pre-loaded with favorable scores that other Season participants never had the opportunity to influence. The Season Final Score would then reflect a crowd of strangers rather than the Group that is actually competing. This undermines the competitive integrity of Standings regardless of whether the submission was intentionally strategic.

Season re-judging from zero puts all Players on equal footing within the Group: the Season Final Score is determined entirely by Judgments submitted by people who encountered the Jump in the Season context. A Player who submits a proven Jump has an informational advantage (they know it's strong) but not a scoring advantage.

The public running average remains a live aggregate of all Judgments ever received, independent of Season scoring.

## Consequences

- The data model must track Judgment provenance: each Judgment must record whether it was submitted in a public-feed context or a Season-linked context, and Season Final Score aggregation must filter accordingly.
- A Jump submitted to a Season may have two visible scores: a public running average (all Judgments) and a Season score (Season-linked Judgments only). The UI must make this distinction legible.
- Season Commissioners may still restrict retroactive submission via House Rules if they want fresh-only Jumps, but the platform default is permissive — only the scoring scope is restricted, not the submission eligibility.
