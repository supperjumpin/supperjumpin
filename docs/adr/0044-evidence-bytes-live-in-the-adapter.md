# Evidence bytes live in the adapter, served as stable URLs; the API stays byte-free

The core API stores Evidence as `evidenceUrls: []string` (`apps/api/openapi.yaml:152`, `dto.go:77`) — never bytes. The Discord adapter owns the bytes: it downloads the uploaded photo from Discord's fresh signed URL at submit time, stores them on local FS, and serves them over a small static file mux, handing the API a stable front-end-agnostic URL. The core stays byte-free; the adapter stays the only one with image-hosting infra.

## Why not the alternatives

Two alternatives were considered:

- **Just store Discord's signed URL and serve it at Reveal time.** Rejected: Discord attachment URLs are signed with a preset expiry (`?ex=...&is=...&hm=...`, see Discord API Reference → "Signed Attachment CDN URLs"). The TTL is not officially documented but is finite. Reveal can be days away (CONTEXT.md: the initiator picks a timeframe from `GET /v1/reveal-timeframes`). A signed URL captured at submit time will be stale by reveal.
- **Use ephemeral Discord attachments and reference `attachment_id`/`message_id` in the stored URL (`discord-attachment://{id}`).** Rejected because it breaks the #319 thesis: the stored URL only makes sense to the Discord adapter — a future web or mobile adapter calling `GET /v1/rounds/{roundId}/recap` could not render it. The API's contract must remain front-end-agnostic.

## The seam

- **Submit (adapter → API):** the `/jump submit` slash command carries the photo as an `attachment` (type 11) option. The bot immediately downloads the bytes from Discord's signed URL (fresh at interaction time), writes to `./.bot-data/evidence/<stableID>.<ext>` (created lazily on bot startup), and serves them over a `http.FileServer` on its own port. The URL handed to `POST /v1/rounds/{id}/jumps` is `http://localhost:<botPort>/evidence/<stableID>.<ext>`.
- **Reveal (adapter reads API):** the bot calls `POST /v1/rounds/{id}/reveal`, then `GET /v1/rounds/{roundId}/recap` (or reads the Jumps from the reveal response). The URLs the API returns are the same stable adapter-hosted URLs. The bot opens the bytes, re-attaches them as `discordgo.File` in the shared announcement embed — at this point Discord uploads a freshly-signed CDN URL with a TTL of seconds-to-minutes, only needing to live long enough for the render to complete.

## Local-first posture, not long-term

The `localhost` URL stored in the core is fine for the proof stage where the bot is the only live consumer and the human-in-loop verification runs on one dev machine. It is **not** a production answer — a web or mobile adapter on another host cannot dereference localhost.

Migration to hosted infra is additive and is covered by ADR-0036 ("Hosted infrastructure will be additive when introduced"): a one-time job will copy FS bytes to S3 / R2 and a later adapter will rewrite existing `evidenceUrls` rows when the bot moves off localhost. The core contract stays unchanged; only adapter deployment config moves.

## Consequences

- The Discord bot is no longer *purely* an HTTP client of the API — it also owns an evidence directory and a tiny static file server. This is more adapter infra than the thinest case but is the price of preserving the front-end-agnostic `evidenceUrls` invariant.
- Persistent adapter state: the evidence directory survives bot restarts. `npm run bot:dev` purges or migrates it; in dev this is a `git clean`-curious directory like `.bot-data/` (gitignore'd, like `.work-issue/`).
- The adapter never links Discord's signed URL anywhere persistent. It is consumed once at submit time, downloaded, then discarded.
- Stale-CDN-URL defense only: at submit time the bot must fetch the bytes before the interaction's signed URL expires (typically sub-minute at the API call). A retry on transient fetch failure is the bot's problem, not the API's.
- New env vars on the bot, not the core: `SUPPERJUMPIN_BOT_EVIDENCE_DIR` (default `.bot-data/evidence`), `SUPPERJUMPIN_BOT_PORT` (default 9999 or unused if served by the same process as the bot's slash-command listener).
- Bot's `go.mod` may need `net/http` static-serving helpers but no new external library — `http.FileServer` is std lib.

## Why ADR-worthy

Hard to reverse (every stored `evidenceUrls` row points at a host; switching hosts later means writing a one-shot migration for existing URLs, which we explicitly plan for); surprising (a reader will ask "why does the bot own a file server"); real trade-off (two documented alternatives rejected for specific reasons). The local-first caveat matches ADR-0036 and is explicitly a proof-stage posture, not a final architecture.

## Status

accepted — settles where Evidence bytes live for the Discord adapter surface, preserves the front-end-agnostic URL contract (#319), and amortizes the hosted-infra decision to a later additive migration.