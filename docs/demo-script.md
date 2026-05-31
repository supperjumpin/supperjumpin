     1|# Supperjumpin Demo Script
     2|
     3|End-to-end walkthrough of the first playable Group Jump loop, covering every feature
     4|that exists in the codebase as of PRD #1.
     5|
     6|## Contents
     7|
     8|1. [Prerequisites](#1-prerequisites)
     9|2. [Architecture overview](#2-architecture-overview)
    10|3. [Start the backend](#3-start-the-backend)
    11|4. [Player auth (dev)](#4-player-auth-dev)
    12|5. [Group lifecycle](#5-group-lifecycle)
    13|6. [Invites & membership](#6-invites--membership)
    14|7. [Season lifecycle](#7-season-lifecycle)
    15|8. [Jump lifecycle](#8-jump-lifecycle)
    16|9. [Judging & scoring](#9-judging--scoring)
    17|10. [Standings](#10-standings)
    18|11. [Season history audit trail](#11-season-history-audit-trail)
    19|12. [Group home (aggregated view)](#12-group-home-aggregated-view)
    20|13. [Mobile app](#13-mobile-app)
    21|14. [Run the test suite](#14-run-the-test-suite)
    22|15. [Tear down](#15-tear-down)
    23|
    24|---
    25|
    26|## 1. Prerequisites
    27|
    28|| Tool     | Version   | Check              |
    29||----------|-----------|--------------------|
    30|| Go       | 1.25+     | `go version`       |
    31|| Node.js  | 22+       | `node --version`   |
    32|| npm      | 10+       | `npm --version`    |
    33|| Docker   | Desktop   | `docker version`   |
    34|| curl     | any       | `curl --version`   |
    35|| jq       | 1.6+      | `jq --version`     |
    36|
    37|One-time JS dependency install:
    38|
    39|```sh
    40|npm install
    41|```
    42|
    43|---
    44|
    45|## 2. Architecture overview
    46|
    47|```
    48|                  ┌──────────────────┐
    49|                  │   Supabase Auth   │
    50|                  │  (magic links)    │
    51|                  └────────┬─────────┘
    52|                           │ bearer token
    53|                  ┌────────▼─────────┐
    54|  ┌───────────────┤  Go REST API     ├───────────────┐
    55|  │               │  :8080           │               │
    56|  │               └────────┬─────────┘               │
    57|  │                        │                          │
    58|  ▼                        ▼                          ▼
    59|┌──────────┐     ┌──────────────────┐     ┌─────────────────────┐
    60|│  Expo RN │     │   PostgreSQL 16  │     │  Object Storage     │
    61|│  Mobile  │     │  (docker) :5432  │     │  (evidence uploads) │
    62|└──────────┘     └──────────────────┘     └─────────────────────┘
    63|```
    64|
    65|- Backend owns all game rules (jump lifecycle, judging eligibility, scoring).
    66|- Mobile is a thin view layer with gesture-driven scoring UI.
    67|- API contract defined in `apps/api/openapi.yaml` (17 endpoints).
    68|- Dev mode uses a static bearer token instead of Supabase.
    69|
    70|---
    71|
    72|## 3. Start the backend
    73|
    74|### 3a. Quick start (single player)
    75|
    76|```sh
    77|npm run demo:api
    78|```
    79|
    80|This starts Postgres via Docker Compose, applies all 9 migrations, and runs the Go API
    81|on `http://localhost:8080` with the dev bearer token `dev-token`.
    82|
    83|The dev token maps to identity `player@example.com`.
    84|
    85|### 3b. Two-player setup (for full demo)
    86|
    87|The full demo needs two players so judging works (you cannot judge your own jump).
    88|Open **three terminals**.
    89|
    90|**Terminal 1 — Postgres & migrations:**
    91|
    92|```sh
    93|docker compose up -d postgres
    94|# wait for pg_isready, then apply migrations
    95|# (simplest: run `npm run demo:api` once and Ctrl-C after migrations)
    96|sleep 10
    97|```
    98|
    99|**Terminal 2 — Player A API (:8080):**
   100|
   101|```sh
   102|DATABASE_URL="postgres://postgres:***@localhost:5432/supperjumpin?sslmode=disable" \
   103|SUPPERJUMPIN_DEV_AUTH_TOKEN=player-a-token \
   104|SUPPERJUMPIN_DEV_AUTH_SUBJECT=dev-subject-a \
   105|SUPPERJUMPIN_DEV_AUTH_EMAIL="alice@example.com" \
   106|PORT=8080 \
   107|go run ./apps/api/cmd/api
   108|```
   109|
   110|**Terminal 3 — Player B API (:8081):**
   111|
   112|```sh
   113|DATABASE_URL="postgres://postgres:***@localhost:5432/supperjumpin?sslmode=disable" \
   114|SUPPERJUMPIN_DEV_AUTH_TOKEN=player-b-token \
   115|SUPPERJUMPIN_DEV_AUTH_SUBJECT=dev-subject-b \
   116|SUPPERJUMPIN_DEV_AUTH_EMAIL="bob@example.com" \
   117|PORT=8081 \
   118|go run ./apps/api/cmd/api
   119|```
   120|
   121|> Both APIs share the same Postgres database. Player B accesses their own endpoints on
   122|> :8081. You can also run a single API with both tokens by modifying `main.go` to read
   123|> multiple env vars, but the two-instance approach requires zero source changes.
   124|
   125|---
   126|
   127|## 4. Player auth (dev)
   128|
   129|Smoke-test that each player is authenticated and has an auto-created Account + Player:
   130|
   131|### Player A
   132|
   133|```sh
   134|curl -s -H "Authorization: Bearer *** http://localhost:8080/v1/me | jq
   135|```
   136|
   137|```json
   138|{
   139|  "account": { "id": "account_...", "email": "alice@example.com" },
   140|  "player": { "id": "player_...", "displayName": "alice" }
   141|}
   142|```
   143|
   144|### Player B
   145|
   146|```sh
   147|curl -s -H "Authorization: Bearer *** http://localhost:8081/v1/me | jq
   148|```
   149|
   150|```json
   151|{
   152|  "account": { "id": "account_...", "email": "bob@example.com" },
   153|  "player": { "id": "player_...", "displayName": "bob" }
   154|}
   155|```
   156|
   157|Save the player IDs for later steps:
   158|
   159|```sh
   160|ALICE=$(curl -s -H "Authorization: Bearer *** http://localhost:8080/v1/me | jq -r '.player.id')
   161|BOB=$(curl -s -H "Authorization: Bearer *** http://localhost:8081/v1/me | jq -r '.player.id')
   162|```
   163|
   164|**Key design point:** `BootstrapIdentity` is idempotent — calling GET /v1/me multiple
   165|times returns the same Account and Player for the same auth identity.
   166|
   167|---
   168|
   169|## 5. Group lifecycle
   170|
   171|### 5a. Player A creates a group
   172|
   173|```sh
   174|curl -s -X POST http://localhost:8080/v1/groups \
   175|  -H "Authorization: Bearer *** \
   176|  -H "Content-Type: application/json" \
   177|  -d '{"name": "Taco Bell Daredevils"}' | jq
   178|```
   179|
   180|Response includes group details, membership role, and (empty) recent jumps + standings:
   181|
   182|```json
   183|{
   184|  "group": { "id": "group_...", "name": "Taco Bell Daredevils" },
   185|  "membership": { "groupId": "group_...", "playerId": "player_...", "role": "Group Admin" },
   186|  "activeSeason": null,
   187|  "recentJumps": [],
   188|  "standings": []
   189|}
   190|```
   191|
   192|Notice `role: "Group Admin"` — the creator automatically gets admin rights.
   193|
   194|Save the group ID:
   195|
   196|```sh
   197|GROUP_ID=$(curl -s -X POST http://localhost:8080/v1/groups \
   198|  -H "Authorization: Bearer *** \
   199|  -H "Content-Type: application/json" \
   200|  -d '{"name": "Taco Bell Daredevils"}' | jq -r '.group.id')
   201|```
   202|
   203|### 5b. Player A lists their groups
   204|
   205|```sh
   206|curl -s -H "Authorization: Bearer *** http://localhost:8080/v1/groups | jq
   207|```
   208|
   209|```json
   210|{
   211|  "memberships": [
   212|    {
   213|      "group": { "id": "group_...", "name": "Taco Bell Daredevils" },
   214|      "membership": { "groupId": "group_...", "playerId": "player_...", "role": "Group Admin" }
   215|    }
   216|  ]
   217|}
   218|```
   219|
   220|Player B's list is empty (not a member yet):
   221|
   222|```sh
   223|curl -s -H "Authorization: Bearer *** http://localhost:8081/v1/groups | jq
   224|```
   225|
   226|```json
   227|{ "memberships": [] }
   228|```
   229|
   230|---
   231|
   232|## 6. Invites & membership
   233|
   234|### 6a. Player A creates an invite
   235|
   236|```sh
   237|INVITE=$(curl -s -X POST http://localhost:8080/v1/groups/$GROUP_ID/invites \
   238|  -H "Authorization: Bearer *** | jq)
   239|
   240|echo "$INVITE" | jq
   241|TOKEN=$(echo "$INVITE" | jq -r '.token')
   242|```
   243|
   244|```json
   245|{
   246|  "id": "invite_...",
   247|  "groupId": "group_...",
   248|  "token": "invite_token_...",
   249|  "createdBy": "player_...",
   250|  "expiresAt": "2026-06-01T..."
   251|}
   252|```
   253|
   254|Invites expire after 7 days. Only group members can create invites (any member, not just admin).
   255|
   256|### 6b. Player B accepts the invite
   257|
   258|```sh
   259|curl -s -X POST http://localhost:8081/v1/invites/$TOKEN/accept \
   260|  -H "Authorization: Bearer *** | jq
   261|```
   262|
   263|Player B now has `role: "Player"` (not Admin) in the group:
   264|
   265|```json
   266|{
   267|  "group": { "id": "group_...", "name": "Taco Bell Daredevils" },
   268|  "membership": { "groupId": "group_...", "playerId": "player_...", "role": "Player" },
   269|  "activeSeason": null,
   270|  "recentJumps": [],
   271|  "standings": []
   272|}
   273|```
   274|
   275|### 6c. Error cases (for demo)
   276|
   277|**Expired invite** — try accepting after `expiresAt`:
   278|```json
   279|// HTTP 410 Gone
   280|"Invite expired"
   281|```
   282|
   283|**Already used invite** — a second attempt:
   284|```json
   285|// HTTP 409 Conflict
   286|"Invite already used"
   287|```
   288|
   289|**Already a member** — try accepting another invite for the same group:
   290|```json
   291|// HTTP 409 Conflict
   292|"Player already has a Group Membership"
   293|```
   294|
   295|**Invalid token**:
   296|```json
   297|// HTTP 404 Not Found
   298|"Invite cannot be accepted"
   299|```
   300|
   301|---
   302|
   303|## 7. Season lifecycle
   304|
   305|### 7a. Player A starts a season
   306|
   307|Deadlines are ISO 8601 timestamps. Set them far enough in the future for the demo.
   308|
   309|```sh
   310|SEASON=$(curl -s -X POST http://localhost:8080/v1/groups/$GROUP_ID/seasons \
   311|  -H "Authorization: Bearer *** \
   312|  -H "Content-Type: application/json" \
   313|  -d '{
   314|    "submissionDeadline": "2026-06-15T23:59:59Z",
   315|    "judgingDeadline": "2026-06-30T23:59:59Z"
   316|  }' | jq)
   317|
   318|echo "$SEASON" | jq
   319|SEASON_ID=$(echo "$SEASON" | jq -r '.activeSeason.id')
   320|```
   321|
   322|```json
   323|{
   324|  "group": { "id": "group_...", "name": "Taco Bell Daredevils" },
   325|  "membership": { "groupId": "group_...", "playerId": "player_...", "role": "Group Admin" },
   326|  "activeSeason": {
   327|    "id": "season_...",
   328|    "groupId": "group_...",
   329|    "commissionerPlayerId": "player_...",
   330|    "status": "Active",
   331|    "submissionDeadline": "2026-06-15T23:59:59Z",
   332|    "judgingDeadline": "2026-06-30T23:59:59Z"
   333|  },
   334|  "recentJumps": [],
   335|  "standings": []
   336|}
   337|```
   338|
   339|The creator becomes **Season Commissioner**. Only one open season per group is allowed.
   340|
   341|### 7b. Duplicate season guard (error case)
   342|
   343|```sh
   344|curl -s -X POST http://localhost:8080/v1/groups/$GROUP_ID/seasons \
   345|  -H "Authorization: Bearer *** \
   346|  -H "Content-Type: application/json" \
   347|  -d '{
   348|    "submissionDeadline": "2026-07-15T23:59:59Z",
   349|    "judgingDeadline": "2026-07-30T23:59:59Z"
   350|  }' | jq
   351|```
   352|
   353|```json
   354|// HTTP 409 Conflict
   355|"Group already has an active or closing Season"
   356|```
   357|
   358|### 7c. Auto-transition (temporal state)
   359|
   360|If `now > submissionDeadline` the season auto-transitions from `Active` → `Judging Grace Period`.
   361|If `now > judgingDeadline` it auto-transitions `Judging Grace Period` → `Finalized`.
   362|
   363|This is checked every time season state is loaded (no cron needed).
   364|
   365|---
   366|
   367|## 8. Jump lifecycle
   368|
   369|The jump lifecycle: `Idea` → `Planned Jump` → `Performed Jump` → `Judged Jump` / `Unwitnessed Jump`
   370|
   371|### 8a. Player A creates an Idea
   372|
   373|Each Idea has a Source (where you buy the food), Destination (where you eat it), and Food.
   374|
   375|```sh
   376|IDEA=$(curl -s -X POST http://localhost:8080/v1/groups/$GROUP_ID/ideas \
   377|  -H "Authorization: Bearer *** \
   378|  -H "Content-Type: application/json" \
   379|  -d '{
   380|    "source": "Taco Bell",
   381|    "destination": "Olive Garden Parking Lot",
   382|    "food": "Crunchwrap Supreme"
   383|  }' | jq)
   384|
   385|echo "$IDEA" | jq
   386|IDEA_ID=$(echo "$IDEA" | jq -r '.id')
   387|```
   388|
   389|```json
   390|{
   391|  "id": "jump_...",
   392|  "groupId": "group_...",
   393|  "playerId": "player_...",
   394|  "seasonId": null,
   395|  "status": "Idea",
   396|  "source": "Taco Bell",
   397|  "destination": "Olive Garden Parking Lot",
   398|  "food": "Crunchwrap Supreme",
   399|  "offSeason": true,
   400|  "finalScore": null
   401|}
   402|```
   403|
   404|Ideas are `offSeason: true` by default and have `seasonId: null`.
   405|
   406|### 8b. Player A promotes the Idea to a Planned Jump
   407|
   408|Since an active season exists and we don't set `offSeason: true`, the jump links to the
   409|active season automatically.
   410|
   411|```sh
   412|PLANNED=$(curl -s -X POST http://localhost:8080/v1/ideas/$IDEA_ID/planned-jump \
   413|  -H "Authorization: Bearer *** \
   414|  -H "Content-Type: application/json" \
   415|  -d '{"offSeason": false}' | jq)
   416|
   417|echo "$PLANNED" | jq
   418|JUMP_ID=$(echo "$PLANNED" | jq -r '.id')
   419|```
   420|
   421|```json
   422|{
   423|  "id": "jump_...",
   424|  "groupId": "group_...",
   425|  "playerId": "player_...",
   426|  "seasonId": "season_...",
   427|  "status": "Planned Jump",
   428|  "source": "Taco Bell",
   429|  "destination": "Olive Garden Parking Lot",
   430|  "food": "Crunchwrap Supreme",
   431|  "offSeason": false,
   432|  "finalScore": null
   433|}
   434|```
   435|
   436|Now `status: "Planned Jump"`, `offSeason: false`, and `seasonId` points to the active season.
   437|
   438|> **Off-Season jumps:** Setting `{"offSeason": true}` keeps the jump outside season
   439|> competition. It can still be performed and judged, but won't appear in standings.
   440|
   441|### 8c. Authorize evidence upload
   442|
   443|Player A authorizes a 15-minute upload window. The response includes a signed `uploadUrl`
   444|and headers — in production this would be a PUT to object storage.
   445|
   446|```sh
   447|AUTH_JSON=$(curl -s -X POST http://localhost:8080/v1/jumps/$JUMP_ID/evidence-upload-authorizations \
   448|  -H "Authorization: Bearer *** \
   449|  -H "Content-Type: application/json" \
   450|  -d '{"contentType": "image/jpeg"}' | jq)
   451|
   452|echo "$AUTH_JSON" | jq
   453|UPLOAD_AUTH_ID=$(echo "$AUTH_JSON" | jq -r '.id')
   454|```
   455|
   456|```json
   457|{
   458|  "id": "evidence_upload_...",
   459|  "jumpId": "jump_...",
   460|  "uploadUrl": "https://storage.supperjumpin.test/uploads/jump_...",
   461|  "uploadMethod": "PUT",
   462|  "uploadHeaders": { "Content-Type": "image/jpeg" },
   463|  "mediaObjectKey": "uploads/jump_.../1",
   464|  "expiresAt": "2026-05-25T..."
   465|}
   466|```
   467|
   468|Only the jump performer can authorize an upload. The auth window lasts 15 minutes.
   469|
   470|### 8d. Submit evidence (perform the jump)
   471|
   472|Player A submits the upload authorization ID and a caption to complete the jump.
   473|
   474|```sh
   475|SUBMISSION=$(curl -s -X POST http://localhost:8080/v1/jumps/$JUMP_ID/evidence \
   476|  -H "Authorization: Bearer *** \
   477|  -H "Content-Type: application/json" \
   478|  -d "{
   479|    \"uploadAuthorizationId\": \"$UPLOAD_AUTH_ID\",
   480|    \"caption\": \"Crunchwrap devoured in the Olive Garden parking lot. Security gave me a look.\"
   481|  }" | jq)
   482|
   483|echo "$SUBMISSION" | jq
   484|```
   485|
   486|```json
   487|{
   488|  "jump": {
   489|    "id": "jump_...",
   490|    "status": "Performed Jump",
   491|    ...
   492|  },
   493|  "evidence": {
   494|    "id": "evidence_...",
   495|    "jumpId": "jump_...",
   496|    "caption": "Crunchwrap devoured in the Olive Garden parking lot. Security gave me a look.",
   497|    "mediaObjectKey": "uploads/jump_.../1",
   498|    "createdAt": "2026-05-25T..."
   499|  }
   500|}
   501|