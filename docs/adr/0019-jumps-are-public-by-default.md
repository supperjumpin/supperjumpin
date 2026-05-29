# Jumps Are Public By Default

Supersedes: ADR-0008 (Stunts Belong to One Group)

For MVP, Jumps exist independently of Groups. A Player performs a Jump and it is visible on a public feed without requiring Group membership. Groups are an optional social overlay — a Player may associate a Jump with a Group for Season competition, but Group membership is not required to post, view, or judge a Jump.

## Decision

- A Jump is owned by the Player who performed it, not by a Group.
- The default visibility of a Jump is public (visible to any authenticated Player).
- Judging on the public feed is open: any Player may Judge a Jump they did not perform.
- Groups remain the context for Seasons, Standings, and Awards. A Jump may be submitted to a Group's Active Season, but this is additive — the Jump exists on the public feed regardless.
- ADR-0008's constraint ("each Stunt belongs to exactly one Group") is superseded. A Jump may belong to zero or one Groups for Season purposes.

## Rationale

The Group-first architecture optimizes for tight social competition but creates a hard onboarding barrier: a new Player must join or create a Group before they can post anything. A public-by-default feed lowers that barrier to zero and creates viral surface area — anyone can see a Jump, Judge it, and be inspired to post their own without first finding a Group to join. Groups and Seasons layer on top as the competitive context once a Player is engaged.
