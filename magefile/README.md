# Mage Build Orchestration

`magefile/` is a small Go module that defines this repo's local development and CI commands. The public entry point is `mage -l` from the repository root.

## Structure

- `*_targets.go`: Mage targets and namespaces exposed to users.
- `*_plan.go`: pure helpers that return command plans.
- `command.go`: the execution adapter for command plans.
- `*_test.go`: tests for command planning and orchestration helpers.

## Common Commands

```sh
mage -l
mage init:all
mage db:reset
mage dev:api
mage test -coverage
mage ci:all
```

## Agent Verification

An unattended runner must allocate an isolated Postgres database whose name ends
in `_test`, inject its URL, and identify each attempt. The runner owns database
provisioning and cleanup, so agent verification cannot stop a developer or
home-server stack.

For a Docker-capable local runner, use the complete disposable lifecycle:

```sh
AGENT_TASK_ID=issue-123 \
AGENT_ATTEMPT_ID=1 \
./scripts/agent-run.sh
```

It allocates a random loopback port, runs verification, and removes its
Postgres container on success, failure, or interruption.

```sh
AGENT_TASK_ID=issue-123 \
AGENT_ATTEMPT_ID=1 \
AGENT_SOURCE_REVISION=$(git rev-parse HEAD) \
SUPPERJUMPIN_TEST_DATABASE_URL=postgres://.../supperjumpin_issue_123_test?sslmode=disable \
./scripts/agent-verify.sh
```

The launcher assigns a Mage compilation cache unique to that task and attempt,
then delegates to `mage agent:verify`. The target runs the canonical lint and
coverage test gates and writes `artifacts/agents/<task>/<attempt>/result.json`.
Task IDs may contain only letters, digits, `.`, `_`, and `-`. CI runners may
supply `GITHUB_SHA` instead of `AGENT_SOURCE_REVISION`; external runners can
inject `SUPPERJUMPIN_TEST_DATABASE_URL` and call `agent-verify.sh` directly.

Keep target comments current so `mage -l` stays self-documenting.
