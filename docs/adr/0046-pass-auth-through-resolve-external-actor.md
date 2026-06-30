# Pass the auth subsystem through ResolveExternalActor — adapter-token + actor-tuple; account-era scaffolding deleted

The competition-era auth subsystem (`accounts` + `auth_identities` + `BootstrapIdentity` + `StaticAuthVerifier` of account-keyed tokens) is vestigial under the pivoted domain. It identifies an **account** that owns a **Player** — but the pivoted domain has no accounts (ADR-0037 deferred them), and Players are Community-scoped via `external_identity`. Slice 1 wired `external_identity` in alongside the old accounts path without removing it; for experiments through Slice 9 the two ran in parallel and tests tolerated the resulting mismatch (an account-Player authenticated by the bearer would then operate on a Community-scoped Player ID resolved separately). This ADR deletes the old path and rewires auth through the Community-scoped identity resolution.

## The seam

### What auth protects changes:
- **Before:** a credential proves *an account owns this Player*. Auth middleware resolves an `Account` + `Player` pair via `BootstrapIdentity`, and that Player ID is used in handlers.
- **After:** a credential proves *an adapter (the Discord bot, v1) is authorized to drive this API*; an accompanying header declares *which platform actor the adapter is vouching for*. Auth middleware resolves the actor to a Community-scoped `(Player, Community)` pair via `ResolveExternalActor`, and that pair is used in handlers.

### The new auth flow

1. **Adapter-level credential:** env var `SUPPERJUMPIN_ADAPTER_TOKEN` (replaces the `SUPPERJUMPIN_DEV_AUTH_TOKEN` semantics in spirit — the env var name in code is a separate, smaller decision folded into the implementation). One shared secret held by the bot. A caller without it is `401`.
2. **Actor tuple on every authenticated request**, carried in the header `X-Adapter-Actor: discord:<guildID>:<userID>`. Read by auth middleware, **not** a per-endpoint OpenAPI parameter.
3. **Auth middleware (`signedInProfile`):** two-step dance:
   1. Verify the adapter bearer against the configured verifier chain. `401` on failure.
   2. Parse the actor tuple from `X-Adapter-Actor`. `400` if missing/malformed.
   3. Call `ResolveExternalActor(platform, serverID, userID, ...)`. The display name fields are best-effort — the actor may already exist (idempotent case) or be new (created case).
   4. Return the resolved `(Player, Community)` pair as the auth profile.
4. **`GET /v1/me`** still exists but returns only the Player (no Account); the Community is also returned because the actor is now Community-scoped.
5. **Both** `optionalProfile` and `signedInProfile` follow the new flow. Public endpoints (`GET /v1/prompt-catalog`, `GET /v1/reveal-timeframes`, `GET /v1/stamp-catalog`) remain unauthenticated — no `X-Adapter-Actor` header needed.

### What gets deleted in the same PR

- Migration: drop `auth_identities` and `accounts` tables; drop `players.account_id` column. (Pre-stable migration discipline per AGENTS.md — fold into the existing `0001_accounts_players.up.sql` and the matching down migration.)
- Code: `internal/httpapi/postgres_store.go::BootstrapIdentity`, the `accounts.sql.go` query file (sqlc-generated), the `accounts.sql` source query, and the `accounts`/`auth_identities` query wrappers. The migration file renames from `0001_accounts_players.up.sql` to `0001_players_communities.up.sql` (or is left as-is, since the rename is cosmetic — preferred: leave the filename to minimize churn).
- `StaticAuthVerifier` map type stays — it's still the dev verifier implementation, but it now maps *adapter tokens* (one entry: `{SUPPERJUMPIN_ADAPTER_TOKEN}`) to a no-op AuthIdentity. The actor comes from the header, not the token.
- `AuthIdentity` struct: loses Email, gains nothing. Becomes a marker that the adapter bearer passed verification. The Community+Player pair comes from `ResolveExternalActor`, not from `AuthIdentity`.
- `Account` DTO and `MeResponse.Account` field are removed from `dto.go`. `MeResponse` gains `Community`.
- All `*_test.go` files that today construct `StaticAuthVerifier{"alice-token": {Provider: "test-prod", Subject: "alice-auth", Email: "alice@example.com"}}` and separately call `store.ResolveExternalActor(...)` are rewritten to pass an adapter bearer + actor header. The `ResolveExternalActor` test setup remains (it's now called *by the middleware*, but tests still need to drive it for first-touch).
- `loggingStore.BootstrapIdentity` and the `bootstrapErr` plumbing in `route_logging_test.go` are replaced with `loggingStore.ResolveExternalActor` / `resolveErr` equivalents.

## What about real OAuth / per-Player Account?

This is explicitly **deferred**, per ADR-0037. The new auth model makes the seam additive: a future Account reintroduction replaces the *adapter-token verifier* with an *OAuth-verifier chain*, and the actor tuple remains. The `external_identity` mapping and `ResolveExternalActor` function don't change when real OAuth lands — they always keyed by `(platform, serverID, userID)` and still do. A later ADR will record the OAuth/Account reintroduction.

## OpenAPI contract change

This is a real change to `openapi.yaml`:
- `components.securitySchemes.bearerAuth` stays as `type: http` `scheme: bearer` but the description changes to "adapter authorization" (the bearer is the adapter's credential, not an account's).
- A new header parameter `X-Adapter-Actor` is declared at the **path-level** (or via a shared component), required on all authenticated routes, not duplicated on each operation. This is the smallest OpenAPI footprint for the contract change.
- The `GET /v1/me` response shape changes: `account` field removed; `community` field added. (Originally this ADR noted the TypeScript client would regenerate via `npm run generate:api-client` and the OpenAPI sync gate would fire; that pipeline is gone as of ADR-0049, so the only consumer-visible change is the spec itself.)

This is the first intentional contract change since the pivot. The contract is being made internally consistent with the pivoted domain (ADR-0037).

## Why not a softer migration

- Keep `BootstrapIdentity` and the accounts tables as a parallel surface "just in case": rejected because the two systems don't reconcile. Their Player IDs are derived from different `stableID` inputs. Maintaining both means forever tolerating the auth-Player ≠ community-Player mismatch (Slice 9's tests demonstrate this). Every future adapter inherits the cognitive load. Deleting now is cheaper than supporting both.
- Add a feature flag to switch between the two: rejected. There are no production users, no compatibility concerns ("this still only lives on my machine"). A feature flag here is process cost without payoff.
- Defer to a follow-on slice (the earlier "Fork C" option the grill considered): rejected by the user on the grounds of long-term leverage, recorded here for posterity.

## Consequences

- This breach of ADR-0041's "consuming, not extending" rule is recorded as a one-paragraph amendment to ADR-0041 (the *only* exception, scoped to "internal inconsistency with the pivoted domain"), not a general precedent. See ADR-0041 amendment in the same slice's PR.
- The Discord bot's auth config becomes: `SUPPERJUMPIN_ADAPTER_TOKEN=<secret>` (shared with the API), and every request sets `X-Adapter-Actor: discord:<guildID>:<userID>`. No per-Player tokens, no deterministically derived HMAC; the actor is vouched for by the adapter, not by a credential the Player holds.
- Test setup style across `apps/api/internal/httpapi/*_test.go` changes: instead of `StaticAuthVerifier{"alice-token": {...}}` + separate `ResolveExternalActor`, tests pass `StaticAuthVerifier{"adapter-token": {}}` + `X-Adapter-Actor: discord:server-1:alice-discord` and the middleware does the resolution. Tests still call `ResolveExternalActor` *before* HTTP requests for cases where they need to assert on first-touch creation, but the *handlers seen by the API* all originate in the header.
- A **new follow-up issue** captures the OAuth/Account reintroduction — additive to *this* auth substrate — as a post-MVP slice. It is marked `ready-for-human` (real OAuth provider integration requires human setup) and depends on #331 for context.

## Status

accepted — concretizes the post-pivot auth subsystem; supersedes the account-based identity flow that was carried forward without cleanup since Slice 0; makes the auth contract internally consistent with ADR-0037.