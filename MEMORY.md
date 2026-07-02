# Project Memory

## 🟢 Current Focus
- **Objective**: Design package complete (issues #50–#67). Ready for implementation backlog (#68) and first tracer bullet.
- **Active Issue**: #119 Public Jump read experience is being handled externally; avoid stepping on Feed/Jump Detail work unless asked.
- **Status**: Groups/Seasons are deprioritized for now. Prefer direct changes over backward-compatibility shims because no compatibility promise exists yet.
- **Agent ID**: Sisyphus
- **Last Updated**: 2026-05-30T20:00:00Z

## ⏳ Activity Timeline (Current Session)
- 2026-06-29T22:00:00Z [OpenCode]: Completed issue #349 design via ADR-0047..0050, posted the ready-for-agent brief on #349, and started the Mage migration with TDD-style helper/plan tests.
- 2026-05-30T20:00:00Z [Sisyphus]: Completed #67 — updated CONTEXT.md and ADRs for accepted decision changes. New ADRs 0025–0027. All five acceptance criteria met.
- 2026-05-30T14:45:00Z [Sisyphus]: Seven expert agents completed growth loop analysis. Consolidated findings now available for review. Key tensions: Judge→Performer transition, monthly Open cadence, share viral mechanics, Guest conversion path, BeReal reciprocity gate gap.
- 2026-05-30T14:00:00Z [Sisyphus]: Drafted MVP Roadmap and Sequencing Plan (#65) -> docs/design/04-mvp-roadmap.md with 4 horizons, explicit cuts table, metrics per stage, and decision gates.
- 2026-05-30T14:00:00Z [Sisyphus]: Resolved Product Vision auth contradiction — Guest Judges judge freely in v1, auth wall is at posting. Updated docs/design/01-product-vision.md.

## ⚠️ Active Hurdles
- **Auth Model Contradiction**: Product Vision says "soft auth gate when trying to Judge"; CONTEXT.md says Guest Judges may Judge without auth in v1. Resolution: Guest Judges judge freely in v1; auth wall is at posting. Product Vision step 3 is aspirational for v2.
- **Monthly Open Too Sparse**: All 5 expert agents agree monthly cadence is too long for retention. Weekly rhythm (prompt, digest, spotlight) needed.
- **Retention Gap**: BeReal's reciprocity gate (post-to-view) is rejected per Design Pillar #5. Replacement mechanics (Judge Back, weekly Prompt, email capture) are identified but not in MVP scope.
- **CONTENT.md duplicate fixed in this session**.

---

## 🏗️ Architecture & Decisions
- **Dual-Track Approach**: Track A (Engine/Foundation) and Track B (User Experience) are running in parallel. Track B depends on Track A's underlying logic.
- **Identity**: Many-to-One mapping (Auth → Account → Player) implemented in `PostgresStore`.
- **Jump Lifecycle**: Draft → Performed Jump (gated by Evidence). Supersedes earlier "Stunt" terminology (ADR-0020).
- **Public-First Architecture**: Jumps are public by default (ADR-0019). Groups are optional overlay for Seasons (v2). The Open is the v1 competitive engine (ADR-0023).
- **Judging Logic**: 
  - Any Player or Guest Judge may Judge any Jump they did not perform.
  - Single-screen tap-to-select with 1-4 named tier labels (ADR-0022).
  - No auth required to Judge in v1; auth gates posting and Account features.
- **Season/Temporal Logic**: 
  - The Open: monthly soft-close, always active, no commissioner (ADR-0023).
  - Group Seasons: bounded periods with commissioner, submission window, judging grace period (v2).
  - Open Standings separate from Season Standings.
- **Build Orchestration**: `magefile/` now owns the repo command surface; the old root Node scripts are being deleted.
- **Bot Client Independence**: `apps/bot-discord` has its own Go API client. Deleting `packages/api-client/` does not remove any runtime consumer.
- **OpenAPI Drift**: The TypeScript sync gate is intentionally gone. Coverage comments and CI now only compare the Go API and Discord bot reports.
- **Migration Boundary**: Once the home-server deployment holds real group data, stop folding DB changes into existing migrations and switch to append-only numbered migrations.

---

## 💡 Working Hypotheses
- **Judge→Performer Bridge**: Taste-based Prompts after 3-5 Judgments, plus Draft creation, are the lowest-friction bridge.
- **Weekly Rhythm Required**: Monthly Open alone cannot sustain retention. Weekly Prompt, Jump of the Week, or checkpoint needed.
- **Share Card Is Critical Growth Surface**: Evidence Dossier design with deadpan-institutional tone makes it conversation fuel, not broadcast.
- **Guest Conversion Needs Scarcity**: 15% conversion requires soft cap (5-10 free Judgments), scoring style teaser, or Open eligibility hook.

---

## 📜 Activity History (Archived Sessions)
- 2026-05-26 [hermes-agent]: Implemented hard-failure API client sync verification (#48).
- 2026-05-25 [prn_dev]: Implemented seasonal boundaries (#21).
- 2026-05-24 [Opencode]: Implemented PanResponder for gesture scoring.
- 2026-05-24 [Opencode]: Implemented `judgments` schema and Store logic.

---

## 📋 PRD Context (Reference)
### PRD #1: Supperjumpin MVP — Public Performance Stage
**Status**: SUPERSEDED by ADR-0019, ADR-0020, ADR-0023 | **Author**: Ben Turney

Public-first social game where Players perform Jumps (food-location stunts), submit Evidence, and receive Judgments from any Player or Guest Judge. The Open provides monthly competition; Groups and Seasons are v2.

**Key Mechanisms:**
- Public feed of Jumps (no Group required)
- Open Judging: any Player or Guest Judge may Judge any Jump
- Monthly Open competition with soft-close Final Scores
- Jump lifecycle: Draft → Performed Jump (Evidence-gated)
- Four-factor scoring: Commitment, Transgression, Creativity, Presentation
- Guest Judges may Judge without Account; conversion to Account optional

---

## 📡 Past Hand-off Notes (Archived)
- **Growth Loop Tensions Identified**:
  - Monthly Open too sparse; weekly rhythm needed
  - Share card needs curiosity gap and "Tap to Judge" CTA
  - Guest conversion requires scarcity or competition hook
  - Retention gap from rejected reciprocity gate needs "Judge Back" + weekly Prompt
  - Content supply at 100-500 Players needs seed content + power-law concentration
