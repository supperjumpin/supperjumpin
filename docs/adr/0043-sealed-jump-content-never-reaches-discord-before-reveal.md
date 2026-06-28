# Sealed Jump content never reaches a Discord channel before Reveal

The core's sealing invariant (ADR-0040 — content hidden, existence visible) is enforced by the **adapter's rendering**, not by Discord's cosmetic spoiler tags or a permission-managed side channel. Pre-Reveal, Jump content does not exist on any Discord channel the Community can read. The bot shows the author an ephemeral confirmation and updates a shared existence-count line.

## The seam

- **Submit flow:** a Jumper uses the `/jump submit` slash command, which carries two options — an `attachment` option (Discord option `type: 11`, native file upload) for the photo and a `string` option for the caption. Discord modals only accept text inputs (no file-upload component exists), so the upload surfaces via the slash command's options, not a modal. On submit the bot calls `POST /v1/rounds/{id}/jumps`, then replies to that user with an **ephemeral** message only they see ("Your Jump is sealed until <reveal time>"). The shared channel receives nothing new from this submission. (Where Evidence bytes live is settled in the follow-on ADR on adapter-hosted Evidence.)
- **Existence-visible:** the bot edits the shared Round announcement's embed — its `Jumpers` line moves from `2 of 5 submitted` to `3 of 5 submitted`. Count is the only channel-visible artifact of a sealed submission.
- **Reveal:** when the bot fires `POST /v1/rounds/{id}/reveal` and the round transitions to revealed, it edits/replaces the announcement to expose the Jumps — each Jump as its own message or embed with caption + attached photo, each rendered with `Stamp` buttons (per ADR-0042).

## Why not the alternatives

- **Spoiler-tagged message in the channel.** Discord spoiler tags are a client-side cosmetic — the content is on Discord's servers and the channel. A modded client, an admin pulling audit logs, or a bot with `read-message-history` can bypass it trivially. It violates the actual invariant, not just its appearance; a "reveal leak" would be embarrassing on day one.
- **Role-locked hidden channel per Round.** Adds a side-channel per Round, fights "membership is the host platform's roster" (ADR-0037) by making the bot a permission system the domain disowns, and still leaks if any role-holder shares. Also fails the same spoilering problem since the content is in the channel.
- **DM threads with each Jumper.** Plausible, but adds a DM-channel-open step per player, breaks for players with DMs disabled, and provides no existence-visible render line at all — every submit would be silent to the channel, killing the anticipation beat CONTEXT.md explicitly protects ("existence fuels anticipation").

## Consequences

- **No content reaches a Discord channel before Reveal.** This is invariant code, not UX preference: the Round announcement message stays content-free until the bot rewrites it on `200` from `POST /v1/rounds/{id}/reveal`.
- The existence-count render means the bot must keep going state for the Round announcement message ID — it edits that message as submissions accrue. This state lives in the adapter only (a `roundAnnouncementMessageID` map keyed by `roundID`), not the core.
- Reveal is the one ceremony that touches content. Because the bot calls the reveal endpoint then immediately edits the announcement, a Reveal crash between the two needs a reconcile path — covered separately in Q5 (the wall-clock + reveal-event UX ADR).
- Discord modals accept text inputs only; photo upload surfaces via the slash command's `attachment` (type 11) option. The submit handler reads the attachment URL from the interaction payload and resolves where the bytes live per the follow-on ADR on adapter-hosted Evidence.

## Status

accepted — concretizes the sealing invariant (ADR-0040) for the Discord adapter surface, and ties the adapter's edit-the-announcement pattern to "existence visible" being the only pre-Reveal channel artifact.