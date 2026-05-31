     1|# Project Memory
     2|
     3|## 🟢 Current Focus
     4|- **Objective**: Resolve Evidence-Gated Execution (#18) - Ensure mobile integration matches backend gating.
     5|- **Active Issue**: #18
     6|- **Status**: Backend verified as complete; Mobile frontend integration missing.
     7|- **Agent ID**: hermes-agent
     8|- **Last Updated**: 2026-05-26T03:15:00Z
     9|
    10|## ⏳ Activity Timeline (Current Session)
    11|- 2026-05-26T02:50:00Z [hermes-agent]: Implemented hard-failure API client sync verification (#48) -> Replaced fake generator with `openapi-typescript` and added CI check to fail on stale types.
    12|- 2026-05-25T02:15:00Z [prn_dev]: Implemented seasonal boundaries (#21) -> Submission window enforcement, auto-transition logic, deadline-based API.
    13|- 2026-05-25T00:55:00Z [prn_dev]: Added accessibility labels to all score adjustment buttons -> Screen reader compatible.
    14|- 2026-05-25T00:54:00Z [prn_dev]: Updated memory protocol with security warning and append-only rule -> Prevents secret leaks and merge conflicts.
    15|- 2026-05-25T00:50:00Z [prn_dev]: Rebased feature branch on main -> Incorporated Ben's PR #32 (Judgment API) and PR #31 (Group home views).
    16|
    17|## ⚠️ Active Hurdles
    18|- **Gesture Sensitivity**: 50px threshold for score changes may need tuning based on user feedback.
    19|- **ScrollView Conflict**: PanResponder may conflict with parent ScrollView; monitoring for issues.
    20|- **PR #35 Pending Review**: Waiting for feedback from @bturney and @codex.
    21|
    22|---
    23|
    24|## 🏗️ Architecture & Decisions
    25|- **Dual-Track Approach**: Track A (Engine/Foundation) and Track B (User Experience) are running in parallel. Track B depends on Track A's underlying logic.
    26|- **Identity**: Many-to-One mapping (Auth → Account → Player) implemented in `PostgresStore`.
    27|- **Stunt Lifecycle**: Idea → Planned → Performed (gated by Evidence).
    28|- **Judging Logic**: Implemented authoritative guards (must be a group member, cannot judge own stunt, stunt must be 'Performed'). Scoring uses an upsert model (one judgment per player per stunt).
    29|- **Temporal Logic**: 
    30|  - Season status auto-transitions: Active → Judging Grace Period → Finalized (based on deadline comparison)
    31|  - Submission window open: Season status == \"Active\" AND now < submissionDeadline
    32|  - Judging window open: Season status IN (\"Active\", \"Judging Grace Period\")
    33|  - Evidence submission rejected after submission deadline (ErrSubmissionWindowClosed → 409)
    34|- **Season Deadlines**: ISO 8601 timestamps stored in DB; enforced at API boundary and store layer.
    35|- **Gesture Layer**: Implemented with PanResponder; vertical swipes adjust scores, API call only on explicit 'Submit' (per PRD #15).
    36|- **Accessibility**: All score adjustment buttons have explicit labels for screen readers.
    37|- **API Client Sync**: Strictly enforced via CI using `openapi-typescript`. Any divergence between `openapi.yaml` and `packages/api-client/src` results in a hard build failure to prevent runtime type mismatches.
    38|
    39|---
    40|
    41|## 💡 Working Hypotheses
    42|- **Tracer Bullet**: By implementing a basic scoring UI first, we can verify the end-to-end loop before adding complex swipe gestures.
    43|- **Gesture Layer**: Swipe gestures modify local frontend state only; no API calls occur until the explicit 'Submit' action (per PRD #15).
    44|
    45|---
    46|
    47|## 📜 Activity History (Archived Sessions)
    48|- 2026-05-24 [Opencode]: Implemented PanResponder for swipe gestures -> True gesture-based scoring complete.
    49|- 2026-05-24 [Opencode]: Added visual styling for gesture rows -> Improved touch targets and feedback.
    50|- 2026-05-24 [Opencode]: Button-based MVP complete -> Ready for gesture enhancement.
    51|- 2026-05-24 [Opencode]: Researched opencode MCP configuration docs -> Correct format identified (command as array, type: local).
    52|- 2026-05-24 [Opencode]: Added gesture score state and judging functions to App.tsx -> Core interaction logic ready.
    53|- 2026-05-24 [Opencode]: Implemented judging UI with score adjustment controls -> MVP for gesture layer complete.
    54|- 2026-05-24 [Opencode]: Added styles for judging card and score rows -> Visual distinction for judging mode.
    55|- 2026-05-24 [Opencode]: Reviewed Ben's PR #32 -> Confirmed temporal window logic is implemented.
    56|- 2026-05-24 [Opencode]: Added review comment to PR #32 -> Aligned on Dual-Track approach and API stability.
    57|- 2026-05-24 [Opencode]: Implemented `judgments` schema and Store logic (Postgres/Memory) -> Core engine for judging is ready.
    58|- 2026-05-24 [Opencode]: Added judging endpoints to `openapi.yaml` and `server.go` -> API surface is defined.
    59|- 2026-05-24 [Opencode]: Fixed API client generator and regenerated types -> Client now supports judging endpoints.
    60|- 2026-05-24 [Opencode]: Opened PR #30 for judging engine -> Merged into main.
    61|- 2026-05-24 [Opencode]: Updated agent memory protocol with archival flow -> Ensures historical continuity.
    62|
    63|---
    64|
    65|## 📋 PRD Context (Reference)
    66|### PRD #1: First Playable Group Stunt Loop
    67|**Status**: OPEN | **Author**: Ben Turney
    68|
    69|Single-player-per-group social game where players perform absurd food-location stunts, share evidence, judge each other's stunts on 4 axes (Commitment, Transgression, Creativity, Documentation), and compete for season standings. Uses Supabase Auth + Postgres + Expo mobile app + Go REST API. MVP excludes progression systems, public feed, auto-scoring, video evidence, push notifications.
    70|
    71|**Key Mechanisms:**
    72|- Groups multi-member with admin/commissioner roles
    73|- Season phases: submission → judging grace period → finalized
    74|- Stunt lifecycle: Idea → Planned → Performed (requires evidence + caption)
    75|- Judging: one judgment per judge, editable during window, locked after close
    76|- Standings based on season-linked judged stunts only; off-season excluded
    77|
    78|---
    79|
    80|## 📡 Past Hand-off Notes (Archived)
    81|- **Next Steps**: 
    82|    - Test end-to-end: Performed Stunt → Judge → Submit → Verify in DB.
    83|    - Tune gesture sensitivity (currently 50px threshold) based on user feedback.
    84|    - Consider adding haptic feedback on score change.
    85|    - Monitor for ScrollView gesture conflicts.
    86|- **Warning**: Ensure the backend's temporal window check (Season status) is respected before allowing submission. Ben's implementation in PR #32 handles this.
    87|