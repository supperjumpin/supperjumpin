# Guest Judge Session Model

## Context

Jumps are public by default (ADR-0019). Any visitor may Judge without creating an Account. This requires a way to track Guest Judge state, including Judgment count and session identity, plus a migration path when a Guest creates an Account.

## Decision

Guest Judges are tracked through a `guest_sessions` table. Each session has a UUID, a `judgment_count`, and an optional `player_id`, set when claimed. Judgments record `guest_session_id` when submitted by a Guest.

A soft cap, default 5 Judgments, limits Guest activity before encouraging Account creation. When a Guest creates an Account, their existing Judgments are migrated: `player_id` is set on each Judgment row, `guest_session_id` is nulled, and `ClaimGuestSession` prevents double-claim.

## Rationale

Guest Judging lowers the onboarding barrier to zero, aligned with ADR-0019's public-first model. The soft cap prevents abuse while still giving Guests a meaningful taste. Session-based tracking is simpler than cookie-based or IP-based approaches and survives browser restarts.

## Consequences

The `guest_sessions` table must be created. The `judgments` table needs `guest_session_id` and `provenance` columns. Partial unique indexes prevent duplicate Judgments from both Player and Guest contexts. The migration from Guest to Player must be atomic to prevent double-claim.

## References

- Technical Design §7
- Backend/Data Architecture (#107) §5.4
