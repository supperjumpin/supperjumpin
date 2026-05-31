# Discussion Record: 86
**Date:** 2026-05-31
**Status:** closed
**Thread:** Discord / B & B's Workshop / supperjumpin / Issue #86 BDD tests

## 🎯 Executive Summary

Recovered from a corrupted terminology sweep (Stunt→Jump, Difficulty→Commitment) that left stale SQL column references and missing DB migrations, then added BDD test coverage. PR #117 closed then reopened for review.

## ✅ Decisions & Agreements

- **Decision:** Replicate migrations 0012–0015 on fresh DB vs. debug each failure in-place → **Reasoning:** Faster iteration; each failure surfaced a distinct bug
- **Decision:** Use `ON DELETE SET NULL` on `evidences.upload_authorization_id` FK vs. CASCADE or reordering → **Reasoning:** Evidences should retain which auth was used for audit; SET NULL keeps the reference without blocking auth cleanup
- **Decision:** Add `IF NOT EXISTS` to 0014 season deadline columns → **Reasoning:** Migration may re-run in environments that already have `0004_seasons_deadlines.sql` on main
- **Decision:** Restore `0004_seasons_deadlines.sql` (was deleted from branch diff to avoid conflict with 0014) → **Reasoning:** Migration 0004 exists on main; it was inadvertently removed when 0014 was added. Verified working copy restored locally but not committed (would cause duplicate on main merge)
- **Decision:** Close then reopen PR #117 → **User directive** (context unclear; PR re-opened for further review)

## 🚧 Open Items / Future Work

- [ ] PR #117 needs reviewer sign-off before merge (diff includes terminology sweep commits from d326979 onward + unrelated churn across mobile/api-client/docs)
- [ ] Remove debug `log.Printf` statements from `server.go` after PR merges (SubmitEvidence, SubmitJudgment, GroupHome error paths)
- [ ] Consider consolidating 0012/0013 into a single migration for cleaner rollback story
- [ ] `conclave-roan.vercel.app` dashboard nav "click does nothing" bug still open (root cause: API calls 404 on wrong domain)

## 📚 Context & References

**PRs:**
- https://github.com/supperjumpin/supperjumpin/pull/117 (open)

**Key Files:**
- `apps/api/internal/httpapi/postgres_store.go` — 5 stale SQL fixes (stunt_id→jump_id, stunts.*→jumps.*)
- `apps/api/internal/httpapi/server.go` — Debug logging on SubmitEvidence/SubmitJudgment/GroupHome
- `apps/api/db/migrations/0012_rename_stunts_to_jumps.sql` — RENAME TABLE stunts→jumps
- `apps/api/db/migrations/0013_terminology_sweep.sql` — Constraints, disputes.stunt_id fix, 'Judged Stunt'→'Judged Jump'
- `apps/api/db/migrations/0014_seasons_deadlines.sql` — Adds submission_deadline/judging_deadline to seasons
- `apps/api/db/migrations/0015_evidence_auth_on_delete_set_null.sql` — FK ON DELETE SET NULL fix
- `apps/api/internal/bdd/jumps.feature` — 5 scenarios
- `apps/api/internal/bdd/judging.feature` — 2 scenarios
- `apps/api/internal/bdd/bdd_test.go` — Full step definitions

**Root causes discovered:**
- 5 SQL column refs missed by original sweep (postgres_store.go lines 819, 1424, 1482, 1486, 1553)
- Migration 0012 forgot to rename `disputes.stunt_id` column
- `seasons` table missing `submission_deadline`/`judging_deadline` columns (code expected them, DB didn't have them)
- FK constraint `evidences_upload_authorization_id_fkey` defaulting to NO ACTION blocked DELETE of consumed upload auth within same transaction

**Test results (verified against local Postgres):**
- Unit tests: 46/46 ✅
- BDD jumps.feature: 5 scenarios ✅
- BDD judging.feature: 2 scenarios ✅

---
*Archived via Hermes discussion-closer skill*