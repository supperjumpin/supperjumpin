---
name: work-issue
description: Coordinate one wave of ready-for-agent GitHub issue work using isolated worker worktrees, a local run ledger, structured worker reports, and coordinator-owned review/PR gates.
---

# Work Issue

Coordinate one wave of GitHub issue work.

This is an orchestrator-only skill. The coordinator does not implement issues directly. Workers implement exactly one issue each in isolated worktrees. The coordinator owns readiness checks, dispatch, run ledger state, GitHub issue/project state, review, PR creation, run summaries, and cleanup.

## Required configuration

Before dispatching workers, require:

```text
.work-issue/operator-config.yaml
```

If this file is missing or invalid, do not dispatch workers. Report the exact problem and tell the operator to run `setup-work-issue`.

Do not inspect or warn about `.gitignore` status. Assume setup handled local ignore hygiene.

## Source of truth

Use GitHub as the durable source of truth for:

- issues
- parent/sub-issue relationships
- blocker relationships
- labels
- project fields
- PR links
- comments
- merged state

Use the local run ledger only as ephemeral coordinator state:

```text
.work-issue/runs/<run-id>/
```

The coordinator is the only writer to the canonical run ledger.

## Target modes

Given a target issue:

- Tree mode: if the target has structured GitHub sub-issues, only structured sub-issues are dispatch candidates.
- Single-issue mode: if the target has no structured sub-issues, the target issue itself is the dispatch candidate.

In both modes, implementation still happens through a worker worktree and coordinator review gate.

## Readiness rule

Only work issues that are `ready-for-agent`.

A dispatch candidate is runnable only if it satisfies all of these:

- open issue
- exactly one canonical triage label
- triage label is `ready-for-agent`
- project `Status` is `Todo`
- no open structured blockers

Anything else is skipped in tree mode or stops the run in single-issue mode.

Do not route other labels to other skills in v1. Just report the reason and stop or skip.

## One wave only

Run exactly one wave per invocation.

Flow:

1. Validate config.
2. Run pre-dispatch sanity checks.
3. Discover target mode and issue graph.
4. Build the ready queue.
5. Dispatch up to `dispatch.max_concurrency` workers.
6. Review workers as they finish.
7. Allow one fix pass per issue per run if coordinator review fails.
8. Open PRs for accepted work if configured.
9. Post a concise run summary comment.
10. Offer explicit cleanup.
11. Stop.

Do not automatically start a second wave after PRs merge or blockers clear. The operator can invoke this skill again.

## Required runbooks

Load the runbooks as needed:

- `runbooks/discover-issue-graph.md`
- `runbooks/classify-ready-work.md`
- `runbooks/dispatch-workers.md`
- `runbooks/review-worker-output.md`
- `runbooks/complete-parent.md`
- `runbooks/cleanup.md`

Use templates from:

- `templates/worker-prompt.md`
- `templates/worker-report.md`
- `templates/ready-queue.md`
- `templates/coordinator-review.md`
- `templates/pull-request-body.md`
- `templates/parent-run-summary-comment.md`
- `templates/completion-summary.md`

## Hard guardrails

- Do not implement code in the coordinator checkout.
- Do not dispatch non-`ready-for-agent` issues.
- Do not dispatch issues with open blockers.
- Do not use issue-body parent links as dispatch membership.
- Do not use issue-body blocker text as authoritative blocker state.
- Do not let multiple workers write in the same checkout.
- Do not let workers update GitHub issue labels, project fields, blockers, or parent relationships.
- Do not push worker branches before coordinator review passes.
- Do not open PRs unless `permissions.coordinator_may_open_prs` is true.
- Do not merge PRs unless `permissions.coordinator_may_merge` is true and the operator explicitly delegated merge authority.
- Do not commit `.work-issue-worker-report.md`.
- Stop if configuration, dispatch instructions, GitHub state, or architecture/product requirements are contradictory or unsafe.

## Worker lifecycle

Every worker follows this lifecycle:

```text
created -> started -> working -> blocked | ready-for-review -> accepted | needs-fix | abandoned
```

Each issue gets one implementation attempt and one fix pass per run.

Workers must write a structured report at this fixed path in their own worktree:

```text
.work-issue-worker-report.md
```

The coordinator imports that report into:

```text
.work-issue/runs/<run-id>/reports/issue-<number>.md
```

## Branch and worktree naming

Use run-scoped worktrees:

```text
worktrees/<run-id>/issue-<number>/
```

Use branches that include the issue number and run slug:

```text
agent/issue-<number>-<run-slug>
```

If paths or branches already exist, stop and ask the operator to clean up or choose a new run.

## GitHub state ownership

The coordinator owns GitHub issue/project state.

On dispatch:

```text
Todo -> In Progress
```

If abandoned before PR creation:

```text
In Progress -> Todo
```

When a linked PR merges:

```text
In Progress -> Done
```

Let GitHub automation close issues and update done state when possible.

## PR ownership

Workers do not open PRs by default.

Default flow:

```text
worker ready-for-review -> coordinator review -> coordinator pushes branch -> coordinator opens PR
```

Coordinator-created PRs are not drafts by default.

Use the skill-local PR body template. Include `Closes #<issue-number>`.

## Completion

At the end of each run, post a concise run summary comment to:

- the parent issue in tree mode
- the target issue in single-issue mode

Include worked issues, PRs opened, skipped issues and reasons, blocked/abandoned issues, and next actions.

Do not paste the full local ledger into GitHub.
