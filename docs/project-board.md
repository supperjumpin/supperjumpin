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

When creating issues from PRDs, add them to this project and set `Priority`, `Area`, `Size`, and `Work Type` before implementation starts.
