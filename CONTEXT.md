# Supperjumpin

Supperjumpin is a social game about inventing and performing absurd food-location combinations inspired by Jon Bois' Supperjumpin article. The game borrows the spirit of the article without treating its exact terminology or rules as binding canon.

## Language

**Jump**:
A playable Supperjumpin attempt centered on taking food associated with one place and consuming or presenting it in another place. A **Jump** starts as a client-only **Draft**, becomes a **Performed Jump** only when **Evidence** is submitted, passes through the **Author Grace Period**, and is then socially judged.
_Avoid_: Stunt, challenge, mission, post

**Source**:
The place, brand, cuisine, event, or context the food in a **Jump** is associated with.
_Avoid_: Restaurant A, origin

**Destination**:
The place, brand, venue, or context where the food in a **Jump** is consumed, displayed, or documented.
_Avoid_: Restaurant B, target

**Food**:
The item or items carried from a **Source** into a **Destination** as part of a **Jump**.
_Avoid_: Meal, order

**Draft**:
A private, local-only Jump concept a **Player** is composing before submitting **Evidence**. A **Draft** may include a **Caption**, **Source**, **Destination**, and **Food** without a photo. A **Draft** never appears on the feed and is not sent to the server until the **Player** submits **Evidence** to create a **Performed Jump** via `POST /v1/jumps`.
_Avoid_: saved jump

**Performed Jump**:
A **Jump** with submitted **Evidence** claiming that the performance happened. During the **Author Grace Period**, the **Player** may edit or retract it; after that, it is locked and can only be escalated to a **Removed Jump**.
_Avoid_: Completed jump

**Judged Jump**:
A **Performed Jump** that has received at least one **Judgment**.
_Avoid_: Rated jump, scored jump


**Judge**:
A **Player** or **Guest Judge** who scores a **Performed Jump** they did not perform. Any **Player** or **Guest Judge** may Judge a **Jump** they did not perform. The performer of a **Jump** is not a **Judge** for that **Jump**.
_Avoid_: Voter, reviewer, rater

**Author Grace Period**:
The 10-minute window after a **Performed Jump** is submitted during which the **Player** may still edit or retract it. The **Judging Window** does not open until the **Author Grace Period** expires.
_Avoid_: Edit window, correction window

**Judging Window**:
The period when a **Performed Jump** can receive **Judgments**, beginning after the **Author Grace Period** expires. On the public feed, the **Judging Window** is open-ended. Season-level judging windows are a v2 concept.
_Avoid_: Voting period, review window

**Judgment**:
A **Judge's** submitted scores for a **Performed Jump**. Each **Judgment** records provenance, either `public`, `open`, or `season`, so public-feed, **Open**, and **Season** scoring can remain separate where required per ADR-0021.
_Avoid_: Vote, review, rating

**Final Score**:
The aggregate score for a **Judged Jump** after its competition period closes, used to update **Standings**. Supperjumpin tracks two score types: a live running average, which updates with each **Judgment** and appears on the feed and detail view; and an **Open Final Score**, computed at monthly **Open** soft-close and stored as `open_final_score`. Public-feed **Judgments** not associated with any competition period contribute only to the live running average and never produce a **Final Score**.
_Avoid_: Total score

**Mission**:
A non-competitive prompt or objective a **Player** can complete for progression or rewards outside competitive scoring. **Missions** may teach the game, encourage participation, or suggest themed **Jumps** without affecting **Standings**.
_Avoid_: Quest, achievement, task

**Action Mission**:
A **Mission** completed through app behavior rather than performing a **Jump**.
_Avoid_: Task

**Guest Judge**:
A visitor who submits a **Judgment** without creating an **Account**. **Guest Judgments** are tracked through `guest_sessions`, contribute to the public running average, and are subject to a soft cap, default 5 **Judgments**, before the server returns `ErrGuestCapReached` and encourages **Account** creation. A **Guest Judge** may create an **Account** at any time to claim their history and become a **Player**; existing **Judgments** are migrated by setting `player_id` and nulling `guest_session_id`.
_Avoid_: Anonymous voter, unregistered judge

**Prompt**:
A reusable **Jump** idea, theme, or constraint that can inspire **Drafts**, **Missions**, or **Bounties**.
_Avoid_: Template, card, challenge

**Bounty**:
A **Mission** with a meaningful reward for completing a **Prompt**. **Bounties** do not affect competitive scoring by default.
_Avoid_: Prize challenge

**Sponsored Bounty**:
A **Bounty** funded or promoted by an external sponsor, such as a restaurant or brand. **Sponsored Bounties** do not affect competitive scoring by default.
_Avoid_: Ad, promotion

**Disqualified Jump**:
A **Performed Jump** removed from **Standings** because it failed **House Rules** or another competition requirement. At MVP, the team may manually exclude a Jump from the **Open** **Standings** without a formal Disqualified status.
_Avoid_: Deleted jump, rejected jump

**Removed Jump**:
A **Jump** hidden from all visibility because of a serious safety, privacy, legal, or platform violation. In v1, a Removed Jump is fully suppressed from the public feed, direct links, and share previews. **Evidence** is preserved in storage for potential appeal review but excluded from all read queries. Deep links return a tombstone page with no **Evidence**, no performer info, and a "Browse Feed" CTA. The performer is notified privately by the team. **Removed Jumps** are distinct from **Disqualified Jumps**, which may remain visible while not affecting **Standings**.
_Avoid_: Deleted jump

**Player**:
A person who participates in Supperjumpin by performing, viewing, or judging **Jumps**.
_Avoid_: User, member, athlete

**Level**:
A non-competitive progression marker earned by a **Player** through **Missions** and participation. **Levels** are separate from **Season Score** and **Standings**.
_Avoid_: Rank, currency

**Account**:
The login identity that owns a **Player**. **Account** identity is separate from in-game **Player** identity so login methods can change without redefining play history.
_Avoid_: User

**Group**:
A v2 concept — an optional bounded set of **Players** who share, view, and judge each other's **Jumps**. Group, Group Membership, Group Admin, and Invite code was removed per ADR-0019; will be rebuilt from scratch for v2.
_Avoid_: League, community, club

**Invite**:
A v2 concept — a way for a **Player** to bring another person into a **Group**. Removed per ADR-0019.

**Share**:
The act of distributing a **Jump** link from the app to an external channel (group chat, social media, etc.). A **Share** surfaces a deep link with a preview card containing the **Evidence** photo, a truncated **Caption**, the running average score, and the **Source**/**Destination**/**Food** summary. The recipient opens the **Jump** detail view directly, where they may **Judge** without creating an **Account**.
_Avoid_: Post, broadcast, distribute

**North Star Metric**:
The primary measure of product health for a given phase. For v1, the North Star is **"Judgments per Jump within 7 days of posting"** — it captures whether the core promise ("take food somewhere it doesn't belong, document it, get judged") is being fulfilled. Supporting health metrics include: **Guest-to-Player conversion rate** (% of **Guest Judges** who create an **Account**), **Share-to-Judge rate** (% of shared links that result in a **Judgment**), and **Jump-to-Open Final Score rate** (% of **Performed Jumps** that earn a score in the monthly **Open**).
_Avoid_: Key metric, KPI, success metric

**Open**:
The platform-run monthly competition, open to all **Players** globally. The **Open** is backed by `opens` and `open_standings`, runs on a fixed calendar cadence, requires no **Season Commissioner**, and soft-closes at month-end. There is no **Submission Window** or **Judging Grace Period**; **Final Scores** are computed from whatever **Judgments** exist at soft-close. A **Jump** may earn an **Open Final Score** independently of any **Season Final Score** it also earns.
_Avoid_: Global Season, public season

**Season**:
A v2 concept — a bounded competition period within a **Group** where **Judged Jumps** contribute to **Standings** and **Awards**. Season code (start, close, finalize) was removed per ADR-0019; `season_id` column retained as nullable provenance field. The **Open** is the v1 competitive context.
_Avoid_: Campaign, tournament

**Standings**:
A ranked view of **Players** by **Final Score** for a given competition period — currently platform-wide for the monthly **Open**.
_Avoid_: Leaderboard

**Award**:
A v2 concept — an end-of-**Season** recognition for a notable pattern or achievement. Not implemented in v1.
_Avoid_: Badge, achievement

**House Rules**:
The boundaries that define what behavior can count as valid Supperjumpin play. A **Jump** may be awkward, absurd, or transgressive, but must not require harassment, deception that causes material harm to a specific person, trespass, behavior that creates imminent risk of physical injury, illegal acts, property damage, animal cruelty, content that contains hate speech or graphic violence, or intentionally preventing a business from operating. The team may remove any **Jump** that violates the spirit of playful, harmless absurdity, even if it does not violate a specific rule.
_Avoid_: Safety policy, moderation policy

**Evidence**:
Material submitted to support that a **Jump** happened as claimed. A minimum submission requires at least one photo and a **Caption**; video, location data, receipts, or additional photos may supplement but are not required. **Evidence** may inform social judgment or later structured scoring.
_Avoid_: Proof, verification

**Caption**:
A **Player's** written context for **Evidence**, explaining what happened and why the **Jump** should count.
_Avoid_: Description, note

**Commitment**:
A scoring factor for how completely the performer sold the bit with a straight face — deadpan execution, institutional seriousness applied to an absurd act.
_Avoid_: Dedication, effort

**Transgression**:
A scoring factor for how strongly a **Jump** violates an expected food or place boundary without becoming harassment, trespass, or harm.
_Avoid_: Rule-breaking, illegality

**Creativity**:
A scoring factor for how novel, thematic, poetic, or absurdly elegant a **Jump** is.
_Avoid_: Originality

**Presentation**:
A scoring factor for how compellingly the **Evidence** captures the **Jump** as a performance.
_Avoid_: Documentation, media quality

**Credibility**:
How well **Evidence** supports that a **Jump** happened as claimed. **Credibility** is distinct from **Presentation**: a **Jump** may be compellingly presented without being fully believable.
_Avoid_: Truth, proof

## Example Dialogue

**Player A**: I performed it: Taco Bell as the Source, Olive Garden parking lot as the Destination, Crunchwrap as the Food. I called `POST /v1/jumps` with photo Evidence and a Caption. It is now a Performed Jump in its Author Grace Period, so I have 10 minutes to edit before the Judging Window opens.

**Player B**: Once the Author Grace Period expires I'll Judge it on Commitment, Transgression, Creativity, and Presentation. I don't need an Account to Judge; as a Guest Judge, my Judgment adds to the public running average.

**Player A**: If enough people Judge it before the monthly Open soft-close, it can earn an Open Final Score. Until then, the feed and detail page show the live running average as Judgments come in.

**Team**: If a Jump violates House Rules in v1, we may remove it entirely as a Removed Jump.
