# apps/bot-discord Guide

Discord bot adapter for Supperjumpin. Talks to `apps/api` over HTTP; the bot owns no game rules, no durable state (except adapter-local artifacts under `.bot-data/`), and no Discord-shaped identifiers in the core. Per ADR-0041, the first front-end; same shape every later adapter (Telegram, web, mobile) must follow.

## Where To Look

| Task | Location | Notes |
|------|----------|-------|
| Entry point / wiring | `cmd/bot/main.go` | Reads env, builds the Wired struct, opens Discord + evidence server |
| Bot core (commands, dispatcher) | `internal/bot/` | Pure bot, no `discordgo` import. Handlers per command, dispatch by `CommandRoute` (slash) or `CustomID` prefix (component) |
| Discord integration | `internal/discord/` | `EventToIncoming` (discordgo → IncomingInteraction), `Responder` (IncomingInteraction → discordgo), `Dispatcher`, env-var `Config`, `Wired` wiring |
| Evidence download + serve | `internal/evidence/` | `Store.Save` downloads from a source URL, content-addresses via SHA256, returns stable `http://localhost:<botPort>/evidence/<hash>.<ext>`. `FileServer` mounted at `/evidence/`. ADR-0044. |
| Reveal scheduler | `internal/scheduler/` | `time.AfterFunc` per active round; file-backed `JSONStateFile` at `<DataDir>/active-reveals.json`; 1-minute `Watchdog` catches missed fires. ADR-0045. |
| Wire it all together | `internal/discord/wire.go::NewWired` | Single composition root — main.go calls this. |
| Bot dev / test commands | `magefile/` (root) | `mage dev:bot` and `mage test` |

## Core Rules

- **Adapter-token + actor-tuple auth** (ADR-0046): every authenticated API call sets `Authorization: Bearer <SUPPERJUMPIN_ADAPTER_TOKEN>` and `X-Adapter-Actor: discord:<guildID>:<userID>`. Public endpoints (`/v1/prompt-catalog`, `/v1/reveal-timeframes`, `/v1/stamp-catalog`) get neither.
- **Pure bot core, thin discord shim**: `internal/bot/` does not import `discordgo`. All Discord types are confined to `internal/discord/`. The translation is in `EventToIncoming` (incoming) and the `Responder` impl (outgoing).
- **Slash commands use `CommandRoute{Name, Subcommand}`** as the dispatch key. Components use `CustomID` prefix (currently only `stamp:` for stamp apply).
- **Hand-rolled API client**: matches the API's stdlib-first convention. No OpenAPI generator. `APIClient.StartRound` / `ApplyReaction` / `SubmitJump` / `ListStampCatalog` are one method per endpoint, hand-synced with `apps/api/openapi.yaml`. Add a `// matches openapi.yaml:<SchemaName>` comment when adding a new request type.
- **Injected clock**: `scheduler` takes a `clock.Clock` (via `github.com/benbjohnson/clock`). Production uses `clock.New()`; tests use `clock.NewMock()`.
- **Adapter-local state** in `.bot-data/` (gitignored): `active-reveals.json` (reveal scheduler), `evidence/<sha256>.<ext>` (downloaded photos). Three artifacts total, all plain JSON or files. No database.

## Avoid

- Importing `discordgo` from `internal/bot/`. All Discord types must live in `internal/discord/`.
- Storing Discord's signed URL anywhere persistent. It expires. Download immediately at submit time.
- Putting game rules in the bot. All rules belong to `apps/api/internal/game/`; the bot is a thin transport.
- Hand-writing the OpenAPI client (mobile did this, the bot doesn't — per ADR-0041, the contract work is upstream).

## Notes

- New env vars (bot only): `SUPPERJUMPIN_BOT_TOKEN` (required), `SUPPERJUMPIN_ADAPTER_TOKEN` (required, shared with API), `SUPPERJUMPIN_API_BASE_URL` (required, e.g. `http://localhost:8080`), `SUPPERJUMPIN_BOT_DATA_DIR` (default `.bot-data`), `SUPPERJUMPIN_BOT_EVIDENCE_ADDR` (default `:9999`), `SUPPERJUMPIN_BOT_EVIDENCE_BASE_URL` (default `http://localhost:9999`), `SUPPERJUMPIN_BOT_APP_ID` (required for command registration; from Discord developer portal), `SUPPERJUMPIN_BOT_GUILD_ID` (optional; if set, commands are registered per-guild, which is faster in dev).
- Sealing invariant (ADR-0043): sealed content never reaches a Discord channel. Pre-reveal submissions send an **ephemeral** reply to the author only; the shared channel sees only the existence-count line.
- Component `custom_id` formats: `stamp:<roundID>:<jumpID>:<stampID>` (stamp apply), `commit:<roundID>` (I'm In). Both parsed in `internal/bot/`.
- Slash commands registered automatically on startup via `Wired.RegisterCommands(appID, guildID)`. Commands: `/round start`, `/round status`, `/jump submit`, `/comment round`, `/comment jump`, `/recap`.
- The reveal scheduler's `OnFire` callback is `RevealActor.Fire`, which calls `POST /v1/rounds/{id}/reveal` then `GET /v1/rounds/{id}/recap` and posts the reveal to the channel. Channel ID is tracked in `RoundRegistry`, populated at `/round start` time.
- A real Discord bot token AND `SUPPERJUMPIN_BOT_APP_ID` are required for end-to-end verification (this slice is `ready-for-human`).
