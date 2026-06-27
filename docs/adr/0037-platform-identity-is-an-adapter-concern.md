# Platform identity and membership are adapter concerns; the core domain is platform-pure

The core domain knows only **Player**, **Community**, **Round**, and **Jump** — addressed by opaque internal IDs. It contains no platform-shaped identifiers. A separate adapter owns the mapping from a platform-native actor to a domain **Player** within a **Community**.

## The seam

- **Core domain:** `Player` and `Community` carry no Discord/Telegram/platform fields. A **Player** belongs to a **Community**; that is all the core knows.
- **Adapter:** an `external_identity` mapping keyed by `(platform, platform_server_id, platform_user_id)` resolves to `(player_id, community_id)`. For the MVP, `platform = discord`, `platform_server_id = guild id`. Adding Telegram (or web/mobile) later adds adapter rows, not core changes.

## Membership and Player creation

- **Membership is delegated to the host platform.** A **Community** corresponds to a single front-end instance (one Discord server for the MVP); you are a member iff the platform's roster says so. The domain does not own a join/invite flow.
- **The core exposes an idempotent "ensure a Player exists for this actor in this Community" operation** and nothing more. It returns the same **Player** whether called for the first or hundredth time. There is no join state machine, no pending/joined status — taking sides on join semantics is exactly what the domain must not do.
- **Ambient vs. explicit join is an adapter/UX choice, deliberately left open to test into.** The adapter may call the ensure-operation on first interaction (ambient), on an explicit `/join` command, or support both at once. The domain is indifferent. (The removed v1 **Invite** from ADR-0019 is not resurrected.)

## Account is deferred, not foreclosed

The MVP has no **Account** concept; **Player** is the top of the identity chain and is owned by the **Community** instance, not by a login. This is safe to defer: adding an **Account** later (an auth-agnostic identity that owns **Players** across logins, needed once native web/mobile login exists) is *additive* — it slots above **Player** without restructuring it. The contract must not assume **Player** is permanently the top of the identity chain.

## Constraints this locks in

- **No cross-platform identity for now.** The same human acting on two different platforms is not reconciled into one **Player**. All **Players** in a **Community** share one front-end. Cross-platform identity is explicitly deferred, not designed-for (avoiding premature generality).
- A **Community** spans exactly one front-end instance in the MVP.

## Why

The owner's goal is one stable API with interchangeable front-ends (Discord bot first, then Telegram/web/mobile), and a trusted-group MVP where everyone is in the same front-end. Delegating membership to the platform roster and keeping platform IDs in an adapter gives the cleanest hexagonal seam (ADR-0027): the core never imports anything platform-shaped, so a new front-end is a new adapter and the domain contract is untouched. Designing cross-platform identity now would be premature generality for a single-host MVP.

## Status

accepted — concretizes the hexagonal boundary (ADR-0027) for identity/membership and replaces the removed Group/Invite model (ADR-0019) with platform-delegated membership.
