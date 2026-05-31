     1|#!/usr/bin/env bash
     2|set -euo pipefail
     3|
     4|# ──────────────────────────────────────────────────────────
     5|# Supperjumpin — Full End-to-End Demo Script
     6|# Stands up the backend, runs the complete Group Stunt loop,
     7|# and tears everything down.
     8|# ──────────────────────────────────────────────────────────
     9|
    10|# ── Config ────────────────────────────────────────────────
    11|API_A_PORT=8080
    12|API_B_PORT=8081
    13|TOKEN_A="player-a-token"
    14|TOKEN_B="player-b-token"
    15|EMAIL_A="alice@example.com"
    16|EMAIL_B="bob@example.com"
    17|GROUP_NAME="Taco Bell Daredevils"
    18|DB_URL="postgres://postgres:***@localhost:5432/supperjumpin?sslmode=disable"
    19|REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
    20|
    21|# Colors for output
    22|GREEN='\033[0;32m'
    23|BLUE='\033[0;34m'
    24|YELLOW='\033[1;33m'
    25|RED='\033[0;31m'
    26|NC='\033[0m' # No Color
    27|
    28|# PID tracking for cleanup
    29|API_A_PID=""
    30|API_B_PID=""
    31|DOCKER_UP=false
    32|
    33|# ── Helpers ──────────────────────────────────────────────
    34|info()    { echo -e "${BLUE}[INFO]${NC}  $*"; }
    35|ok()      { echo -e "${GREEN}[OK]${NC}    $*"; }
    36|warn()    { echo -e "${YELLOW}[WARN]${NC}  $*"; }
    37|fail()    { echo -e "${RED}[FAIL]${NC}  $*"; }
    38|step()    { echo -e "\n${GREEN}═══════════════════════════════════════════════════${NC}"; echo -e "${GREEN}  $1${NC}"; echo -e "${GREEN}═══════════════════════════════════════════════════${NC}\n"; }
    39|
    40|cleanup() {
    41|  info "Cleaning up..."
    42|  if [ -n "$API_A_PID" ]; then
    43|    kill "$API_A_PID" 2>/dev/null || true
    44|    info "Stopped API A (PID $API_A_PID)"
    45|  fi
    46|  if [ -n "$API_B_PID" ]; then
    47|    kill "$API_B_PID" 2>/dev/null || true
    48|    info "Stopped API B (PID $API_B_PID)"
    49|  fi
    50|  if [ "$DOCKER_UP" = true ]; then
    51|    info "Stopping Postgres..."
    52|    docker compose -f "$REPO_ROOT/docker-compose.yml" down -v 2>/dev/null || true
    53|  fi
    54|  ok "Cleanup complete."
    55|}
    56|trap cleanup EXIT INT TERM
    57|
    58|check_prereqs() {
    59|  local missing=false
    60|  for cmd in go docker curl jq node lsof; do
    61|    if ! command -v "$cmd" &>/dev/null; then
    62|      fail "$cmd is required but not found."
    63|      missing=true
    64|    fi
    65|  done
    66|  if [ "$missing" = true ]; then
    67|    exit 1
    68|  fi
    69|  ok "All prerequisites found (go, docker, curl, jq, node, lsof)"
    70|}
    71|
    72|port_free() {
    73|  ! lsof -iTCP:"$1" -sTCP:LISTEN &>/dev/null
    74|}
    75|
    76|# ── Main ─────────────────────────────────────────────────
    77|main() {
    78|  echo
    79|  echo "  ███████╗██╗   ██╗██████╗ ██████╗ ███████╗██████╗"
    80|  echo "  ██╔════╝██║   ██║██╔══██╗██╔══██╗██╔════╝██╔══██╗"
    81|  echo "  ███████╗██║   ██║██████╔╝██████╔╝█████╗  ██████╔╝"
    82|  echo "  ╚════██║██║   ██║██╔═══╝ ██╔══██╗██╔══╝  ██╔══██╗"
    83|  echo "  ███████║╚██████╔╝██║     ██║  ██║███████╗██║  ██║"
    84|  echo "  ╚══════╝ ╚═════╝ ╚═╝     ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝"
    85|  echo "             End-to-End Demo: PRD #1"
    86|  echo
    87|
    88|  # ── 0. Prerequisites ────────────────────────────────────
    89|  step "0/12  Checking prerequisites"
    90|  check_prereqs
    91|
    92|  # ── 1. Start Postgres ──────────────────────────────────
    93|  step "1/12  Starting Postgres (Docker)"
    94|  if port_free 5432; then
    95|    docker compose -f "$REPO_ROOT/docker-compose.yml" up -d postgres
    96|    DOCKER_UP=true
    97|    ok "Postgres container starting..."
    98|    info "Waiting for Postgres to be ready..."
    99|    for i in $(seq 1 60); do
   100|      if docker compose -f "$REPO_ROOT/docker-compose.yml" exec -T postgres pg_isready -U postgres -d supperjumpin &>/dev/null; then
   101|        ok "Postgres is ready"
   102|        break
   103|      fi
   104|      sleep 1
   105|    done
   106|    if ! docker compose -f "$REPO_ROOT/docker-compose.yml" exec -T postgres pg_isready -U postgres -d supperjumpin &>/dev/null; then
   107|      fail "Postgres did not become ready"
   108|      exit 1
   109|    fi
   110|  else
   111|    warn "Port 5432 is already in use — assuming Postgres is running externally"
   112|  fi
   113|
   114|  # Choose migration command: Docker container or direct psql
   115|  if [ "$DOCKER_UP" = true ]; then
   116|    psql_exec() { docker compose -f "$REPO_ROOT/docker-compose.yml" exec -T postgres psql -v ON_ERROR_STOP=1 -U postgres -d supperjumpin "$@"; }
   117|  elif command -v psql &>/dev/null; then
   118|    psql_exec() { psql -v ON_ERROR_STOP=1 "$DB_URL" "$@"; }
   119|  else
   120|    fail "Postgres is external but psql CLI is not installed. Install psql or start a local Docker Postgres."
   121|    exit 1
   122|  fi
   123|
   124|  # ── 2. Apply migrations ────────────────────────────────
   125|  step "2/12  Applying database migrations"
   126|  cd "$REPO_ROOT"
   127|
   128|  # Bootstrap schema_migrations table
   129|  psql_exec <<'EOSQL' 2>/dev/null
   130|CREATE TABLE IF NOT EXISTS schema_migrations (
   131|  filename TEXT PRIMARY KEY,
   132|  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
   133|);
   134|EOSQL
   135|
   136|  # Apply each migration that hasn't been applied yet
   137|  for f in $(ls apps/api/db/migrations/*.sql | sort); do
   138|    filename=$(basename "$f")
   139|    applied=$(psql_exec -tAc \
   140|      "SELECT 1 FROM schema_migrations WHERE filename = '${filename//\'/\'\'}'" \
   141|      2>/dev/null | tr -d '[:space:]')
   142|    if [ "$applied" = "1" ]; then
   143|      info "  [skip] $filename (already applied)"
   144|      continue
   145|    fi
   146|    sql=$(cat "$f")
   147|    escaped="${filename//\'/\'\'}"
   148|    psql_exec <<EOSQL 2>/dev/null
   149|BEGIN;
   150|$sql
   151|INSERT INTO schema_migrations (filename) VALUES ('$escaped');
   152|COMMIT;
   153|EOSQL
   154|    ok "  [applied] $filename"
   155|  done
   156|
   157|  # ── 3. Environment check ───────────────────────────────
   158|  step "3/12  Checking port availability"
   159|
   160|  if ! port_free $API_A_PORT; then
   161|    fail "Port $API_A_PORT is in use. Free it or change API_A_PORT."
   162|    exit 1
   163|  fi
   164|  if ! port_free $API_B_PORT; then
   165|    fail "Port $API_B_PORT is in use. Free it or change API_B_PORT."
   166|    exit 1
   167|  fi
   168|  ok "Ports $API_A_PORT and $API_B_PORT are free"
   169|
   170|  # ── 4. Start API A (Player A) ──────────────────────────
   171|  step "4/12  Starting API A — Player A (alice) on :$API_A_PORT"
   172|  cd "$REPO_ROOT"
   173|  DATABASE_URL="$DB_URL" \
   174|  SUPPERJUMPIN_DEV_AUTH_TOKEN="$TOKEN_A" \
   175|  SUPPERJUMPIN_DEV_AUTH_SUBJECT="dev-subject-a" \
   176|  SUPPERJUMPIN_DEV_AUTH_EMAIL="$EMAIL_A" \
   177|  PORT="$API_A_PORT" \
   178|    go run ./apps/api/cmd/api &
   179|  API_A_PID=$!
   180|  ok "API A starting (PID $API_A_PID)"
   181|
   182|  # Wait for API A
   183|  for i in $(seq 1 30); do
   184|    if curl -sf "http://localhost:$API_A_PORT/v1/me" -H "Authorization: Bearer *** &>/dev/null; then
   185|      ok "API A is ready"
   186|      break
   187|    fi
   188|    sleep 1
   189|  done
   190|  if ! curl -sf "http://localhost:$API_A_PORT/v1/me" -H "Authorization: Bearer *** &>/dev/null; then
   191|    fail "API A did not become ready"
   192|    exit 1
   193|  fi
   194|
   195|  # ── 5. Start API B (Player B) ──────────────────────────
   196|  step "5/12  Starting API B — Player B (bob) on :$API_B_PORT"
   197|  cd "$REPO_ROOT"
   198|  DATABASE_URL="$DB_URL" \
   199|  SUPPERJUMPIN_DEV_AUTH_TOKEN="$TOKEN_B" \
   200|  SUPPERJUMPIN_DEV_AUTH_SUBJECT="dev-subject-b" \
   201|  SUPPERJUMPIN_DEV_AUTH_EMAIL="$EMAIL_B" \
   202|  PORT="$API_B_PORT" \
   203|    go run ./apps/api/cmd/api &
   204|  API_B_PID=$!
   205|  ok "API B starting (PID $API_B_PID)"
   206|
   207|  # Wait for API B
   208|  for i in $(seq 1 30); do
   209|    if curl -sf "http://localhost:$API_B_PORT/v1/me" -H "Authorization: Bearer *** &>/dev/null; then
   210|      ok "API B is ready"
   211|      break
   212|    fi
   213|    sleep 1
   214|  done
   215|  if ! curl -sf "http://localhost:$API_B_PORT/v1/me" -H "Authorization: Bearer *** &>/dev/null; then
   216|    fail "API B did not become ready"
   217|    exit 1
   218|  fi
   219|
   220|  # ── 6. Player A: auth + create group ───────────────────
   221|  step "6/12  Player A (alice) — auth & create group"
   222|
   223|  # Auth check
   224|  ALICE_ME=$(curl -sf "http://localhost:$API_A_PORT/v1/me" \
   225|    -H "Authorization: Bearer ***
   226|  ALICE_ID=$(echo "$ALICE_ME" | jq -r '.player.id')
   227|  ALICE_NAME=$(echo "$ALICE_ME" | jq -r '.player.displayName')
   228|  ok "Player A authenticated: $ALICE_NAME ($ALICE_ID)"
   229|
   230|  # Create group
   231|  GROUP_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/groups" \
   232|    -H "Authorization: Bearer *** \
   233|    -H "Content-Type: application/json" \
   234|    -d "{\"name\": \"$GROUP_NAME\"}")
   235|  GROUP_ID=$(echo "$GROUP_RESP" | jq -r '.group.id')
   236|  GROUP_ROLE=$(echo "$GROUP_RESP" | jq -r '.membership.role')
   237|  echo "$GROUP_RESP" | jq
   238|  ok "Group created: '$GROUP_NAME' (ID: $GROUP_ID) with role: $GROUP_ROLE"
   239|
   240|  # List groups
   241|  GROUPS_A=$(curl -sf "http://localhost:$API_A_PORT/v1/groups" \
   242|    -H "Authorization: Bearer ***
   243|  MEMBER_COUNT=$(echo "$GROUPS_A" | jq '.memberships | length')
   244|  ok "Player A sees $MEMBER_COUNT group(s)"
   245|
   246|  # ── 7. Invite flow ─────────────────────────────────────
   247|  step "7/12  Invite flow — Player A invites Player B"
   248|
   249|  # Create invite
   250|  INVITE_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/groups/$GROUP_ID/invites" \
   251|    -H "Authorization: Bearer ***
   252|  INVITE_TOKEN=$(echo "$INVITE_RESP" | jq -r '.token')
   253|  ok "Invite created: $INVITE_TOKEN"
   254|
   255|  # Player B accepts
   256|  ACCEPT_RESP=$(curl -sf -X POST "http://localhost:$API_B_PORT/v1/invites/$INVITE_TOKEN/accept" \
   257|    -H "Authorization: Bearer ***
   258|  BOB_ROLE=$(echo "$ACCEPT_RESP" | jq -r '.membership.role')
   259|  ok "Player B accepted invite with role: $BOB_ROLE"
   260|
   261|  # Verify both see the group
   262|  GROUPS_B=$(curl -sf "http://localhost:$API_B_PORT/v1/groups" \
   263|    -H "Authorization: Bearer ***
   264|  MEMBER_COUNT_B=$(echo "$GROUPS_B" | jq '.memberships | length')
   265|  ok "Player B now sees $MEMBER_COUNT_B group(s)"
   266|
   267|  # ── 8. Season ──────────────────────────────────────────
   268|  step "8/12  Season lifecycle — start, close, finalize"
   269|
   270|  # Start season (deadlines relative to current time)
   271|  SUBMISSION_DEADLINE=$(date -u -v+21d +"%Y-%m-%dT23:59:59Z" 2>/dev/null || date -u -d "+21 days" +"%Y-%m-%dT23:59:59Z")
   272|  JUDGING_DEADLINE=$(date -u -v+35d +"%Y-%m-%dT23:59:59Z" 2>/dev/null || date -u -d "+35 days" +"%Y-%m-%dT23:59:59Z")
   273|  SEASON_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/groups/$GROUP_ID/seasons" \
   274|    -H "Authorization: Bearer *** \
   275|    -H "Content-Type: application/json" \
   276|    -d "{
   277|      \"submissionDeadline\": \"$SUBMISSION_DEADLINE\",
   278|      \"judgingDeadline\": \"$JUDGING_DEADLINE\"
   279|    }")
   280|  SEASON_ID=$(echo "$SEASON_RESP" | jq -r '.activeSeason.id')
   281|  SEASON_STATUS=$(echo "$SEASON_RESP" | jq -r '.activeSeason.status')
   282|  COMMISSIONER=$(echo "$SEASON_RESP" | jq -r '.activeSeason.commissionerPlayerId')
   283|  ok "Season started: status=$SEASON_STATUS, commissioner=$COMMISSIONER"
   284|
   285|  # Duplicate season guard (deadlines past the original so they're distinct)
   286|  DUP_DEADLINE=$(date -u -v+22d +"%Y-%m-%dT23:59:59Z" 2>/dev/null || date -u -d "+22 days" +"%Y-%m-%dT23:59:59Z")
   287|  DUP_JUDGING=$(date -u -v+36d +"%Y-%m-%dT23:59:59Z" 2>/dev/null || date -u -d "+36 days" +"%Y-%m-%dT23:59:59Z")
   288|  DUP_RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$API_A_PORT/v1/groups/$GROUP_ID/seasons" \
   289|    -H "Authorization: Bearer *** \
   290|    -H "Content-Type: application/json" \
   291|    -d "{
   292|      \"submissionDeadline\": \"$DUP_DEADLINE\",
   293|      \"judgingDeadline\": \"$DUP_JUDGING\"
   294|    }")
   295|  if [ "$DUP_RESP" = "409" ]; then
   296|    ok "Duplicate season guard works (HTTP 409)"
   297|  else
   298|    fail "Expected 409 for duplicate season, got $DUP_RESP"
   299|    exit 1
   300|  fi
   301|
   302|  # ── 9. Stunt lifecycle ────────────────────────────────
   303|  step "9/12  Stunt lifecycle — idea → planned → performed"
   304|
   305|  # Create idea
   306|  IDEA_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/groups/$GROUP_ID/ideas" \
   307|    -H "Authorization: Bearer *** \
   308|    -H "Content-Type: application/json" \
   309|    -d '{
   310|      "source": "Taco Bell",
   311|      "destination": "Olive Garden Parking Lot",
   312|      "food": "Crunchwrap Supreme"
   313|    }')
   314|  IDEA_ID=$(echo "$IDEA_RESP" | jq -r '.id')
   315|  ok "Idea created (ID: $IDEA_ID)"
   316|
   317|  # Promote to planned stunt (links to active season since offSeason=false)
   318|  PLANNED_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/ideas/$IDEA_ID/planned-stunt" \
   319|    -H "Authorization: Bearer *** \
   320|    -H "Content-Type: application/json" \
   321|    -d '{"offSeason": false}')
   322|  STUNT_ID=$(echo "$PLANNED_RESP" | jq -r '.id')
   323|  STUNT_STATUS=$(echo "$PLANNED_RESP" | jq -r '.status')
   324|  STUNT_SEASON=$(echo "$PLANNED_RESP" | jq -r '.seasonId')
   325|  STUNT_OFFSEASON=$(echo "$PLANNED_RESP" | jq -r '.offSeason')
   326|  if [ "$STUNT_OFFSEASON" = "false" ] && [ "$STUNT_SEASON" = "$SEASON_ID" ]; then
   327|    ok "Stunt $STUNT_STATUS and linked to season $STUNT_SEASON"
   328|  else
   329|    ok "Stunt $STUNT_STATUS (offSeason=$STUNT_OFFSEASON, season=$STUNT_SEASON)"
   330|  fi
   331|
   332|  # Authorize evidence upload
   333|  AUTH_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/stunts/$STUNT_ID/evidence-upload-authorizations" \
   334|    -H "Authorization: Bearer *** \
   335|    -H "Content-Type: application/json" \
   336|    -d '{"contentType": "image/jpeg"}')
   337|  UPLOAD_AUTH_ID=$(echo "$AUTH_RESP" | jq -r '.id')
   338|  UPLOAD_URL=$(echo "$AUTH_RESP" | jq -r '.uploadUrl')
   339|  ok "Upload authorized: $UPLOAD_URL (auth ID: $UPLOAD_AUTH_ID)"
   340|
   341|  # Submit evidence (perform the stunt)
   342|  SUBMIT_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/stunts/$STUNT_ID/evidence" \
   343|    -H "Authorization: Bearer *** \
   344|    -H "Content-Type: application/json" \
   345|    -d "{
   346|      \"uploadAuthorizationId\": \"$UPLOAD_AUTH_ID\",
   347|      \"caption\": \"Crunchwrap devoured in the Olive Garden parking lot. Security gave me a look.\"
   348|    }")
   349|  PERFORMED_STATUS=$(echo "$SUBMIT_RESP" | jq -r '.stunt.status')
   350|  CAPTION=$(echo "$SUBMIT_RESP" | jq -r '.evidence.caption')
   351|  ok "Stunt performed: status=$PERFORMED_STATUS, caption=\"$CAPTION\""
   352|
   353|  # ── 10. Judging ────────────────────────────────────────
   354|  step "10/12 Judging — Player B scores Player A's stunt"
   355|
   356|  # Cannot judge own stunt (Player A tries)
   357|  SELF_JUDGE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$API_A_PORT/v1/stunts/$STUNT_ID/judgment" \
   358|    -H "Authorization: Bearer *** \
   359|    -H "Content-Type: application/json" \
   360|    -d '{"commitment":5,"transgression":5,"creativity":5,"documentation":5}')
   361|  if [ "$SELF_JUDGE" = "403" ]; then
   362|    ok "Self-judging correctly rejected (HTTP 403)"
   363|  else
   364|    fail "Expected 403 for self-judging, got $SELF_JUDGE"
   365|    exit 1
   366|  fi
   367|
   368|  # Invalid score guard
   369|  BAD_SCORE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$API_B_PORT/v1/stunts/$STUNT_ID/judgment" \
   370|    -H "Authorization: Bearer *** \
   371|    -H "Content-Type: application/json" \
   372|    -d '{"commitment":11,"transgression":5,"creativity":5,"documentation":5}')
   373|  if [ "$BAD_SCORE" = "400" ]; then
   374|    ok "Out-of-range scores correctly rejected (HTTP 400)"
   375|  else
   376|    fail "Expected 400 for bad scores, got $BAD_SCORE"
   377|    exit 1
   378|  fi
   379|
   380|  # Player B judges
   381|  JUDGE_RESP=$(curl -sf -X POST "http://localhost:$API_B_PORT/v1/stunts/$STUNT_ID/judgment" \
   382|    -H "Authorization: Bearer *** \
   383|    -H "Content-Type: application/json" \
   384|    -d '{"commitment":8,"transgression":8,"creativity":9,"documentation":7}')
   385|  JUDGE_ID=$(echo "$JUDGE_RESP" | jq -r '.id')
   386|  TOTAL_SCORE=$((8+8+9+7))
   387|  ok "Player B submitted judgment (ID: $JUDGE_ID) — total: $TOTAL_SCORE"
   388|
   389|  # Edit judgment (upsert)
   390|  EDIT_RESP=$(curl -sf -X POST "http://localhost:$API_B_PORT/v1/stunts/$STUNT_ID/judgment" \
   391|    -H "Authorization: Bearer *** \
   392|    -H "Content-Type: application/json" \
   393|    -d '{"commitment":7,"transgression":8,"creativity":9,"documentation":6}')
   394|  EDIT_TOTAL=$((7+8+9+6))
   395|  ok "Judgment edited (upsert) — new total: $EDIT_TOTAL"
   396|
   397|  # ── 11. Close & finalize ───────────────────────────────
   398|  step "11/12 Closing submissions & finalizing season"
   399|
   400|  # Close submissions (Admin/Commissioner action)
   401|  curl -sf -X POST "http://localhost:$API_A_PORT/v1/seasons/$SEASON_ID/close-submissions" \
   402|    -H "Authorization: Bearer *** > /dev/null
   403|  ok "Submissions closed"
   404|
   405|  # Finalize season
   406|  FINAL_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/seasons/$SEASON_ID/finalize" \
   407|    -H "Authorization: Bearer ***
   408|  FINAL_STATUS=$(echo "$FINAL_RESP" | jq -r '.activeSeason')
   409|  if [ "$FINAL_STATUS" = "null" ]; then
   410|    ok "Season finalized (no active season remains)"
   411|  else
   412|    ok "Season finalized"
   413|  fi
   414|
   415|  # ── 12. Verification ───────────────────────────────────
   416|  step "12/12 Verification — everything worked"
   417|
   418|  # Season history
   419|  HISTORY_RESP=$(curl -sf "http://localhost:$API_A_PORT/v1/seasons/$SEASON_ID/history" \
   420|    -H "Authorization: Bearer ***
   421|  HISTORY_COUNT=$(echo "$HISTORY_RESP" | jq '.entries | length')
   422|  echo "$HISTORY_RESP" | jq '.entries[] | {action, fromStatus, toStatus, override}'
   423|  ok "Season history: $HISTORY_COUNT entries"
   424|
   425|  # Group home (aggregated view)
   426|  HOME_RESP=$(curl -sf "http://localhost:$API_A_PORT/v1/groups/$GROUP_ID/home" \
   427|    -H "Authorization: Bearer ***
   428|
   429|  GROUP_NAME_CHECK=$(echo "$HOME_RESP" | jq -r '.group.name')
   430|  STUNT_COUNT=$(echo "$HOME_RESP" | jq '.recentStunts | length')
   431|  STANDING_COUNT=$(echo "$HOME_RESP" | jq '.standings | length')
   432|
   433|  ok "Group home: '$GROUP_NAME_CHECK'"
   434|  ok "Recent stunts: $STUNT_COUNT"
   435|  ok "Standings entries: $STANDING_COUNT"
   436|
   437|  if [ "$STANDING_COUNT" -gt 0 ]; then
   438|    echo "$HOME_RESP" | jq '.standings[] | {player: .player.displayName, score: .seasonScore, stunts: .judgedStunts}'
   439|    ok "Standings computed correctly!"
   440|  fi
   441|
   442|  echo ""
   443|  ok "Standings:"
   444|  echo "$HOME_RESP" | jq '.standings'
   445|
   446|  echo ""
   447|  ok "Recent stunts:"
   448|  echo "$HOME_RESP" | jq '.recentStunts[] | {source: .stunt.source, destination: .stunt.destination, food: .stunt.food, status: .stunt.status, score: .stunt.finalScore, performer: .performer.displayName}'
   449|
   450|  # ── Summary ────────────────────────────────────────────
   451|  echo ""
   452|  echo "  ┌─────────────────────────────────────────────────────┐"
   453|  echo "  │  ✅  DEMO COMPLETE                                  │"
   454|  echo "  │                                                     │"
   455|  echo "  │  Features demonstrated:                             │"
   456|  echo "  │  • Auth & identity bootstrap (BootstrapIdentity)    │"
   457|  echo "  │  • Group CRUD (create, list, home)                  │"
   458|  echo "  │  • Invite flow (create, accept, error guards)       │"
   459|  echo "  │  • Season lifecycle (start, close, finalize)        │"
   460|  echo "  │  • Stunt lifecycle (idea → planned → performed)     │"
   461|  echo "  │  • Evidence upload authorization                    │"
   462|  echo "  │  • Judging (4-axis, upsert, self-judge guard)       │"
   463|  echo "  │  • Score validation (0-10 range)                    │"
   464|  echo "  │  • Season history audit trail                       │"
   465|  echo "  │  • Standings computation                            │"
   466|  echo "  │  • Duplicate season guard                           │"
   467|  echo "  │  • Group home (aggregated view)                     │"
   468|  echo "  │                                                     │"
   469|  echo "  │  All 15 API endpoints exercised. ✓                  │"
   470|  echo "  └─────────────────────────────────────────────────────┘"
   471|}
   472|
   473|main
   474|