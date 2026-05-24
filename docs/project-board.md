# Project Board

Supperjumpin work is tracked on the GitHub Project board:

https://github.com/orgs/supperjumpin/projects/1

## Fields

- **Status**: Kanban lane. Use `Todo`, `In Progress`, or `Done`.
- **Priority**: `P0`, `P1`, or `P2`.
- **Area**: `Product`, `Mobile`, `API`, `Data`, `Auth`, `Evidence`, `Contract`, `Infra`, or `Future`.
- **Size**: `XS`, `S`, `M`, `L`, or `XL`.
- **Work Type**: `PRD`, `Feature`, `Tech Decision`, `Chore`, or `Bug`.

## Recommended Views

GitHub does not currently expose project view creation through `gh`, so create these in the project UI.

- **Board**: board layout, grouped by `Status`. Primary day-to-day Kanban view.
- **Backlog**: table layout filtered to `Status:Todo`, showing `Priority`, `Area`, `Size`, `Assignees`, and `Labels`.
- **Ready for Agent**: table or board filtered to label `ready-for-agent` and not `Status:Done`.
- **By Area**: board layout grouped by `Area`, useful for seeing Mobile/API/Data/Auth balance.
- **My Work**: table filtered to current assignee and not `Status:Done`.
- **Decisions and Chores**: table filtered to `Work Type:Tech Decision` or `Work Type:Chore`.

## Workflows

The project has the default GitHub workflows enabled, including auto-add, auto-add sub-issues, item closed, pull request merged, pull request linked to issue, and auto-close issue. Keep these enabled unless they start creating noise.

When creating issues from PRDs, add them to this project and set all custom fields before implementation starts: `Status`, `Priority`, `Area`, `Size`, and `Work Type`. Sub-issues may auto-add through the project workflow, but still verify field values after creation.

Default new implementation issues to:

- **Assignee**: current GitHub user (`@me`) unless the user specifies otherwise.
- **Status**: `Todo`.
- **Priority**: `P1`, unless the issue is sequence-changing or blocks multiple workstreams (`P0`) or is low urgency (`P2`).
- **Area**: the primary area touched by the vertical slice. Pick one even when the slice crosses layers.
- **Size**: estimate implementation size as `XS`, `S`, `M`, `L`, or `XL`.
- **Work Type**: `Feature` for PRD implementation slices, `PRD` for parent product specs, `Tech Decision` for architecture decisions, `Chore` for maintenance, and `Bug` for regressions.

Use GitHub's relationship fields in addition to project fields:

- **Parent issue** is populated when implementation slices are attached as sub-issues of a PRD.
- **Sub-issues progress** is populated automatically on the parent PRD from its sub-issues.
- **Blocked by / blocking** should be set for dependency ordering, not only described in the issue body.

## PRD Breakdown

Use GitHub sub-issues for implementation slices that directly deliver a parent PRD. The parent PRD issue should track product intent and roll up progress; sub-issues should be vertical, independently mergeable slices rather than tiny task checklist items.

Use normal first-order issues instead of sub-issues for cross-cutting tech decisions, infrastructure chores, bugs, or future ideas that may support multiple PRDs.

For each PRD breakdown:

- Create sub-issues in dependency order.
- Attach each created issue to the PRD as a GitHub sub-issue.
- Add explicit blocker relationships between sub-issues when one cannot start before another lands.
- Assign each issue to the current operator by default.
- Set every custom project field before reporting completion.
