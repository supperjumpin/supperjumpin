## Agent skills

### Issue tracker

Issues and PRDs are tracked in GitHub Issues for `supperjumpin/supperjumpin`. See `docs/agents/issue-tracker.md`.

### Triage labels

Use the default five-label triage vocabulary. See `docs/agents/triage-labels.md`.

### Domain docs

This is a single product context monorepo: read root `CONTEXT.md` and root `docs/adr/` when present. See `docs/agents/domain.md`.

### Pre-MVP compatibility stance

This project is pre-MVP. Existing code has no compatibility guarantees — it is safe to delete or rewrite anything that does not serve the current goal. Do not preserve code for backward compatibility unless doing so costs less than deleting it. Once the system matures past MVP, compatibility promises will tighten; until then, treat the codebase as disposable.

**This does not mean lowering the bar on main.** `main` must always build and pass tests. The flexibility is about not carrying dead code or preserving interfaces for hypothetical consumers; it is not about shipping broken code.

### Worktrees

Create issue worktrees under `worktrees/issue-<number>` inside this repository checkout. Do not create sibling worktrees outside the repo unless the user explicitly asks, because external paths may trigger OpenCode permission prompts.
