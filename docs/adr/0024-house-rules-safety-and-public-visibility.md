# House Rules, Safety, and Public Visibility Boundaries

## Context

Jumps are public by default (ADR-0019). Any authenticated Player or Guest Judge may view and Judge a Jump without Group membership. Share links with preview cards distribute Jumps externally. This creates a public user-generated content product with no pre-moderation and no access gates.

The product vision (docs/design/01-product-vision.md) establishes "Absurdity within Boundaries" as a design pillar: the game rewards transgressive humor and deadpan commitment, but strictly enforces House Rules. The Transgression scoring axis structurally rewards escalation, creating tension with safety boundaries.

Issue #63 asked for: distinguishing good transgression from unacceptable behavior, addressing bystander privacy / public filming / trespass / harassment / business disruption / unsafe behavior, matching the safety model to the public visibility model, specifying minimum viable Dispute / Disqualified Jump / Removed Jump behavior, and calling out any public-distribution risks that must block MVP.

## Decision

1. **House Rules are defined in the domain model** and updated in CONTEXT.md:
   - "Harmful deception" clarified to "deception that causes material harm to a specific person"
   - "Unsafe behavior" sharpened to "behavior that creates imminent risk of physical injury"
   - "Disrupting a business" clarified to "intentionally preventing a business from operating"
   - Added: property damage, animal cruelty, hate speech, graphic violence
   - Added: "spirit of the game" clause giving team discretion for edge cases
   - Removed: filming identifiable non-consenting bystanders
2. **Dispute / Report flow** exists as a simple Player-facing "Report" button with 4 categories + "Other," manual team adjudication. No formal Dispute tooling is built at MVP.
3. **Removed Jump** suppresses a Jump fully from all visibility (feed, direct links, share previews) for serious violations. The performer is notified privately.
4. **Disqualified Jump** is deferred to v2. The governance model for adjudicating competitive violations will be designed when Groups are specified.
5. **No pre-moderation** at MVP. Post-moderation via Report flow + manual team removal is sufficient at seed scale.

## Rationale

**Bystander filming removed**: TikTok, Instagram, and other UGC platforms normalize incidental stranger faces in public photos. Prohibiting this creates a rule that will be routinely broken and rarely enforced, undermining the rules that matter. A soft "be mindful" nudge in the upload flow replaces the hard prohibition.

**Disqualified Jump deferred**: v1 has no Groups, no Seasons, no Season Commissioner, and no Group Admin. The fantasy commissioner model may not survive Group design. Competitive integrity issues (staged Jumps) are handled manually by the team at MVP.

**Post-moderation over pre-moderation**: Pre-moderation creates a cold-start death spiral where an empty feed kills engagement. At seed scale (10–100 Players), manual review is viable. Auto-hide on multiple reports and basic image scanning should be added before public growth.

**Report categories**: Four categories (Harassment/Targeting, Unsafe/Illegal, Privacy Violation, Misleading/Staged) plus "Other" give the team signal without overbuilding. Categories 1–3 map to Removed Jump; category 4 (game integrity) is handled manually without a formal Disqualified state in v1.

## Consequences

- The Transgression scoring axis remains structurally misaligned with House Rules. This is accepted risk at MVP; mitigation is observation and potential soft caps or safety reminders post-launch.
- Manual moderation breaks down around 1,000 active Players. Auto-hide on reports and basic image scanning (AWS Rekognition / Google Vision) must be added before reaching that scale.
- The data model should reserve space for a future Disqualified Jump status without requiring migration.
- **Admin removal tool and share link tombstoning are P0 must-build before public launch.** Auto-hide on reports and rate limiting are P1 backlog. Image scanning is P2 for post-MVP.

## Open Questions

- What is the exact threshold for "auto-hide on multiple reports" — e.g., 2 or 3 reports from **distinct Players**?
- Should high-Transgression Jumps trigger a safety reminder at submission time?
- At what Player count does the team commit to automated image scanning?
- The governance model for Disqualified Jump and Group moderation will be designed when Groups are specified in v2.

## Status

Closes issue #63. All acceptance criteria satisfied. Public-distribution risks that must block MVP are identified and prioritized.
