# Issue tracker: GitHub

Issues and PRDs for this repo live as GitHub issues. Use the `gh` CLI for all operations.

## Conventions

- **Create an issue**: `gh issue create --title "..." --body "..."`. Use a heredoc for multi-line bodies.
- **Read an issue**: `gh issue view <number> --comments`, filtering comments by `jq` and also fetching labels.
- **List issues**: `gh issue list --state open --json number,title,body,labels,comments --jq '[.[] | {number, title, body, labels: [.labels[].name], comments: [.comments[].body]}]'` with appropriate `--label` and `--state` filters.
- **Comment on an issue**: `gh issue comment <number> --body "..."`
- **Apply / remove labels**: `gh issue edit <number> --add-label "..."` / `--remove-label "..."`
- **Close**: `gh issue close <number> --comment "..."`

Infer the repo from `git remote -v` - `gh` does this automatically when run inside a clone.

## When a skill says "publish to the issue tracker"

Create a GitHub issue in `supperjumpin/supperjumpin`.

Add implementation issues and PRDs to the Supperjumpin GitHub Project when practical: https://github.com/orgs/supperjumpin/projects/1. Project field conventions are documented in `docs/project-board.md`.

When breaking down a PRD, prefer GitHub sub-issues under the PRD issue for implementation slices that directly deliver that PRD. Use normal first-order issues for cross-cutting tech decisions, infrastructure chores, bugs, or future ideas that may support multiple PRDs.

## When a skill says "fetch the relevant ticket"

Run `gh issue view <number> --comments`.
