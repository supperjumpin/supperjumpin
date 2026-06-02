# Project Board

Supperjumpin work is tracked on the GitHub Project board:

https://github.com/orgs/supperjumpin/projects/1

## Fields

**Status** — Kanban lane. Keep accurate through the issue lifecycle:
- `Todo` (default when added to the project)
- `In Progress` (set when actively working)
- `Done` (moved automatically if PR references issue number)

**Priority** — Relative importance, set during triage (leave unset until then):
- `P0` — Sequence-changing blocker. Blocks multiple workstreams or a critical path. Late P0s cause other work to slip or ship broken. Requires explicit team awareness.
- `P1` — Active delivery. WIP-able in the current or next cycle without triage escalation.
- `P2` — Nice-to-have. No downstream dependency. May sit without causing problems.

**Area** — Layer of the stack the issue primarily touches (pick one):
- `App` — React Native / Expo frontend
- `API` — Go backend, REST handlers
- `Data` — DB schema, sqlc queries, migrations
- `Auth` — Authentication, account management
- `Infra` — Deployment, CI/CD, hosting, Docker

**Size** — Implementation effort, AI-calibrated. Sizes also determine decomposition policy (see [Decomposition Policy](#decomposition-policy) below):
- `XS` — One small code/doc change, one PR. ~15–30 min.
- `S` — One vertical implementation issue, agent-ready. ~1–2 hours.
- `M` — Crosses a couple layers. ~3–5 hours. **Allowed only if it is one coherent vertical slice with a single acceptance path.** Otherwise split.
- `L` — Would take a whole focused day. **Always split before implementation.**
- `XL` — Needs decomposition. Run the PRD → issues breakdown flow below. **Never assigned directly.**

## Title Prefixes

Issue titles use a prefix to signal the type of work (replaces a dedicated Work Type field):

- `PRD:` — Parent product spec
- `Spike:` — Investigation or tech decision
- `Docs:` — Documentation
- `Bug:` — Regression or defect
- *(no prefix)* — Default feature implementation

## Workflows

The project has the default GitHub workflows enabled, including auto-add, auto-add sub-issues, item closed, pull request merged, pull request linked to issue, and auto-close issue. Keep these enabled unless they start creating noise.

When creating issues from PRDs, add them to this project and set all custom fields before implementation starts: `Status`, `Priority`, `Area`, and `Size`. Sub-issues may auto-add through the project workflow, but still verify field values after creation.

Default new implementation issues to:

- **Status**: `Todo`.
- **Priority**: unset (determined during triage).
- **Area**: the primary layer touched by the slice. Pick one even when the slice crosses layers.
- **Size**: estimate implementation size per the definitions above.

Use GitHub's relationship fields in addition to project fields:

- **Parent issue** is populated when implementation slices are attached as sub-issues of a PRD.
- **Sub-issues progress** is populated automatically on the parent PRD from its sub-issues.
- **Blocked by / blocking** should be set for dependency ordering, not only described in the issue body.

## Milestones

Milestones represent release scopes (MVP, v1.1, etc.), not individual PRDs. A milestone may contain multiple PRDs and their sub-issues.

When a PRD is scoped to a release, assign the milestone to the PRD issue and to every sub-issue under it. Milestone assignment does not cascade automatically — set it explicitly on each sub-issue when breaking down a PRD. The milestone progress bar is the primary signal for whether a release is ready to ship.

Create milestones in the GitHub UI or with `gh api repos/supperjumpin/supperjumpin/milestones --method POST -f title="..."`.

## PRD Breakdown

Use GitHub sub-issues for implementation slices that directly deliver a parent PRD. The parent PRD issue should track product intent and roll up progress; sub-issues should be vertical, independently mergeable slices rather than tiny task checklist items.

Use normal first-order issues instead of sub-issues for cross-cutting tech decisions, infrastructure chores, bugs, or future ideas that may support multiple PRDs.

For each PRD breakdown:

- Create sub-issues in dependency order.
- Attach each created issue to the PRD as a GitHub sub-issue.
- Add explicit blocker relationships between sub-issues when one cannot start before another lands.
- Set every custom project field before reporting completion.

## Decomposition Policy

Sizes determine when an issue must be decomposed into smaller sub-issues before implementation. The rules are:

| Size | Decomposition requirement | `ready-for-agent` eligible? |
|------|--------------------------|-----------------------------|
| XS   | None — one PR, one issue | Yes |
| S    | None — one vertical slice, already agent-ready | Yes |
| M    | **Allowed only if one coherent vertical slice with a single acceptance path.** If the issue touches multiple subsystems, has independent sub-tasks, or has more than one acceptance path, split it. | Yes, if it passes the coherence test |
| L    | **Always split before implementation.** Break into XS/S sub-issues with explicit dependency ordering. | No — decompose first |
| XL   | **Never assigned directly.** Must flow through the PRD breakdown workflow. | No — decompose first |

**Guidance for agents breaking down L/XL issues:**

1. Identify the thin, independently testable vertical slices — each should produce a meaningful increment even if later slices are delayed.
2. Order slices by dependency: foundational changes first (schema, core logic), then progressive feature layers.
3. Each sub-issue must be XS or S by the definitions above. If a sub-issue comes out M, split it again.
4. Set blocker relationships (`Blocked by`) between sub-issues so the board shows the dependency chain.
5. Only apply `ready-for-agent` to sub-issues that individually pass the size test. Decompose the parent issue; assign and label the children.
