# Market Research: Adjacent Products and Positioning
_Issue #62 — Research spike_

---

## Competitive Landscape

### BeReal
- Rewards authentic, unguarded self-documentation against a norm of curated feeds; daily push notification creates the ritual.
- Supperjumpin differs: a Jump is deliberate transgression, not involuntary candor. Evidence is a scored artifact, not a raw snapshot.
- Trajectory: 73M MAU in Aug 2022 → 16M by March 2025. Ritual-novelty plateaus fast.
- **Risk for Supperjumpin:** when BeReal introduced brand integrations, loyalists felt the "no ads" contract was broken. A Sponsored Bounty model faces the same backlash if brand presence feels injected rather than native to the game mechanic.

### Jackbox Games
- Rewards group performance and social judgment in a co-present session; users show up because others are already in the room.
- Supperjumpin differs: asynchronous and persistent — Judges arrive on their own schedule, not locked into a session window.
- Jackbox's engagement lives or dies at gatherings. Its absurd-performance model requires a closed social container that licenses silliness.
- **Risk:** without that container, solo Judging must be intrinsically compelling. If Judging feels like rating posts rather than rendering game verdicts, the social engine stalls. ADR-0018 locks public open judging — the closed-container safety net is deliberately absent.

### TikTok Challenges (as format)
- Rewards participation in a moment of cultural proximity — doing the thing everyone else is doing and being seen doing it.
- Supperjumpin differs: structured scoring (Difficulty, Transgression, Creativity, Presentation) creates a persistent evaluative layer TikTok challenges lack.
- **Risk:** the Transgression dimension structurally rewards escalation. The fire noodle challenge (1.5M+ videos, documented harm) is a clear precedent. Content moderation and House Rules are load-bearing, not cosmetic.

### Wordle / Ritual Social Games
- Rewards daily low-friction participation with a shareable output that creates common cultural reference.
- Supperjumpin differs: a Jump requires real-world logistics — sourcing food, traveling, capturing Evidence. Entry cost is orders of magnitude higher than a two-minute daily puzzle.
- **Risk:** Wordle's habit loop requires frictionless daily reset. Supperjumpin's natural cadence is probably weekly or occasional, not daily. The sharing moment has to carry habit work that Wordle offloads to repetition.

### Strava
- Rewards documented physical achievement with peer recognition. Group activities receive substantially more engagement than solo ones.
- Judging model is congratulatory (kudos), not adversarial-creative. Supperjumpin's Judges evaluate and score — closer to a competitive league than a kudos feed.
- Strava has GPS as an authenticity substrate — activity data is hard to fake.
- **Risk:** Supperjumpin's Evidence (photo + caption) is easy to stage. If staged Jumps circulate without consequence, Credibility degrades and the Judging economy loses meaning. Dispute mechanics are the only check.

### Food Creator Trends
- Rewards personality-driven discovery content; audiences follow creators, not formats. Saturated but massive.
- Supperjumpin differs: the game layer exists independent of creator identity. Any Player's Jump can be compelling regardless of follower count.
- **Risk:** without the game layer being immediately legible, a Jump is just food content in an already saturated format. New users scrolling the public feed must understand what they're looking at and why it's funny within the first session.

---

## Behavioral Analogues

### Which analogue best predicts day-one Player behavior?

No single analogue covers the full tension (public performance + open stranger judging). The closest composite:

- **TikTok challenges** best predict spread mechanics and participation format once seeded — the format behavior (film yourself doing a food thing, share it) is already trained at scale.
- **Wordle** best predicts the shareable artifact design — the result (Evidence + score breakdown) needs to be forwardable as a self-contained unit that makes sense to someone who has never heard of Supperjumpin.
- **BeReal** is the most instructive warning: its stickiest mechanic was a reciprocity gate (post to unlock others' content). Supperjumpin has no such gate by design (ADR-0018: open judging). This means BeReal's cold-start retention mechanism is unavailable here — the product must find another reason for first-time Players to return after posting their first Jump. See HITL questions below.
- **Jackbox** is a warning, not a model: absurd performance requires a closed social container; ADR-0018's public-by-default removes it.

### Is "Judge a stranger's Jump" a realistic cold-start behavior?

- Reddit research: ~73% of votes are cast without deep engagement with content — judging strangers is behaviorally natural when the surface is legible at a glance.
- Position bias is strong: lower-ranked content gets ~40% less engagement regardless of quality. First Jumps in a feed set the engagement floor.
- Herding effect: early positive scoring increases final rating probability by ~24.6%. First Judges heavily shape perceived quality.
- The four-factor rubric (Difficulty, Transgression, Creativity, Presentation) adds cognitive load relative to a simple upvote. Multi-axis judging is known to reduce participation vs. single-axis. _(No direct analogue found for multi-factor stranger scoring at cold start — unverified.)_
- Without social connection, Judges need intrinsic entertainment value in the content itself. A confusing or boring Jump generates no judging, not negative judging.

### Evidence quality expectations

- Authentic casual UGC consistently outperforms polished content in engagement. The platform will likely self-calibrate toward phone-quality video.
- Quality floor is set by the first visible content a new user sees. If early seeded content is high-production, day-one Players will feel underqualified to submit.
- Users exposed to significantly higher-quality peer content show declining engagement — quality gap suppresses participation.
- TikTok challenge spread depends on content being "easy to replicate." Overproduced Evidence signals the Jump is harder to attempt.
- BeReal's explicit anti-curation stance (no filters, dual camera, time-pressured) suggests constraining quality upward can be a feature — it removes the "my video isn't good enough" exit ramp.
- **Highest risk:** inconsistent quality across Players. If some submit cinematic Evidence and others submit shaky phone video, the Presentation axis will mechanically penalize casual submitters and suppress casual participation over time.

### Sharing surface

- Wordle's viral growth was driven by shareable result artifacts (emoji grids) sent via DMs and group chats — direct share, not feed discovery, was the primary initial vector.
- Instagram DM sends are described as the most powerful discovery signal for reaching new audiences.
- A Jump has natural direct-share utility: "look at this ridiculous thing someone did" is a strong message to send a specific friend.
- TikTok's For You Page amplification is a secondary-phase mechanic — it rewards content that already has direct-share velocity. Ambient feed discovery follows successful direct sharing; it doesn't replace it.
- The shareable artifact (Evidence + score breakdown) needs to be forwardable as a self-contained unit intelligible to someone who has never heard of Supperjumpin.

---

## Why Supperjumpin Is Timely Now

- TikTok has trained audience behavior at scale — filming yourself doing a food thing and sharing it is normalized, not novel.
- Food content is the most-watched category on short-form video; distribution into that audience is frictionless in a way it was not pre-TikTok.
- Creator-monetization infrastructure has matured enough that a Sponsored Bounty model has prior art and willing brand buyers, particularly restaurant chains reallocating budgets from macro to niche activations.
- The specific combination — absurdist rule-breaking + food + peer judgment + IRL documentation — has no direct incumbent.

## Cultural and Platform Risks

- **Platform distribution:** if primary acquisition is short-form video, TikTok's regulatory uncertainty (US ban risk) is a single-point-of-failure for top-of-funnel. _(Whether this materializes as a real distribution cutoff — unverified.)_
- **Transgression escalation:** the scoring system structurally rewards pushing limits. Moderation at scale is expensive and a known UGC failure mode. House Rules and Dispute mechanics are load-bearing.
- **Audience fatigue:** challenge formats have visible boom-bust cycles. Supperjumpin's durability depends on the game layer providing enough variety to outlast any single Jump archetype going stale.
- **Staged Evidence:** unlike Strava (GPS substrate), Supperjumpin has no hard authenticity check. Credibility degrades if staged Jumps circulate without Dispute consequences.

---

## Implications

### Visibility
The public-by-default model (ADR-0018) removes the closed-container safety net that makes absurd social performance comfortable in Jackbox. First-session Players must feel licensed to perform publicly for strangers without that container. How the feed is introduced — what content appears first, whether early norms are visibly set — is the primary visibility design challenge.

### Evidence Quality
Seed early content at casual phone-video quality, not polished production. The platform norm is set by what Players see first. Protect casual submitters from being mechanically penalized by the Presentation axis — the quality floor should be designed, not emergent.

### Sharing Format
The Jump result artifact (Evidence + four-factor score breakdown) should be the primary viral surface — designed to be forwarded to someone who has never heard of Supperjumpin and immediately make them laugh or want to try it. Direct share is the cold-start growth vector, not algorithmic feed.

### Creator and Sponsor Path
- The Sponsored Bounty primitive maps cleanly to how restaurants already pay micro-creators for city-specific activations.
- BeReal's backlash history is the specific warning: brand presence must be legible as in-game content (a Bounty Players are invited to participate in), not injected advertising.
- Performance-plus-volume deal structures (base + outcome-based upside) align sponsor economics with game participation rather than impressions.

---

## HITL Decisions — Not Research Conclusions

These questions are raised by the research but cannot be resolved by it:

1. **Reciprocity gate vs. open judging.** BeReal's stickiest retention mechanic was requiring users to post before seeing others' content. ADR-0018 locks open judging — any Player may Judge without posting. This is the right product call for the public-feed vision, but it removes the cold-start return incentive BeReal relied on. What replaces it? This is a product design decision, not a research finding.

2. **Transgression ceiling and moderation bar.** The Transgression scoring axis structurally rewards escalation. Where is the ceiling, and who enforces it? House Rules and Dispute mechanics are currently the only check. Whether those mechanics are sufficient at cold-start scale — before norms are established — is a decision about moderation investment, not a research conclusion.

3. **Evidence quality floor governance.** Research suggests inconsistent quality will suppress casual participation over time. Whether the product should actively constrain Evidence quality (à la BeReal's anti-curation stance), let it self-regulate, or rely on the Presentation axis to signal norms is a product and community design decision.

---

_Research conducted by Trend Researcher and UX Researcher agents in parallel. Sources cited inline by each agent. Claims flagged (unverified) where no grounded source was found._
