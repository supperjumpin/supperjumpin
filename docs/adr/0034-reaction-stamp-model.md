# Reactions are typed stamps over a data-driven catalog

The on-stage interaction at a **Round** reveal is a **Reaction**: a **Player** applying a **Stamp** to a **Jump**. The contract shape is stable; the stamp vocabulary is tunable data.

## Spine (stable contract)

- A **Reaction** is `(player, jump, stampKind, revealContext)`, submitted at or after the synchronized reveal.
- A **Reaction** is the *only* expressive act on a **Jump** (alongside free-form **Comments**). It feeds **Lore** and the **Recap** and never produces a score or ranking. (Originally this ADR framed **Reaction** as distinct-from-but-coexisting-with a backstage **Judgment**; ADR-0035 subsequently deleted the competition apparatus entirely, so there is no **Judgment** to distinguish it from.)
- The stable identity of a stamp is its **stance** (e.g. approval, appetite, chaos, lore, certification, affectionate-failure), addressed by id — not its display string.
- Reactions feed **Lore** by density (per ADR-0033).

## Edge (tunable data, never in `openapi.yaml`)

- The **Stamp** catalog rows — label, glyph/emoji, copy, ordering, active flag, optional lore-weight — live as seeded data, not a Go enum or OpenAPI enum. Labels are expected to churn continuously (slang expiry, seasonal stamps, community in-jokes); that churn must never require an OpenAPI or Go enum change. (Originally the rationale cited the OpenAPI sync gate; that gate is gone as of ADR-0049, but the structural reason — labels are edge, stance is spine — still holds.)

## Also decided

- **Threaded comments** remain a *separate* free-form channel alongside stamps (the riff surface), not modeled as Reactions. Stamps are the countable signal; comments are freeform comedy.
- **No head-to-head vote or leaderboard** at the reveal — it would reintroduce the competence-exposure and competition the #314 pivot deleted (ADR-0035).

## Why

Two independent expert analyses of Jackbox party-game design converged: the fun lives in the reveal "theater" and the (async) emcee — which for us is the **Recap** — not in the reaction button itself. Stamps must be short, evergreen, one-tap, and survive being pressed dozens of times a night. A hardcoded enum would force a contract change (and CI gate trip) on every label tweak, violating the goal of a stable API; free-form-only reactions can't be aggregated into typed **Lore** density. Therefore: stamp *kind/stance* is spine, stamp *label* is edge.

## Status

accepted
