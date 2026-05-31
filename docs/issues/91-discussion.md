# Discussion Record: Issue #91
**Date:** 2026-05-30
**Status:** resolved
**Thread:** 1510036274944151552

## 🎯 Executive Summary
Implementation of a full-stack "canon rename" replacing the term **Stunt** with **Jump** and **Documentation** with **Presentation** across the entire `supperjumpin` codebase to align with ADR-0020.

## ✅ Decisions & Agreements
- **Decision:** Adopt "Hard Pivot" (Direct Migration) instead of a phased rollout $\rightarrow$ **Reasoning:** Project is in early local development with empty tables; risk of downtime is non-existent.
- **Decision:** Execute changes in architectural chunks (DB $\rightarrow$ Backend $\rightarrow$ API $\rightarrow$ Frontend $\rightarrow$ Docs) $\rightarrow$ **Reasoning:** Ensures logical cohesion and prevents massive systemic breakage by updating dependencies in order.
- **Decision:** Implement a flat SQL migration for schema changes $\rightarrow$ **Reasoning:** Direct table/column renames in Postgres are efficient for early-stage prototypes.

## 🚧 Open items / Future Work
- [ ] Verify full mobile app flow with the new terminology in a live build.

## 📚 Context & References
- **Key Files:** 
    - `/home/bigocb/supperjumpin/apps/api/db/migrations/0012_rename_stunts_to_jumps.sql` (The primary schema migration)
    - `/home/bigocb/supperjumpin/apps/api/openapi.yaml` (Updated API contract)
    - `/home/bigocb/supperjumpin/packages/api-client/src/generated.d.ts` (Regenerated TS types)
- **Notes:** Discovered that Postgres inline CHECK constraints must be dropped and recreated by name during renames; captured the necessary constraint names from `0006_evidence.sql`.

---
*Archived via Hermes discussion-closer skill*
