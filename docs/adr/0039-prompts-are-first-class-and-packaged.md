# Prompts are first-class reusable resources, organized into Packs

The **Prompt** is the spine primitive of the ritual (per the #314 brief, it moves "from supporting primitive to spine"). It is modeled as a first-class, reusable resource — not a freetext field on a **Round** — and Prompts are grouped into **Prompt Packs**.

## Shape

- A **Prompt** is its own entity with stable identity, authored and stored independently, then *attached* to a **Round**. Many **Rounds** (across **Communities** and time) may reference the same **Prompt**. It carries metadata: the prompt copy, theme, and a cost tier (e.g. $0 fridge bit vs. real-world travel).
- A **Prompt Pack** is a first-class grouping of **Prompts** — a curated themed deck (the Cards Against Humanity model: a Pack is a deck, a **Prompt** is a card). **A Prompt belongs to exactly one Pack**, which keeps a Pack a clean unit for curation now and entitlements later.

## Authorship and scope (v1)

- **The catalog is platform-curated and global.** All **Prompts**/**Packs** are platform-authored; every **Community** draws from the same global catalog. Prompt quality is the brief's #1 named risk, and a single-owner curated catalog is the strongest lever on it at MVP.
- **Community authoring is a deliberate seam, not built yet.** The "Cards Against Humanity blank cards" capability — a **Community** writing its own **Prompts** — is designed for but deferred. Because **Prompt**/**Pack** are already first-class, adding a `communityId` scope (null = global) later is additive, no contract reshape.

## Selection: per-Round, organizer-chosen or random (not a platform calendar)

How a **Prompt** attaches to a **Round** is an **orchestration concern, not core domain**: a **Round** simply holds a `promptId`. When an initiator starts a **Round**, they either **pick a Prompt** from the catalog or request a **random** one. Random selection is an orchestration affordance (choose a random catalog **Prompt**), not a new domain concept.

A global, synchronized **prompt calendar** (Wordle-style — every **Community** gets the same **Prompt** this week) was considered and **deferred**. It is a platform-scale feature whose payoff (a shared cultural moment) only arrives with many **Communities**; the MVP is a few trusted groups who start **Rounds** when ready. Because selection lives in orchestration and the core only stores `promptId`, adding a platform **Schedule/Calendar** later is additive — the seam is deliberately left open to test, without building it now.

## Monetization is deferred, not foreclosed

The **Pack** is the natural future revenue unit (prompt packs, custom-prompt entitlements, paywalled decks). **Pricing, entitlements, and any paywall are explicitly out of the MVP domain.** The Pack grouping is justified independently of revenue: curation itself wants to be organized into themed decks, and "here's this week's deck" is good product on its own. Modeling the grouping now (vs. retrofitting it) is the cheaper time to do it.

## Status

accepted
