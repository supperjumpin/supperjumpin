# Jumps Are Public By Default — Groups, Seasons, and Invites Removed

⚠️ **Partially superseded.** The *public-by-default* decision is reversed by ADR-0036 (trusted-group MVP — there is no public feed). The *Groups/Seasons/Invites removal* still stands, but the container is now replaced by the **Community** model (ADR-0037), not left absent. Preserved as historical record.

Supersedes: ADR-0008 (Stunts Belong to One Group)

For MVP, Jumps exist independently of Groups. A Player performs a Jump and it is visible on a public feed without requiring any Group, Season, Invite, or Dispute infrastructure. The Group, Season, Invite, and Dispute code was removed entirely per issue #225 — these features will be rebuilt from scratch when scoped for v2.

## Decision

- A Jump is owned by the Player who performed it, not by a Group.
- Groups, Seasons, Invites, and Disputes are removed from the codebase. No Group membership check gates posting, viewing, or judging.
- The default visibility of a Jump is public (visible to any authenticated Player).
- Judging is open: any Player or Guest Judge may Judge a Jump they did not perform. No Group membership required.
- The `season_id` column on `jumps` is retained as a nullable provenance field only — no Season lifecycle code (start, close, finalize) exists.
- ADR-0008's constraint ("each Stunt belongs to exactly one Group") is superseded. A Jump belongs to no Groups.
- Evidence is submitted inline during Jump creation. The separate evidence upload-authorization flow is removed.

## Rationale

The Group-first architecture optimizes for tight social competition but creates a hard onboarding barrier: a new Player must join or create a Group before they can post anything. A public-by-default feed lowers that barrier to zero and creates viral surface area — anyone can see a Jump, Judge it, and be inspired to post their own without first finding a Group to join. Removing the Group/Season code entirely avoids maintaining dead code against a future rebuild.
