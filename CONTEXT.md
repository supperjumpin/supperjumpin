# Supperjumpin

Supperjumpin is a recurring food-bit ritual, inspired by Jon Bois' Supperjumpin article, played within a group of people who already know each other. Each **Round**, the group gets a shared **Prompt**; everyone privately performs an absurd food-location **Jump**; all performances reveal together; the group reacts with expressive **Stamps** and free-form **Comments**; and a **Recap** packages the result. The game borrows the spirit of the article without treating its exact terminology or rules as binding canon.

There is no scoring, no competition, and no ranking — the payoff is shared laughter and accumulated **Lore**.

## Language

### The ritual

**Community**:
The group of **Players** who play together — people who already know each other, all on a single front-end. The **Community** is the durable container every **Round** happens inside. Membership is the host platform's roster (e.g. who is in the Discord server); the domain does not own a join or invite flow. A **Community** spans exactly one front-end instance.
_Avoid_: Group, League, club, pod, server

**Round**:
One cycle of the ritual within a **Community**, and the central thing the system orchestrates. A **Community** has at most one active **Round** at a time, so "the current Round" is unambiguous. A **Round** moves through: **Players** **commit** to its **Prompt** (becoming **Jumpers**); **Jumpers** privately **submit** sealed **Jumps**; a **Reveal** flips all sealed **Jumps** to visible at once; **Reactions** and **Comments** accrue; a **Recap** is produced. Any **Player** in the **Community** may start a **Round** (the *initiator* is simply whoever did — not a privileged role); since only one **Round** is active at a time, starting is unavailable while one is in flight. The initiator chooses the **Round**'s reveal time from a menu of timeframes; cadence and the available timeframes are tunable, not fixed.
_Avoid_: Game, match, session, week

**Commit**:
The act by which a **Player** becomes a **Jumper** in a **Round** — the visible "I'm In" beat, taken before any **Jump** exists. Committing creates anticipation and is distinct from submitting: a **Jumper** who commits but never submits is a real, distinguishable state (an **Ghost Jumper**), not the same as someone who never played.
_Avoid_: Join, sign up, RSVP

**Ghost Jumper**:
A **Jumper** who committed to a **Round** but never submitted a **Jump** before the **Reveal**. Tracked because the gap between committing and delivering is part of the comedy — the **Recap** may affectionately note who was in but no-showed.
_Avoid_: No-show, dropout, quitter

**Reveal**:
The moment a **Round**'s sealed **Jumps** all become visible at once. The **Reveal** fires when the **Round**'s reveal condition is met; in v1 that condition is a scheduled time, so the **Reveal** is a known, anticipated event the whole **Community** shares. Before the **Reveal**, a submitted **Jump** is *sealed*: its content (**Evidence**, **Caption**) is hidden from everyone but its author, but its existence is visible — the **Community** sees how many **Jumpers** have submitted. Hidden content preserves simultaneous interpretation; visible existence fuels anticipation.
_Avoid_: Unlock, release, publish

**Prompt**:
The shared theme or constraint a **Community** performs for a given **Round** — the setup the **Jumpers** each interpret. The same **Prompt** producing many different **Jumps** is the source of the collision and comedy. A **Prompt** is a first-class, reusable resource (not freetext on a **Round**): authored once, attached to many **Rounds**, carrying its copy, theme, and cost tier. In v1 all **Prompts** are platform-authored and global; **Community**-authored **Prompts** are a deferred seam. When an initiator starts a **Round** they either pick a **Prompt** from the catalog or get a random one; there is no synchronized global prompt calendar in v1.
_Avoid_: Template, card, challenge, mission

**Prompt Pack**:
A curated, themed collection of **Prompts** — the deck a **Round**'s **Prompt** is drawn from (the Cards Against Humanity model: a **Pack** is a deck, a **Prompt** is a card). A **Prompt** belongs to exactly one **Pack**. **Packs** are how the catalog is organized for curation and are the natural future unit for prompt packs and entitlements, though no pricing or paywall exists in the domain yet.
_Avoid_: Deck, set, collection, bundle

**Recap**:
The artifact produced after a **Round**'s **Reveal** that packages the **Round**'s **Jumps**, the standout **Stamps** and **Comments**, any resurfaced **Lore**, and affectionate notes on **Ghost Jumpers** into the **Community**'s payoff. The **Recap** is the asynchronous stand-in for a party-game emcee: it narrates the **Round** and is the core moment that drives return. Its format and narrator voice are presentation, tunable, not domain concepts.
_Avoid_: Summary, report, digest, scrapbook

### Performing a Jump

**Jump**:
A Supperjumpin performance: taking food associated with one place and consuming or presenting it in another. To **Jump** is the core verb of the game. Within a **Round**, a **Jumper** submits a **Jump** with **Evidence** and a **Caption**; it stays sealed until the synchronized reveal.
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

**Evidence**:
Material submitted to show that a **Jump** happened as claimed. A minimum submission requires at least one photo and a **Caption**; video or additional photos may supplement but are not required.
_Avoid_: Proof, verification

**Caption**:
A **Player's** written context for a **Jump** — what happened and why it lands.
_Avoid_: Description, note

### People

**Player**:
The persistent in-game identity of a person who participates in Supperjumpin, belonging to one **Community**. A **Player** does not by itself imply participation in any particular **Round** — a **Player** may watch and react to a **Round** without becoming a **Jumper** in it. The adapter resolves a platform actor to a **Player**; the core domain holds no platform identifiers.
_Avoid_: User, member, athlete

**Jumper**:
A **Player** who has entered a specific **Round** by committing to its **Prompt** — the active, per-**Round** role. "Six Jumpers this week" counts the **Players** who jumped, not those who only watched or reacted. A **Jumper** is someone who jumps.
_Avoid_: Entrant, participant, contestant

### Reacting

**Reaction**:
A **Player** applying a **Stamp** to a **Jump** at or after a **Round**'s synchronized reveal. **Reactions** are the front-of-house, one-tap expressive interaction; they feed **Lore** by density and give the **Recap** its texture. There is no scored alternative — a **Reaction** is expressive and never produces a score or ranking.
_Avoid_: Vote, rating, like, judgment

**Stamp**:
The kind of a **Reaction** — the stance a **Player** expresses on a **Jump** (e.g. approval, appetite, chaos, lore, certification, affectionate failure). A **Stamp**'s stable identity is its stance; its display label, glyph, and copy are tunable data, not fixed terms, and are expected to change often. The set of **Stamps** is a seeded catalog, not a hard-coded enum.
_Avoid_: Emoji, badge, score

**Comment**:
A **Player**'s free-form written riff on a **Jump** or **Round** — the open comedy surface alongside **Stamps**. **Comments** are freeform and uncounted; **Stamps** carry the countable signal that feeds **Lore**.
_Avoid_: Reply, post, note

**Lore**:
A **Community**'s durable shared memory: standout moments, quotes, and bits that have accrued across **Rounds**. **Lore** is **emergent and derived** — computed from **Reaction**/**Stamp** density, never nominated, voted on, or written to directly. There is no "canonize" action; the **Recap** is where **Lore** resurfaces. **Lore** is keyed to moments, not to **Players** — it is never a per-**Player** tally. A **Community** may give its own **Lore** a display name; that name is presentation, not a domain concept.
_Avoid_: Canon, Hall of Fame, Highlights, Greatest Hits

## Example Dialogue

**Player A**: This **Round**'s **Prompt** is "fine dining from your fridge." I jumped it — Taco Bell as the **Source**, my dining room as the **Destination**, a Crunchwrap plated under a cloche as the **Food** — and submitted my **Evidence** photo and **Caption**. It stays sealed until reveal.

**Player B**: I didn't jump this week, but at reveal I'll be reacting. When everyone's **Jumps** drop together I'll **Stamp** the ones that get me and riff in the **Comments**.

**Player A**: The **Stamps** and **Comments** are the whole payoff — no scores, no winner. The bits that get the most reaction become our **Lore**, and tomorrow's **Recap** will resurface them.
