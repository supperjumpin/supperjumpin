---
name: work-issue
description: Work issue trees in parallel from a parent GitHub issue and its sub-issues. Use when the user says "work issue" or wants agents to execute a PRD, issue tree, or blocked/unblocked issue graph until acceptance criteria are met.
---

# Work Issue

Coordinate a parent PRD through dependency-ordered implementation waves. The coordinator is the user's interface; workers each own one issue in one isolated git worktree.

## Guardrails

- Use GitHub as source of truth for parent/sub-issue relationships, blocker relationships, labels, assignees, PR links, and project fields.
- Start only issues labeled `ready-for-agent` with `Status: Todo` and no open blockers.
- Stop when there are no unblocked `ready-for-agent` issues. Do not work `ready-for-human`, `needs-triage`, `needs-info`, `wontfix`, or unlabeled issues.
- Use one git worktree and one branch per worker. Never let multiple workers write in the same checkout.
- Workers implement exactly one issue and must not broaden scope beyond that issue's acceptance criteria.
- Workers must load and follow the `tdd` skill for red-green-refactor implementation.
- Do not merge PRs without explicit user approval unless the user delegated merge authority.
- Stop if a worker discovers an unresolved architecture, product, or dependency conflict.

## Quick Start

When invoked, ask for any missing run controls:

- Parent issue number or URL.
- Maximum concurrent workers. Default: 3.
- Whether workers may open PRs. Default: yes.
- Whether the coordinator may merge approved PRs. Default: no.

Then inspect `docs/agents/issue-tracker.md`, `docs/project-board.md`, `docs/agents/triage-labels.md`, `docs/agents/domain.md`, `CONTEXT.md`, and relevant ADRs in `docs/adr/`.

## Discover The Issue Graph

Fetch parent, sub-issues, blockers, assignees, labels, and project data.

```bash
gh issue view <parent> --comments --json number,title,body,labels,assignees,projectItems,url
gh api repos/supperjumpin/supperjumpin/issues/<parent>/sub_issues
gh api graphql -f query='query($owner:String!, $repo:String!, $number:Int!) { repository(owner:$owner, name:$repo) { issue(number:$number) { number title subIssues(first:100) { nodes { number title state labels(first:20) { nodes { name } } assignees(first:20) { nodes { login } } blockedBy(first:20) { nodes { number state } } projectItems(first:20) { nodes { id project { title } fieldValues(first:50) { nodes { ... on ProjectV2ItemFieldSingleSelectValue { field { ... on ProjectV2SingleSelectField { name } } name } } } } } } } } } }' -f owner=supperjumpin -f repo=supperjumpin -F number=<parent>
```

A ready issue satisfies all of these:

- It is a sub-issue of the parent PRD.
- It is open.
- It has `ready-for-agent`.
- Its project `Status` is `Todo`.
- Every issue in `blockedBy` is closed or has a merged linked PR that closes it.

Label routing:

- `ready-for-agent`: eligible for worker waves when unblocked.
- `ready-for-human`: report in the human lane and stop if no agent-ready work remains.
- `needs-triage`: report as needing the `triage` skill before implementation.
- `needs-info`: report as blocked on external or user input.
- `wontfix`: exclude from worker waves and summarize only if relevant.
- Missing triage label: report as a tracker hygiene issue; do not start a worker.

## Run A Wave

For each ready issue up to the concurrency limit:

```bash
git fetch origin
git worktree add worktrees/issue-<number> -b agent/issue-<number> origin/main
```

Keep worktrees under `worktrees/` inside the current repository checkout so OpenCode can access them without external-directory permission prompts during AFK runs.

If there are zero ready issues, do not launch workers. Report why the remaining open sub-issues are not runnable: blocked by open issues, `ready-for-human`, `needs-triage`, `needs-info`, missing triage label, or already in progress.

Launch one worker per worktree. If the runtime has a Task/subagent tool, use it; otherwise ask the user to start separate OpenCode sessions in each worktree. Give each worker this prompt, replacing placeholders:

```md
You are implementing GitHub issue #<number> in supperjumpin/supperjumpin.

First, load and follow the `tdd` skill. Use red-green-refactor with vertical tracer bullets.

Read issue #<number>, parent issue #<parent>, CONTEXT.md, relevant ADRs, and docs/agents guidance.

Work only in this git worktree: <worktree-path>.

Rules:
- Implement only issue #<number>.
- Preserve Supperjumpin glossary terms.
- Satisfy every acceptance criterion.
- Add or update behavior tests through public interfaces.
- Run relevant checks.
- Open a PR from branch agent/issue-<number> linked to issue #<number>.
- In the PR body, map changes to acceptance criteria and list test results.
- Update project status to In Progress while working and report if blocked.
- Stop and report if the issue needs a product or architecture decision.
```

## Between Waves

- Monitor worker results and PR links.
- Review each PR for scope, tests, acceptance criteria coverage, and glossary/ADR alignment.
- Ask the user before merging unless merge authority was delegated.
- After a PR merges, update the issue/project status if automation did not.
- Rebase or recreate downstream worktrees on latest `origin/main` before starting dependent work.
- Recompute the ready queue from GitHub relationships, then start the next wave.

## Completion

The parent PRD is complete only when:

- Every sub-issue is closed.
- The parent's sub-issues progress is complete.
- No sub-issue is blocked by an open issue.
- Project fields are set on all issues.
- The coordinator can summarize which acceptance criteria were satisfied by which PRs.
