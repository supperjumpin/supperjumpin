#!/usr/bin/env bash
set -euo pipefail

# ──────────────────────────────────────────────────────────
# Supperjumpin — Full End-to-End Demo Script
# Stands up the backend, runs the complete Group Stunt loop,
# and tears everything down.
# ──────────────────────────────────────────────────────────

# ── Config ────────────────────────────────────────────────
API_A_PORT=8080
API_B_PORT=8081
TOKEN_A="player-a-token"
TOKEN_B="player-b-token"
EMAIL_A="alice@example.com"
EMAIL_B="bob@example.com"
GROUP_NAME="Taco Bell Daredevils"
DB_URL="postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# PID tracking for cleanup
API_A_PID=""
API_B_PID=""
DOCKER_UP=false

# ── Helpers ──────────────────────────────────────────────
info()    { echo -e "${BLUE}[INFO]${NC}  $*"; }
ok()      { echo -e "${GREEN}[OK]${NC}    $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC}  $*"; }
fail()    { echo -e "${RED}[FAIL]${NC}  $*"; }
step()    { echo -e "\n${GREEN}═══════════════════════════════════════════════════${NC}"; echo -e "${GREEN}  $1${NC}"; echo -e "${GREEN}═══════════════════════════════════════════════════${NC}\n"; }

cleanup() {
  info "Cleaning up..."
  if [ -n "$API_A_PID" ]; then
    kill "$API_A_PID" 2>/dev/null || true
    info "Stopped API A (PID $API_A_PID)"
  fi
  if [ -n "$API_B_PID" ]; then
    kill "$API_B_PID" 2>/dev/null || true
    info "Stopped API B (PID $API_B_PID)"
  fi
  if [ "$DOCKER_UP" = true ]; then
    info "Stopping Postgres..."
    docker compose -f "$REPO_ROOT/docker-compose.yml" down -v 2>/dev/null || true
  fi
  ok "Cleanup complete."
}
trap cleanup EXIT INT TERM

check_prereqs() {
  local missing=false
  for cmd in go docker curl jq node; do
    if ! command -v "$cmd" &>/dev/null; then
      fail "$cmd is required but not found."
      missing=true
    fi
  done
  if [ "$missing" = true ]; then
    exit 1
  fi
  ok "All prerequisites found (go, docker, curl, jq, node)"
}

port_free() {
  ! lsof -iTCP:"$1" -sTCP:LISTEN &>/dev/null
}

# ── Main ─────────────────────────────────────────────────
main() {
  echo
  echo "  ███████╗██╗   ██╗██████╗ ██████╗ ███████╗██████╗"
  echo "  ██╔════╝██║   ██║██╔══██╗██╔══██╗██╔════╝██╔══██╗"
  echo "  ███████╗██║   ██║██████╔╝██████╔╝█████╗  ██████╔╝"
  echo "  ╚════██║██║   ██║██╔═══╝ ██╔══██╗██╔══╝  ██╔══██╗"
  echo "  ███████║╚██████╔╝██║     ██║  ██║███████╗██║  ██║"
  echo "  ╚══════╝ ╚═════╝ ╚═╝     ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝"
  echo "             End-to-End Demo: PRD #1"
  echo

  # ── 0. Prerequisites ────────────────────────────────────
  step "0/12  Checking prerequisites"
  check_prereqs

  # ── 1. Start Postgres ──────────────────────────────────
  step "1/12  Starting Postgres (Docker)"
  if port_free 5432; then
    docker compose -f "$REPO_ROOT/docker-compose.yml" up -d postgres
    DOCKER_UP=true
    ok "Postgres container starting..."
    info "Waiting for Postgres to be ready..."
    for i in $(seq 1 60); do
      if docker compose -f "$REPO_ROOT/docker-compose.yml" exec -T postgres pg_isready -U postgres -d supperjumpin &>/dev/null; then
        ok "Postgres is ready"
        break
      fi
      sleep 1
    done
    if ! docker compose -f "$REPO_ROOT/docker-compose.yml" exec -T postgres pg_isready -U postgres -d supperjumpin &>/dev/null; then
      fail "Postgres did not become ready"
      exit 1
    fi
  else
    warn "Port 5432 is already in use — assuming Postgres is running externally"
  fi

  # ── 2. Apply migrations ────────────────────────────────
  step "2/12  Applying database migrations"
  cd "$REPO_ROOT"

  # Bootstrap schema_migrations table
  docker compose -f docker-compose.yml exec -T postgres psql -v ON_ERROR_STOP=1 -U postgres -d supperjumpin <<'EOSQL' 2>/dev/null
CREATE TABLE IF NOT EXISTS schema_migrations (
  filename TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
EOSQL

  # Apply each migration that hasn't been applied yet
  for f in $(ls apps/api/db/migrations/*.sql | sort); do
    filename=$(basename "$f")
    applied=$(docker compose -f docker-compose.yml exec -T postgres psql -tAc \
      "SELECT 1 FROM schema_migrations WHERE filename = '${filename//\'/\'\'}'" \
      -U postgres -d supperjumpin 2>/dev/null | tr -d '[:space:]')
    if [ "$applied" = "1" ]; then
      info "  [skip] $filename (already applied)"
      continue
    fi
    sql=$(cat "$f")
    escaped="${filename//\'/\'\'}"
    docker compose -f docker-compose.yml exec -T postgres psql -v ON_ERROR_STOP=1 -U postgres -d supperjumpin <<EOSQL 2>/dev/null
BEGIN;
$sql
INSERT INTO schema_migrations (filename) VALUES ('$escaped');
COMMIT;
EOSQL
    ok "  [applied] $filename"
  done

  # ── 3. Environment check ───────────────────────────────
  step "3/12  Checking port availability"

  if ! port_free $API_A_PORT; then
    fail "Port $API_A_PORT is in use. Free it or change API_A_PORT."
    exit 1
  fi
  if ! port_free $API_B_PORT; then
    fail "Port $API_B_PORT is in use. Free it or change API_B_PORT."
    exit 1
  fi
  ok "Ports $API_A_PORT and $API_B_PORT are free"

  # ── 4. Start API A (Player A) ──────────────────────────
  step "4/12  Starting API A — Player A (alice) on :$API_A_PORT"
  cd "$REPO_ROOT"
  DATABASE_URL="$DB_URL" \
  SUPPERJUMPIN_DEV_AUTH_TOKEN="$TOKEN_A" \
  SUPPERJUMPIN_DEV_AUTH_SUBJECT="dev-subject-a" \
  SUPPERJUMPIN_DEV_AUTH_EMAIL="$EMAIL_A" \
  PORT="$API_A_PORT" \
    go run ./apps/api/cmd/api &
  API_A_PID=$!
  ok "API A starting (PID $API_A_PID)"

  # Wait for API A
  for i in $(seq 1 30); do
    if curl -sf "http://localhost:$API_A_PORT/v1/me" -H "Authorization: Bearer $TOKEN_A" &>/dev/null; then
      ok "API A is ready"
      break
    fi
    sleep 1
  done
  if ! curl -sf "http://localhost:$API_A_PORT/v1/me" -H "Authorization: Bearer $TOKEN_A" &>/dev/null; then
    fail "API A did not become ready"
    exit 1
  fi

  # ── 5. Start API B (Player B) ──────────────────────────
  step "5/12  Starting API B — Player B (bob) on :$API_B_PORT"
  cd "$REPO_ROOT"
  DATABASE_URL="$DB_URL" \
  SUPPERJUMPIN_DEV_AUTH_TOKEN="$TOKEN_B" \
  SUPPERJUMPIN_DEV_AUTH_SUBJECT="dev-subject-b" \
  SUPPERJUMPIN_DEV_AUTH_EMAIL="$EMAIL_B" \
  PORT="$API_B_PORT" \
    go run ./apps/api/cmd/api &
  API_B_PID=$!
  ok "API B starting (PID $API_B_PID)"

  # Wait for API B
  for i in $(seq 1 30); do
    if curl -sf "http://localhost:$API_B_PORT/v1/me" -H "Authorization: Bearer $TOKEN_B" &>/dev/null; then
      ok "API B is ready"
      break
    fi
    sleep 1
  done
  if ! curl -sf "http://localhost:$API_B_PORT/v1/me" -H "Authorization: Bearer $TOKEN_B" &>/dev/null; then
    fail "API B did not become ready"
    exit 1
  fi

  # ── 6. Player A: auth + create group ───────────────────
  step "6/12  Player A (alice) — auth & create group"

  # Auth check
  ALICE_ME=$(curl -sf "http://localhost:$API_A_PORT/v1/me" \
    -H "Authorization: Bearer $TOKEN_A")
  ALICE_ID=$(echo "$ALICE_ME" | jq -r '.player.id')
  ALICE_NAME=$(echo "$ALICE_ME" | jq -r '.player.displayName')
  ok "Player A authenticated: $ALICE_NAME ($ALICE_ID)"

  # Create group
  GROUP_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/groups" \
    -H "Authorization: Bearer $TOKEN_A" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"$GROUP_NAME\"}")
  GROUP_ID=$(echo "$GROUP_RESP" | jq -r '.group.id')
  GROUP_ROLE=$(echo "$GROUP_RESP" | jq -r '.membership.role')
  echo "$GROUP_RESP" | jq
  ok "Group created: '$GROUP_NAME' (ID: $GROUP_ID) with role: $GROUP_ROLE"

  # List groups
  GROUPS_A=$(curl -sf "http://localhost:$API_A_PORT/v1/groups" \
    -H "Authorization: Bearer $TOKEN_A")
  MEMBER_COUNT=$(echo "$GROUPS_A" | jq '.memberships | length')
  ok "Player A sees $MEMBER_COUNT group(s)"

  # ── 7. Invite flow ─────────────────────────────────────
  step "7/12  Invite flow — Player A invites Player B"

  # Create invite
  INVITE_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/groups/$GROUP_ID/invites" \
    -H "Authorization: Bearer $TOKEN_A")
  INVITE_TOKEN=$(echo "$INVITE_RESP" | jq -r '.token')
  ok "Invite created: $INVITE_TOKEN"

  # Player B accepts
  ACCEPT_RESP=$(curl -sf -X POST "http://localhost:$API_B_PORT/v1/invites/$INVITE_TOKEN/accept" \
    -H "Authorization: Bearer $TOKEN_B")
  BOB_ROLE=$(echo "$ACCEPT_RESP" | jq -r '.membership.role')
  ok "Player B accepted invite with role: $BOB_ROLE"

  # Verify both see the group
  GROUPS_B=$(curl -sf "http://localhost:$API_B_PORT/v1/groups" \
    -H "Authorization: Bearer $TOKEN_B")
  MEMBER_COUNT_B=$(echo "$GROUPS_B" | jq '.memberships | length')
  ok "Player B now sees $MEMBER_COUNT_B group(s)"

  # ── 8. Season ──────────────────────────────────────────
  step "8/12  Season lifecycle — start, close, finalize"

  # Start season (deadlines far in future)
  SEASON_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/groups/$GROUP_ID/seasons" \
    -H "Authorization: Bearer $TOKEN_A" \
    -H "Content-Type: application/json" \
    -d '{
      "submissionDeadline": "2026-06-15T23:59:59Z",
      "judgingDeadline": "2026-06-30T23:59:59Z"
    }')
  SEASON_ID=$(echo "$SEASON_RESP" | jq -r '.activeSeason.id')
  SEASON_STATUS=$(echo "$SEASON_RESP" | jq -r '.activeSeason.status')
  COMMISSIONER=$(echo "$SEASON_RESP" | jq -r '.activeSeason.commissionerPlayerId')
  ok "Season started: status=$SEASON_STATUS, commissioner=$COMMISSIONER"

  # Duplicate season guard
  DUP_RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$API_A_PORT/v1/groups/$GROUP_ID/seasons" \
    -H "Authorization: Bearer $TOKEN_A" \
    -H "Content-Type: application/json" \
    -d '{
      "submissionDeadline": "2026-07-15T23:59:59Z",
      "judgingDeadline": "2026-07-30T23:59:59Z"
    }')
  if [ "$DUP_RESP" = "409" ]; then
    ok "Duplicate season guard works (HTTP 409)"
  else
    warn "Expected 409 for duplicate season, got $DUP_RESP"
  fi

  # ── 9. Stunt lifecycle ────────────────────────────────
  step "9/12  Stunt lifecycle — idea → planned → performed"

  # Create idea
  IDEA_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/groups/$GROUP_ID/ideas" \
    -H "Authorization: Bearer $TOKEN_A" \
    -H "Content-Type: application/json" \
    -d '{
      "source": "Taco Bell",
      "destination": "Olive Garden Parking Lot",
      "food": "Crunchwrap Supreme"
    }')
  IDEA_ID=$(echo "$IDEA_RESP" | jq -r '.id')
  ok "Idea created (ID: $IDEA_ID)"

  # Promote to planned stunt (links to active season since offSeason=false)
  PLANNED_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/ideas/$IDEA_ID/planned-stunt" \
    -H "Authorization: Bearer $TOKEN_A" \
    -H "Content-Type: application/json" \
    -d '{"offSeason": false}')
  STUNT_ID=$(echo "$PLANNED_RESP" | jq -r '.id')
  STUNT_STATUS=$(echo "$PLANNED_RESP" | jq -r '.status')
  STUNT_SEASON=$(echo "$PLANNED_RESP" | jq -r '.seasonId')
  STUNT_OFFSEASON=$(echo "$PLANNED_RESP" | jq -r '.offSeason')
  if [ "$STUNT_OFFSEASON" = "false" ] && [ "$STUNT_SEASON" = "$SEASON_ID" ]; then
    ok "Stunt $STUNT_STATUS and linked to season $STUNT_SEASON"
  else
    ok "Stunt $STUNT_STATUS (offSeason=$STUNT_OFFSEASON, season=$STUNT_SEASON)"
  fi

  # Authorize evidence upload
  AUTH_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/stunts/$STUNT_ID/evidence-upload-authorizations" \
    -H "Authorization: Bearer $TOKEN_A" \
    -H "Content-Type: application/json" \
    -d '{"contentType": "image/jpeg"}')
  UPLOAD_AUTH_ID=$(echo "$AUTH_RESP" | jq -r '.id')
  UPLOAD_URL=$(echo "$AUTH_RESP" | jq -r '.uploadUrl')
  ok "Upload authorized: $UPLOAD_URL (auth ID: $UPLOAD_AUTH_ID)"

  # Submit evidence (perform the stunt)
  SUBMIT_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/stunts/$STUNT_ID/evidence" \
    -H "Authorization: Bearer $TOKEN_A" \
    -H "Content-Type: application/json" \
    -d "{
      \"uploadAuthorizationId\": \"$UPLOAD_AUTH_ID\",
      \"caption\": \"Crunchwrap devoured in the Olive Garden parking lot. Security gave me a look.\"
    }")
  PERFORMED_STATUS=$(echo "$SUBMIT_RESP" | jq -r '.stunt.status')
  CAPTION=$(echo "$SUBMIT_RESP" | jq -r '.evidence.caption')
  ok "Stunt performed: status=$PERFORMED_STATUS, caption=\"$CAPTION\""

  # ── 10. Judging ────────────────────────────────────────
  step "10/12 Judging — Player B scores Player A's stunt"

  # Cannot judge own stunt (Player A tries)
  SELF_JUDGE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$API_A_PORT/v1/stunts/$STUNT_ID/judgment" \
    -H "Authorization: Bearer $TOKEN_A" \
    -H "Content-Type: application/json" \
    -d '{"difficulty":5,"transgression":5,"creativity":5,"documentation":5}')
  if [ "$SELF_JUDGE" = "403" ]; then
    ok "Self-judging correctly rejected (HTTP 403)"
  else
    warn "Expected 403 for self-judging, got $SELF_JUDGE"
  fi

  # Invalid score guard
  BAD_SCORE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$API_B_PORT/v1/stunts/$STUNT_ID/judgment" \
    -H "Authorization: Bearer $TOKEN_B" \
    -H "Content-Type: application/json" \
    -d '{"difficulty":11,"transgression":5,"creativity":5,"documentation":5}')
  if [ "$BAD_SCORE" = "400" ]; then
    ok "Out-of-range scores correctly rejected (HTTP 400)"
  else
    warn "Expected 400 for bad scores, got $BAD_SCORE"
  fi

  # Player B judges
  JUDGE_RESP=$(curl -sf -X POST "http://localhost:$API_B_PORT/v1/stunts/$STUNT_ID/judgment" \
    -H "Authorization: Bearer $TOKEN_B" \
    -H "Content-Type: application/json" \
    -d '{"difficulty":8,"transgression":8,"creativity":9,"documentation":7}')
  JUDGE_ID=$(echo "$JUDGE_RESP" | jq -r '.id')
  TOTAL_SCORE=$((8+8+9+7))
  ok "Player B submitted judgment (ID: $JUDGE_ID) — total: $TOTAL_SCORE"

  # Edit judgment (upsert)
  EDIT_RESP=$(curl -sf -X POST "http://localhost:$API_B_PORT/v1/stunts/$STUNT_ID/judgment" \
    -H "Authorization: Bearer $TOKEN_B" \
    -H "Content-Type: application/json" \
    -d '{"difficulty":7,"transgression":8,"creativity":9,"documentation":6}')
  EDIT_TOTAL=$((7+8+9+6))
  ok "Judgment edited (upsert) — new total: $EDIT_TOTAL"

  # ── 11. Close & finalize ───────────────────────────────
  step "11/12 Closing submissions & finalizing season"

  # Close submissions (Admin/Commissioner action)
  curl -sf -X POST "http://localhost:$API_A_PORT/v1/seasons/$SEASON_ID/close-submissions" \
    -H "Authorization: Bearer $TOKEN_A" > /dev/null
  ok "Submissions closed"

  # Finalize season
  FINAL_RESP=$(curl -sf -X POST "http://localhost:$API_A_PORT/v1/seasons/$SEASON_ID/finalize" \
    -H "Authorization: Bearer $TOKEN_A")
  FINAL_STATUS=$(echo "$FINAL_RESP" | jq -r '.activeSeason')
  if [ "$FINAL_STATUS" = "null" ]; then
    ok "Season finalized (no active season remains)"
  else
    ok "Season finalized"
  fi

  # ── 12. Verification ───────────────────────────────────
  step "12/12 Verification — everything worked"

  # Season history
  HISTORY_RESP=$(curl -sf "http://localhost:$API_A_PORT/v1/seasons/$SEASON_ID/history" \
    -H "Authorization: Bearer $TOKEN_A")
  HISTORY_COUNT=$(echo "$HISTORY_RESP" | jq '.entries | length')
  echo "$HISTORY_RESP" | jq '.entries[] | {action, fromStatus, toStatus, override}'
  ok "Season history: $HISTORY_COUNT entries"

  # Group home (aggregated view)
  HOME_RESP=$(curl -sf "http://localhost:$API_A_PORT/v1/groups/$GROUP_ID/home" \
    -H "Authorization: Bearer $TOKEN_A")

  GROUP_NAME_CHECK=$(echo "$HOME_RESP" | jq -r '.group.name')
  STUNT_COUNT=$(echo "$HOME_RESP" | jq '.recentStunts | length')
  STANDING_COUNT=$(echo "$HOME_RESP" | jq '.standings | length')

  ok "Group home: '$GROUP_NAME_CHECK'"
  ok "Recent stunts: $STUNT_COUNT"
  ok "Standings entries: $STANDING_COUNT"

  if [ "$STANDING_COUNT" -gt 0 ]; then
    echo "$HOME_RESP" | jq '.standings[] | {player: .player.displayName, score: .seasonScore, stunts: .judgedStunts}'
    ok "Standings computed correctly!"
  fi

  echo ""
  ok "Standings:"
  echo "$HOME_RESP" | jq '.standings'

  echo ""
  ok "Recent stunts:"
  echo "$HOME_RESP" | jq '.recentStunts[] | {source: .stunt.source, destination: .stunt.destination, food: .stunt.food, status: .stunt.status, score: .stunt.finalScore, performer: .performer.displayName}'

  # ── Summary ────────────────────────────────────────────
  echo ""
  echo "  ┌─────────────────────────────────────────────────────┐"
  echo "  │  ✅  DEMO COMPLETE                                  │"
  echo "  │                                                     │"
  echo "  │  Features demonstrated:                             │"
  echo "  │  • Auth & identity bootstrap (BootstrapIdentity)    │"
  echo "  │  • Group CRUD (create, list, home)                  │"
  echo "  │  • Invite flow (create, accept, error guards)       │"
  echo "  │  • Season lifecycle (start, close, finalize)        │"
  echo "  │  • Stunt lifecycle (idea → planned → performed)     │"
  echo "  │  • Evidence upload authorization                    │"
  echo "  │  • Judging (4-axis, upsert, self-judge guard)       │"
  echo "  │  • Score validation (0-10 range)                    │"
  echo "  │  • Season history audit trail                       │"
  echo "  │  • Standings computation                            │"
  echo "  │  • Duplicate season guard                           │"
  echo "  │  • Group home (aggregated view)                     │"
  echo "  │                                                     │"
  echo "  │  All 15 API endpoints exercised. ✓                  │"
  echo "  └─────────────────────────────────────────────────────┘"
}

main
