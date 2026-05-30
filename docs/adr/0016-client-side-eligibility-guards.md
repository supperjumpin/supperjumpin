# Client-Side Eligibility Guards

The mobile app may mirror backend eligibility rules as client-side UX guards, but only when all three conditions hold: the rule is stable and unlikely to change independently on the backend, it can be checked cheaply from already-loaded local state, and short-circuiting produces a meaningfully better user experience than waiting for an API error response.

The backend remains the authoritative source of truth for all game rules per ADR-0002. Client-side guards are a UX courtesy, not a substitute for backend enforcement. A guard that passes client-side can still be rejected by the backend.

The canonical example is preventing a Player from judging their own Stunt: the performer identity is already in local state, the rule is stable, and showing an inline message is clearly better than a round-trip rejection. Complex Season state transitions or role-based permission checks do not meet the bar and should not be duplicated client-side.

Additional guards for v1: (1) Guest Judge soft cap: after 5 Judgments, the client should show an auth gate encouraging Account creation before the server returns `ErrGuestCapReached`. (2) Grace Period countdown: the client may display a countdown timer based on `grace_period_expires_at` and disable the Judge button until the window opens, avoiding a round-trip rejection. Both guards meet the stability and local-state criteria: the Guest cap is a server-enforced rule with a stable default, and the Grace Period expiry is already in the Jump payload. Guest Judge eligibility remains server-authoritative; the client guard only improves the pre-error path.
