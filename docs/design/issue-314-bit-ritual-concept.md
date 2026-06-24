# Issue #314 Bit Ritual Concept Brief

> **Status:** Draft exploration for [issue #314](https://github.com/supperjumpin/supperjumpin/issues/314), not a final PRD or build handoff. Use this as input for a fresh grilling pass, then rewrite the PRD and relitigate the current backlog against the resulting product spine.
>
> **Review posture:** The core concept should be challenged hard. Edge details that can be A/B tested should remain configurable rather than treated as settled doctrine.

- **Codename:** `daily-collision` (placeholder)
- **Mode:** refinement
- **Last updated / current state:** 2026-06-23; issue #314 is under ideation. Direction locked: Supperjumpin should center on a bit ritual, not competition. Hybrid cadence is the default hypothesis to validate. Creation is pod-native and public-legible: close friend pods are the ritual container, but shareable artifacts must travel beyond the pod. Scoring moves backstage; expressive comedy reactions are on-stage. The weekly recap/canon artifact is the core weekly success moment. Recap model locked: private episode/scrapbook for the pod, public postcard/trailer for outsiders. Public artifacts are share-forward by default with pre-share review and redaction.
- **Validated against:** repo commit `ad96df5` on 2026-06-23; issue #314, `CONTEXT.md`, product vision, UX design, growth-loop analysis, and Open competitiveness analysis reviewed.
- **Confidence verdict:** 8/10 that #314 identifies the right problem and bit ritual is the right center; 6/10 that hybrid cadence is the right solution without playtest evidence. Would move up if small groups create callbacks, repeat performances, and anticipation for the next Prompt; down if the loop produces polite participation but no group lore.

## The concept

- **One-line promise:** Supperjumpin is a recurring food-bit ritual where close friend pods perform absurd collisions, react expressively, and produce artifacts funny enough to make outsiders want their own pod.
- **Problem + who feels it:**
- Current design puts too much fun in four-axis Judging, which is structurally administrative even when polished.
- The likely early Player wants group-chat absurdity, performer payoff, and reactions to their bit more than a precise score.
- Existing design docs already identify the retention gap: direct share lacks synchronicity, the Open is a weak retention engine, and global competition alone is not enough.
- **Beachhead persona:** People who already do food/location bits for a small group chat and enjoy narrating the audacity afterward. Secondary audience: casual reactors who may never perform daily but can make the performer payoff feel alive.
- **The wedge / differentiator:** A shared Prompt inside a close friend pod creates comparison and collision: the same constraint produces many interpretations, expressive stamps turn reacting into another comedy surface, and the weekly artifact makes the pod's lore legible enough to travel.
- **Market evidence:** Internal research notes compare Supperjumpin against Wordle, BeReal, Strava, TikTok, and leaderboard products. These are directional, not direct demand proof. No external demand validation for daily food-bit prompts has been completed yet.
- **Competitive landscape:**
- Social feeds can host funny food bits, but they do not create a shared daily constraint or structured reaction surface.
- Challenge apps can create prompts, but often over-optimize for tasks, streaks, or creator polish rather than small-group deadpan absurdity.
- Games with leaderboards create competition, but existing analysis says raw global Standings are unlikely to be the fun center.
- **Monetization / sustainability model:** Not decided. For this phase, sustainability means a small manually curated playtest and low-cost operation; no monetization surface before product/retention proof.

## The gate

- **Success metric:** Group-level repeat ritual retention: in a manual playtest, at least one seed group with 6-10 participants completes the ritual in 3 of 4 consecutive weeks, with 3+ people contributing each week and non-performers still reacting.
- **Aha / activation moment:** A performer sees the weekly recap/canon artifact, recognizes their bit as part of the pod's lore, and wants to share it or return for the next Prompt.
- **Kill criterion / smallest proof:** If a 2-4 week group-chat playtest produces no durable callbacks, no anticipation for the next Prompt, and fewer than 3 repeat performers, shelve the pivot or redesign the prompt/reaction model before writing more app code.

## Scope & roadmap

- **Scope IN (v1 candidate, not locked):** Close friend pods as creation context, pod-local weekly scrapbook/episode recap, public-legible postcard/trailer artifact, public-legible Jump artifacts, weekly main Prompt, daily lightweight reaction/recap beats, cost-tiered prompt calendar, expressive verdict-stamps/reactions as the on-stage interaction, backstage scoring/structure for recap texture, performer-facing reaction wall, lightweight manual curation, instrumentation for prompt participation and repeat group ritual.
- **Scope OUT / deferred:**
- The Open as primary retention engine — likely preserved only if it supports the prompt loop rather than leading it.
- Group Seasons — still v2 unless the playtest proves friend-group containment is essential.
- Private league walls — pod context should make creation fun, but artifacts need to be shareable and understandable outside the pod.
- Public feed as the early creative container — deferred until pod ritual proves fun; public can be an inspiration/showcase/distribution layer, not the place that makes the bit work.
- Missions, Levels, Bounties, Sponsored Bounties — premature progression/economy layers before fun is proven.
- Complex scoring UI — rubric moves backstage and must not become the first-class interaction.
- Algorithmic feed — direct share and weekly collision artifacts are the discovery thesis.
- **Locked decisions:**
- LOCKED: This is a product pivot decision, not a feature request — issue #314 may reframe existing backlog and ADR priorities.
- LOCKED: Cheapest validation comes before more hardening — group chat + spreadsheet is a better next artifact than more infrastructure.
- LOCKED: Feel is load-bearing — if the product does not make performing and reacting feel funny, the architecture does not matter.
- LOCKED: Supperjumpin is a bit ritual first, not a competition-first product — competition may season the ritual but cannot be the primary surface.
- LOCKED: Daily should not mean daily performance — the default hypothesis is weekly main Prompt plus daily lightweight reactions, recaps, and callbacks.
- LOCKED: Creation is pod-native and public-legible — close friend pods provide shared context, callbacks, and social bravery, but the output must travel outside the pod as a shareable artifact.
- LOCKED: Scores move backstage — the on-stage interaction is expressive comedy reactions/verdict-stamps, while structured dimensions may survive only as internal recap texture.
- LOCKED: Reaction language should not preserve the old formal Judgment vocabulary by default — react like a funny friend in the pod, not a Judge, court, critic, or review app.
- LOCKED: The weekly recap/canon artifact is the core weekly success moment — it packages the pod's collisions, stamps, callbacks, and standout bits into the payoff that drives return and sharing.
- LOCKED: Recap artifacts split by audience — the pod-local recap is the full private episode/scrapbook; the public share is a curated postcard/trailer that makes outsiders understand the ritual and want their own pod.
- LOCKED: Public postcard drafts are share-forward by default — names/images/captions/stamps/selected quotes can be included unless marked pod-only, but external sharing requires a quick review/redaction step.
- **Open decision-forks:**
- Reaction vocabulary: playtest a small stamp set that covers approval, chaos, craving, commitment, lore, and affectionate failure without becoming generic emoji or rubric homework.
- Public showcase shape: what can outsiders see/react to, and what remains pod-local, so growth is possible without turning the product into a generic public content feed?
- Public artifact permissions details: what exact fields are always excluded from public postcards, and who can redact on whose behalf?
- **Phased roadmap:**
- Phase 1 (manual proof): Run a 2-4 week group-chat playtest with weekly main Prompts, daily lightweight reaction/recap beats, manual reaction collection, a pod-local weekly recap, and a public-safe postcard artifact.
- Phase 2 (decision): Review repeat group ritual, reaction specificity, prompt quality, cost fatigue, group lore, recap/share comprehension, and scoring visibility.
- Phase 3 (design rewrite): Update product vision, UX, and ADRs around the chosen spine; relitigate backlog.
- Phase 4 (build): Only then sequence app changes.
- **Risks + mitigations:** Biggest risk is pod lore, not feasibility. Mitigate by testing whether the ritual creates callbacks, teasing, anticipation, and repeat performances before app code. Secondary risks are prompt quality, cost fatigue, safety/transgression escalation, weak pod formation, share artifacts that are too inside-baseball to travel, and design drift back into administrative scoring or growth hacking.

## Tech & constraints

- **Tech approach:** Existing Expo mobile app, Go API, Postgres, OpenAPI-generated client. Do not commit new build architecture until the playtest selects the product spine.
- **Design load-bearing? (+ direction if yes):** Yes. Target feel: group-chat absurdity packaged with arcade immediacy, not legal/court bureaucracy. The app should make reactions feel like a close room losing it at the bit, while keeping House Rules clear enough to prevent harmful escalation. Acceptance criteria: the first screen explains the active Prompt in one glance; daily beats feel like the pod staying alive, not homework; reacting feels like writing a punchline or stamping a verdict, not evaluating a form; the performer payoff is reaction-dense; the weekly share artifact makes the pod's collisions legible to a stranger without a tutorial.
- **Hard constraints:** House Rules remain load-bearing because Transgression incentives can escalate. Cost must stay low enough that daily ritual does not imply daily purchase/travel. Guest/casual reaction must stay low-friction.
- **Core domain primitives:** Prompt, Jump, Evidence, Caption, Source, Destination, Food, Reaction/Verdict Stamp, backstage Judgment structure, Share, House Rules. Prompt moves from supporting primitive to spine.
- **Source-of-truth rules:** Product truth should be validated by manual playtest first; code and ADRs follow the locked product spine, not vice versa.
- **Experimentability rule:** Nail the core loop first, then tune the edges. Decisions that can reasonably be A/B tested should be designed as configurable surfaces, not hard-coded product doctrine. Core locks are the ritual spine, pod-native creation, public-legible artifacts, backstage scoring, expressive reactions, and recap payoff. Tunable surfaces include reaction label variants, postcard permissions copy/defaults, CTA wording, recap layout, prompt cadence details, and public artifact detail level.

### Recap artifact model

- **Pod-local recap:** the full private episode. Use a story/carousel/scrapbook shape with the week's Prompt, the spread of submissions, stamp clusters, funny disagreements, callbacks, character beats, a canon moment, and a next-week hook. It should feel like the pod's private episode, not a report.
- **Public share artifact:** a curated postcard/trailer from the pod. Use one to five public-safe highlights, the Prompt, a stamp cloud or canon line, light pod branding, and a CTA to `Start this Prompt with your pod` or `Steal this Prompt`. It should drive pod replication, not awkward requests to join the original pod.
- **Permissions model:** the public postcard draft is share-forward by default, not opt-in-only. Names, images, captions, stamps, and selected quotes may appear in the draft unless marked pod-only. Before external sharing, the sharer gets a quick review/redaction step. Exact locations, full member lists, backstage scoring, low-participation signals, and surveillance-feeling timestamps stay private by default.
- **Experiment posture:** postcard permissions are an edge surface, not a core doctrine. Start with a defensible share-forward default plus redaction, but design the surface so opt-out wording, default inclusion, quote handling, and review friction can be tested without changing the core ritual.
- **Keep pod-local:** full comments, private captions, raw inside jokes, who did not participate, full member lists, timestamps/locations that feel surveillant, backstage scoring logic, and any teasing that only works because the pod has trust.
- **Make public-legible:** the Prompt, a clear ritual frame, curated visual moments, public-safe captions/quotes, opt-in attribution, and enough mystery to feel like a postcard from an inside joke.
- **Avoid:** winners, leaderboards, point totals, engagement language, exporting the private recap wholesale, generic social-card polish, and making the artifact more about Supperjumpin branding than the pod's voice.

### Reaction vocabulary direction

- **Working term:** stamps or reactions. Avoid `Judgment`, `verdict`, `ruling`, `court`, `jury`, `case`, `appeal`, and review-app language as core terms. Mock-official language can appear sparingly as a joke, not as the system model.
- **Principles:** friend-pod first; expressive over evaluative; short enough to tap repeatedly; ownable weirdness without disposable meme slang; food/pod/lore identity over generic internet reactions; warm roasts only, no hostile dunking.
- **Starter set to playtest:** Certified, Unhinged, Would Bite, Respect, Lore, Condolences.
- **More personality-forward candidates:** That's Lore, Respectfully Feral, Need This Immediately, Unhinged Pairing, Group Chat Certified, Suspiciously Perfect, No Crumbs Left, Pod Canon, Certified Bit.
- **Reject for now:** Five Stars (review-app gravity), The Court Approves (keeps the old court frame alive), Yummy (too generic), Slay/Skull/W/Rizz/etc. (meme expiry), Bad/Fail/Cringe (too hostile for close friend pods).
- **Playtest rule:** If people pause to decode a stamp, feel like they are grading, or stop riffing after tapping, the vocabulary is wrong. If a stamp becomes a callback in the recap, it is working.

## Refinement-only

- **Parity contract:** Preserve the core Supperjumpin identity: take food somewhere it does not belong, document it, and get judged/reacted to. Preserve House Rules, direct share, and the existing domain language unless playtest evidence shows they block fun. Public-stage visibility remains possible as a showcase/distribution layer, but close friend pods are the early creation context. Preserve four-axis scoring only as backstage structure if it creates better recap/reaction texture; do not preserve it as a primary surface merely because it is implemented or documented.
- **Current state — done vs not-done:**
- Done: domain language, Open concept, Judging model, public-by-default decision, backend/API scaffolding, mobile prototype shell, design docs, growth and competitiveness analyses.
- Not done: proof that Players return, proof that Judging is fun, proof that share artifacts convert, proof that prompts create better collisions than open-ended posting.
- **Pre-mortem:** The pivot fails if it keeps old scoring screens and simply adds prompts on top. It also fails if reactions become generic emoji, if daily beats become homework, if prompt quality is inconsistent, if cheap fridge bits are not funny, if the group creates no lore, or if safety constraints become invisible while transgression incentives remain visible.

## Validated facts vs assumptions

- **Verified:** Issue #314 explicitly frames this as an undecided product pivot. Existing docs identify direct-share explanation gaps, weak global Open retention, and the lack of a strong habitual return mechanism. The domain model already includes Prompt. Current v1 docs place Judging and the Open near the center of the experience.
- **Still a bet:** Weekly Prompts plus daily lightweight beats will be fun rather than chores. Expressive verdict-stamps will outperform visible rubric scoring. Backstage scoring can add texture without contaminating the on-stage experience. Cheap $0 food bits can land as hard as real-world audacity. A weekly collision artifact will be self-explanatory to strangers. The early community wants a bit ritual more than a competition. Small groups can create enough lore to carry retention. Growth can emerge from fun and shareable artifacts without designing the core loop around growth hacks.

## Naming

- Codename: `daily-collision` · Display name: TBD · Slug/domain/bundle-id: unchanged/TBD

## Handoff

> Concept direction partially locked. Next step: hand this brief to a fresh agent to grill the product spine, then update issue #314 / the PRD and determine how the current issue backlog changes. Do not hand this to build execution until that pass is complete.
