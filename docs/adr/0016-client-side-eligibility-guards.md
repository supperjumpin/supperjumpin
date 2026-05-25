# Client-Side Eligibility Guards

The mobile app may mirror backend eligibility rules as client-side UX guards, but only when all three conditions hold: the rule is stable and unlikely to change independently on the backend, it can be checked cheaply from already-loaded local state, and short-circuiting produces a meaningfully better user experience than waiting for an API error response.

The backend remains the authoritative source of truth for all game rules per ADR-0002. Client-side guards are a UX courtesy, not a substitute for backend enforcement. A guard that passes client-side can still be rejected by the backend.

The canonical example is preventing a Player from judging their own Stunt: the performer identity is already in local state, the rule is stable, and showing an inline message is clearly better than a round-trip rejection. Complex Season state transitions or role-based permission checks do not meet the bar and should not be duplicated client-side.
