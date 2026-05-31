# Discussion Record: Issue #48 — CI: verify generated API client sync
**Date:** 2026-05-26
**Status:** Closed
**Thread:** Discord #48 thread (ID: 1508627348642201812)

## 🎯 Executive Summary
Discussed and implemented CI enforcement for OpenAPI client sync. Discovered that the generated `openapi-typescript` types (nested under `components.schemas.*`) broke the mobile app's existing imports, exposing a TypeScript type error at `App.tsx:342`. Solved with a two-PR split: #52 (type fix) then #51 (CI infrastructure).

## ✅ Decisions & Agreements
- **Decision:** Separate the type fix (PR #52) from the CI infrastructure (PR #51) → **Reasoning:** Keeps chores and bug fixes cleanly isolated; PR #51 can focus on the CI check without mixing concerns.
- **Decision:** Hard-failure CI check (build breaks if generated client diverges from `openapi.yaml`) → **Reasoning:** Silent drift is worse than a one-time fixup. User prefers explicit feedback over auto-commit workflows.
- **Decision:** Rebase PR #51 onto main after PR #52 merged (force push) → **Reasoning:** Resolves merge conflicts and picks up the type fix.
- **Decision:** Delete dead `scripts/generate.mjs` (package.json now calls `openapi-typescript` CLI directly) → **Reasoning:** Review feedback from bturney; keeping both is a maintenance risk.
- **Decision:** Wait for bturney's approval to merge (respecting CHANGES_REQUESTED) → **Reasoning:** He requested changes, proper to let him verify before merging.
- **Decision:** Bump `package-lock.json` and commit regenerated types as part of the fix → **Reasoning:** Ensures lockfile is in sync with new dep tree.

## 🚧 Open items / Future Work
- [ ] The orphaned path `supperjumpin/packages/api-client/scripts/generate.mjs` was cleaned from the PR branch but the nested `supperjumpin/` directory structure may still exist on disk. Worth a one-time cleanup sweep.
- [ ] Monitor for similar type breakage when regenerating API types — the `openapi-typescript` nested schema format differs from the old flat type exports.

## 📚 Context & References
- **Issue:** #48 — CI: verify generated API client is in sync with OpenAPI spec
- **PRs:** [#51](https://github.com/supperjumpin/supperjumpin/pull/51) (CI sync — merged), [#52](https://github.com/supperjumpin/supperjumpin/pull/52) (type fix — merged)
- **Key Files:**
  - `.github/workflows/ci.yml` — added `Verify API client sync` step
  - `packages/api-client/package.json` — added `openapi-typescript` dep
  - `packages/api-client/src/generated.d.ts` — regenerated types
  - `apps/mobile/App.tsx` — added `: PerformedStuntView` type annotation
  - `MEMORY.md` — fixed corruption (ScrollView truncation, bold formatting)
- **Reviewers:** bturney (MEMBER) — flagged 4 issues, all resolved; final review: APPROVED
- **Notes:** The TypeScript type error TS7006 was exposed by the new type generator nesting types differently. PR #52 fixed App.tsx. The branch had a `revert unrelated changes` commit that accidentally undid the fix and had to be squashed out during rebase.

---
*Archived via Hermes discussion-closer skill*
