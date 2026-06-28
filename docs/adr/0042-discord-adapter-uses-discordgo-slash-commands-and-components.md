# The Discord adapter uses discordgo over slash commands and message components

The Discord bot uses the `discordgo` library exclusively through Discord's **slash commands** and **message components** (buttons, select menus). It does not parse text commands, and it does not use message reactions as an interaction surface.

## The seam

- **Library:** `github.com/bwmarrin/discordgo`. It is the lowest-opinion Go library — Gateway/WebSocket + REST + interactions — and matches the core's std-library-first convention (ADR-0027). A more opinionated framework would impose a routing model that fights the thin-transport rule.
- **Interaction surface:**
  - **Slash commands** for verbs: `/round start`, `/round status`, `/recap`, etc. Discord validates and discovers them; the bot authors no parser and owns no `!prefix` conventions.
  - **Buttons** for one-tap beats: `I'm In` on the Round announcement (→ `POST /v1/rounds/{id}/commits`), `Stamp` taps on each revealed Jump (→ `POST /v1/rounds/{id}/jumps/{id}/reactions`). Each button's `custom_id` carries the targeting IDs.
  - **Select menus** for data-driven choices: the Stamp catalog (`GET /v1/stamp-catalog`, ADR-0034) and the reveal-timeframe menu (`GET /v1/reveal-timeframes`, ADR-0039). The bot renders fetched data into Discord UI; it carries no catalog copy.

## Why not the alternatives

- **Text commands** (`!start round`): the bot would own a parser, prefix collisions, and ambiguity recovery that Discord's slash-command surface already solves. No UX payoff for the maintenance cost.
- **Message reactions (emoji taps) for Stamps:** reactions are emoji-shaped and exist outside the component lifecycle. They cannot cleanly carry a stable `stampId`, they cannot be scoped per-Jump on a multi-Jump reveal message without ambiguity, and Discord does not surface "which user reacted with which emoji on which sub-message" cleanly. Buttons with `custom_id` carrying `jumpId + stampId` are unambiguous and match the API contract.

## Consequences

- The bot registers slash commands per-guild (dev) or globally (later) via Discord's application-command API; this registration is part of the bot's startup, not the core's concern.
- "Sealed content hidden, existence visible" (ADR-0040) is enforced by the bot's *rendering* — pre-Reveal it posts an ephemeral to the author and an existence-count line to the channel; post-Reveal it edits the shared message to show the content. No content arrives in the channel before the Reveal fires.
- Stamp and timeframe catalog drift: changes to the core's seeded data flow through automatically on each Round because the bot fetches per-use, not caches.

## Status

accepted — locks the Discord interaction surface for the first front-end adapter (ADR-0041).