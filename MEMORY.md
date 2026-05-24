# Project Memory

## ? PRD #1: First Playable Group Stunt Loop
**Status**: OPEN | **Author**: Ben Turney

Single-player-per-group social game where players perform absurd food-location stunts, share evidence, judge each other's stunts on 4 axes (Difficulty, Transgression, Creativity, Documentation), and compete for season standings. Uses Supabase Auth + Postgres + Expo mobile app + Go REST API. MVP excludes progression systems, public feed, auto-scoring, video evidence, push notifications.

---
### Key Mechanisms
- Groups multi-member with admin/commissioner roles
- Season phases: submission -> judging grace period -> finalized
- Stunt lifecycle: Idea -> Planned -> Performed (requires evidence + caption)
- Judging: one judgment per judge, editable during window, locked after close
- Standings based on season-linked judged stunts only; off-season excluded

---
## ? Current Focus
- **Objective**: Implement Peer Judging Loop (T-A #19)
- **Active Issue**: #19 
- **Status**: In-progress (Store logic complete, API Client generation blocked)

## ⏳ Activity Timeline
- 2026-05-24 [Opencode]: Implemented `judgments` schema and Store logic (Postgres/Memory) -> Core engine for judging is ready.
- 2026-05-24 [Opencode]: Added judging endpoints to `openapi.yaml` and `server.go` -> API surface is defined.
- 2026-05-24 [Opencode]: Attempted API client regeneration -> Blocked by syntax error/strict checks in `generate.mjs`.

## 🏗️ Architecture & Decisions
- **Dual-Track Approach**: Track A (Engine/Foundation) and Track B (User Experience) are running in parallel. Track B depends on Track A's underlying logic.
- **Identity**: Many-to-One mapping (Auth -> Account -> Player) implemented in `PostgresStore`.
- **Stunt Lifecycle**: Idea -> Planned -> Performed (gated by Evidence).
- **Judging Logic**: Implemented authoritative guards (must be a group member, cannot judge own stunt, stunt must be 'Performed'). Scoring uses an upsert model (one judgment per player per stunt).

## ⚠️ Hurdles & Gotchas
- **API Client Generator**: The `packages/api-client/scripts/generate.mjs` script has a strict whitelist of `operationId`s. Adding new endpoints requires updating this list manually.
- **Concurrency**: Solved via Postgres `ON CONFLICT` for judgment updates.

## 💡 Working Hypotheses
- **Tracer Bullet**: By implementing a basic scoring API first, we can verify the end-to-end loop before adding the complex gesture-driven UI of #20.

## 📡 Session Wrap-up & Hand-off
- **Completed in this session**:
    - `judgments` table migration.
    - Store interface updates and implementations.
    - HTTP handlers for judging.
    - Updated `openapi.yaml`.
- **Next Steps**: Fix the `generate.mjs` script to allow the new judging operations and successfully regenerate the client.
- **Warning**: Check for exact matches of `operationId` in `openapi.yaml` vs `generate.mjs` to avoid the current "not found" error.
