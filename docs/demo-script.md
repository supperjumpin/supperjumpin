# Supperjumpin Demo Script

End-to-end walkthrough of the first playable Group Stunt loop, covering every feature
that exists in the codebase as of PRD #1.

## Contents

1. [Prerequisites](#1-prerequisites)
2. [Architecture overview](#2-architecture-overview)
3. [Start the backend](#3-start-the-backend)
4. [Player auth (dev)](#4-player-auth-dev)
5. [Group lifecycle](#5-group-lifecycle)
6. [Invites & membership](#6-invites--membership)
7. [Season lifecycle](#7-season-lifecycle)
8. [Stunt lifecycle](#8-stunt-lifecycle)
9. [Judging & scoring](#9-judging--scoring)
10. [Standings](#10-standings)
11. [Season history audit trail](#11-season-history-audit-trail)
12. [Group home (aggregated view)](#12-group-home-aggregated-view)
13. [Mobile app](#13-mobile-app)
14. [Run the test suite](#14-run-the-test-suite)
15. [Tear down](#15-tear-down)

---

## 1. Prerequisites

| Tool     | Version   | Check              |
|----------|-----------|--------------------|
| Go       | 1.25+     | `go version`       |
| Node.js  | 22+       | `node --version`   |
| npm      | 10+       | `npm --version`    |
| Docker   | Desktop   | `docker version`   |
| curl     | any       | `curl --version`   |
| jq       | 1.6+      | `jq --version`     |

One-time JS dependency install:

```sh
npm install
```

---

## 2. Architecture overview

```
                  ┌──────────────────┐
                  │   Supabase Auth   │
                  │  (magic links)    │
                  └────────┬─────────┘
                           │ bearer token
                  ┌────────▼─────────┐
  ┌───────────────┤  Go REST API     ├───────────────┐
  │               │  :8080           │               │
  │               └────────┬─────────┘               │
  │                        │                          │
  ▼                        ▼                          ▼
┌──────────┐     ┌──────────────────┐     ┌─────────────────────┐
│  Expo RN │     │   PostgreSQL 16  │     │  Object Storage     │
│  Mobile  │     │  (docker) :5432  │     │  (evidence uploads) │
└──────────┘     └──────────────────┘     └─────────────────────┘
```

- Backend owns all game rules (stunt lifecycle, judging eligibility, scoring).
- Mobile is a thin view layer with gesture-driven scoring UI.
- API contract defined in `apps/api/openapi.yaml` (17 endpoints).
- Dev mode uses a static bearer token instead of Supabase.

---

## 3. Start the backend

### 3a. Quick start (single player)

```sh
npm run demo:api
```

This starts Postgres via Docker Compose, applies all 9 migrations, and runs the Go API
on `http://localhost:8080` with the dev bearer token `dev-token`.

The dev token maps to identity `player@example.com`.

### 3b. Two-player setup (for full demo)

The full demo needs two players so judging works (you cannot judge your own stunt).
Open **three terminals**.

**Terminal 1 — Postgres & migrations:**

```sh
docker compose up -d postgres
# wait for pg_isready, then apply migrations
# (simplest: run `npm run demo:api` once and Ctrl-C after migrations)
sleep 10
```

**Terminal 2 — Player A API (:8080):**

```sh
DATABASE_URL="postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable" \
SUPPERJUMPIN_DEV_AUTH_TOKEN=player-a-token \
SUPPERJUMPIN_DEV_AUTH_SUBJECT=dev-subject-a \
SUPPERJUMPIN_DEV_AUTH_EMAIL="alice@example.com" \
PORT=8080 \
go run ./apps/api/cmd/api
```

**Terminal 3 — Player B API (:8081):**

```sh
DATABASE_URL="postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable" \
SUPPERJUMPIN_DEV_AUTH_TOKEN=player-b-token \
SUPPERJUMPIN_DEV_AUTH_SUBJECT=dev-subject-b \
SUPPERJUMPIN_DEV_AUTH_EMAIL="bob@example.com" \
PORT=8081 \
go run ./apps/api/cmd/api
```

> Both APIs share the same Postgres database. Player B accesses their own endpoints on
> :8081. You can also run a single API with both tokens by modifying `main.go` to read
> multiple env vars, but the two-instance approach requires zero source changes.

---

## 4. Player auth (dev)

Smoke-test that each player is authenticated and has an auto-created Account + Player:

### Player A

```sh
curl -s -H "Authorization: Bearer player-a-token" http://localhost:8080/v1/me | jq
```

```json
{
  "account": { "id": "account_...", "email": "alice@example.com" },
  "player": { "id": "player_...", "displayName": "alice" }
}
```

### Player B

```sh
curl -s -H "Authorization: Bearer player-b-token" http://localhost:8081/v1/me | jq
```

```json
{
  "account": { "id": "account_...", "email": "bob@example.com" },
  "player": { "id": "player_...", "displayName": "bob" }
}
```

Save the player IDs for later steps:

```sh
ALICE=$(curl -s -H "Authorization: Bearer player-a-token" http://localhost:8080/v1/me | jq -r '.player.id')
BOB=$(curl -s -H "Authorization: Bearer player-b-token" http://localhost:8081/v1/me | jq -r '.player.id')
```

**Key design point:** `BootstrapIdentity` is idempotent — calling GET /v1/me multiple
times returns the same Account and Player for the same auth identity.

---

## 5. Group lifecycle

### 5a. Player A creates a group

```sh
curl -s -X POST http://localhost:8080/v1/groups \
  -H "Authorization: Bearer player-a-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "Taco Bell Daredevils"}' | jq
```

Response includes group details, membership role, and (empty) recent stunts + standings:

```json
{
  "group": { "id": "group_...", "name": "Taco Bell Daredevils" },
  "membership": { "groupId": "group_...", "playerId": "player_...", "role": "Group Admin" },
  "activeSeason": null,
  "recentStunts": [],
  "standings": []
}
```

Notice `role: "Group Admin"` — the creator automatically gets admin rights.

Save the group ID:

```sh
GROUP_ID=$(curl -s -X POST http://localhost:8080/v1/groups \
  -H "Authorization: Bearer player-a-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "Taco Bell Daredevils"}' | jq -r '.group.id')
```

### 5b. Player A lists their groups

```sh
curl -s -H "Authorization: Bearer player-a-token" http://localhost:8080/v1/groups | jq
```

```json
{
  "memberships": [
    {
      "group": { "id": "group_...", "name": "Taco Bell Daredevils" },
      "membership": { "groupId": "group_...", "playerId": "player_...", "role": "Group Admin" }
    }
  ]
}
```

Player B's list is empty (not a member yet):

```sh
curl -s -H "Authorization: Bearer player-b-token" http://localhost:8081/v1/groups | jq
```

```json
{ "memberships": [] }
```

---

## 6. Invites & membership

### 6a. Player A creates an invite

```sh
INVITE=$(curl -s -X POST http://localhost:8080/v1/groups/$GROUP_ID/invites \
  -H "Authorization: Bearer player-a-token" | jq)

echo "$INVITE" | jq
TOKEN=$(echo "$INVITE" | jq -r '.token')
```

```json
{
  "id": "invite_...",
  "groupId": "group_...",
  "token": "invite_token_...",
  "createdBy": "player_...",
  "expiresAt": "2026-06-01T..."
}
```

Invites expire after 7 days. Only group members can create invites (any member, not just admin).

### 6b. Player B accepts the invite

```sh
curl -s -X POST http://localhost:8081/v1/invites/$TOKEN/accept \
  -H "Authorization: Bearer player-b-token" | jq
```

Player B now has `role: "Player"` (not Admin) in the group:

```json
{
  "group": { "id": "group_...", "name": "Taco Bell Daredevils" },
  "membership": { "groupId": "group_...", "playerId": "player_...", "role": "Player" },
  "activeSeason": null,
  "recentStunts": [],
  "standings": []
}
```

### 6c. Error cases (for demo)

**Expired invite** — try accepting after `expiresAt`:
```json
// HTTP 410 Gone
"Invite expired"
```

**Already used invite** — a second attempt:
```json
// HTTP 409 Conflict
"Invite already used"
```

**Already a member** — try accepting another invite for the same group:
```json
// HTTP 409 Conflict
"Player already has a Group Membership"
```

**Invalid token**:
```json
// HTTP 404 Not Found
"Invite cannot be accepted"
```

---

## 7. Season lifecycle

### 7a. Player A starts a season

Deadlines are ISO 8601 timestamps. Set them far enough in the future for the demo.

```sh
SEASON=$(curl -s -X POST http://localhost:8080/v1/groups/$GROUP_ID/seasons \
  -H "Authorization: Bearer player-a-token" \
  -H "Content-Type: application/json" \
  -d '{
    "submissionDeadline": "2026-06-15T23:59:59Z",
    "judgingDeadline": "2026-06-30T23:59:59Z"
  }' | jq)

echo "$SEASON" | jq
SEASON_ID=$(echo "$SEASON" | jq -r '.activeSeason.id')
```

```json
{
  "group": { "id": "group_...", "name": "Taco Bell Daredevils" },
  "membership": { "groupId": "group_...", "playerId": "player_...", "role": "Group Admin" },
  "activeSeason": {
    "id": "season_...",
    "groupId": "group_...",
    "commissionerPlayerId": "player_...",
    "status": "Active",
    "submissionDeadline": "2026-06-15T23:59:59Z",
    "judgingDeadline": "2026-06-30T23:59:59Z"
  },
  "recentStunts": [],
  "standings": []
}
```

The creator becomes **Season Commissioner**. Only one open season per group is allowed.

### 7b. Duplicate season guard (error case)

```sh
curl -s -X POST http://localhost:8080/v1/groups/$GROUP_ID/seasons \
  -H "Authorization: Bearer player-a-token" \
  -H "Content-Type: application/json" \
  -d '{
    "submissionDeadline": "2026-07-15T23:59:59Z",
    "judgingDeadline": "2026-07-30T23:59:59Z"
  }' | jq
```

```json
// HTTP 409 Conflict
"Group already has an active or closing Season"
```

### 7c. Auto-transition (temporal state)

If `now > submissionDeadline` the season auto-transitions from `Active` → `Judging Grace Period`.
If `now > judgingDeadline` it auto-transitions `Judging Grace Period` → `Finalized`.

This is checked every time season state is loaded (no cron needed).

---

## 8. Stunt lifecycle

The stunt lifecycle: `Idea` → `Planned Stunt` → `Performed Stunt` → `Judged Stunt` / `Unjudged Stunt`

### 8a. Player A creates an Idea

Each Idea has a Source (where you buy the food), Destination (where you eat it), and Food.

```sh
IDEA=$(curl -s -X POST http://localhost:8080/v1/groups/$GROUP_ID/ideas \
  -H "Authorization: Bearer player-a-token" \
  -H "Content-Type: application/json" \
  -d '{
    "source": "Taco Bell",
    "destination": "Olive Garden Parking Lot",
    "food": "Crunchwrap Supreme"
  }' | jq)

echo "$IDEA" | jq
IDEA_ID=$(echo "$IDEA" | jq -r '.id')
```

```json
{
  "id": "stunt_...",
  "groupId": "group_...",
  "playerId": "player_...",
  "seasonId": null,
  "status": "Idea",
  "source": "Taco Bell",
  "destination": "Olive Garden Parking Lot",
  "food": "Crunchwrap Supreme",
  "offSeason": true,
  "finalScore": null
}
```

Ideas are `offSeason: true` by default and have `seasonId: null`.

### 8b. Player A promotes the Idea to a Planned Stunt

Since an active season exists and we don't set `offSeason: true`, the stunt links to the
active season automatically.

```sh
PLANNED=$(curl -s -X POST http://localhost:8080/v1/ideas/$IDEA_ID/planned-stunt \
  -H "Authorization: Bearer player-a-token" \
  -H "Content-Type: application/json" \
  -d '{"offSeason": false}' | jq)

echo "$PLANNED" | jq
STUNT_ID=$(echo "$PLANNED" | jq -r '.id')
```

```json
{
  "id": "stunt_...",
  "groupId": "group_...",
  "playerId": "player_...",
  "seasonId": "season_...",
  "status": "Planned Stunt",
  "source": "Taco Bell",
  "destination": "Olive Garden Parking Lot",
  "food": "Crunchwrap Supreme",
  "offSeason": false,
  "finalScore": null
}
```

Now `status: "Planned Stunt"`, `offSeason: false`, and `seasonId` points to the active season.

> **Off-Season stunts:** Setting `{"offSeason": true}` keeps the stunt outside season
> competition. It can still be performed and judged, but won't appear in standings.

### 8c. Authorize evidence upload

Player A authorizes a 15-minute upload window. The response includes a signed `uploadUrl`
and headers — in production this would be a PUT to object storage.

```sh
AUTH_JSON=$(curl -s -X POST http://localhost:8080/v1/stunts/$STUNT_ID/evidence-upload-authorizations \
  -H "Authorization: Bearer player-a-token" \
  -H "Content-Type: application/json" \
  -d '{"contentType": "image/jpeg"}' | jq)

echo "$AUTH_JSON" | jq
UPLOAD_AUTH_ID=$(echo "$AUTH_JSON" | jq -r '.id')
```

```json
{
  "id": "evidence_upload_...",
  "stuntId": "stunt_...",
  "uploadUrl": "https://storage.supperjumpin.test/uploads/stunt_...",
  "uploadMethod": "PUT",
  "uploadHeaders": { "Content-Type": "image/jpeg" },
  "mediaObjectKey": "uploads/stunt_.../1",
  "expiresAt": "2026-05-25T..."
}
```

Only the stunt performer can authorize an upload. The auth window lasts 15 minutes.

### 8d. Submit evidence (perform the stunt)

Player A submits the upload authorization ID and a caption to complete the stunt.

```sh
SUBMISSION=$(curl -s -X POST http://localhost:8080/v1/stunts/$STUNT_ID/evidence \
  -H "Authorization: Bearer player-a-token" \
  -H "Content-Type: application/json" \
  -d "{
    \"uploadAuthorizationId\": \"$UPLOAD_AUTH_ID\",
    \"caption\": \"Crunchwrap devoured in the Olive Garden parking lot. Security gave me a look.\"
  }" | jq)

echo "$SUBMISSION" | jq
```

```json
{
  "stunt": {
    "id": "stunt_...",
    "status": "Performed Stunt",
    ...
  },
  "evidence": {
    "id": "evidence_...",
    "stuntId": "stunt_...",
    "caption": "Crunchwrap devoured in the Olive Garden parking lot. Security gave me a look.",
    "mediaObjectKey": "uploads/stunt_.../1",
    "createdAt": "2026-05-25T..."
  }
}
```

The stunt transitions from `Planned Stunt` → `Performed Stunt`. Evidence is stored with
the caption and media key.

> **Submission window enforcement:** If the season's submission deadline has passed,
> this step returns `HTTP 409 Conflict` with `"Submission Window closed"`.

---

## 9. Judging & scoring

### 9a. Player B judges Player A's stunt

Player B scores the stunt on four axes (each 0–10):

```sh
JUDGMENT=$(curl -s -X POST http://localhost:8081/v1/stunts/$STUNT_ID/judgment \
  -H "Authorization: Bearer player-b-token" \
  -H "Content-Type: application/json" \
  -d '{
    "difficulty": 7,
    "transgression": 8,
    "creativity": 9,
    "documentation": 6
  }' | jq)

echo "$JUDGMENT" | jq
```

```json
{
  "id": "judgment_...",
  "stuntId": "stunt_...",
  "playerId": "player_...",
  "difficulty": 7,
  "transgression": 8,
  "creativity": 9,
  "documentation": 6
}
```

Returns `201 Created` on first judgment, `200 OK` on edits (upsert).

### 9b. Error cases

**Cannot judge own stunt:**
```json
// HTTP 403 Forbidden
"Judge required"
```

**Scores outside 0–10:**
```json
// HTTP 400 Bad Request
"Judgment scores must be between 0 and 10"
```

**Judging window closed (season finalized):**
```json
// HTTP 409 Conflict
"Judging Window closed"
```

### 9c. Edit a judgment

Player B can adjust their scores as long as the judging window is open:

```sh
curl -s -X POST http://localhost:8081/v1/stunts/$STUNT_ID/judgment \
  -H "Authorization: Bearer player-b-token" \
  -H "Content-Type: application/json" \
  -d '{
    "difficulty": 8,
    "transgression": 8,
    "creativity": 9,
    "documentation": 7
  }' | jq
```

Returns `200 OK` (not `201 Created`) on edits.

---

## 10. Standings

### 10a. Close submissions

The Season Commissioner (or Group Admin) transitions the season to the judging grace period:

```sh
curl -s -X POST http://localhost:8080/v1/seasons/$SEASON_ID/close-submissions \
  -H "Authorization: Bearer player-a-token" | jq
```

Now `status: "Judging Grace Period"`. Submissions are locked but judging continues.

### 10b. Finalize the season

```sh
FINALIZE=$(curl -s -X POST http://localhost:8080/v1/seasons/$SEASON_ID/finalize \
  -H "Authorization: Bearer player-a-token" | jq)

echo "$FINALIZE" | jq
```

On finalization, the season transitions to `Finalized`, and:

- **Judged stunts** (those with at least one judgment) → `Final Score` = average of all
  judgment scores across all four axes → status becomes `Judged Stunt`.
- **Performed stunts without judgments** → status becomes `Unjudged Stunt`, no final score.

Standings now populate:

```json
{
  "group": { "id": "group_...", "name": "Taco Bell Daredevils" },
  ...
  "standings": [
    {
      "player": { "id": "player_...", "displayName": "alice" },
      "seasonScore": 32,
      "judgedStunts": 1
    }
  ]
}
```

`seasonScore` = (7+8+9+6) = 30 / 1 judgment = average 30... wait, let me recalculate.
With scores Difficulty=8, Transgression=8, Creativity=9, Documentation=7, that's
8+8+9+7 = 32 total / 1 judgment = 32. So `seasonScore` = 32.

Actually wait, looking at the `finalScoreForStunt` function:

```go
func (s *MemoryStore) finalScoreForStunt(stuntID string) (int, bool) {
    total := 0
    count := 0
    for _, judgment := range s.judgments {
        if judgment.StuntID != stuntID {
            continue
        }
        total += judgment.Difficulty + judgment.Transgression + judgment.Creativity + judgment.Documentation
        count++
    }
    if count == 0 {
        return 0, false
    }
    return total / count, true
}
```

So finalScore is the sum of all four axes averaged across judgments. With one judgment of 8+8+9+7=32, finalScore = 32.

### 10c. Standings computation rules

- Only the **latest season's** stunts count (determined by season creation order, not lexicographic ID).
- Only **Judged Stunts** count (not Unjudged, Performed, or Off-Season).
- Standings sorted by `SeasonScore` descending, then display name ascending.
- **Empty standings** for groups with no Finalized Season.

---

## 11. Season history audit trail

Every season transition is logged with actor, role, override flag, and timestamps:

```sh
curl -s -H "Authorization: Bearer player-a-token" \
  http://localhost:8080/v1/seasons/$SEASON_ID/history | jq
```

```json
{
  "entries": [
    {
      "id": "season_history_...",
      "seasonId": "season_...",
      "action": "Submissions Closed",
      "actorPlayerId": "player_...",
      "actorRole": "Group Admin",
      "override": false,
      "fromStatus": "Active",
      "toStatus": "Judging Grace Period",
      "createdAt": "..."
    },
    {
      "id": "season_history_...",
      "seasonId": "season_...",
      "action": "Season Finalized",
      "actorPlayerId": "player_...",
      "actorRole": "Group Admin",
      "override": false,
      "fromStatus": "Judging Grace Period",
      "toStatus": "Finalized",
      "createdAt": "..."
    }
  ]
}
```

When a Group Admin (not the Season Commissioner) performs a transition, `override: true`
is set, providing an audit trail for emergency actions.

---

## 12. Group home (aggregated view)

The group home endpoint returns everything in one call — group details, membership, active
season, recent performed stunts, and standings:

```sh
curl -s -H "Authorization: Bearer player-a-token" \
  http://localhost:8080/v1/groups/$GROUP_ID/home | jq
```

```json
{
  "group": {
    "id": "group_...",
    "name": "Taco Bell Daredevils"
  },
  "membership": {
    "groupId": "group_...",
    "playerId": "player_...",
    "role": "Group Admin"
  },
  "activeSeason": {
    "id": "season_...",
    "status": "Finalized",
    ...
  },
  "recentStunts": [
    {
      "stunt": {
        "id": "stunt_...",
        "status": "Judged Stunt",
        "source": "Taco Bell",
        "destination": "Olive Garden Parking Lot",
        "food": "Crunchwrap Supreme",
        "finalScore": 32,
        ...
      },
      "performer": { "id": "player_...", "displayName": "alice" },
      "evidence": {
        "id": "evidence_...",
        "caption": "Crunchwrap devoured in the Olive Garden parking lot...",
        "mediaObjectKey": "uploads/stunt_.../1",
        "createdAt": "..."
      }
    }
  ],
  "standings": [
    {
      "player": { "id": "player_...", "displayName": "alice" },
      "seasonScore": 32,
      "judgedStunts": 1
    }
  ]
}
```

---

## 13. Mobile app

The Expo React Native mobile app renders the same flow with a gesture-driven UI.

### Setup

```sh
cp apps/mobile/.env.example apps/mobile/.env
```

Edit `apps/mobile/.env` with your Supabase project URL, anon key, and API base URL.

### Start

```sh
npm run demo:mobile
```

### Key UI features

| Feature | Implementation |
|---------|---------------|
| Auth | Supabase magic link sign-in via email |
| Group creation | Name input, POST to backend |
| Group list | Shows memberships from GET /v1/groups |
| Invite creation | Generates token, shows invite code |
| Season start | Picks deadlines, POST to backend |
| Idea capture | Source, Destination, Food inputs |
| Stunt detail | Performer, caption, media key |
| Gesture scoring | PanResponder swipe on each axis, +/- buttons |
| Score submit | Explicit Submit button, shows scores |
| Accessibility | `accessibilityLabel` on all score buttons |

---

## 14. Run the test suite

The Go backend test suite covers all endpoints with both in-memory and Postgres stores,
including concurrency tests.

```sh
# All Go tests
npm run api:test

# With verbose output
go test -v ./apps/api/...

# Run specific test
go test -v -run TestAcceptInvite ./apps/api/...

# Postgres-backed tests (requires DATABASE_URL)
DATABASE_URL="postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable" \
  go test -v -run TestPostgres ./apps/api/...
```

Test count: ~80 test functions across 2 files (`me_test.go`, `groups_test.go`).

---

## 15. Tear down

Stop the API processes (Ctrl-C in each terminal).

```sh
# Stop Postgres and remove data
docker compose down -v

# Or just stop it (preserve data)
docker compose stop
```

---

## Reference: API endpoint map

| Method | Path | Step |
|--------|------|------|
| `GET /v1/me` | Auth / bootstrap | 4 |
| `POST /v1/groups` | Create group | 5a |
| `GET /v1/groups` | List groups | 5b |
| `GET /v1/groups/{groupId}/home` | Group home | 12 |
| `POST /v1/groups/{groupId}/invites` | Create invite | 6a |
| `POST /v1/invites/{token}/accept` | Accept invite | 6b |
| `POST /v1/groups/{groupId}/seasons` | Start season | 7a |
| `POST /v1/seasons/{seasonId}/close-submissions` | Close submissions | 10a |
| `POST /v1/seasons/{seasonId}/finalize` | Finalize season | 10b |
| `GET /v1/seasons/{seasonId}/history` | Season history | 11 |
| `POST /v1/groups/{groupId}/ideas` | Create idea | 8a |
| `POST /v1/ideas/{ideaId}/planned-stunt` | Plan stunt | 8b |
| `POST /v1/stunts/{stuntId}/evidence-upload-authorizations` | Authorize upload | 8c |
| `POST /v1/stunts/{stuntId}/evidence` | Submit evidence | 8d |
| `POST /v1/stunts/{stuntId}/judgment` | Submit judgment | 9 |

---

## End-to-end verification checklist

- [ ] API starts and responds to GET /v1/me
- [ ] Player can create a group and become Group Admin
- [ ] Player can list their group memberships
- [ ] Group member can create an invite
- [ ] Another player can accept the invite
- [ ] Expired/used/invalid invites return correct errors
- [ ] Group member can start a season with deadlines
- [ ] Multiple open seasons are rejected
- [ ] Group member can create an idea
- [ ] Idea can be promoted to planned stunt (season-linked or off-season)
- [ ] Performer can authorize evidence upload
- [ ] Performer can submit evidence with caption
- [ ] Group member can judge a performed stunt on 4 axes
- [ ] Self-judging is rejected
- [ ] Out-of-range scores are rejected
- [ ] Judgments can be edited (upsert)
- [ ] Season commissioner can close submissions
- [ ] Season commissioner can finalize season
- [ ] Finalized season locks standings
- [ ] Judged stunts get a final score in standings
- [ ] Season history shows all transitions
- [ ] Group home returns complete view
