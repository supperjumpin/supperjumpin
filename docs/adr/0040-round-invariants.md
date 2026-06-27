# Round invariants: one active Round per Community; existence visible, content sealed

Two domain invariants on the **Round** aggregate (ADR-0032/0038).

## One active Round per Community at a time

A **Community** has at most one active **Round** in flight. "The current Round" is therefore an unambiguous reference for the API and every front-end.

Chosen because it matches the "this week's **Prompt**" ritual and keeps the **Community**'s attention on a single shared collision — the whole point of a synchronized reveal. It is also the more restrictive invariant: relaxing it later (e.g. themed side-Rounds) is additive, whereas assuming concurrency now and tightening later would be a breaking change.

## Sealed means content-hidden, existence-visible

Before a **Reveal**, a submitted **Jump**'s *content* (**Evidence**, **Caption**) is hidden from everyone but its author, but its *existence* is visible — the **Community** can see "3 of 6 **Jumpers** have submitted."

- **Existence visible** fuels the "I'm In" anticipation beat and sets up the **Ghost Jumper** distinction (committed vs. delivered).
- **Content hidden** preserves simultaneous interpretation — nobody can copy the first good idea — which is what makes the **Reveal** a genuine collision.

Hiding existence as well would kill the anticipation for no real gain among trusted friends.

## Status

accepted
