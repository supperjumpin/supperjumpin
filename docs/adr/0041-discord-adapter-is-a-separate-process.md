# The Discord adapter is a separate process over the HTTP API, not an in-process transport

The first front-end adapter — the Discord bot — lives in `apps/bot-discord/` and runs as its own process, talking to the core API over HTTP exactly the way any future front-end (Telegram, web, mobile) will. It is not embedded inside `apps/api`.

## The seam

- **Package:** `apps/bot-discord/` with its own `cmd/bot/main.go`, run via `npm run bot:dev`. The API process and the bot process are independently restartable.
- **Transport:** the bot is an HTTP client of `apps/api`. It calls the same OpenAPI-published endpoints any other client would; the core carries no Discord-shaped identifiers in its process memory beyond the `external_identity` rows the bot wrote (ADR-0037).
- **No core-side code for the adapter.** Adding the bot consumes the contract; it does not extend it. This is the bar every later adapter must clear.

## Why not in-process

Two alternatives were considered and rejected:

- **In-process sub-router inside `apps/api/internal/`.** Faster to build (no HTTP hop), but it breaches the #319 thesis the moment Discord identifiers share a binary with the core. It also tempts future adapters to bypass the OpenAPI contract and call `game.*` directly — the exact drift the OpenAPI sync gate is meant to prevent.
- **Shared library with the bot calling `game.*` directly (same process, same binary, separate `cmd`).** Same drift risk, and harder to walk back than the HTTP seam because the call boundary is implicit.

The independent-process, HTTP-client shape makes "swap the front-end" a literal replacement (`rm -rf apps/bot-discord` and write `apps/bot-telegram`), keeps logs separable, and lets the bot migrate to its own host later without touching the core. The HTTP hop is negligible for a chat-rate workload.

## Caveat: consuming, not extending — one scoped exception

The "consuming, not extending" rule above is the default. There is exactly one exception, recorded here so the rule itself is not silently eroded:

> **When the existing OpenAPI/auth contract is internally inconsistent with the pivoted domain, the PR that surfaces the inconsistency may resolve it in core, in the same PR.**

The first (and so far only) use of this exception is ADR-0046: the competition-era `accounts`/`auth_identities`/`BootstrapIdentity` auth subsystem was carried forward without cleanup since Slice 0, and Slice 1's `external_identity` seam ran in parallel without replacing it. The two paths produced *different* Player IDs for the same human (different `stableID` inputs) and tests tolerated the mismatch through Slice 9. ADR-0046 deletes the old path and rewires auth through `ResolveExternalActor`. This is recorded as a one-time use of the exception scoped to "internal inconsistency with the pivoted domain," not a general precedent — adding normal adapter-driven features (slash-command surfaces, UX decisions, adapter-only state) remains consuming-only and does not invoke this caveat.

## Consequences

- The bot authenticates like any other client using `SUPPERJUMPIN_DEV_AUTH_TOKEN` for dev. Per-Player bearer tokens (one Player per Discord actor) are an open question for the next slice, not pre-decided here.
- A new `apps/bot-discord/AGENTS.md` is required by the root maintenance contract.
- The bot owns its own dependencies (Discord library, scheduler). The core's `go.mod` stays Discord-free.

## Status

accepted — concretizes the ports-and-adapters thesis (ADR-0027) for the first front-end, and sets the shape every later adapter (Telegram, web, mobile) must follow.