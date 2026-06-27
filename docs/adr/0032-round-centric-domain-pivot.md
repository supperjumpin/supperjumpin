# Round becomes the aggregate root; Jump stays the verb

Per issue #314, Supperjumpin pivots from a feed-era competition centered on the **Jump** (a standalone performed attempt) and the **Open** (global monthly competition) to a pod-native **bit ritual**. The new aggregate root is the **Round**: one **Prompt**, one community of **Jumpers**, sealed **Submissions** revealed together, expressive **Reactions**, and a **Recap**. **Jump** is preserved as the verb/identity of the act ("take food where it doesn't belong"), not as the orchestrated aggregate.

We chose Round-centric over keeping Jump as the atomic posted artifact because the defining mechanics of the pivot — sealed-until-reveal submissions and synchronized reveal — are Round-level states that become awkward to model on a free-standing feed-era Jump. We accept significant domain churn now (touching CONTEXT.md and ADRs 0019/0023/0026) in exchange for a domain and API contract that is stable for front-end adapters to build against.

## Status

accepted — supersedes the Open-as-competitive-context framing in ADR-0023 and the feed-centric framing in ADR-0019 where they conflict.
