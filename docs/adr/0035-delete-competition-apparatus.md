# The competition apparatus is deleted, not demoted

The #314 pivot removes scoring and competition from Supperjumpin entirely. **Reactions**, **Stamps**, **Lore**, and the **Recap** are the only feedback surface. There is no backstage scoring.

## What this deletes

The following concepts are removed from the v1 domain:

- **Judgment** and the four scoring axes (Commitment, Transgression, Creativity, Presentation) and Credibility
- **Judge** and **Guest Judge** (and the guest-session cap model)
- **Final Score** / **Open Final Score** / live running average
- **The Open** (platform-run monthly competition)
- **Standings**
- **Judging Window** and **Author Grace Period** (the grace period existed to gate the Judging Window; with no judging, it has no purpose — submission edit-ability, if any, becomes a Round-reveal-timing concern, not a judging-grace concern)

## What replaces it

A **Player** expresses themselves on a **Jump** only through **Reactions** (typed **Stamps**) and free-form **Comments**. Durable recognition is **Lore**, which is emergent and derived from stamp density (ADR-0033), surfaced in the **Recap**. Nothing produces a numeric score or a ranking.

## Why

The owner's direction is "focus on the fun." The #314 concept brief lists "the Open as primary retention engine" as Scope OUT and flags "backstage scoring can add texture without contaminating the fun" as an *unproven bet*. Keeping scoring backstage carries contamination risk (the fun layer drifting back into a scoring form) for a benefit that was never validated. Deleting it outright yields the smallest, most stable API contract and the clearest product spine — which directly serves the goal of an API that does not change often.

## Status

accepted — supersedes ADR-0021 (season scoring), ADR-0022 (judgment interaction model), ADR-0023 (the Open), ADR-0025 (guest judge session model), ADR-0026 (Open monthly competition data model), and the Judgment/Standings references in ADR-0034. ADR-0013/0017 (gesture-driven scoring) are moot. Reverses the scoring-centric portions of the prior product model.
