#!/usr/bin/env bash
set -euo pipefail

usage() {
  printf 'Usage: %s <prd-issue>\n' "$0"
  printf '\n'
  printf 'Runs ONE Ralph HITL iteration. Watch what it does, then run again.\n'
  printf '\n'
  printf 'Arguments:\n'
  printf '  prd-issue       GitHub issue URL or number (e.g. 42 or https://github.com/org/repo/issues/42)\n'
  printf '\n'
  printf 'Environment overrides:\n'
  printf '  RALPH_REPO=supperjumpin/supperjumpin\n'
  printf '  OPENCODE_AGENT=<agent-name>\n'
  printf '  OPENCODE_MODEL=<provider/model>\n'
  printf '\n'
  printf 'Examples:\n'
  printf '  %s 42\n' "$0"
  printf '  %s https://github.com/org/repo/issues/42\n' "$0"
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -ne 1 ]]; then
  usage >&2
  exit 1
fi

prd_arg="$1"
repo="${RALPH_REPO:-supperjumpin/supperjumpin}"

# Normalize PRD: extract issue number from URL or bare number
if [[ "$prd_arg" =~ github\.com/.*/issues/([0-9]+) ]]; then
  prd_number="${BASH_REMATCH[1]}"
elif [[ "$prd_arg" =~ ^#?([0-9]+)$ ]]; then
  prd_number="${BASH_REMATCH[1]}"
else
  printf 'Invalid PRD argument: "%s". Pass a GitHub issue URL or number.\n' "$prd_arg" >&2
  exit 1
fi

if ! git rev-parse --show-toplevel >/dev/null 2>&1; then
  printf 'Ralph must be run from inside a git repository.\n' >&2
  exit 1
fi

branch="$(git branch --show-current)"
if [[ -z "$branch" ]]; then
  printf 'Refusing to run on a detached HEAD. Create/check out a feature branch first.\n' >&2
  exit 1
fi

if [[ "$branch" == "main" || "$branch" == "master" ]]; then
  printf 'Refusing to run on %s. This repo requires feature branches for all changes.\n' "$branch" >&2
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  printf 'gh is not installed or not on PATH.\n' >&2
  exit 1
fi

if ! gh auth status >/dev/null 2>&1; then
  printf 'gh is not authenticated. Run gh auth login before using Ralph.\n' >&2
  exit 1
fi

if ! command -v opencode >/dev/null 2>&1; then
  printf 'opencode is not installed or not on PATH.\n' >&2
  exit 1
fi

printf '== Ralph HITL iteration (PRD #%s) ==\n' "$prd_number"

prompt=$(cat <<PROMPT
GitHub PRD: $repo#$prd_number

1. Read AGENTS.md, CONTEXT.md, docs/agents/issue-tracker.md, docs/project-board.md, the PRD issue, and its comments.
2. Explicitly fetch the PRD's GitHub sub-issues/parent relationships; do not infer them from the issue body or comments alone.
3. If sub-issues already exist, choose the next unfinished ready sub-issue from that list. Do not decompose the PRD again.
4. If no sub-issues exist, decide the next unfinished piece of work; create sub-issues only if the PRD clearly needs decomposition before implementation.
5. Use PRD issue comments as the progress log. Read prior Ralph progress comments before deciding what remains.
6. If there is no remaining work, or the next work needs a human decision, add a short PRD comment and stop.
7. Otherwise, execute the chosen work by using /tdd. Pass the chosen issue/PRD context into the skill.
8. After the skill finishes, comment on the PRD or relevant sub-issue with concise progress, result, verification, and commit hash; then make one commit for the completed slice.

Nonnegotiables: follow repo AGENTS.md files, do not push, do not rebase, do not touch unrelated user changes, do not close issues directly unless the issue explicitly says to.
PROMPT
)

if [[ -n "${OPENCODE_AGENT:-}" && -n "${OPENCODE_MODEL:-}" ]]; then
  opencode run --agent "$OPENCODE_AGENT" --model "$OPENCODE_MODEL" "$prompt"
elif [[ -n "${OPENCODE_AGENT:-}" ]]; then
  opencode run --agent "$OPENCODE_AGENT" "$prompt"
elif [[ -n "${OPENCODE_MODEL:-}" ]]; then
  opencode run --model "$OPENCODE_MODEL" "$prompt"
else
  opencode run "$prompt"
fi

printf 'HITL iteration finished. Review the commit and PRD #%s comments.\n' "$prd_number"
