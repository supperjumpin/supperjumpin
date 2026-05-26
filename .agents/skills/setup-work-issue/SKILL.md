---
name: setup-work-issue
description: Configure local runtime settings for the work-issue skill. Use before work-issue when .work-issue/operator-config.yaml is missing, invalid, or needs to change.
---

# Setup Work Issue

Configure local, gitignored operator settings for the `work-issue` orchestration skill.

This skill is setup-only. It does not dispatch workers, modify implementation files, or work GitHub issues.

## Responsibilities

Use this skill to:

- Explain the supported dispatch modes.
- Interview the operator about their runtime setup.
- Create or update `.work-issue/operator-config.yaml`.
- Validate the operator config.
- Optionally update `.gitignore` to ignore `.work-issue/`.

Do not silently choose runtime, provider, model, or merge settings. Recommend values when useful, but ask the operator before writing config.

## Required setup output

The setup output is:

```text
.work-issue/operator-config.yaml
```

This file is local operator configuration and should not be committed.

Use `templates/operator-config.example.yaml` as the starting point.

## Dispatch modes

The operator must choose exactly one dispatch mode:

- `manual_sessions`: The coordinator creates worktrees and prompt files. The operator starts each worker session manually.
- `in_session_subagents`: The current runtime can start worker subagents inside the coordinator session.
- `background_sessions`: The runtime can launch and monitor independent background worker sessions.
- `non_interactive_batch`: The runtime can execute a worker prompt non-interactively and exit.
- `external_service`: An external service, queue, CI job, or custom orchestrator accepts worker tasks and reports results.

The selected mode must provide `dispatch.instructions`. Treat these instructions as the runtime contract. If the instructions are unsafe, contradictory, missing required details, or impossible in the current environment, stop and ask the operator to fix the config.

## Setup flow

1. If `.work-issue/operator-config.yaml` exists, read and validate it.
2. If it is valid, summarize the current settings and ask whether the operator wants changes.
3. If it is missing or invalid, ask for setup details.
4. Ask for dispatch mode first.
5. Ask only the follow-up questions needed for the selected mode.
6. Ask for `max_concurrency`; default to `2`; allowed range is `1` through `3`.
7. Ask whether the selected runtime supports worker model selection.
8. If model selection is supported, ask for `models.worker`.
9. Ask PR and merge permissions.
10. Show the final YAML.
11. Write `.work-issue/operator-config.yaml` only after operator approval.
12. Check whether `.work-issue/` is ignored by git.
13. If not ignored, offer to patch `.gitignore`.
14. Patch `.gitignore` only after operator approval.

## Safe defaults

Recommend these defaults unless the operator says otherwise:

```yaml
dispatch:
  max_concurrency: 2

permissions:
  coordinator_may_open_prs: true
  coordinator_may_merge: false
```

Do not invent:

- dispatch mode
- runtime-specific instructions
- model/provider names
- merge authority

## Validation rules

A valid config must satisfy:

- `dispatch.mode` is one of the supported dispatch modes.
- `dispatch.max_concurrency` is `1`, `2`, or `3`.
- `dispatch.instructions` is non-empty.
- `dispatch.supports_model_selection` is boolean.
- If `dispatch.supports_model_selection` is `true`, `models.worker` is required and non-empty.
- `permissions.coordinator_may_open_prs` is boolean.
- `permissions.coordinator_may_merge` is boolean.

## Existing config behavior

If `.work-issue/operator-config.yaml` already exists:

- Do not overwrite it.
- Read it.
- Validate it.
- Summarize the current settings.
- Identify missing or invalid fields.
- Propose targeted edits.
- Apply edits only after operator approval.

## Gitignore behavior

If `.work-issue/` is not ignored, propose adding:

```gitignore
# Local work-issue orchestration state
.work-issue/
```

Use the repo's existing local/agent section if one exists.

Do not require this patch for setup success unless the operator wants it.
