# Project Memory

## 🟢 Current Focus
- **Objective**: Resolve Evidence-Gated Execution (#18) - Ensure mobile integration matches backend gating.
- **Active Issue**: #18
- **Status**: Backend verified as complete; Mobile frontend integration missing.
- **Agent ID**: hermes-agent
- **Last Updated**: 2026-05-26T03:15:00Z

## ⏳ Activity Timeline (Current Session)
- 2026-05-26T02:50:00Z [hermes-agent]: Implemented hard-failure API client sync verification (#48) -> Replaced fake generator with `openapi-typescript` and added CI check to fail on stale types.
- 2026-05-25T02:15:00Z [prn_dev]: Implemented seasonal boundaries (#21) -> Submission window enforcement, auto-transition logic, deadline-based API.
- 2026-05-25T00:55:00Z [prn_dev]: Added accessibility labels to all score adjustment buttons -> Screen reader compatible.
- 2026-05-25T00:54:00Z [prn_dev]: Updated memory protocol with security warning and append-only rule -> Prevents secret leaks and merge conflicts.
- 2026-05-25T00:50:00Z [prn_dev]: Rebased feature branch on main -> Incorporated Ben's PR #32 (Judgment API) and PR #31 (Group home views).

## ⚠️ Active Hurdles
- **Gesture Sensitivity**: 50px threshold for score changes may need tuning based on user feedback.
- **ScrollView Conflict**: PanResponder may conflict with parent ScrollView; monitoring for issues.
- **PR #35 Pending Review**: Waiting for feedback from @bturney and @codex.

---

## 🏗️ Architecture & Decisions
- **Dual-Track Approach**: Track A (Engine/Foundation) and Track B (User Experience) are running in parallel. Track B depends on Track A's underlying logic.
- **Identity**: Many-to-One mapping (Auth → Account → Player) implemented in `PostgresStore`.
- **Stunt Lifecycle**: Idea → Planned → Performed (gated by Evidence).
- **Judging Logic**: Implemented authoritative guards (must be a group member, cannot judge own stunt, stunt must be 'Performed'). Scoring uses an upsert model (one judgment per player per stunt).
- **Temporal Logic**: 
  - Season status auto-transitions: Active → Judging Grace Period → Finalized (based on deadline comparison)
  - Submission window open: Season status == \"Active\" AND now < submissionDeadline
  - Judging window open: Season status IN (\"Active\", \"Judging Grace Period\")
  - Evidence submission rejected after submission deadline (ErrSubmissionWindowClosed → 409)
- **Season Deadlines**: ISO 8601 timestamps stored in DB; enforced at API boundary and store layer.
- **Gesture Layer**: Implemented with PanResponder; vertical swipes adjust scores, API call only on explicit 'Submit' (per PRD #15).
- **Accessibility**: All score adjustment buttons have explicit labels for screen readers.
- **API Client Sync**: Strictly enforced via CI using `openapi-typescript`. Any divergence between `openapi.yaml` and `packages/api-client/src` results in a hard build failure to prevent runtime type mismatches.

---

## 💡 Working Hypotheses
- **Tracer Bullet**: By implementing a basic scoring UI first, we can verify the end-to-end loop before adding complex swipe gestures.
- **Gesture Layer**: Swipe gestures modify local frontend state only; no API calls occur until the explicit 'Submit' action (per PRD #15).

---

## 📜 Activity History (Archived Sessions)
- 2026-05-24 [Opencode]: Implemented PanResponder for swipe gestures -> True gesture-based scoring complete.
- 2026-05-24 [Opencode]: Added visual styling for gesture rows -> Improved touch targets and feedback.
- 2026-05-24 [Opencode]: Button-based MVP complete -> Ready for gesture enhancement.
- 2026-05-24 [Opencode]: Researched opencode MCP configuration docs -> Correct format identified (command as array, type: local).
- 2026-05-24 [Opencode]: Added gesture score state and judging functions to App.tsx -> Core interaction logic ready.
- 2026-05-24 [Opencode]: Implemented judging UI with score adjustment controls -> MVP for gesture layer complete.
- 2026-05-24 [Opencode]: Added styles for judging card and score rows -> Visual distinction for judging mode.
- 2026-05-24 [Opencode]: Reviewed Ben's PR #32 -> Confirmed temporal window logic is implemented.
- 2026-05-24 [Opencode]: Added review comment to PR #32 -> Aligned on Dual-Track approach and API stability.
- 2026-05-24 [Opencode]: Implemented `judgments` schema and Store logic (Postgres/Memory) -> Core engine for judging is ready.
- 2026-05-24 [Opencode]: Added judging endpoints to `openapi.yaml` and `server.go` -> API surface is defined.
- 2026-05-24 [Opencode]: Fixed API client generator and regenerated types -> Client now supports judging endpoints.
- 2026-05-24 [Opencode]: Opened PR #30 for judging engine -> Merged into main.
- 2026-05-24 [Opencode]: Updated agent memory protocol with archival flow -> Ensures historical continuity.

---

## 📋 PRD Context (Reference)
### PRD #1: First Playable Group Stunt Loop
**Status**: OPEN | **Author**: Ben Turney

Single-player-per-group social game where players perform absurd food-location stunts, share evidence, judge each other's stunts on 4 axes (Difficulty, Transgression, Creativity, Documentation), and compete for season standings. Uses Supabase Auth + Postgres + Expo mobile app + Go REST API. MVP excludes progression systems, public feed, auto-scoring, video evidence, push notifications.

**Key Mechanisms:**
- Groups multi-member with admin/commissioner roles
- Season phases: submission → judging grace period → finalized
- Stunt lifecycle: Idea → Planned → Performed (requires evidence + caption)
- Judging: one judgment per judge, editable during window, locked after close
- Standings based on season-linked judged stunts only; off-season excluded

---

## 📡 Past Hand-off Notes (Archived)
- **Next Steps**: 
    - Test end-to-end: Performed Stunt → Judge → Submit → Verify in DB.
    - Tune gesture sensitivity (currently 50px threshold) based on user feedback.
    - Consider adding haptic feedback on score change.
    - Monitor for Scroll la gest conflict.
    - Warning: Ensure the backend's temporal window check (Season status) is respected before allowing submission. Ben's implementation in PR #32 handles this.
