# Supperjumpin Design Package

Parent tracker: #50

This directory holds the full design document set for the Supperjumpin redesign. The redesign is driven by two foundational decisions made before writing began:

- **Jump** replaces "Stunt" as the canonical term for a playable attempt (CONTEXT.md)
- **Jumps are public by default** — Groups are an optional competitive overlay, not a required context (ADR-0019)

## Document Set

| Document | Status | Must Decide | Must Not Decide | Needs |
|---|---|---|---|---|
| [Product Vision](./01-product-vision.md) | Done | Primary audience; what success looks like at launch | UX flows, data model | Nothing |
| [Product/UX Design](./02-product-ux-design.md) | Not started | Core loop; feed model; onboarding path; screen inventory | Data schema, API shape | Product Vision |
| [Backend/Data Architecture](./03-backend-data-architecture.md) | Not started | Data model changes from Group-first to public-first; API contract changes; what survives vs. gets reworked | UX flows, screen design | Product/UX Design |
| [MVP Roadmap](./04-mvp-roadmap.md) | Done | Which features ship in MVP vs. later; ordering | Implementation details | Backend/Data Architecture |
| [Implementation Backlog](./05-implementation-backlog.md) | Not started | Nothing — translates roadmap into independently-grabbable issues | Feature scope | MVP Roadmap |

ADRs are written inline as decisions crystallize in each upstream document, not as a single terminal step.

## Dependency Order

```
Product Vision
  → Product/UX Design
      → Backend/Data Architecture
          → MVP Roadmap
              → Implementation Backlog
```

## Major Decision Areas

### 1. Audience & Core Loop
_Owner: Product Vision + Product/UX Design_

- Who is the primary audience: friend-group competitors, viral strangers, or both?
- What does the core loop look like for a first-time Player with no Group?
- What triggers a Player to post their first Jump vs. just browse?

### 2. Feed & Discovery
_Owner: Product/UX Design_

- How does the public feed work — chronological, algorithmic, friend-weighted?
- How does a Player find Jumps to Judge?
- What does "going viral" look like in this product?

### 3. Judging Model
_Owner: Product/UX Design → Backend/Data Architecture_

- Who can Judge a Jump on the public feed?
- Does a Jump need a minimum number of Judgments before a Final Score is calculated?
- Do the four scoring factors (Difficulty, Transgression, Creativity, Documentation) survive unchanged in a public-feed context?

### 4. Group & Season Relationship to Public Feed
_Owner: Backend/Data Architecture_

- How does a Player submit a Jump to a Group's Season after posting it publicly?
- Can a Jump contribute to multiple Groups' Seasons, or still at most one?
- What changes in the data model when a Jump no longer requires a Group?

### 5. Onboarding & Virality Mechanics
_Owner: Product/UX Design_

- Can a non-authenticated visitor view the public feed?
- What is the minimum action to post a first Jump?
- How does the app share Jumps externally (link previews, social share)?

### 6. MVP Scope Boundary
_Owner: MVP Roadmap_

- Which of Groups, Seasons, Missions, Bounties, and Levels are in MVP?
- What is the smallest set of features that proves the core loop?
