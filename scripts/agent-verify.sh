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
export MAGEFILE_CACHE="$repo_root/artifacts/agents/$AGENT_TASK_ID/$AGENT_ATTEMPT_ID/mage-cache"

cd "$repo_root"
exec mage agent:verify
