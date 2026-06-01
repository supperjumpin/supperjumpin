# Backend/Data Architecture

_Part of the [Supperjumpin Design Package](./README.md). Depends on: [Product/UX Design](./02-product-ux-design.md) (#106). Parent tracker: #107._

## 1. Architecture Overview

The backend follows a **ports-and-adapters (hexagonal) architecture** with three concentric layers: domain core, application services, and infrastructure adapters. The domain core (`internal/game/`) owns all game rules and defines repository interfaces (ports). Infrastructure adapters (`internal/httpapi/`) provide concrete implementations for persistence (Postgres via sqlc) and transport (HTTP REST). Application services sit between them, orchestrating domain commands and cross-cutting concerns.

### 1.1 Layer Structure

```
┌──────────────────────────────────────────────────────┐
│  Infrastructure Adapters                             │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │
│  │ HTTP REST    │  │ PostgresStore│  │ Object Store │ │
│  │ (server.go)  │  │ (postgres_   │  │ (Evidence)   │ │
│  │              │  │  store.go)   │  │             │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬──────┘ │
│         │                 │                  │       │
│  ┌──────┴─────────────────┴──────────────────┴─────┐ │
│  │  Application Services (app layer)               │ │
│  │  - Command orchestration                        │ │
│  │  - DTO assembly / read model queries            │ │
│  │  - Authorization checks (delegated to domain)   │ │
│  └──────────────────────┬─────────────────────────┘ │
│                         │                             │
│  ┌──────────────────────┴─────────────────────────┐ │
│  │  Domain Core (internal/game/)                  │ │
│  │  - Entities: Jump, Judgment, Evidence, etc.    │ │
│  │  - Repository interfaces (ports)               │ │
│  │  - Game rules, invariants, state transitions   │ │
│  └────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────┘
```

### 1.2 Current State vs. Target State

The codebase already has a nascent hexagonal structure, but several gaps must be closed:

| Aspect | Current | Target |
|--------|---------|--------|
| Domain boundaries | Entities in `internal/game/`; some business logic in top-level functions | Formalize domain commands per aggregate (Jump, Judgment, Season, Dispute) |
| Application service layer | Implicit — handler helpers in `store.go` call domain functions and assemble DTOs | Explicit `Application` struct per aggregate with command methods and read model queries |
| Repository interfaces | `Persistence` monolith combining all repo interfaces + transport-layer DTO queries | Split into granular ports per aggregate: `JumpWriteRepo`, `JumpReadRepo`, `JudgmentWriteRepo`, etc. |
| Read/write separation | Single `Persistence` interface used for both | Separate read models (feed queries, standings) from write repositories |
| Transport DTOs | Mixed into `store.go` alongside MemoryStore | Move DTOs to `internal/httpapi/dto.go`; keep store focused on adapter logic |
| Error mapping | `mapGameErr` bridges domain errors to HTTP errors in `store.go` | Keep error mapping in HTTP handler layer, not store layer |

### 1.3 Aggregate Roots and Bounded Contexts

The domain is organized around these aggregate roots:

| Aggregate | Identity | Key Invariants | v1 Scope |
|-----------|----------|----------------|----------|
| **Jump** | `id` | Lifecycle states, Author Grace Period, visibility rules | Full |
| **Judgment** | `id` | One per Judge per Jump, 1–4 scale, provenance | Full (with Guest support) |
| **Evidence** | `id` | One per Jump, upload authorization required | Full |
| **Dispute** | `id` | Category validation, resolution authority | Simplified (Report only) |
| **Open** | `year-month` | Monthly calendar, soft-close, Standings | Full |
| **Season** | `id` | One active per Group, Commissioner authority | Deferred to v2 |
| **Group** | `id` | Membership, invites | Deferred to v2 |

In v1, Season and Group aggregates are **frozen** — the existing code and schema survive but receive no new features. The Open is a new aggregate introduced in v1.

---

## 2. Jump Lifecycle — States, Transitions, and Persistence

The Jump lifecycle is the central state machine. Each transition has specific persistence and business-rule implications.

### 2.1 Jump States

| State | Description | Visible on Feed | Can Receive Judgments | Persistence Notes |
|-------|-------------|-----------------|----------------------|-------------------|
| **Draft** | Local-only concept; never sent to server | No | No | Not persisted server-side. Exists only on the client. |
| **Performed Jump** | Evidence submitted; in Author Grace Period | Yes (with "Editing" badge for performer; "Judging Window opens in [countdown]" for others) | No — Judging Window blocked during Grace Period | `jumps.status = 'Performed Jump'`, `jumps.grace_period_expires_at = created_at + 10min` |
| **Performed Jump** (post-grace) | Grace Period expired; Judging Window open | Yes | Yes | Same row; Grace Period expiry is computed, not stored as a status change |
| **Judged Jump** | Has received at least one Judgment | Yes | Yes (Judging Window remains open on public feed) | `jumps.status = 'Judged Jump'`, `jumps.final_score` set only at competition close |
| **Unwitnessed Jump** | Judging Window closed with zero Judgments (Season context only) | Yes | No | Only meaningful in Season context; on public feed, the Judging Window is open-ended |
| **Removed Jump** | Hidden due to safety/legal violation | No (tombstoned) | No | `jumps.status = 'Removed Jump'`; Evidence suppressed from all queries |
| **Disqualified Jump** | Excluded from Standings but visible | Yes | No | Deferred to v2 (requires Group/Season governance model) |

### 2.2 State Transition Diagram

```
Draft (client-only)
  │
  ▼ submit Evidence
Performed Jump (Grace Period active)
  │
  │ grace_period_expires_at passes
  ▼
Performed Jump (Judging Window open)
  │
  ├────────────────── first Judgment received ──────► Judged Jump
  │                                                    │
  │                                                    │ (Judging Window stays open on public feed)
  │
  ├── Report ► Team removal ──────────────────────► Removed Jump (tombstoned)
  │
  └── (Season context: Judging Window closes) ───► Unwitnessed Jump (v2)

Disqualified Jump (v2) branches from Judged Jump via Dispute resolution.
```

### 2.3 Persistence Implications

**New column: `grace_period_expires_at`**

```sql
ALTER TABLE jumps ADD COLUMN grace_period_expires_at TIMESTAMPTZ NOT NULL DEFAULT (created_at + interval '10 minutes');
```

This replaces the computed approach. Storing the expiry timestamp explicitly enables:
- Indexable queries: "find Jumps whose Grace Period has expired but status hasn't been updated"
- No clock-skew issues between application and database
- Clear semantics for the Author Grace Period query

**Status constraint update:**

The `jumps.status` CHECK constraint must accommodate the new v1 states. The `Idea` and `Planned Stunt` states are deprecated (v1 Players submit Performed Jumps directly per MVP Roadmap), and `Disqualified Jump` is reserved for v2:

```sql
CHECK (status IN ('Performed Jump', 'Judged Jump', 'Unwitnessed Jump', 'Removed Jump', 'Disqualified Jump'))
```

`Draft` is not a server-side status and is not stored.

**Removed Jump tombstoning:**

When a Jump transitions to `Removed Jump`:
- The row stays in `jumps` with `status = 'Removed Jump'`
- The related `evidences.media_object_key` is retained (for potential appeal) but excluded from all read queries
- A tombstone record is served for deep links: "This Jump has been removed"
- All feed queries filter `WHERE status != 'Removed Jump'`

---

## 3. Data Model Changes — Group-First to Public-First

### 3.1 Table Rename: `stunts` → `jumps`

ADR-0020 mandates the rename from Stunt to Jump. Since there is no production data, the table is renamed outright in a new migration:

```sql
ALTER TABLE stunts RENAME TO jumps;
-- All foreign key columns: stunt_id → jump_id
-- All indexes: stunts_* → jumps_*
-- All constraints: stunts_* → jumps_*
```

Every `stunt_id` foreign key column across the schema is renamed to `jump_id`. This includes:
- `evidence_upload_authorizations.stunt_id` → `jump_id`
- `evidences.stunt_id` → `jump_id`
- `judgments.stunt_id` → `jump_id`
- `disputes.stunt_id` → `jump_id`

Index and constraint names follow the same rename pattern.

### 3.2 Modified Tables

| Table | Change | Rationale |
|-------|--------|-----------|
| `jumps` | `group_id TEXT NOT NULL` → `group_id TEXT` (nullable) | Jumps are public by default (ADR-0019); a Jump may belong to zero Groups |
| `jumps` | Add `grace_period_expires_at TIMESTAMPTZ NOT NULL` | Author Grace Period tracking (Product/UX Design §1.4) |
| `jumps` | Status CHECK: remove `Idea`, `Planned Stunt`; add v1 states | v1 has no pre-commitment flow; Players submit Performed Jumps directly |
| `jumps` | `final_score INT` → `open_final_score INT` + `season_final_score INT` | A Jump may earn separate Open and Season Final Scores (ADR-0023) |
| `judgments` | `player_id TEXT NOT NULL` → `player_id TEXT` (nullable) | Guest Judges submit Judgments without a `player_id` |
| `judgments` | `difficulty` → `commitment`, `documentation` → `presentation` (done) | ADR-0020, ADR-0022 |
| `judgments` | CHECK constraints: 0–10 → 1–4 scale | ADR-0022: forced-choice 1–4 named-tier scale |
| `judgments` | Add `guest_session_id TEXT` | Links Guest Judge Judgments to their session |
| `judgments` | Add `provenance TEXT NOT NULL` | Tracks whether Judgment was submitted in public-feed or Season-linked context (ADR-0021) |
| `judgments` | Add `open_month TEXT` | Tracks which Open month a Judgment belongs to (ADR-0023) |
| `judgments` | UNIQUE constraint: `(jump_id, player_id)` → `(jump_id, player_id)` where `player_id IS NOT NULL` + `(jump_id, guest_session_id)` where `guest_session_id IS NOT NULL` | Guest Judges and authenticated Players have separate uniqueness scopes |

### 3.3 New Tables

| Table | Purpose | Key Columns |
|-------|---------|-------------|
| `guest_sessions` | Tracks Guest Judge identity across Judgments within a session | `id TEXT PK`, `device_fingerprint TEXT`, `judgment_count INT NOT NULL DEFAULT 0`, `created_at TIMESTAMPTZ`, `soft_capped_at TIMESTAMPTZ` |
| `opens` | Monthly competition tracking; one row per calendar month | `year_month TEXT PK` (e.g., `'2026-05'`), `status TEXT NOT NULL CHECK (status IN ('Active', 'Closed'))`, `soft_closed_at TIMESTAMPTZ`, `created_at TIMESTAMPTZ` |
| `open_standings` | Snapshot of Player rankings per Open month | `year_month TEXT REFERENCES opens(year_month)`, `player_id TEXT REFERENCES players(id)`, `open_score INT NOT NULL`, `judged_jumps INT NOT NULL`, `rank INT NOT NULL`, `PRIMARY KEY (year_month, player_id)` |

### 3.4 Deprecated Tables (v1 → v2 freeze)

These tables survive in the schema but receive no new features or API endpoints in v1:

| Table | Status | Notes |
|-------|--------|-------|
| `groups` | Frozen | No v1 endpoints beyond existing schema; Groups are v2 |
| `group_memberships` | Frozen | Same |
| `invites` | Frozen | v1 uses Jump share links, not Group invites |
| `seasons` | Frozen | The Open replaces Seasons in v1 |
| `season_history` | Frozen | Same |

No data is deleted. The tables remain queryable for potential v2 reactivation, and the existing domain code in `internal/game/season.go` and `internal/game/group.go` is preserved but not extended.

### 3.5 Unchanged Tables

| Table | Notes |
|-------|-------|
| `accounts` | Unchanged |
| `auth_identities` | Unchanged |
| `players` | Unchanged |
| `evidence_upload_authorizations` | Column rename `stunt_id` → `jump_id` only |
| `evidences` | Column rename `stunt_id` → `jump_id` only |
| `disputes` | Column rename `stunt_id` → `jump_id` only; Report flow simplified per ADR-0024 |

---

## 4. API Contract Changes

All changes follow ADR-0004 (REST/OpenAPI contract). The API versioning remains `/v1/`.

### 4.1 New Endpoints

| Method | Path | Purpose | Auth Required |
|--------|------|---------|---------------|
| `GET` | `/v1/feed` | Public chronological feed of Jumps (paginated, 20 per page) | No |
| `GET` | `/v1/jumps/{jumpID}` | Jump detail with Evidence, running average, score breakdown | No |
| `POST` | `/v1/jumps` | Create a Performed Jump (photo + Caption + Source/Destination/Food) | Yes |
| `POST` | `/v1/jumps/{jumpID}/evidence` | Authorize and submit Evidence for a Jump | Yes |
| `POST` | `/v1/jumps/{jumpID}/judgments` | Submit a Judgment (4 factors, 1–4 scale) | No (Guest allowed) |
| `POST` | `/v1/jumps/{jumpID}/retract` | Retract a Jump during Author Grace Period | Yes (performer only) |
| `PATCH` | `/v1/jumps/{jumpID}` | Edit Caption during Author Grace Period | Yes (performer only) |
| `GET` | `/v1/opens/{yearMonth}/standings` | Open Standings for a given month | No |
| `GET` | `/v1/players/{playerID}` | Player profile with Jump history | No |
| `POST` | `/v1/jumps/{jumpID}/reports` | Report a Jump for moderation (4 categories + Other) | Yes |
| `POST` | `/v1/guest-sessions` | Initialize a Guest Judge session | No |

### 4.2 Modified Endpoints

| Method | Path | Change | Rationale |
|--------|------|--------|-----------|
| `POST /v1/jumps/{jumpID}/judgments` | Request body: `difficulty` → `commitment`, `documentation` → `presentation` (done), scale 1–4 | ADR-0022 |
| `POST /v1/jumps/{jumpID}/judgments` | New optional `guestSessionId` field | Guest Judges without `player_id` |
| `GET /v1/jumps/{jumpID}` | Response includes `openFinalScore`, `seasonFinalScore` (nullable), `gracePeriodExpiresAt` | Multi-score model, Grace Period |

### 4.3 Deprecated Endpoints (v1 → v2 freeze)

These endpoints survive in the router but are not extended. They may be removed or reworked in v2:

| Method | Path | Status |
|--------|------|--------|
| `POST` | `/v1/groups` | Frozen — no v1 changes |
| `GET` | `/v1/groups` | Frozen |
| `GET` | `/v1/groups/{groupID}/home` | Frozen |
| `POST` | `/v1/groups/{groupID}/invites` | Frozen |
| `POST` | `/v1/invites/{token}/accept` | Frozen |
| `POST` | `/v1/groups/{groupID}/seasons` | Frozen |
| `POST` | `/v1/seasons/{seasonID}/close-submissions` | Frozen |
| `POST` | `/v1/seasons/{seasonID}/finalize` | Frozen |
| `GET` | `/v1/seasons/{seasonID}/history` | Frozen |
| `POST` | `/v1/groups/{groupID}/ideas` | Frozen — v1 has no Idea/Planned Jump flow |
| `POST` | `/v1/stunts/{stuntID}/plan` | Deprecated — replaced by `POST /v1/jumps` |

### 4.4 Feed Query Design

The public feed endpoint (`GET /v1/feed`) requires efficient querying of Jumps without a Group filter. Key design decisions:

**Indexing strategy:**

```sql
-- Primary feed query: recent Performed/Judged Jumps, excluding Removed
CREATE INDEX jumps_feed_idx ON jumps (created_at DESC)
    WHERE status IN ('Performed Jump', 'Judged Jump');

-- Player profile: Jumps by a specific Player
CREATE INDEX jumps_player_created_idx ON jumps (player_id, created_at DESC)
    WHERE status IN ('Performed Jump', 'Judged Jump');
```

The partial index on `jumps_feed_idx` filters out Removed Jumps at the index level, avoiding a full-table scan for every feed page. The `created_at DESC` ordering matches the reverse-chronological feed requirement.

**Pagination:**

Cursor-based pagination using `created_at` as the cursor. Each page returns 20 Jumps. The cursor is the `created_at` of the last Jump on the current page, and the next page queries `WHERE created_at < :cursor`.

This is preferred over offset-based pagination because:
- New Jumps posted between page fetches don't cause duplicates or missed items
- Performance is constant regardless of page depth (index seeks, not scans)

### 4.5 Jump Detail Response Shape

The Jump detail response must carry multiple scores and time-sensitive state:

```json
{
  "id": "jump_abc",
  "performer": { "id": "player_1", "displayName": "Alice" },
  "source": "Taco Bell",
  "destination": "Olive Garden",
  "food": "Crunchwrap Supreme",
  "caption": "Crunchwrap devoured in the Olive Garden parking lot...",
  "evidence": {
    "mediaObjectKey": "uploads/abc.jpg",
    "caption": "Crunchwrap devoured in the Olive Garden parking lot..."
  },
  "status": "Judged Jump",
  "gracePeriodExpiresAt": null,
  "runningAverage": 3.2,
  "judgmentCount": 12,
  "openFinalScore": null,
  "seasonFinalScore": null,
  "openMonth": "2026-05",
  "createdAt": "2026-05-30T12:00:00Z"
}
```

- `runningAverage` is the live aggregate of all Judgments ever received (public + Season-linked).
- `openFinalScore` is set when the Open month soft-closes; null until then.
- `seasonFinalScore` is set when a Season finalizes (v2); null in v1.
- `gracePeriodExpiresAt` is non-null during the 10-minute Author Grace Period; null after expiry.

---

## 5. Judging — Eligibility, Scoring, and Unwitnessed Jumps

### 5.1 Judging Eligibility Rules

| Rule | Condition | Store-Layer Implication |
|------|-----------|----------------------|
| Not own Jump | `judgments.player_id != jumps.player_id` | Query `jumps.player_id` before inserting Judgment |
| Judging Window open | `jumps.grace_period_expires_at < NOW()` | Check `grace_period_expires_at` in domain command |
| Not already judged | No existing Judgment for same `(jump_id, player_id)` or `(jump_id, guest_session_id)` | UNIQUE constraint + domain check |
| Guest soft cap | `guest_sessions.judgment_count < server_configured_cap` (default 5) | Check `guest_sessions.judgment_count` before inserting; increment on success |
| Jump not Removed | `jumps.status != 'Removed Jump'` | Filter in feed and detail queries |

### 5.2 Score Factor Migration

The scoring factors change in two ways per ADR-0022:

| Old Factor | New Factor | Old Scale | New Scale |
|------------|------------|-----------|-----------|
| difficulty | commitment | 0–10 | 1–4 |
| transgression | transgression | 0–10 | 1–4 |
| creativity | creativity | 0–10 | 1–4 |
| documentation | presentation | 0–10 | 1–4 |

**Schema change:**

```sql
ALTER TABLE judgments RENAME COLUMN difficulty TO commitment;
ALTER TABLE judgments RENAME COLUMN documentation TO presentation;

ALTER TABLE judgments DROP CONSTRAINT judgments_difficulty_check;
ALTER TABLE judgments ADD CONSTRAINT judgments_commitment_check CHECK (commitment >= 1 AND commitment <= 4);

ALTER TABLE judgments DROP CONSTRAINT judgments_transgression_check;
ALTER TABLE judgments ADD CONSTRAINT judgments_transgression_check CHECK (transgression >= 1 AND transgression <= 4);

ALTER TABLE judgments DROP CONSTRAINT judgments_creativity_check;
ALTER TABLE judgments ADD CONSTRAINT judgments_creativity_check CHECK (creativity >= 1 AND creativity <= 4);

ALTER TABLE judgments DROP CONSTRAINT judgments_documentation_check;
ALTER TABLE judgments ADD CONSTRAINT judgments_presentation_check CHECK (presentation >= 1 AND presentation <= 4);
```

**Running average calculation:**

With the 1–4 scale, the running average for a Jump is:

```
running_average = AVG(
  (commitment + transgression + creativity + presentation) / 4.0
) across all Judgments for the Jump
```

Each Judgment produces a composite score of 1.0–4.0 (the mean of its four factors). The running average is the mean of all composite scores. This produces a single number in the 1.0–4.0 range, displayed to one decimal place (e.g., "3.2").

### 5.3 Judgment Provenance

Per ADR-0021, each Judgment must track its context:

```sql
ALTER TABLE judgments ADD COLUMN provenance TEXT NOT NULL
    CHECK (provenance IN ('public', 'season', 'open'));
ALTER TABLE judgments ADD COLUMN open_month TEXT;
ALTER TABLE judgments ADD COLUMN season_id TEXT REFERENCES seasons(id);
```

| `provenance` Value | Meaning | When Set |
|--------------------|---------|----------|
| `public` | Judgment submitted on the public feed, not associated with any competition | Default when Jump has no Season or Open context |
| `open` | Judgment submitted during an Open month | Set when `created_at` falls within an active Open's calendar month |
| `season` | Judgment submitted while Jump is Season-linked | Set when Jump has an active Season association (v2) |

**Open Final Score calculation:**

When an Open month soft-closes, the Open Final Score for each Jump is:

```sql
SELECT AVG((commitment + transgression + creativity + presentation) / 4.0)
FROM judgments
WHERE jump_id = :jump_id
  AND provenance IN ('open', 'public')
  AND created_at < :soft_closed_at;
```

All public and open-provenance Judgments received before the soft-close contribute. Season-provenance Judgments are excluded.

**Season Final Score calculation (v2):**

```sql
SELECT AVG((commitment + transgression + creativity + presentation) / 4.0)
FROM judgments
WHERE jump_id = :jump_id
  AND provenance = 'season'
  AND season_id = :season_id;
```

Only Season-linked Judgments contribute, per ADR-0021.

### 5.4 Guest Judge Data Model

Guest Judges submit Judgments without creating an Account. They are tracked by session:

**`guest_sessions` table:**

```sql
CREATE TABLE guest_sessions (
    id TEXT PRIMARY KEY,
    device_fingerprint TEXT,
    judgment_count INT NOT NULL DEFAULT 0,
    soft_capped_at TIMESTAMPTZ,
    claimed_player_id TEXT REFERENCES players(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

**Flow:**

1. Guest opens the app → client calls `POST /v1/guest-sessions` → receives `guest_session_id`
2. Guest submits a Judgment → `POST /v1/jumps/{jumpID}/judgments` with `guestSessionId` in the body
3. Server checks `guest_sessions.judgment_count` against soft cap (default 5)
4. If under cap: insert Judgment with `guest_session_id`, increment `judgment_count`
5. If at cap: return 403 with `X-Guest-Cap-Reached: true` header
6. Guest creates Account → `guest_sessions.claimed_player_id` is set; future Judgments use `player_id` instead of `guest_session_id`

**Guest-to-Player migration:**

When a Guest Judge creates an Account, their existing Guest Judgments are migrated:
- `judgments.player_id` is set to the new Player's ID
- `judgments.guest_session_id` is set to NULL
- The `guest_sessions.claimed_player_id` is set to the new Player's ID

This preserves the Judgment history and running averages that the Guest contributed.

**Uniqueness constraint:**

A Guest Judge may submit one Judgment per Jump (just like authenticated Players):

```sql
-- Replace the old UNIQUE (stunt_id, player_id) with:
ALTER TABLE judgments DROP CONSTRAINT judgments_stunt_id_player_id_key;
-- New partial unique indexes:
CREATE UNIQUE INDEX judgments_jump_player_unique ON judgments (jump_id, player_id)
    WHERE player_id IS NOT NULL;
CREATE UNIQUE INDEX judgments_jump_guest_unique ON judgments (jump_id, guest_session_id)
    WHERE guest_session_id IS NOT NULL;
```

### 5.5 Unwitnessed Jump Behavior

An Unwitnessed Jump is a Performed Jump whose Judging Window closed without any Judgments. On the public feed, the Judging Window is open-ended — it never closes. Therefore:

- **Public feed:** Unwitnessed Jump status does not apply. A Jump on the public feed with zero Judgments remains a `Performed Jump` indefinitely; it simply has no running average.
- **Season context (v2):** When a Season's Judging Grace Period closes, any Jump with zero Season-provenance Judgments transitions to `Unwitnessed Jump`. This status affects Season Standings but not public feed visibility.

The `Unwitnessed Jump` status is reserved in the schema for v2 but is not assigned by any v1 code path.

---

## 6. Group, Visibility, Season, and Standings — v1 vs. v2

### 6.1 v1 Scope

| Concept | v1 Status | Notes |
|---------|-----------|-------|
| Public feed | Implemented | `GET /v1/feed` — chronological, no Group filter |
| Jump visibility | Public by default | A Jump is visible to anyone unless Removed (ADR-0019) |
| Guest Judge visibility | Full | Guests can see and Judge any non-Removed Jump |
| The Open | Implemented | Monthly competition, soft-close, Standings |
| Groups | Frozen schema, no UI | Tables exist; no Group CRUD endpoints are extended |
| Seasons | Frozen schema, no UI | Tables exist; no Season lifecycle endpoints are extended |
| Season Standings | Not implemented | Replaced by Open Standings in v1 |
| Group Admin / Season Commissioner | Not implemented | v2 concepts |

### 6.2 v2 Scope (Design Reservations)

| Concept | v2 Addition | Schema Reservation |
|---------|-------------|-------------------|
| Group Home | New endpoints for Group CRUD, membership, invites | `groups`, `group_memberships`, `invites` tables already exist |
| Season lifecycle | Start/close/finalize endpoints re-enabled | `seasons`, `season_history` tables already exist |
| Season Standings | Computed from `provenance = 'season'` Judgments | `judgments.season_id` column reserved |
| Season Commissioner | Authority model for Season lifecycle | `seasons.commissioner_player_id` column already exists |
| Group Admin | Override authority over active Season | `group_memberships.role = 'Group Admin'` already exists |
| Disqualified Jump | Formal status for competitive violations | `jumps.status = 'Disqualified Jump'` reserved in CHECK constraint |
| Dispute tooling | Formal Dispute lifecycle with resolution | `disputes` table already exists; resolution flow is manual in v1 |

### 6.3 Jump-Group Association

In v1, `jumps.group_id` is `NULL` for all Jumps. The column is made nullable but not removed, because v2 will re-introduce Group association when Seasons are activated.

In v2, when a Player submits a Jump to a Group's Active Season:
- `jumps.group_id` is set to the Group's ID
- `jumps.season_id` is set to the active Season's ID
- Future Judgments on that Jump from within the Season context are recorded with `provenance = 'season'`
- The Jump remains visible on the public feed regardless of Group association

---

## 7. Evidence Upload Flow

The Evidence upload flow follows ADR-0007 (direct uploads to object storage).

### 7.1 Flow

```
1. Player selects photo in Create Jump screen
2. Client calls POST /v1/jumps/{jumpID}/evidence-authorization
   → Server validates: Player owns the Jump, Jump is in Grace Period or pre-Evidence
   → Server generates: media_object_key, signed upload URL, expiration
   → Server persists: evidence_upload_authorizations row
   → Server returns: { uploadUrl, uploadMethod, uploadHeaders, mediaObjectKey, expiresAt }
3. Client uploads photo directly to object storage using signed URL
4. Client calls POST /v1/jumps/{jumpID}/evidence
   → Server validates: authorization exists, not expired, media exists in storage
   → Server persists: evidences row, updates Jump status to 'Performed Jump'
   → Server returns: EvidenceSubmission (Jump + Evidence)
```

### 7.2 Storage Design

| Concern | Design |
|---------|--------|
| Object key format | `uploads/{jump_id}/{uuid}.{ext}` — scoped to Jump for tombstoning |
| Storage backend | S3-compatible object storage (e.g., R2, S3, MinIO for dev) |
| Upload method | `PUT` with signed URL (pre-signed, 15-minute expiry) |
| Content type validation | Server validates `Content-Type` header is `image/*` before issuing authorization |
| Size limit | 10 MB per upload (server-configurable) |

### 7.3 Validation

| Rule | Enforcement |
|------|-------------|
| One Evidence per Jump | `evidences.jump_id` is `UNIQUE` |
| Upload authorization required | `evidences.upload_authorization_id` is `NOT NULL` and `UNIQUE` |
| Authorization expiry | Server checks `evidence_upload_authorizations.expires_at > NOW()` before accepting Evidence submission |
| Player owns the Jump | Domain command validates `jumps.player_id == authenticated_player_id` |

### 7.4 Tombstoning

When a Jump is Removed:
- The `evidences` row is preserved (for potential appeal review)
- The object storage object is **not deleted** (preserved for team review)
- All read queries exclude Evidence for Removed Jumps at the application layer
- Deep links return a tombstone page: "This Jump has been removed" — no Evidence, no performer info
- The `media_object_key` is retained so the team can review the content if needed

---

## 8. Hexagonal Architecture — Ports and Adapters

### 8.1 Domain Core (`internal/game/`)

The domain core defines repository interfaces (ports) and domain commands. Each aggregate gets its own focused port.

**Current state:** Six repository interfaces exist in `internal/game/`:
- `GroupRepository`
- `StuntPlanningRepository` (will become `JumpRepository`)
- `EvidenceRepository`
- `JudgmentRepository`
- `SeasonRepository`
- `DisputeRepository`

**Target state:** Split into read/write ports per aggregate for CQRS alignment:

```go
// Write ports (commands)
type JumpWriteRepo interface {
    InsertJump(ctx context.Context, jump JumpSnapshot) error
    UpdateJumpStatus(ctx context.Context, jumpID, status string) error
    SetJumpOpenFinalScore(ctx context.Context, jumpID string, score int) error
}

// Read ports (queries)
type JumpReadRepo interface {
    Jump(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
    Feed(ctx context.Context, cursor time.Time, limit int) ([]JumpSnapshot, error)
    JumpsByPlayer(ctx context.Context, playerID string, cursor time.Time, limit int) ([]JumpSnapshot, error)
}

// Judgment ports
type JudgmentWriteRepo interface {
    InsertJudgment(ctx context.Context, judgment JudgmentSnapshot) error
    JudgmentCountByGuestSession(ctx context.Context, guestSessionID string) (int, error)
}

type JudgmentReadRepo interface {
    JudgmentsForJump(ctx context.Context, jumpID string) ([]JudgmentSnapshot, error)
    RunningAverage(ctx context.Context, jumpID string) (float64, error)
}

// Open ports
type OpenWriteRepo interface {
    UpsertOpen(ctx context.Context, yearMonth string) error
    SetOpenFinalScores(ctx context.Context, yearMonth string) error
}

type OpenReadRepo interface {
    OpenStandings(ctx context.Context, yearMonth string) ([]StandingEntry, error)
    CurrentOpen(ctx context.Context) (OpenSnapshot, error)
}

// Guest session port
type GuestSessionRepo interface {
    CreateGuestSession(ctx context.Context, id string) error
    GuestSession(ctx context.Context, id string) (GuestSessionSnapshot, bool, error)
    IncrementJudgmentCount(ctx context.Context, id string) error
    ClaimGuestSession(ctx context.Context, id, playerID string) error
}
```

The existing `Persistence` monolith interface is decomposed. Adapters implement only the ports they need. This allows PostgresStore to implement all ports while MemoryStore can implement a subset for testing.

### 8.2 Application Service Layer

A formal application service layer is introduced between the domain and adapters. Each service method is a single use case:

```go
type JumpService struct {
    JumpWrite    JumpWriteRepo
    JumpRead     JumpReadRepo
    EvidenceRepo EvidenceRepository
    Now          func() time.Time
}

func (s *JumpService) CreatePerformedJump(ctx context.Context, input CreateJumpInput) (JumpSnapshot, error) { ... }
func (s *JumpService) RetractJump(ctx context.Context, jumpID, playerID string) error { ... }
func (s *JumpService) EditJumpCaption(ctx context.Context, jumpID, playerID, caption string) error { ... }

type JudgmentService struct {
    JudgmentWrite JudgmentWriteRepo
    JudgmentRead  JudgmentReadRepo
    JumpRead      JumpReadRepo
    GuestSession  GuestSessionRepo
    Now           func() time.Time
}

func (s *JudgmentService) SubmitJudgment(ctx context.Context, input SubmitJudgmentInput) (JudgmentResult, error) { ... }

type OpenService struct {
    OpenWrite    OpenWriteRepo
    OpenRead     OpenReadRepo
    JudgmentRead JudgmentReadRepo
    JumpRead     JumpReadRepo
    Now          func() time.Time
}

func (s *OpenService) SoftCloseMonth(ctx context.Context, yearMonth string) error { ... }
func (s *OpenService) ComputeStandings(ctx context.Context, yearMonth string) ([]StandingEntry, error) { ... }
```

### 8.3 Infrastructure Adapters

**HTTP adapter (`internal/httpapi/server.go`):**

- Handles HTTP routing, auth verification, request validation
- Calls application service methods
- Maps domain errors to HTTP status codes
- Assembles response DTOs

**Persistence adapter (`internal/httpapi/postgres_store.go`):**

- Implements all repository ports using sqlc-generated queries
- Each port method is a thin wrapper over a sqlc query
- No business logic in the adapter

**Memory adapter (`internal/httpapi/store.go`):**

- Implements repository ports with in-memory maps
- Used for testing and development
- Currently combines DTO definitions with adapter logic; DTOs will be extracted to `internal/httpapi/dto.go`

### 8.4 Error Mapping

Domain errors are defined in `internal/game/` as sentinel errors. The HTTP adapter maps them:

| Domain Error | HTTP Status | Client Message |
|-------------|-------------|----------------|
| `ErrJumpNotFound` | 404 | "Jump not found" |
| `ErrJudgingWindowClosed` | 403 | "Judging Window closed" |
| `ErrGracePeriodActive` | 403 | "Judging Window opens in [MM:SS]" |
| `ErrInvalidJudgmentScore` | 400 | "Judgment scores must be between 1 and 4" |
| `ErrAlreadyJudged` | 409 | "You have already entered your Judgment" |
| `ErrSelfJudging` | 403 | "You cannot Judge your own Jump" |
| `ErrGuestCapReached` | 403 | "Guest Judgment cap reached" |
| `ErrJumpRemoved` | 410 | "This Jump has been removed" |

---

## 9. ADRs and CONTEXT.md Requiring Updates

### 9.1 ADRs Requiring Updates

| ADR | Update Required | Reason |
|-----|-----------------|--------|
| ADR-0005 (Postgres/sqlc) | Note table rename `stunts` → `jumps`; note new tables `guest_sessions`, `opens`, `open_standings` | Schema evolution |
| ADR-0007 (Direct Evidence Uploads) | Add tombstoning behavior for Removed Jumps; clarify object key format scoped by `jump_id` | Safety model (ADR-0024) |
| ADR-0008 (Stunts Belong to One Group) | Mark as fully superseded by ADR-0019; note `group_id` is now nullable | ADR-0019 |
| ADR-0011 (Season Close and Judging Grace Period) | Clarify that Author Grace Period is a separate concept from Season's Judging Grace Period | Product/UX Design §1.4 |
| ADR-0016 (Client-Side Eligibility Guards) | Add Guest Judge eligibility; add Grace Period countdown guard | ADR-0022, Product/UX Design |

### 9.2 CONTEXT.md Sections Requiring Updates

| Section | Update Required |
|---------|-----------------|
| **Jump** definition | Add Author Grace Period mention; clarify that Draft is client-only |
| **Author Grace Period** | Already present; verify alignment with `grace_period_expires_at` column |
| **Judgment** definition | Note Guest Judge eligibility; note provenance tracking |
| **Guest Judge** definition | Note `guest_sessions` table; note soft cap enforcement |
| **Unwitnessed Jump** definition | Clarify that this is a v2/Season concept only; public feed has open-ended Judging Window |
| **Removed Jump** definition | Confirm tombstoning behavior aligns with ADR-0024 |
| **Open** definition | Note `opens` and `open_standings` tables; note soft-close mechanism |
| **Final Score** definition | Note three score types: running average (live), Open Final Score (monthly), Season Final Score (v2) |

### 9.3 New ADRs to Write

| Proposed ADR | Topic | Trigger |
|--------------|-------|---------|
| ADR-0025 | Guest Judge session model | This document §5.4 |
| ADR-0026 | Open monthly competition data model | This document §3.3, ADR-0023 |
| ADR-0027 | Hexagonal architecture formalization | This document §8 |

---

## 10. References

| Document | Relationship |
|----------|-------------|
| [Product/UX Design](./02-product-ux-design.md) (#106) | Upstream dependency: UX decisions inform data model, API contract, and Jump lifecycle states |
| [MVP Roadmap](./04-mvp-roadmap.md) | Scope boundary: defines what is v1 vs. v2 |
| ADR-0019 | Jumps are public by default — drives nullable `group_id`, public feed, open Judging |
| ADR-0020 | Rename Stunt → Jump, Documentation → Presentation — drives table/column renames |
| ADR-0021 | Season scoring excludes pre-existing public Judgments — drives `provenance` column |
| ADR-0022 | Judgment interaction model — drives 1–4 scale, factor names, confirmation flow |
| ADR-0023 | The Open — drives `opens` table, `open_month` tracking, soft-close mechanism |
| ADR-0024 | House Rules and safety — drives Removed Jump tombstoning, Report flow |
| ADR-0007 | Direct Evidence Uploads — drives storage, validation, authorization flow |
| ADR-0004 | REST/OpenAPI contract — drives endpoint design |
| ADR-0005 | Postgres/sqlc persistence — drives persistence adapter implementation |
| ADR-0002 | Backend owns game rules — drives domain-first design |
| CONTEXT.md | Domain language — all terminology must align |

---

_Document status: Complete. Parent tracker: #107. Depends on: Product/UX Design #106, ADR-0019, ADR-0020, ADR-0021, ADR-0022, ADR-0023, ADR-0024, MVP Roadmap._
