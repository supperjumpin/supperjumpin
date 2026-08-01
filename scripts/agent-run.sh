#!/usr/bin/env bash
set -euo pipefail

safe_component() {
  [[ "$1" =~ ^[A-Za-z0-9._-]+$ ]] && [[ "$1" != "." ]] && [[ "$1" != ".." ]]
}

: "${AGENT_TASK_ID:?AGENT_TASK_ID is required}"
: "${AGENT_ATTEMPT_ID:?AGENT_ATTEMPT_ID is required}"

if ! safe_component "$AGENT_TASK_ID" || ! safe_component "$AGENT_ATTEMPT_ID"; then
  echo "AGENT_TASK_ID and AGENT_ATTEMPT_ID may contain only letters, digits, dot, underscore, or hyphen" >&2
  exit 2
fi

repo_root=$(git rev-parse --show-toplevel)
checksum=$(printf '%s' "$AGENT_TASK_ID:$AGENT_ATTEMPT_ID" | cksum)
checksum=${checksum%% *}
container_name="supperjumpin-agent-$checksum"
database_name="supperjumpin_agent_${checksum}_test"
image=${SUPPERJUMPIN_AGENT_POSTGRES_IMAGE:-postgres:16}
container_started=false

cleanup() {
  if [[ "$container_started" == true ]]; then
    docker stop "$container_name" >/dev/null 2>&1 || true
  fi
}

trap cleanup EXIT
trap 'exit 130' INT TERM

docker run -d --rm --name "$container_name" \
  -e "POSTGRES_DB=$database_name" \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 127.0.0.1::5432 \
  "$image" >/dev/null
container_started=true

for attempt in $(seq 1 30); do
  if docker exec "$container_name" pg_isready -U postgres -d "$database_name" >/dev/null 2>&1; then
    break
  fi
  if [[ "$attempt" == "30" ]]; then
    echo "temporary Postgres did not become ready in 30 seconds" >&2
    exit 1
  fi
  sleep 1
done

port=$(docker inspect --format '{{(index (index .NetworkSettings.Ports "5432/tcp") 0).HostPort}}' "$container_name")
if [[ -z ${AGENT_SOURCE_REVISION:-} && -z ${GITHUB_SHA:-} ]]; then
  export AGENT_SOURCE_REVISION
  AGENT_SOURCE_REVISION=$(git -C "$repo_root" rev-parse HEAD)
fi
export SUPPERJUMPIN_TEST_DATABASE_URL="postgres://postgres:postgres@127.0.0.1:${port}/${database_name}?sslmode=disable"

"$repo_root/scripts/agent-verify.sh"
