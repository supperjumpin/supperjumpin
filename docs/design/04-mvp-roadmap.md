# MVP Roadmap and Sequencing Plan

_Part of the [Supperjumpin Design Package](./README.md). Feeds: [Implementation Backlog](./05-implementation-backlog.md)._

## Overview

This roadmap turns the [product vision](./01-product-vision.md) into staged, verifiable work. Each stage has a clear scope boundary, a success metric, and an explicit decision gate before the next stage begins. The goal is to ship the smallest viable loop first, validate it with real Players, then expand — never the reverse.

The roadmap is organized into four horizons:

1. **First Playable Loop** — the smallest end-to-end experience that proves the core mechanic works
2. **MVP** — the full v1 product, scoped to validate product-market fit
3. **Post-MVP Experiments** — growth, retention, safety, and competition experiments that build on a working core
4. **Later** — monetization, sponsorship, and creator paths that require scale and trust

---

## First Playable Loop

**Goal:** A single Player can perform a Jump, submit Evidence, and receive at least one Judgment from another person within 24 hours.

**Scope:**

- Public feed of Jumps (seeded with team-generated content)
- Jump creation: Source, Destination, Food, Caption, photo Evidence
- Tap-to-select Judging on four factors (Commitment, Transgression, Creativity, Presentation)
- Guest Judges may Judge without creating an Account
- Running average score displayed per Jump
- Share link with preview card for external distribution

**Explicitly out of scope for First Playable:**

- Authentication (Guest Judges only; no Accounts)
- The Open (monthly competition)
- Standings of any kind
- Groups
- Report/Remove flow
- Push notifications
- Deep links from share cards

**Success metric:**

> **Judgments per Jump within 7 days of posting ≥ 1.0**

If a Jump cannot attract a single Judgment within a week, the core promise is unfulfilled. This is the only metric that matters at this stage.

**Decision gate:** Run the First Playable Loop with 10–20 seed Players for 2 weeks. If the metric is not met, do not expand scope. Diagnose whether the problem is: (a) not enough Judges, (b) Jumps are not compelling, (c) the Judging interface is too heavy, or (d) the share loop is broken. Fix the bottleneck before proceeding.

---

## MVP (v1)

**Goal:** A self-sustaining public performance stage where Players routinely perform Jumps, Judge others, and share results — validated by the North Star metric.

**Scope:**

### Core Loop

- **Authentication:** One-tap social login (no email verification on first tap). Auth gates contribution (Judging, posting), not consumption.
- **Jump lifecycle:**
  1. Player creates a Performed Jump with Evidence (photo + Caption)
  2. 10-minute Author Grace Period for edits/retraction
  3. Judging Window opens
  4. Any authenticated Player or Guest Judge may Judge
  5. Running average updates in real time
- **Judging interface:** Single-screen tap-to-select with named tier labels (1–4) on all four factors. Confirmation screen reads as a filing receipt. No gestures, no sliders, no sequential screens.
- **Feed:** Chronological public feed of all Jumps. No algorithmic ranking, no follower weighting. The feed surfaces what exists; share velocity drives discovery.

### The Open

- Platform-run monthly competition
- Fixed calendar cadence, always active
- Soft-closes at month-end: Final Scores computed from whatever Judgments exist
- Any Player with at least one Performed Jump and at least one Judgment in the calendar month competes
- Open Standings and Awards (monthly reset)
- Separate from any future Group/Season Standings

### Safety

- House Rules defined in domain model (CONTEXT.md)
- Player-facing Report button with 4 categories + "Other"
- Manual team adjudication at MVP scale
- Removed Jump suppresses content fully from feed, links, and share previews
- Admin removal tool and share link tombstoning are P0 must-build before public launch

### Growth

- Share link with preview card (Evidence photo, truncated Caption, running average, Source/Destination/Food summary)
- Deep link opens Jump detail view directly
- Guest Judges may Judge without Account creation
- No invite flow, no referral mechanics, no viral nudges beyond the share artifact

### Explicitly out of scope for MVP (deferred to v2 or later)

| Feature | Deferred to | Rationale |
|---|---|---|
| Private Jumps | v2 (if ever) | Public-by-default is the product promise |
| Groups | v2 | Lightweight social circles only; no Seasons, Standings, or Awards |
| Group Seasons | v2 | The Open is the v1 competitive engine |
| Season Commissioner | v2 | No Groups means no commissioner model |
| Group Admin | v2 | No formal administration in v1 Groups |
| Group Standings / Awards | v2 | The Open provides competitive payoff |
| Missions | Post-MVP | Progression systems after retention is established |
| Bounties | Post-MVP | Same as Missions |
| Levels | Post-MVP | Same as Missions |
| Sponsored Bounties | Later | No monetization before product-market fit |
| Disputes / Disqualification tooling | v2 | Manual team moderation sufficient at seed scale |
| Auto-hide on reports | Post-MVP | Manual review viable below ~1,000 active Players |
| Image scanning (AWS Rekognition / Google Vision) | Post-MVP | Scale threshold not reached |
| Follower graph | Never (non-goal) | Asymmetric follow turns game into clout accumulation |
| Like button | Never (non-goal) | Short-circuits four-factor Judgment system |
| Permanent global leaderboard | Never (non-goal) | The Open resets monthly; no unwinnable distance |
| Idea / Planned Jump pre-commitment | v2 | Players submit Performed Jumps directly in v1 |
| Push notifications | Post-MVP | Retention experiment, not core loop |
| Algorithmic feed ranking | Later | Chronological feed preserves authenticity |

**North Star metric for MVP:**

> **Judgments per Jump within 7 days of posting ≥ 2.0**

Supporting health metrics:
- **Guest-to-Player conversion rate** ≥ 15%
- **Share-to-Judge rate** ≥ 10%
- **Jump-to-Open Final Score rate** ≥ 50%

**Decision gate:** Run MVP with 100–500 Players across 2 monthly Open cycles. If the North Star metric is not met, do not proceed to Post-MVP experiments. Diagnose and fix the core loop first.

---

## Post-MVP Experiments

These are sequenced *after* the core loop is validated. Each experiment is independent and can be run in parallel with others, but none should start before the MVP North Star is met.

### 1. Growth Experiments

| Experiment | Hypothesis | Metric | Success Threshold |
|---|---|---|---|
| Share card with score breakdown | Shareable artifact with four-factor scores drives more external distribution | Share-to-Judge rate | ≥ 15% |
| "Rate this Jump" public link | Non-Players can react without signing up; lowers friction to first Judgment | Guest-to-Player conversion from public link | ≥ 20% |
| Prompt-of-the-week push | Weekly curated Prompt reduces blank-page syndrome and spurs creation | Jumps created within 48h of Prompt | ≥ 30% of active Players |
| Season countdown viral nudge | "Submission Window closes in 3 days" urgency drives performance spike | Jumps created in final 72h of Open | ≥ 40% of monthly total |

### 2. Retention Experiments

| Experiment | Hypothesis | Metric | Success Threshold |
|---|---|---|---|
| Push notification: "Your Jump was Judged" | Dopamine hit on feedback drives re-engagement | D7 return rate after first Judgment received | ≥ 40% |
| Push notification: "Open closes in 48h" | Urgency drives lapsed Players back | Reactivation rate (14+ days inactive → active) | ≥ 15% |
| Group resurrection bot | If a Group has no activity for 2 weeks, bot posts a Prompt | Group re-activation rate | ≥ 25% |
| Referral leaderboard | Gamify invites separately from Standings | Invites sent per active Player per month | ≥ 0.5 |

### 3. Safety Experiments

| Experiment | Hypothesis | Metric | Success Threshold |
|---|---|---|---|
| Auto-hide on multiple reports | 2+ distinct Player reports auto-hide a Jump pending review | Time-to-hide for reported content | < 1 hour |
| Safety reminder at submission | High-Transgression Jumps trigger a "be mindful" nudge | Report rate on high-Transgression Jumps | ≤ 5% |
| Rate limiting on Judging | Prevent bot/drive-by Judging | Judgments per unique Judge per day | ≤ 50 |

### 4. Competition Experiments

| Experiment | Hypothesis | Metric | Success Threshold |
|---|---|---|---|
| Group Seasons (v2) | Friends want to compete in bounded Seasons with Standings and Awards | Group creation rate among active Players | ≥ 20% |
| Season Commissioner | One Player sets Submission Window and rules; adds social energy | Season completion rate (started → finalized) | ≥ 70% |
| Awards beyond Standings | End-of-Season recognitions for Creativity, Transgression, etc. | Player satisfaction with Awards | Qualitative ≥ 4/5 |

**Decision gate:** Each experiment runs for 1–2 Open cycles (1–2 months). If an experiment fails its success threshold, kill it or redesign. Do not accumulate failed experiments.

---

## Later

These are intentionally placed after product-market fit is established and the core loop is self-sustaining. They are not experiments; they are business-model and product-expansion bets.

### Monetization: Sponsored Bounties

- External brands fund Bounties: "Perform a Jump with [Brand] as Source and any coffee shop as Destination"
- Player receives reward; brand receives impressions and UGC
- Sponsored Bounties never affect Season Score or Standings
- Requires: organic Jump creation at scale, content quality baseline, advertiser-friendly brand safety

### Creator Path

- Featured Jumps, Player highlight reels, public profiles
- Requires: enough high-quality content to curate, community norms established, moderation tooling mature

### Public Discovery

- Search Jumps by Source, Destination, Food, or location
- "Jumps near me" or "Jumps involving Taco Bell"
- Requires: public content volume, location data consent model, privacy review

### Levels and Progression

- Non-competitive progression via Missions and participation
- Separate from Season Score and Standings
- Requires: retention established; progression must add meaning, not grind

---

## Sequencing Summary

```
First Playable Loop
  → Validate: Judgments per Jump ≥ 1.0
    → MVP (v1)
      → Validate: Judgments per Jump ≥ 2.0
        → Post-MVP Experiments (parallel)
          → Growth: share cards, public links, Prompts, urgency nudges
          → Retention: push notifications, Group bots, referral gamification
          → Safety: auto-hide, submission nudges, rate limiting
          → Competition: Group Seasons, Commissioner, Awards
            → Later (after product-market fit)
              → Sponsored Bounties
              → Creator path
              → Public discovery
              → Levels and progression
```

## Metrics by Stage

| Stage | Primary Metric | Target | Supporting Metrics |
|---|---|---|---|
| First Playable | Judgments per Jump (7 days) | ≥ 1.0 | — |
| MVP | Judgments per Jump (7 days) | ≥ 2.0 | Guest-to-Player ≥ 15%, Share-to-Judge ≥ 10%, Jump-to-Open ≥ 50% |
| Growth Experiments | Share-to-Judge rate | ≥ 15% | Guest conversion ≥ 20%, Prompt response ≥ 30% |
| Retention Experiments | D7 return rate | ≥ 40% | Reactivation ≥ 15%, Group re-activation ≥ 25% |
| Safety Experiments | Time-to-hide | < 1 hour | Report rate ≤ 5% |
| Competition Experiments | Group creation rate | ≥ 20% | Season completion ≥ 70% |

## Decision Gates

1. **First Playable → MVP:** Core loop validated with real Players. Do not build auth, The Open, or safety tooling until a Guest-only loop works.
2. **MVP → Post-MVP:** North Star metric met across 2 Open cycles. Do not experiment on growth or retention until the core loop sustains itself.
3. **Post-MVP → Later:** Product-market fit demonstrated by sustained North Star metric + positive supporting metrics. Do not monetize or build creator paths until organic content supply is healthy.

## Open Questions

These questions are raised by the roadmap but cannot be resolved without data from the First Playable Loop:

1. **What is the minimum viable Group size?** The roadmap assumes Groups are v2, but if the First Playable shows that Players only Judge friends, the Group timeline may need to accelerate.
2. **What is the actual viral coefficient (K)?** The roadmap assumes direct share is the primary growth vector. If K < 0.3, the product may need paid acquisition or a different loop.
3. **What is the Judging labor ceiling?** If Players perform more Jumps than the community can Judge, the loop breaks. The roadmap assumes 2+ Judgments per Jump is achievable; if not, the Judging interface or incentive model must change.
