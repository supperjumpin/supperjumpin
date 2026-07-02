# The Discord bot owns Reveal scheduling; core stays stateless about who fires it

The core's Reveal is a **condition it evaluates** (ADR-0038): any caller of `POST /v1/rounds/{roundId}/reveal` triggers the check `now >= revealBy`. The core does not own a scheduler, wake anyone, or push callbacks out. The Discord bot owns the timer that fires `POST /v1/rounds/{id}/reveal` at the `revealBy` time the initiator picked, and owns the recovery path if it misses.

## The seam

- **In-process scheduler:** the bot holds an in-memory `time.AfterFunc` per active Round, armed at `StartRound` time using `revealBy` from the API response. On fire, the bot calls `POST /v1/rounds/{id}/reveal` and edits the announcement. No external cron, no daemon, no OS-level timer — `time.AfterFunc` is stdlib Go and the bot already runs as a long-lived process.
- **Bot startup recovery:** on boot, the bot reads `./.bot-data/active-reveals.json` and replays:
  - `revealBy` in the future → re-arm `time.AfterFunc`.
  - `revealBy` in the past → fire `POST /v1/rounds/{id}/reveal` immediately, edit the announcement with a self-aware message ("Reveal a bit late — the emcee was getting coffee"). The arriving-late bit is comedy, not corruption; the round transitions the same way it would have on time.
- **State file updated at start, drained at reveal:** written whenever the bot hears a successful `StartRound`; cleared whenever a Reveal fires. The file is plain JSON, gitignore'd, survives bot restart.
- **Watchdog (defense-in-depth):** a 1-minute `time.Ticker` goroutine walks the in-memory map and, for any timer that should have fires but didn't, fires Reveal manually. Catches Go runtime hiccups or blocked event loops without changing the recovery contract.

## Why not the alternatives

- **Core schedules Reveal.** Rejected: ADR-0038 explicitly keeps the core as a condition-evaluator, not a scheduler. Adding cron-shaped state to the API breaks the thin-domain seam and forces every future adapter (Telegram/web/mobile) to inherit the API's scheduler.
- **External cron / systemd timer / `atd`.** Rejected: host-dependency breaks local-first (ADR-0036), and the bot already has a Go runtime with `time.AfterFunc` available — reaching outside stdlib is pure cost.
- **Ad-hoc `/reveal-now` slash command only, no scheduler.** Rejected: violates #331's explicit AC ("Reveal fires automatically at the scheduled time, so the moment happens even if the initiator is offline"). Defeats the ritual.
- **HTTP callback from the API to the bot.** Rejected: an adapter should not have an inbound surface the core knows about; this is the #319 thesis applied to scheduling too (API → adapter is "tell me what to do", adapter → API is "I'm doing X", never the reverse).

## What this implies about the missing list endpoint

The API has **no `GET /v1/rounds` listing** (`apps/api/openapi.yaml` carries `POST /v1/rounds` and per-round `GET`s only). The bot's startup-recovery therefore relies on **file-backed adapter state** (`.bot-data/active-reveals.json`), *not* on introspecting the core. Two implications:

- **If the state file is intact at startup**, recovery is full — every in-flight Round gets re-armed or fired-late.
- **If the state file is lost** (manual delete, disk error), the bot has no way to discover active Rounds to recover. That round stays unrevealed; the next `/round start` would 409 against the one-active invariant, leaving the Community cold-stuck until someone uses a hypothetical `/reveal-now` slash command (planned, but outside this ADR) or the round is force-completed through admin tooling. This failure mode is **accepted for the MVP / `ready-for-human` verification scope** — the file-backed store is local to the dev machine, near-zero chance of silent loss.
- **Extending the API with a list endpoint to fix this** is rejected here because it breaks the "consuming, not extending" rule from ADR-0041, drags list-pagination/filter schema decisions into the API surface, and is sized for a post-MVP slice when the bot is going to need enumerability for other features (status, history) anyway. (Originally the rationale cited the OpenAPI sync gate; that gate is gone as of ADR-0049, but the structural reason — dragging schema decisions into a slice that doesn't need them — still holds.)

## Consequences

- New bot env vars (not core): `SUPPERJUMPIN_BOT_DATA_DIR` (default `.bot-data`), admission already implied by ADR-0044.
- Adapter owns three durable artifacts: evidence files (ADR-0044), `active-reveals.json`, and the Round-announcement message ID map (ADR-0043). All in one directory, all gitignore'd, all plain JSON-or-files. The adapter's state surface is three files, not a database.
- The bot is no longer purely stateless; it has restart-relevant state. This is the same posture as ADR-0044: more adapter infra than the purely-thin case, paid for because it preserves the front-end-agnostic core contract.
- Late-reveal recovery prose is the bot's voice/UX decision, not the core's. The emcee-coffee line is illustrative, not mandated.
- Future "initiator-triggered Reveal" / "threshold-triggered Reveal" variants (ADR-0038) slot in as additional callers of `POST /v1/rounds/{id}/reveal` — same contract, different trigger. No scheduler change.

## Status

accepted — settles who owns Reveal timing (adapter), how the bot survives restarts (file-backed state), and keeps the core stateless per ADR-0038.