     1|# Gesture Scoring UX Pattern (supersedes ADR-0013)
     2|
     3|ADR-0013 described gesture-driven scoring shortcuts as a planned UX pattern. That pattern is now implemented in the prototype: `App.tsx` uses `PanResponder` to populate score values locally, with an explicit confirmation step before the Judgment is submitted to the backend. The durable decision is captured here; ADR-0013 is superseded.
     4|
     5|The confirmed decision: gesture interactions are shortcuts that populate local score state only. A Judgment is submitted to the backend only when the Player explicitly confirms. Unconfirmed gesture values are local state and can be cleared or manually adjusted before submission. This preserves the four-factor scoring model (Commitment, Transgression, Creativity, Documentation) and requires no changes to the API contract.
     6|
     7|The spec-level detail in ADR-0013 — interaction model, undo behavior, cross-platform drag semantics — belongs in a feature issue rather than an ADR, and will be tracked there as the frontend architecture is built out properly under #39.
     8|