## Agent skills

### Issue tracker

Issues and PRDs are tracked in GitHub Issues for `supperjumpin/supperjumpin`. See `docs/agents/issue-tracker.md`.

### Triage labels

Use the default five-label triage vocabulary. See `docs/agents/triage-labels.md`.

### Domain docs

This is a single product context monorepo: read root `CONTEXT.md` and root `docs/adr/` when present. See `docs/agents/domain.md`.

### Worktrees

Create issue worktrees under `worktrees/issue-<number>` inside this repository checkout. Do not create sibling worktrees outside the repo unless the user explicitly asks, because external paths may trigger OpenCode permission prompts.
