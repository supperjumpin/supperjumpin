# Product Vision

_Part of the [Supperjumpin Design Package](./README.md). Feeds: [Product/UX Design](./02-product-ux-design.md)._

## MVP Lane

**Supperjumpin is a game where you take food somewhere it doesn't belong, document it, and get judged on how well you pulled it off.**

## Target Early Player

A person who does absurd food bits for a small, specific audience — the group chat, not the algorithm. They have already done a version of a Jump in real life and told the story. They care more about the reaction of the people who get it than about reach. They would do the bit even if it never left the group chat.

**Seed community screen:** Does the candidate think it's funny to do a bit for 5 people who will never share it? If yes, recruit them. If the scenario feels deflating, they are not the right early performer.

## First 5-Minute Experience

1. **Unauthenticated visitor opens app** — lands directly on a public feed of seeded Jumps; no login gate.
2. **Browses 2–3 Jumps** — sees Source, Destination, Food, Caption, Evidence; learns the game from examples.
3. **Hits a soft auth gate when they try to Judge** — one-tap social login; no email verification on first tap.
4. **Judges 1–2 Jumps** — gesture-based scoring on four axes; feels like participation, not a form.
5. **Invited to post their own** — CTA surfaces after the first Judgment, not on app open.

Auth gates contribution, not consumption. The first ask is to Judge (low cost), not to post (high cost).

## Group-First vs. Public Visibility

Jumps are public by default (ADR-0019). Any authenticated Player may Judge any Jump they did not perform — Group membership is not required (CONTEXT.md updated). Groups are the competitive overlay for Seasons and Standings, not the access gate.

**Private Jumps are a non-goal for MVP.** The product promise is public performance and open Judgment. Player-controlled visibility is a different product.

## Non-Goals

These are explicit scope boundaries that prevent Supperjumpin from becoming a generic social feed or a fantasy-sports clone:

- **No private Jumps.** Every posted Jump is public. Visibility controls are future work.
- **No global Standings.** A global leaderboard makes competition unwinnable before it starts. Standings are always scoped to a Group Season.
- **No follower graph.** Asymmetric follow/following turns the game into clout accumulation. The Group is the right social unit.
- **No Like button alongside scoring.** A Like would short-circuit the four-factor Judgment system. One interaction model: gesture scoring.
- **No Seasons, Standings, or Awards in v1.** The competitive layer is the retention mechanic, not the acquisition hook. It is validated after the Judgment loop proves its value.
- **No Missions, Bounties, or Levels in v1.** Progression systems belong after retention is established, not before.
- **No Disputes or Disqualification tooling in v1.** Manual moderation by the team is sufficient at MVP scale.
- **No Sponsored Bounties.** No monetization surface before product-market fit.
- **No Idea or Planned Jump pre-commitment flow for public-feed posts.** Players submit Performed Jumps with Evidence directly. The planning layer is a later optimization for Group coordination.

## Reopened Assumptions

The following assumptions in CONTEXT.md or existing ADRs were examined and either updated or explicitly confirmed during this session:

| Assumption | Status | Resolution |
|---|---|---|
| Judge, Judged Jump, Judging Window were Group-scoped | **Updated** | All three definitions now reflect public, open Judging. Judging Window is open-ended on the public feed; Season-scoped within a Judging Grace Period. |
| Example Dialogue implied Group membership required before any action | **Updated** | Rewritten to show public-first flow; Group/Season is additive. |
| "Friends-only social performance" as a distinct MVP lane | **Closed** | Not a separate lane — friends-only visibility is a future privacy option, not the MVP promise. |
| Auth gates the feed | **Flagged as gap** | Current build requires login before viewing the feed. Unauthenticated feed access is the highest-priority first-session change for the UX Design doc. |
