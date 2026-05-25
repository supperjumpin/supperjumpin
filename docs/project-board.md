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

**Size** — Implementation effort, AI-calibrated:
- `XS` — Quick fix, single file. ~15–30 min.
- `S` — Self-contained feature. ~1–2 hours.
- `M` — Crosses a couple layers. ~3–5 hours.
- `L` — Would take a whole focused day.
- `XL` — Needs decomposition. Run the PRD → issues breakdown flow below.

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

- **Assignee**: current GitHub user (`@me`) unless the user specifies otherwise.
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
- Assign each issue to the current operator by default.
- Set every custom project field before reporting completion.
