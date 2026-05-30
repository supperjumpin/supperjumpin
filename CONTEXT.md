# Supperjumpin

Supperjumpin is a social game about inventing and performing absurd food-location combinations inspired by Jon Bois' Supperjumpin article. The game borrows the spirit of the article without treating its exact terminology or rules as binding canon.

## Language

**Jump**:
A playable Supperjumpin attempt centered on taking food associated with one place and consuming or presenting it in another place. A **Jump** is performed, submitted with **Evidence**, and socially judged.
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
A private, local-only Jump concept a **Player** is composing before submitting **Evidence**. A **Draft** may include a **Caption**, **Source**, **Destination**, and **Food** without a photo. A **Draft** never appears on the feed and is not sent to the server until the **Player** submits **Evidence** to create a **Performed Jump**.
_Avoid_: Idea, Planned Jump, saved jump

**Performed Jump**:
A **Jump** with submitted **Evidence** claiming that the performance happened. During the **Author Grace Period**, the **Player** may edit or retract it; after that, it is locked and can only be removed via a **Dispute** or escalation to a **Removed Jump**.
_Avoid_: Completed jump

**Judged Jump**:
A **Performed Jump** that has received at least one **Judgment**.
_Avoid_: Rated jump, scored jump


**Judge**:
A **Player** who scores another **Player's** **Performed Jump**. Any authenticated **Player** may Judge a **Jump** they did not perform; **Group** membership is not required. The performer of a **Jump** is not a **Judge** for that **Jump**.
_Avoid_: Voter, reviewer, rater

**Author Grace Period**:
The 10-minute window after a **Performed Jump** is submitted during which the **Player** may still edit or retract it. The **Judging Window** does not open until the **Author Grace Period** expires.
_Avoid_: Edit window, correction window

**Judging Window**:
The period when a **Performed Jump** can receive **Judgments**, beginning after the **Author Grace Period** expires. On the public feed, the **Judging Window** is open-ended. Within a **Season**, the **Judging Window** closes with the **Judging Grace Period**.
_Avoid_: Voting period, review window

**Judgment**:
A **Judge's** submitted scores for a **Performed Jump**.
_Avoid_: Vote, review, rating

**Final Score**:
The aggregate score for a **Judged Jump** after its competition period closes, used to update **Standings**. A **Jump** may produce an **Open Final Score** (at monthly **Open** soft-close), a **Season Final Score** (when a **Season** finalizes), or both independently. Public-feed **Judgments** not associated with any competition period contribute only to the live running average and never produce a **Final Score**.
_Avoid_: Total score

**Season Score**:
The accumulated score a **Player** earns in a **Season** from non-disqualified **Judged Jumps**.
_Avoid_: Points, rating

**Mission**:
A non-competitive prompt or objective a **Player** can complete for progression or rewards outside **Season Score**. **Missions** may teach the game, encourage participation, or suggest themed **Jumps** without affecting **Standings**.
_Avoid_: Quest, achievement, task

**Action Mission**:
A **Mission** completed through app behavior rather than performing a **Jump**.
_Avoid_: Task

**Guest Judge**:
A visitor who submits a **Judgment** without creating an **Account**. **Guest Judgments** are stored by device/session and contribute to the public running average. A **Guest Judge** may create an **Account** at any time to claim their history and become a **Player**. In v1, all Judging is available to **Guest Judges**; a soft auth cap may be introduced in v2 based on data.
_Avoid_: Anonymous voter, unregistered judge

**Prompt**:
A reusable **Jump** idea, theme, or constraint that can inspire **Drafts**, **Missions**, or **Bounties**.
_Avoid_: Template, card, challenge

**Bounty**:
A **Mission** with a meaningful reward for completing a **Prompt**. **Bounties** do not affect **Season Score** by default.
_Avoid_: Prize challenge

**Sponsored Bounty**:
A **Bounty** funded or promoted by an external sponsor, such as a restaurant or brand. **Sponsored Bounties** do not affect **Season Score** by default.
_Avoid_: Ad, promotion

**Unwitnessed Jump**:
A **Performed Jump** whose **Judging Window** closed without any submitted **Judgments**, so it does not affect **Standings**.
_Avoid_: Unjudged jump, unrated jump

**Dispute**:
A **Player-raised** challenge that a **Performed Jump** may not satisfy **House Rules**, **Credibility**, or claimed **Source**, **Destination**, or **Food**. In v1, Disputes are filed via a simple Report form with 3–4 high-level categories (derived from House Rules) plus an optional text field; adjudication is manual by the team. Formal Dispute tooling is v2.
_Avoid_: Report, flag, appeal

**Disqualified Jump**:
A **Performed Jump** removed from **Standings** because it failed **House Rules** or another competition requirement. A v2 concept; the governance model for adjudicating Disqualified Jumps will be designed when Groups are specified. At MVP, the team may manually exclude a Jump from the **Open** **Standings** without a formal Disqualified status.
_Avoid_: Deleted jump, rejected jump

**Removed Jump**:
A **Jump** hidden from all visibility because of a serious safety, privacy, legal, or platform violation. In v1, a Removed Jump is fully suppressed from the public feed, direct links, and share previews. The performer is notified privately by the team. **Removed Jumps** are distinct from **Disqualified Jumps**, which may remain visible while not affecting **Standings**.
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
An optional bounded set of **Players** who share, view, and judge each other's **Jumps**. In v1, a **Group** is a lightweight social circle (e.g., a group chat) with no formal administration, **Season** requirements, or competitive infrastructure. In v2, **Groups** gain **Seasons**, **Standings**, **Awards**, and the **Season Commissioner** role. **Jumps** exist on a public feed independently of **Groups** in v1.
_Avoid_: League, community, club

**Group Membership**:
The relationship between a **Player** and a **Group**, including that **Player's** participation status or role within the **Group**.
_Avoid_: Membership, subscription

**Group Admin**:
A **Player** with durable authority over a **Group**, including emergency override authority over an **Active Season**. A v2 feature; v1 **Groups** have no formal administration.
_Avoid_: Owner, moderator

**Invite**:
A way for a **Player** to bring another person into a **Group**. In v1, this is an informal share of a **Jump** link to a group chat. In v2, **Invites** become a formal **Group Membership** mechanism with join codes or admin approval.
_Avoid_: Invitation link, referral

**Share**:
The act of distributing a **Jump** link from the app to an external channel (group chat, social media, etc.). A **Share** surfaces a deep link with a preview card containing the **Evidence** photo, a truncated **Caption**, the running average score, and the **Source**/**Destination**/**Food** summary. The recipient opens the **Jump** detail view directly, where they may **Judge** without creating an **Account**.
_Avoid_: Post, broadcast, distribute

**North Star Metric**:
The primary measure of product health for a given phase. For v1, the North Star is **"Judgments per Jump within 7 days of posting"** — it captures whether the core promise ("take food somewhere it doesn't belong, document it, get judged") is being fulfilled. Supporting health metrics include: **Guest-to-Player conversion rate** (% of **Guest Judges** who create an **Account**), **Share-to-Judge rate** (% of shared links that result in a **Judgment**), and **Jump-to-Open Final Score rate** (% of **Performed Jumps** that earn a score in the monthly **Open**).
_Avoid_: Key metric, KPI, success metric

**Share**:
The act of distributing a **Jump** link from the app to an external channel (group chat, social media, etc.). A **Share** surfaces a deep link with a preview card containing the **Evidence** photo, a truncated **Caption**, the running average score, and the **Source**/**Destination**/**Food** summary. The recipient opens the **Jump** detail view directly, where they may **Judge** without creating an **Account**.
_Avoid_: Post, broadcast, distribute

**North Star Metric**:
The primary measure of product health for a given phase. For v1, the North Star is **"Judgments per Jump within 7 days of posting"** — it captures whether the core promise ("take food somewhere it doesn't belong, document it, get judged") is being fulfilled. Supporting health metrics include: **Guest-to-Player conversion rate** (% of **Guest Judges** who create an **Account**), **Share-to-Judge rate** (% of shared links that result in a **Judgment**), and **Jump-to-Open Final Score rate** (% of **Performed Jumps** that earn a score in the monthly **Open**).
_Avoid_: Key metric, KPI, success metric

**Open**:
The platform-run monthly competition, open to all **Players** globally. The **Open** runs on a fixed calendar cadence, requires no **Season Commissioner**, and soft-closes at month-end — **Final Scores** are computed from whatever **Judgments** exist at that moment. A **Jump** may earn an **Open Final Score** independently of any **Season Final Score** it also earns.
_Avoid_: Global Season, public season

**Season**:
A bounded competition period within a **Group** where **Judged Jumps** contribute to **Standings** and **Awards**. A **Group** has at most one active or closing **Season** at a time. Distinct from the **Open**, which is platform-run and requires no **Group** or **Season Commissioner**. **Seasons** are a v2 feature; the **Open** is the v1 competitive context.
_Avoid_: Campaign, tournament

**Season Commissioner**:
The **Player** who starts a **Season** and holds season-scoped authority similar to a fantasy sports commissioner.
_Avoid_: Season owner, league manager, commissioner

**Active Season**:
The current **Season** in a **Group** where **Players** may submit **Jumps** for competition.
_Avoid_: Current season

**Submission Window**:
The phase of an **Active Season** when **Players** may submit **Jumps** for competition.
_Avoid_: Season window

**Judging Grace Period**:
The phase after a **Season's** **Submission Window** closes when no new competition **Jumps** may be submitted, but existing **Performed Jumps** may still receive **Judgments**.
_Avoid_: Overtime, judging window extension

**Finalized Season**:
A **Season** whose **Standings** are locked after the **Judging Grace Period** ends.
_Avoid_: Ended season, closed season, archived season

**Standings**:
A ranked view of **Players** by **Final Score** for a given competition period — either within a **Group** for a **Season**, or platform-wide for an **Open**.
_Avoid_: Leaderboard

**Award**:
An end-of-**Season** recognition for a notable pattern or achievement, which may be based on total score, a specific scoring factor, or a pattern of play that the scoring model cannot capture.
_Avoid_: Badge, achievement

**Unwitnessed Performance**:
An **Award** given at **Season** close to a **Player** whose **Season**-linked **Jump** closed as an **Unwitnessed Jump**. Recognizes commitment without an audience. Does not affect **Season Score** or **Standings**.
_Avoid_: Consolation prize

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

**Player A**: I performed it — Taco Bell as the Source, Olive Garden parking lot as the Destination, Crunchwrap as the Food. I submitted photo Evidence and a Caption. It is now a Performed Jump in its Author Grace Period — I have 10 minutes to edit before the Judging Window opens.

**Player B**: Once the Author Grace Period expires I'll Judge it on Difficulty, Transgression, Creativity, and Presentation. I don't need to be in a Group to Judge — any Player can. If the Evidence looks staged, I may lower Credibility or raise a Dispute. My Judgment adds to the public running average.

**Player A**: Later, I joined a Group with an Active Season and submitted the same Jump to the Season. Only Judgments submitted while it is Season-linked count toward the Season Final Score — the public Judgments it already received stay on the public running average and do not carry over.

**Team**: If a Jump violates House Rules, we may remove it entirely (Removed Jump) or, in v2 when Groups launch, mark it as Disqualified so it stays visible but does not affect Season Standings.
