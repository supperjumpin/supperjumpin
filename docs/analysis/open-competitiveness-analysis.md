# Analysis: Is The Open Compelling Enough?

**Research question:** Does a monthly, soft-close, no-commissioner Open drive retention and re-engagement at 100–500 Players, or does it produce weak Standings nobody cares about?

**Scope:** This analysis evaluates the v1 competitive structure (The Open, per ADR-0023) against four research questions and the Player Signal hypothesis. It draws on competitive product research, behavioral economics literature, and documented leaderboard design patterns.

---

## 1. Executive Summary

**The Open, as currently specified, carries meaningful competitive-design risk.** The combination of (a) a monthly cadence, (b) a soft-close with no hard deadline, (c) no social container, and (d) a global competitive pool at 100–500 scale conflicts with well-documented patterns that make leaderboards motivating rather than demotivating.

However, **none of these risks are fatal**, and the Open has a strategic advantage that most competitions lack: it can be redesigned to generate *player signal* about what the community values, not merely to crown winners. The recommendation is to **ship the Open as specified for the first cycle, but instrument it aggressively and be prepared to pivot the competitive mechanics within 30–60 days based on behavioral data.**

**Bottom line:** The Open is a reasonable MVP starting point, but the current design is more of a *signal-collection surface* than a *retention engine*. Do not expect it to carry retention on its own.

---

## 2. Monthly Leaderboards Without Social Containers

### 2.1 The Research Consensus

Leaderboards without social context are one of the most thoroughly documented failure modes in gamification. The pattern is consistent across fitness apps, learning apps, B2B products, and games:

> "A leaderboard where the top 100 users are strangers who do not know each other is not a leaderboard. It is a wall of names." — EngageFabric

> "Global leaderboards create what Punit calls an 'overwhelming' experience... You're lost in a sea of strangers. The top spots are occupied by people with tens of thousands of points—a gap so large it feels insurmountable." — Trophy.so / Programiz case study

> "A standard first-place-to-last-place ranking only energizes a small group near the top. Everyone else sees a giant gap, feels they have already lost, and stops caring." — Yu-kai Chou

The mechanics are well understood:
- **Global leaderboards motivate ~1% of users** (those who can realistically reach the top).
- **For the other 99%, rank confirms defeat** and accelerates churn.
- **Social context transforms competition** from abstract numbers into personal stakes.

### 2.2 What Works Instead

Products that drive sustained engagement through competition share one trait: **the reference group is small enough that winning feels attainable.**

| Product | Competitive Structure | Key Design Choice |
|---|---|---|
| **Duolingo Leagues** | Weekly 30-person cohorts | Random assignment by recent engagement; promotion/demotion tiers |
| **Strava Segments** | Hyper-local route leaderboards | Only athletes who ran the *same stretch of road* compete |
| **Fitbit Friend Leaderboards** | Friend-group only | 5–15 people you know personally |
| **Legendary: Game of Heroes** | Weekly leagues of ~150 players | Promotion/relegation across 4 tiers; guild leaderboards parallel |
| **Nike Run Club** | Friend + challenge-based | Distance goals with visible progress; social validation (cheers) |

Duolingo is the most directly relevant analogue. Its Chief Product Officer, Jorge Mazal, explicitly states that the key innovation was **"match users with others who had similar engagement levels the week before."** The result: learning time jumped 17% immediately, and highly engaged learners tripled.

### 2.3 Implications for The Open

At 100–500 Players, a monthly global Open Standing is **better than a 100,000-user global board** but still suffers from the same structural problem:

- A Player ranked #87 of 400 knows they will not win.
- There is no friend-group context, no local rivalry, no social stakes.
- The monthly cadence means a Player who falls behind in week 1 has little reason to re-engage in week 3.

**Verdict:** A monthly global leaderboard without social segmentation is unlikely to be a strong retention driver at MVP scale. It may work as a *ritual* (monthly score reveal), but not as a *competition*.

---

## 3. The Soft-Close Pattern

### 3.1 What the Research Says About Deadlines

Hard deadlines create urgency. Soft deadlines deflate it. The evidence is overwhelming:

> "Scarcity drives action. 'Complete this achievement by December 31st' prompts action. 'Complete this achievement anytime' can wait. The deadline creates urgency that permanent mechanics lack." — Trophy.so

> "No urgency means no countdown, no shared push and no reason to act today. There's no reason to engage today instead of next week." — BuddyBoss

> "Deadlines create urgency. Urgency creates action." — BuddyBoss

> "Soft deadlines with extensions... train users that deadlines aren't real. Repeated extensions destroy trust." — EventXGames

Academic research confirms this. Alcalde (2024), analyzing optimal contest design, concludes: **"the designer should announce the contest immediately with a short deadline to promote intense competition."**

### 3.2 The FOMO Architecture in Games

Successful live-service games structure urgency explicitly:
- **Fortnite** seasonal battle passes: items become permanently inaccessible after season end.
- **Genshin Impact** limited-time banner events: rare characters available for narrow windows.
- **Destiny 2** Seasonal Triumphs: complex challenges with tight deadlines, clan communities pressuring each other.

These systems create what researchers call **"calendarized obligation"** — players return not because of novelty but because of anxiety about exclusion.

### 3.3 The Counter-Argument: Returnable Rewards

Disney Dreamlight Valley's **Star Path** offers a healthier middle path: missed rewards return later. This transforms the emotional meaning of missing an event from **permanent loss** to **delay**. The design insight:

> "A player who knows missed rewards can come back later is more likely to keep playing casually during a busy month, rather than quitting in frustration after falling behind for a few days."

This is relevant to The Open because the soft-close is *implicitly* a return-friendly system: a Player can always compete next month. But without any explicit signaling that "next month is a fresh start," the soft-close may simply feel like **absence of stakes** rather than **forgiveness**.

### 3.4 Implications for The Open

The Open's soft-close ("Final Scores computed from whatever Judgments exist") has three problems:

1. **No shared countdown moment.** There is no "last 48 hours" push because there is nothing to push *toward* — scores simply materialize at month-end.
2. **No Judging urgency.** Judges have no reason to Judge a Jump today rather than next week, which means Jumps may sit with few Judgments until the arbitrary month-end snapshot.
3. **No reciprocity gate.** Unlike Group Seasons where peer pressure drives participation, the Open has no social container to create obligation.

**Verdict:** The soft-close is the right choice for operational simplicity (no Commissioner), but it is a *weak* competitive mechanic. It does not create urgency, and urgency is what drives re-engagement.

---

## 4. What Makes Monthly Competitions Meaningful at Small Scale

### 4.1 The Attainability Problem

At 100–500 users, the fundamental question is not "how often does the competition reset?" but **"does every participant believe they have a realistic path to recognition?"**

Research on gamification psychology identifies **perceived attainability** as the strongest predictor of motivation:

> "When users believe they can win, they engage. When they don't, they leave." — Trophy.so

Duolingo explicitly rejected monthly leaderboards for this reason:

> "Pick anything longer (14 days, 30 days) and the prize fades from view as users churn out before the reward arrives." — Duolingo Leagues design documentation

At 100–500 Players, a monthly global board means:
- Top 10 get recognition; bottom 490 do not.
- A new Player joining in week 3 has ~1 week to accumulate Judgments against Players who have been active all month.
- There is no "tier" system where a mid-performing Player can still feel like they won something.

### 4.2 Segmentation Is Required

The fix, documented across every successful competitive product, is **segmentation**:

| Segmentation Type | How It Helps at Small Scale |
|---|---|
| **Skill/engagement tiers** | Bronze/Silver/Gold leagues group similar Players; everyone competes against peers |
| **Time-based cohorts** | Weekly resets give new Players a fresh start every 7 days |
| **Social/friend groups** | Competing against people you know is more motivating than competing against strangers |
| **Local/contextual** | Strava Segments let an average runner be #1 on their local hill |

The Social Games Playbook explicitly recommends:

> "Member cap: 30–50. Below 10 the guild dies; above 100 the social fabric thins."

At 100–500 Players, the Open would need to be **subdivided into ~10–20 cohorts of ~25–50 Players each** to create meaningful competition.

### 4.3 The "Second Game" Problem

Adrian Crook's "Second Game Strategy" essay (2026) argues that modern F2P games must answer: **"What makes this world worth checking in on even when nothing is on fire?"**

> "A pressure-heavy economy asks, 'What happens if the player skips today?' A second-game economy asks, 'What makes this world worth checking in on even when nothing is on fire?'"

The Open, as a monthly soft-close competition, leans toward the "second game" end of the spectrum — low pressure, low obligation. This is not inherently bad, but it means **the Open cannot be the primary retention driver.** Something else must make Players want to return: push notifications on Judgments received, shareable artifacts, curiosity about new Jumps, or social ritual.

### 4.4 Implications for The Open

At 100–500 Players, a monthly global Open Standing is structurally weak because:
- The competitive pool is too large for meaningful rivalry.
- The timeframe is too long for sustained engagement.
- There is no segmentation to create winnable sub-competitions.

**Verdict:** The Open needs either (a) segmentation into smaller cohorts, (b) a shorter cadence (weekly), or (c) a different competitive mechanic entirely (e.g., challenges, prompts, or streaks) to be compelling at small scale.

---

## 5. Examples of Successful and Failed Monthly Competition Structures

### 5.1 Successful Structures

#### Duolingo Leagues (Weekly, Not Monthly — But the Best Analogue)
- **Structure:** 30-person weekly cohorts, 10-tier promotion/demotion ladder.
- **Why it works:** Small cohorts + skill-based matchmaking + weekly resets + demotion threat.
- **Result:** +17% learning time immediately; 3x highly engaged learners.
- **Lesson for The Open:** Weekly is better than monthly; small cohorts are better than global; demotion creates loss aversion.

#### Strava Segments + Challenges
- **Structure:** Hyper-local route leaderboards (Segments) + opt-in monthly distance/elevation Challenges.
- **Why it works:** Segments make competition contextually relevant; Challenges are opt-in public commitments.
- **Result:** 130M+ athletes; 14 billion kudos in 2025.
- **Lesson for The Open:** Local/contextual competition scales across ability levels; opt-in challenges create accountability without coercion.

#### Tiny Tower + Athlos Leaderboards
- **Structure:** Casual single-player game with added leaderboard tasks (build floors, elevator assistance, watch ads).
- **Why it works:** Leaderboards added without changing core gameplay; 2-day tournaments scheduled weekly.
- **Result:** +24% IAP revenue during peak event; +42% ad revenue; +8% D7 retention.
- **Lesson for The Open:** Competitive features can work even in non-competitive games if they are lightweight and frequent.

#### Legendary: Game of Heroes
- **Structure:** Weekly events with individual leagues of ~150 players across 4 tiers + global guild leaderboards.
- **Why it works:** Promotion/relegation creates week-over-week objectives; guild pressure drives participation.
- **Result:** Outperformed comparable rivals despite older tech and weaker IP.
- **Lesson for The Open:** Tiered competition with consequences (promotion/demotion) sustains engagement far better than static rankings.

### 5.2 Failed Structures

#### Programiz Global Leaderboard
- **Structure:** Global coding challenge leaderboard; users ranked by XP worldwide.
- **Why it failed:** Users felt "lost in a sea of strangers"; top spots had insurmountable gaps.
- **Result:** Engagement spiked at launch, then collapsed within weeks.
- **Lesson for The Open:** Global leaderboards without social context or segmentation are demotivating for the majority.

#### Most "Bolts-On" Leaderboards
- **Pattern:** Product adds a global leaderboard as a gamification afterthought.
- **Symptom:** High engagement at launch, then usage collapses to <10% returning-user rate within a month.
- **Root cause:** No social gravity; no achievable targets; no reason to care about rank.
- **Lesson for The Open:** A leaderboard is not a feature — it is a social system. If the social context is missing, the leaderboard is just a database query.

### 5.3 The Middle Ground: Mob Control's Bot-Filled Leagues

Voodoo's Mob Control uses an instructive hack: **its leaderboards are filled with bots.**

> "Users are given a bucket of 'players' that are all within a reasonable distance to allow for meaningful position jumps and attainable short/mid-term goals."

This is ethically questionable but revealing: **the product team recognized that a real leaderboard with real players would be demotivating for most users, so they manufactured attainability.**

**Lesson for The Open:** If the real competitive pool is too sparse or too skewed, the product may need to *engineer* attainability — through segmentation, tiering, or even synthetic competition — rather than relying on a raw global ranking.

---

## 6. The Player Signal Opportunity

### 6.1 What Is Player Signal?

Adrian Crook's "Player Signal Is the New UA Advantage" (2026) argues that studios should not use paid user acquisition to discover what players want. Instead, they should build **signal loops** that teach the team what players value *before* scaling spend.

> "The better question is no longer, 'Can we afford UA?' It is: 'What player Signal can we own before we amplify the product with UA?'"

> "A practical studio ritual is a player-signal review with one rule: every signal cluster must produce either a product action, a live-ops action, a campaign action, or a conscious decision to ignore it."

### 6.2 How The Open Can Generate Signal

The Open is uniquely positioned to generate high-quality behavioral signal because:

1. **Judgment patterns reveal preference.** Which Jumps get Judged most? Which scoring factors correlate with share velocity? What Sources/Destinations/Foods are most popular?
2. **Participation patterns reveal engagement drivers.** Do Players post more Jumps in the first week of the month (fresh start energy) or the last week (deadline urgency)? Do Push notifications on Judgments received drive more return than Open Standings updates?
3. **Cohort behavior reveals competitive design.** If 80% of Players post in the final 72 hours, the monthly cadence is too long. If most Jumps get <2 Judgments, the judging incentive is insufficient.
4. **Segmentation signal reveals social structure.** Do Players who share Jumps to group chats get more Judgments? Do Players who Judge early shape final scores significantly? This informs whether Group Seasons should be accelerated.

### 6.3 Designing The Open for Signal, Not Just Competition

If The Open is treated as a **signal-generation system** rather than a **competition mechanic**, the design priorities shift:

| Traditional Competition Design | Signal-First Design |
|---|---|
| Maximize number of Players competing | Maximize number of *meaningful interactions* (Judgments, Shares) |
| Optimize for clear winners | Optimize for observable behavioral patterns |
| Drive retention through rivalry | Drive retention through feedback loops (Judgment notifications, share velocity) |
| Monthly Standings as payoff | Monthly Standings as *ritual* that surfaces community norms |

**The Open's true value may not be crowning a monthly winner.** It may be creating a predictable, bounded window that generates clean behavioral data about:
- What makes a Jump "good" (correlation between scores and share rate)
- What drives Judging (notification timing, Jump recency, social connection)
- What drives posting (Prompts, peer pressure, Open participation)

### 6.4 Instrumentation Recommendations

To capture this signal, the Open should track:

1. **Judgment velocity:** Judgments per Jump per day, segmented by day-of-month.
2. **Posting cadence:** % of Jumps posted in first week vs. last week of month.
3. **Share-to-Judge funnel:** For shared Jumps, what % of recipients Judge? What % create Accounts?
4. **Score distribution:** Mean, median, and variance of Final Scores. Are scores compressed (everyone is similar) or spread out?
5. **Standing engagement:** How many Players check Standings? How often? Does Standing visibility correlate with re-engagement?
6. **Cohort dynamics:** If segmented into leagues, do smaller cohorts show higher engagement?

---

## 7. Synthesis: Is The Open Compelling Enough?

### 7.1 The Honest Assessment

**No — not if judged as a pure competition mechanic.**

A monthly, soft-close, global Open at 100–500 Players lacks the structural elements that make competitions motivating:
- ❌ No social container (no peer pressure, no rivalry)
- ❌ No hard deadline (no urgency, no FOMO)
- ❌ No segmentation (most Players have no realistic path to recognition)
- ❌ Monthly cadence is too long for sustained engagement
- ❌ No promotion/demotion or tiering (no long-term progression)

**However, the Open is not wrong — it is incomplete.**

### 7.2 What The Open Does Well

The Open succeeds at four things that are genuinely valuable for MVP:

1. **Operational simplicity.** No Commissioner, no Submission Window, no Judging Grace Period. At seed scale, this is the right call.
2. **Universal access.** Any Player with a Jump and a Judgment competes. No Group coordination required.
3. **Ritual payoff.** A monthly score reveal is a *ritual* — even if the competition is weak, the ritual creates a cadence.
4. **Signal generation.** The bounded monthly window produces clean data about player behavior that an unbounded public feed cannot.

### 7.3 The Retention Risk

The Product Vision (docs/design/01-product-vision.md) correctly identifies the highest-risk question:

> "What replaces the reciprocity gate for cold-start retention? BeReal required posting before seeing content — its stickiest retention mechanic is unavailable here... Whether the Open's monthly cadence, push notifications on Judgments received, or something else fills this gap is the single highest-risk product question for v1 retention."

**The Open alone does not fill this gap.** A Player who posts a Jump, receives 1–2 Judgments, and sees a middling monthly Standing has no strong reason to return. The retention work must be done by:
- Push notifications on Judgments received (dopamine feedback)
- Shareable artifacts (viral loop)
- Prompts or weekly nudges (creation stimulus)
- The intrinsic quality of the content (entertainment value)

### 7.4 The Pivot Path

If the Open underperforms after 1–2 cycles, the research points to clear alternatives:

| Problem | Evidence-Based Fix |
|---|---|
| Monthly is too long | Shift to **weekly mini-Opens** or add **weekly Prompt challenges** alongside the monthly Open |
| Global is too impersonal | Segment Players into **~30-person cohorts** by engagement level, geography, or join date |
| Soft-close lacks urgency | Add **countdown notifications** ("48 hours left to earn your Open score") |
| No social pressure | Add **friend-group Standings** or **team challenges** even before full Group Seasons |
| No long-term progression | Add **tiered leagues** (Bronze/Silver/Gold) with promotion/demotion |

---

## 8. Recommendations

### 8.1 Ship The Open As Specified — But With Eyes Open

The Open is the right MVP competitive structure for operational reasons. Do not delay MVP to build a more complex competition system. **But:** set expectations correctly. The Open is a *signal surface* and a *ritual*, not a *retention engine*.

### 8.2 Instrument Aggressively

Track the metrics that reveal whether the Open is working as a competition or merely existing as a feature:
- **Standing check rate:** % of eligible Players who view Open Standings
- **Judgment velocity by week-of-month:** Does activity spike near month-end?
- **Posting cadence:** Is there a "deadline effect" in the final week?
- **Score correlation with re-engagement:** Do Players who score higher return more often?
- **Cohort comparison:** If A/B testing leagues vs. global, which drives more Judgments?

### 8.3 Treat the First Cycle as a Competitive Design Experiment

Run the first Open (Month 1) as a **bare-bones experiment**. In Month 2, add one urgency mechanic:
- **Option A:** "Open closes in 48 hours" push notification to all Players with a Jump.
- **Option B:** Weekly "Prompt of the Week" mini-challenge with a 7-day window.
- **Option C:** Segment Standings into ~5 cohorts of ~20–100 Players each (depending on total active count).

Measure which intervention moves the North Star metric (Judgments per Jump).

### 8.4 Do Not Rely on The Open for Retention

The Product Vision is correct: the core retention driver is the quality of the content and the feedback loop (Judgments received), not the competitive structure. The Open is a *payoff* for Players who are already engaged. It is not a *hook* for Players who are not.

Invest retention energy in:
1. **Push notifications on Judgments received** (immediate dopamine)
2. **Shareable score artifacts** (viral loop)
3. **Prompts / weekly themes** (creation stimulus)
4. **Guest-to-Player conversion flow** (growth)

### 8.5 Accelerate Group Seasons If Signal Demands It

The research is clear: **social containers make competition work.** If the Open data shows that Players who share Jumps to friends get 3x more Judgments, or that Players consistently ask about "private leaderboards," this is signal that Group Seasons should move up from v2.

The ADR-0009 (Season Commissioner) already includes this guardrail:

> "The long-term viability of the Season Commissioner model has not been validated. Before Groups and Seasons are implemented, this decision should be regrilled against whatever is learned from the Open."

**Use the Open to learn whether Players want competition at all, and if so, what kind.** Then build the competitive structure that matches the behavior, not the one that was planned in advance.

### 8.6 The Player Signal Mandate

Every Open cycle should produce at least one product decision. Examples:
- "Players post 60% of Jumps in the final week → add mid-month Prompt nudges."
- "Jumps with Transgression scores >3 get 2x more Shares → lean into transgressive Prompts."
- "Players who check Standings have 40% higher D7 retention → make Standings more visible in the feed."
- "Top 10% of Players dominate Judging → add Guest Judge onboarding to broaden the judge pool."

If the Open is not producing actionable signal, it is not worth the engineering investment.

---

## 9. Conclusion

The Open, as currently designed, is **not compelling enough to drive retention and re-engagement on its own.** A monthly, soft-close, global leaderboard without social segmentation conflicts with well-documented best practices for competitive design.

**However, it does not need to be compelling on its own.** At MVP, the Open serves three valuable roles:
1. **Operational simplicity** — no Commissioner, no complex phases.
2. **Universal access** — any Player can compete without Group coordination.
3. **Signal generation** — a bounded window that produces clean behavioral data.

The recommendation is to **ship it, instrument it, and treat it as a learning surface.** Expect the first cycle to validate that the current design is weak as a competition. Use that data to pivot toward a segmented, tiered, or more frequent competitive structure in Month 2–3.

**The Open's success should not be measured by how many Players care about winning it.** It should be measured by how much it teaches the team about what Players care about — and how quickly those lessons turn into product changes.

---

## Sources and References

- Duolingo Leagues design documentation (duolingo.deconstructoroffun.com)
- "Why Most Leaderboards Fail" — Yu-kai Chou (yukaichou.com)
- "Why Global Leaderboards Fail" — Trophy.so
- "Strava Gamification Strategy" — Trophy.so
- "Player Signal Is the New UA Advantage" — Adrian Crook (blogarama.com/careers-and-industries-blogs)
- "The Second Game Strategy" — Adrian Crook (adriancrook.com)
- "Social Features in Mobile Games" — MAF (maf.ad)
- "Mid-Core Success Part 3: Social" — GameAnalytics / Deconstructor of Fun
- "Leaderboards in Gamification: 5 Real App Examples" — Trophy.so
- "Using Competition to Drive Player Engagement" — SuperScale / Athlos
- "Bolting on Leaderboards Without Social Context" — EngageFabric
- Alcalde, J. "Contests with endogenous deadlines" — Journal of Economics & Management Strategy
- "FOMO as an Engagement Strategy in Video Game Design" — Design The Game
- "Star Path and the End of Live-Service FOMO" — gamingbox.store
- "Voodoo's Breakthrough with Mob Control" — Reverse Nerf
- "Legendary: Game of Heroes — A Master Class in Live Operations" — Deconstructor of Fun
- Supperjumpin ADRs: 0009, 0012, 0019, 0021, 0023
- Supperjumpin Product Vision (docs/design/01-product-vision.md)
- Supperjumpin MVP Roadmap (docs/design/04-mvp-roadmap.md)
- Supperjumpin Market Research (docs/research/issue-62-market-research.md)
