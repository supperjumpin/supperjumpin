# Mage as the build orchestrator

The repo's build, test, and dev commands are orchestrated by `mage` (magefile.org). Build targets are `.go` files in `magefile/` at the repo root, run as `mage <target>`. This replaces the prior `package.json` + `scripts/*.mjs` setup that wrapped `go`, `docker compose`, and `sqlc` calls in Node.

The trade-off: we commit to Go as the only language and add `mage` as the one extra binary on `PATH` (it's a single `go install` away, no other runtime dep). The win: build orchestration lives in testable Go, `mage -l` is the single discoverability surface, and "what does CI do?" is answerable by reading `magefile/ci.go` instead of a YAML file that can drift from local.

## Why Mage over the alternatives

- **Mage (chosen):** targets are `.go` files. Same language as the rest of the codebase. Type-safe shell calls via `os/exec`. Self-documenting via `mage -l`. New dep is one `go install` away.
- **Task (taskfile.dev):** YAML targets, popular in the Go community, has `task --list`. Rejected because it adds a binary dep we don't already have and the YAML can't be unit-tested.
- **Makefile:** zero new deps, but Make syntax is dated, discoverability requires a hand-written `help` target, and the shell logic can't be tested in isolation.
- **Just:** make-replacement with cleaner syntax, smaller community. Same downsides as Make for our purposes.

## Target surface

Targets are namespaced in Go, so `mage -l` shows them grouped:

- `mage init` — first-time setup hints
- `mage db:up` / `mage db:down` / `mage db:reset` / `mage db:migrate` — local Postgres lifecycle
- `mage generate:sqlc` — sqlc generate
- `mage test` — `go test` across both apps, accepts `-coverage` flag
- `mage dev:api` / `mage dev:bot` — run one service in the foreground
- `mage build:api` / `mage build:bot` — multi-stage Dockerfile builds
- `mage ci:all` / `mage ci:lint` / `mage ci:test` / `mage ci:build` / `mage ci:comment` — the CI pipeline and coverage-comment helpers, also runnable locally
- `mage docs` — serve Swagger UI for `apps/api/openapi.yaml`

`mage up` is composed of `mg.Deps(DbUp, StartAPI, StartBot)` — the targeted functions are the building blocks, the whole-stack ones are pure composition. Single source of truth, no copy-paste orchestration. The two layers ("I'm hacking, just go" vs "I'm doing one specific thing") are both first-class.

## What gets deleted in the same slice

- `package.json` at the repo root (workspaces, scripts, `engines.node`).
- `package-lock.json`.
- The entire `scripts/` directory (22 files: `api-dev.mjs`, `api-test.mjs`, `bot-dev.mjs`, `db-lifecycle.mjs`, `db-reset.mjs`, `db-migrate.mjs`, `db-helpers.mjs`, `tool-versions.mjs`, `run-sqlc-generate.mjs`, `test-coverage.mjs`, `coverage-diff.mjs`, `docs-server.mjs`, `setup.mjs`, `demo-full.sh`, and their `*.test.mjs` companions).
- The `engines.node` constraint and the `node: ">=24.16.0"` line in tool-versions.

After this lands, **Node is no longer required for the project at all** — the bot has no `package.json`, the api-client is gone (ADR-0049), the `scripts/` directory is gone. This is intentional: the pivot is to a Go-only project plus a Discord adapter.

## CI integration

`.github/workflows/ci.yml` becomes a thin wrapper that calls `mage ci:lint`, `mage ci:test`, and `mage ci:build`. The three sub-targets are public so the workflow can run them in parallel jobs; `mage ci:all` runs them sequentially via `mg.SerialDeps`. Local `mage ci:all` and GitHub CI run the same steps — no "works locally, fails in CI" drift.

## Status

accepted — supersedes the npm-based orchestration that grew up around the deleted mobile app; makes the build surface Go-only, self-documenting, and unit-testable in the same language as the rest of the codebase.
