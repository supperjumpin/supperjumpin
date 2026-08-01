# Project Guide

Go backend in `apps/api`, Discord adapter in `apps/bot-discord`, and Mage build orchestration in `magefile/`. `CONTEXT.md` owns domain language; ADRs in `docs/adr/` own architectural decisions.

## Where To Look

| Task | Location | Notes |
|------|----------|-------|
| Game rules | `apps/api/internal/game/` | Pure functions, no HTTP/DB imports |
| Add/modify API endpoint | `apps/api/internal/httpapi/server.go` | Routes + handler closures |
| Change API contract | `apps/api/openapi.yaml` | No generated client remains in-repo; keep the spec aligned with handlers and DTOs |
| Modify DB schema | `apps/api/db/migrations/*.sql` | Fold into existing until home-server deployment holds real group data; after that, append numbered migrations |
| Prompt/Pack catalog | `apps/api/internal/game/prompts.go` + `apps/api/db/queries/prompts.sql` + `GET /v1/prompt-catalog` | Copy is data, not contract shape (ADR-0039) |
| Discord bot | `apps/bot-discord/cmd/bot/main.go` + `apps/bot-discord/internal/` | Thin HTTP client of the API. Owns `apps/bot-discord/.bot-data/` (evidence, scheduler state). `mage dev:bot` / `mage test`. |
| Build/test orchestration | `magefile/` | Pure Go command plans + Mage targets. `mage -l` is the discoverability surface |
| Check CI pipeline | `.github/workflows/ci.yml` | Go + Mage |
| Domain terminology | `CONTEXT.md` | Authoritative vocabulary |

## Core Rules

- **Hexagonal architecture**: Domain logic in `internal/game/` (pure functions, injected repo interfaces). Transport in `internal/httpapi/` (routing, JSON, DTO conversion). Domain must never import `net/http` or `database/sql`.
- **Repository-per-flow**: Each `game/*.go` defines its own small repository interface. `PostgresStore` implements the composed interface; tests use per-test fakes or Postgres-backed fixtures.
- **Result pattern**: Domain functions return `XxxResult{Allowed, Created, Err}` structs. `Allowed=false` maps to HTTP 403.
- **Snapshot pattern**: Read-only views use `XxxSnapshot` structs.
- **Stable IDs**: `stableID(kind, value)` generates SHA256-based deterministic IDs, not UUIDs.
- **Clock injection**: `func() time.Time` is injected for testability.
- **Pre-stable migrations**: Fold schema changes into existing migration files until the home-server deployment holds real group data. After real home-server data exists, never rewrite an applied migration; add a new numbered migration for every schema change.
- **Hand-rolled mocks**: Go tests use inline `mock*Repo` structs with function fields. No testify/mock or gomock.
- **Co-located tests**: `*_test.go` alongside source. The Mage module follows the same pattern in `magefile/*_test.go`.
- **Coverage commands**: `mage test -coverage` emits per-service coverage files under `coverage/`, appends to `GITHUB_STEP_SUMMARY`, and feeds the non-blocking PR coverage comment.

## Avoid

- Putting business logic in HTTP handlers (`server.go`). All rules belong in `internal/game/`.
- Creating standalone DB migration files before DB stability, unless the home-server deployment already holds real group data.
- Adding ESLint/Prettier/golangci-lint configs without team discussion.

## Git Workflow

- **Feature branches for all changes** — never push to main directly.
- **Main history should stay clean**: prefer squash merge so each commit on `main` represents one coherent body of work.
- **PR branches are disposable**: commit freely while iterating; typo fixes and review follow-ups do not need cleanup on the branch.
- **Squash titles should be outcome-first**: prefer `feat(area): outcome (#issue)`, `fix(area): outcome (#issue)`, `docs(area): outcome (#issue)`, `test(area): outcome (#issue)`, or `chore(area): outcome (#issue)` over slice/process phrasing.
- **Update from main only when needed**: do it for mergeability, CI, or conflict resolution, not as a routine step before every push.
- **Use the least painful branch update strategy**: merge, rebase, or recreate from fresh `main` depending on the branch state.
- **Only push after**: required build/tests pass.

## Maintenance Contract

These files are maintained by convention, not automation. Follow these rules in every PR:

- **Update in the same PR**: If your PR changes something documented in any AGENTS.md — a convention, pattern, file location, or transitional state — update that AGENTS.md in the same PR.
- **Open a new AGENTS.md on new boundaries**: When you add a new package or app, create a corresponding AGENTS.md. 30-80 lines, reference the parent AGENTS.md, never repeat parent content.

## Commands

```sh
# Development
mage db:up                # Start Docker Compose Postgres service
mage db:down              # Stop Docker Compose Postgres without deleting data
mage db:reset             # Recreate local dev DB and reapply migrations
mage db:migrate           # Apply migrations to the local Docker Postgres only
mage dev:api              # Run API against existing DB
mage dev:bot              # Run Discord bot (needs SUPPERJUMPIN_BOT_TOKEN in env)
mage docs                 # Serve Swagger UI for apps/api/openapi.yaml
mage init:all             # Install local tool binaries and print the common next steps
mage agent:verify         # Run canonical gates against a runner-isolated test database

# Testing
mage test                 # Run Go API + bot tests (API tests prep Postgres-backed _test DB)
mage test -coverage       # Same, plus coverage files + summaries

# sqlc query layer
mage generate:sqlc        # sqlc generate → apps/api/internal/db/

# CI and images
mage ci:lint              # Go vet across the app modules
mage ci:test              # Test path CI uses, with coverage
mage ci:build             # Docker image builds for API + bot
mage ci:all               # Sequential lint + test + build
mage build:api            # Build Dockerfile.api as supperjumpin-api:dev
mage build:bot            # Build Dockerfile.bot as supperjumpin-bot:dev
```

## Notes

- `SUPPERJUMPIN_DATABASE_URL` is mandatory for the Go binary.
- Install Mage with `go install github.com/magefile/mage@v1.17.2` and ensure `~/go/bin` is on `PATH` before using the command table below.
- Infrastructure is local-first: Docker Postgres, local dev bearer auth. Hosted infrastructure will be additive when introduced.
- Auth is local-first: `SUPPERJUMPIN_ADAPTER_TOKEN` defaults to `dev-token` in `mage dev:api` and `mage dev:bot`.
- `mage test` resets a `_test`-suffixed database and applies migrations before running API tests, then runs the bot tests in-process.
- `mage test -coverage` writes `coverage/api.coverprofile`, `coverage/bot.coverprofile`, `coverage/api-report.json`, and `coverage/bot-report.json`, and appends summaries when `GITHUB_STEP_SUMMARY` is set.
- `mage agent:verify` requires `AGENT_TASK_ID`, `AGENT_ATTEMPT_ID`, a source revision, and a runner-provided `SUPPERJUMPIN_TEST_DATABASE_URL` ending in `_test`; it never starts or stops a shared stack.
- Node is no longer required for this repo. `sqlc`, `migrate`, and `mage` are the only local helper binaries beyond Go and Docker.
- No production deployment configs exist in this repo.
- Issues tracked in GitHub Issues (`supperjumpin/supperjumpin`).
- Triage labels: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`.
