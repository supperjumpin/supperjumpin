# Stunts Belong to One Group

⚠️ **Superseded by ADR-0019 (Jumps Are Public By Default) and ADR-0020 (Rename Stunt → Jump).** Jumps no longer require Group membership. The `group_id` column is now nullable; in v1, it is always NULL. This ADR is preserved as a historical record.

For the MVP, each Stunt belongs to exactly one Group, and its judging, House Rules interpretation, Season contribution, and Standings impact are scoped to that Group. Reusing the same performance across multiple Groups is deferred because cross-group scoring would create ambiguity around duplicate points, Evidence disputes, and whether one performance can contribute to multiple competitions.
