# Discussion Record: Issue #18 (Evidence-Gated Execution)
**Date:** 2026-05-26
**Status:** Archived

## 🎯 Executive Summary
Audited the implementation of Evidence-Gated Execution. Confirmed that backend logic is fully implemented and robust, but the Mobile UI is completely missing the evidence submission flow.

## ✅ Decisions & Agreements
- **Decision:** The primary technical debt for Issue #18 is located entirely in the **Mobile UI/Integration layer** $\rightarrow$ **Reasoning:** Backend audit confirmed `AuthorizeEvidenceUpload` and `SubmitEvidence` are correctly gating transitions to "Performed Stunt," but `apps/mobile/App.tsx` lacks the corresponding API calls and UI elements.
- **Decision:** Transition from "Background Monitoring" promises to "On-Demand Verification" $\rightarrow$ **Reasoning:** To avoid "promise hallucination" regarding agent capabilities, the agent will now execute a formal PR/CI health check skill upon request rather than claiming to watch in the background.

## 🚧 Open items / Future Work
- [ ] **Mobile Implementation:** Implement the UI flow to request upload authorization, upload media to the provided URL, and call the `submitEvidence` endpoint.
- [ ] **GitHub Sync:** Programmatically update the issue tracker once `gh` CLI authentication is restored.

## 📚 Context & References
- **PRs:** 
  - PR #52 (Merged): Resolved `implicit any` type debt in `App.tsx` for `performedStunt`.
  - PR #51 (Pending): API client sync infrastructure.
- **Key Files:** 
  - `apps/api/internal/httpapi/postgres_store.go` (Backend gating logic)
  - `apps/mobile/App.tsx` (Frontend gap)
  - `docs/issue_18_findings.md` (Detailed audit report)
- **Notes:** Backend uses a specific window for evidence submission; the mobile app must handle the authorization $\rightarrow$ upload $\rightarrow$ submission sequence.

---
*Archived via Hermes discussion-closer skill*
