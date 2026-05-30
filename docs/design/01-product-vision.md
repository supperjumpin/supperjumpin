# Product Vision

_Part of the [Supperjumpin Design Package](./README.md). Feeds: [Product/UX Design](./02-product-ux-design.md)._

## MVP Identity

**Supperjumpin is a game where you take food somewhere it doesn't belong, document it, and get judged on how well you pulled it off.**

In MVP, Supperjumpin is:
- A **public performance stage** where every Jump is visible to anyone — no private posts, no follower gates
- A **social judgment engine** where Players evaluate each other on Commitment, Transgression, Creativity, and Presentation
- A **group-friendly overlay** where friends can run Seasons and Standings, but Groups are never required to participate
- A **low-barrier, high-commitment** experience: browsing and Judging are effortless; posting a Jump requires real-world effort and Evidence
- A **direct-share artifact** designed to be forwarded to someone who has never heard of the game and understood in one glance

## Design Pillars

These are player-experience promises, not business goals. Every product and UX decision in MVP should trace back to at least one pillar.

### 1. Performance over Consumption
Players are performers first, spectators second. The act of creating, executing, and documenting a Jump is the primary source of value — not passive scroll. The feed exists to surface performances worth reacting to, not to maximize time-on-app. The content a Player sees should make them want to Judge, not keep watching.

### 2. Judgment as Play
Judging is not a review or a survey; it is a participatory act of evaluating absurdity, commitment, and creativity. The interface presents all four factors on a single screen with tap-to-select tier labels — Judges give a verdict, not fill a form. After selecting all four verdicts, a confirmation screen displays the full ruling before submission. The confirmation reads as a filing receipt, not a celebration: the act of Judging should feel like rendering a decision, not gamifying a reaction. Every Judgment is a small performance in itself.

### 3. Public Stage, Private Circles
Every Jump is a public performance by default. Private Groups exist only to add competitive layers on top of the public feed — they never gate access to content or Judging. The public stage builds the audience through the **Open** (the platform-run monthly competition that gives every Player a competitive payoff moment); private circles build the rivalry through Group Seasons (v2). No Jump, Judgment, or score is hidden behind a Group wall.

### 4. Absurdity within Boundaries
The game rewards transgressive humor, deadpan commitment, and awkward situations, but it strictly enforces House Rules. A Jump can be uncomfortable or weird; it cannot require harassment, trespass, unsafe behavior, or harm. The boundary is part of the design, not an afterthought. The Transgression scoring axis structurally rewards escalation — moderation and Dispute mechanics are load-bearing from day one, not features to add later.

### 5. Low Friction, High Commitment
The first session should require zero setup: open the app, see Jumps, understand the game. The first ask is to Judge (low cost), not to post (high cost). Guest Judges can evaluate without creating an Account — no auth gate on consumption. Posting a Jump requires real-world effort — buying food, going somewhere, documenting it — which makes the performance meaningful and the Judgment consequential. This asymmetry means the growth loop depends on Judging being intrinsically rewarding enough to convert passive visitors into active performers.

## Target Early Player

A person who does absurd food bits for a small, specific audience — the group chat, not the algorithm. They have already done a version of a Jump in real life and told the story. They care more about the reaction of the people who get it than about reach. They would do the bit even if it never left the group chat.

**Seed community screen:** Does the candidate think it's funny to do a bit for 5 people who will never share it? If yes, recruit them. If the scenario feels deflating, they are not the right early performer.

## First 5-Minute Experience

1. **Unauthenticated visitor opens app** — lands directly on a public feed of seeded Jumps; no login gate.
2. **Browses 2–3 Jumps** — sees Source, Destination, Food, Caption, Evidence; learns the game from examples.
3. **Hits a soft auth gate when they try to Judge** — one-tap social login; no email verification on first tap.
4. **Judges 1–2 Jumps** — tap-to-select verdicts on all four factors (Commitment, Transgression, Creativity, Presentation), each on a 1–4 forced-choice scale with named tier labels, then confirms the ruling on a filing-receipt screen.
5. **Invited to post their own** — CTA surfaces after the first Judgment, not on app open.

Auth gates contribution, not consumption. The first ask is to Judge (low cost), not to post (high cost).

## Group-First vs. Public Visibility

Jumps are public by default (ADR-0019). Any authenticated Player may Judge any Jump they did not perform — Group membership is not required (CONTEXT.md). Guest Judges may also Judge without creating an Account. Groups are the competitive overlay for Seasons and Standings, not the access gate.

**Private Jumps are a non-goal for MVP.** The product promise is public performance and open Judgment. Player-controlled visibility is a different product.

## The Open: v1 Competitive Context

The platform runs a monthly competition called **the Open** (ADR-0023), distinct from Group Seasons:

- The Open is platform-run and requires no Season Commissioner or Group membership.
- It runs on a fixed monthly calendar cadence and is always active — no cold-start coordination problem.
- The Open soft-closes at month-end: Final Scores are computed from whatever Judgments exist at that moment.
- Any Player may compete by having at least one Performed Jump with at least one Judgment in the calendar month.
- Open Standings and Awards are separate from Season Standings and Awards. A Jump may earn both.

The Open solves the payoff gap for Players who have not joined a Group. It is the v1 competitive engine — Group Seasons (with Commissioner, Submission Windows, Judging Grace Periods) are v2.

**The Open replaces the earlier assumption that v1 would have no competitive structure.** It gives Players a reason to return after their first Jump without requiring them to coordinate a Group.

## Growth Model

### North Star Metric

**Judgments per Jump within 7 days of posting** — the primary measure of product health for v1. It captures whether the core promise ("take food somewhere it doesn't belong, get judged") is being fulfilled.

Supporting health metrics:
- **Guest-to-Player conversion rate** — % of Guest Judges who create an Account
- **Share-to-Judge rate** — % of shared links that result in a Judgment
- **Jump-to-Open Final Score rate** — % of Performed Jumps that earn a score in the monthly Open

### Primary Growth Vector: Direct Share

The Jump result artifact (Evidence + four-factor score breakdown) is the primary viral surface. It is designed to be forwarded to someone who has never heard of Supperjumpin and immediately make them laugh or want to try it (research issue #62). A Share surfaces a deep link with a preview card containing the Evidence photo, a truncated Caption, the running average score, and the Source/Destination/Food summary. The recipient opens the Jump detail view directly, where they may Judge without creating an Account.

Direct share is the cold-start growth vector, not algorithmic feed discovery. The feed rewards what already has share velocity — it does not replace it.

## Non-Goals

These are explicit scope boundaries that prevent Supperjumpin from becoming a generic social feed or a fantasy-sports clone:

- **No private Jumps.** Every posted Jump is public. Visibility controls are future work.
- **No permanent global leaderboard.** The Open's Standings reset monthly — no lifetime rank creates unwinnable distance. Group Standings are v2.
- **No follower graph.** Asymmetric follow/following turns the game into clout accumulation. The Group is the right social unit.
- **No Like button alongside scoring.** A Like would short-circuit the four-factor Judgment system. One interaction model: the tap-to-select verdict interface.
- **No Group Seasons, Group Standings, or Group Awards in v1.** The competitive layer is the Open (v1) and is validated before investing in the Group infrastructure that Seasons require.
- **No Missions, Bounties, or Levels in v1.** Progression systems belong after retention is established, not before.
- **No Disputes or Disqualification tooling in v1.** Manual moderation by the team is sufficient at MVP scale.
- **No Sponsored Bounties.** No monetization surface before product-market fit.
- **No Idea or Planned Jump pre-commitment flow for public-feed posts.** Players submit Performed Jumps with Evidence directly. The planning layer is a later optimization for Group coordination.

## Market Research Implications

The following findings from the competitive and cultural research (issue #62) directly inform the product vision:

### Evidence Quality Floor
Seed content at casual phone-video quality, not polished production. The norm is set by what Players see first. The Presentation scoring axis will naturally penalize low-quality Evidence, so the quality floor must be designed — inconsistent quality between Players will suppress casual participation over time.

### Direct Share Is the Growth Surface
Wordle's viral growth was driven by shareable result artifacts forwarded via DMs and group chats. Supperjumpin's Jump artifact follows the same pattern: "look at this ridiculous thing someone did" is a strong message to send a specific friend. The shareable artifact must be intelligible to someone who has never heard of the game.

### Transgression Escalation Is Structural
The Transgression scoring axis rewards pushing limits. TikTok challenge history (fire noodle challenge: 1.5M+ videos, documented harm) shows this pattern has real-world consequences. House Rules and moderation investment are load-bearing, not cosmetic. The game must enforce boundaries without deflating the absurdist spirit.

## Decisions

The following contested points were resolved during the grilling and design sessions that produced this document:

| Question | Decision | Source |
|---|---|---|
| MVP lane | Public performance stage with open Judging; Groups are optional overlay | Issue #56 |
| Stunt vs. Jump | Canon term is Jump; code and schema rename is implementation work | ADR-0020 |
| Documentation vs. Presentation | Scoring factor renamed to Presentation; sharpens the distinction from Credibility | ADR-0020 |
| Core play lifecycle | Players submit Performed Jumps with Evidence directly; no pre-commitment; Author Grace Period (10 min) before Judging Window opens | Issue #58, ADR-0022 |
| Judging interaction | Single-screen tap-to-select with confirmation; no gestures, no sliders, no sequential screens | ADR-0022 |
| Scoring factors | Commitment, Transgression, Creativity, Presentation; 1–4 forced-choice with named tier labels | ADR-0022 |
| Group role in Judging | Group membership not required; any Player or Guest Judge may Judge any Jump | ADR-0019, ADR-0022 |
| Season relationship to public feed | Season scoring uses only Judgments submitted while Jump is Season-linked; pre-existing public Judgments excluded | ADR-0021 |
| v1 competitive structure | The Open (platform-run monthly competition); Group Seasons are v2 | ADR-0023 |
| Guest Judges | Allowed in v1; Judgments stored by device/session; can claim history by creating an Account | CONTEXT.md |
| Private Jumps | Non-goal for MVP; public-by-default is the product promise | ADR-0019 |
| Auth model | Auth gates contribution (Judging), not consumption (viewing). Guest viewing requires no login | Issue #56 |
| Tone register | Deadpan-institutional for game terms; Jump is the only term that should feel playful | ADR-0020 |

## Open Questions

These questions are raised by the research and design decisions but cannot be resolved without product experience and data:

1. **What replaces the reciprocity gate for cold-start retention?** BeReal required posting before seeing content — its stickiest retention mechanic is unavailable here. ADR-0019 locks open Judging, so first-time Players have no structural reason to return after their first visit beyond the quality of the content itself. Whether the Open's monthly cadence, push notifications on Judgments received, or something else fills this gap is the single highest-risk product question for v1 retention.

2. **What is the transgression ceiling, and who enforces it at scale?** The scoring system structurally rewards escalation. At cold-start scale — before community norms are established — House Rules and manual moderation are the only checks. Whether automated moderation tooling, score caps, or community enforcement is sufficient before dedicated tooling (Disputes, Disqualification) ships is an open investment decision.

3. **How should the product constrain Evidence quality?** Research suggests inconsistent quality will suppress casual participation over time (issue #62). Whether the product should actively constrain quality (à la BeReal's anti-curation stance), let it self-regulate, or rely on the Presentation axis to signal norms is a product and community design decision that will be refined after observing early Player behavior.
