# PROJECT KNOWLEDGE BASE

## OVERVIEW

Go backend (`apps/api`) + Discord bot adapter (`apps/bot-discord`) + generated TypeScript client (`packages/api-client`). Domain language lives in root `CONTEXT.md` and ADRs in `docs/adr/`.

## STRUCTURE

```
.
├── apps/
│   ├── api/              # Go backend (see apps/api/AGENTS.md)
│   └── bot-discord/      # Discord bot adapter (see apps/bot-discord/AGENTS.md)
├── packages/
│   └── api-client/       # Generated TS client
├── docs/
│   ├── adr/              # Architecture Decision Records
│   ├── agents/           # Agent-specific docs (issue tracker, triage labels)
│   └── design/           # Product/UX/technical design docs
├── scripts/              # Local dev and code-generation helpers
├── worktrees/            # In-repo Git worktrees (e.g., issue-<number>)
├── .agents/ / .claude/   # Agent skills (see skills-lock.json)
├── .work-issue/          # Work-issue skill operator config
├── CONTEXT.md            # Authoritative domain language dictionary
└── AGENTS.md             # This file
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Game rules | `apps/api/internal/game/` | Pure functions, no HTTP/DB imports |
| Add/modify API endpoint | `apps/api/internal/httpapi/server.go` | Routes + handler closures |
| Change API contract | `apps/api/openapi.yaml` → regenerate client | CI enforces sync |
| Modify DB schema | `apps/api/db/migrations/*.sql` | Pre-stable: fold into existing |
| Prompt/Pack catalog | `apps/api/internal/game/prompts.go` + `apps/api/db/queries/prompts.sql` + `GET /v1/prompt-catalog` | `ListCatalog` / `SelectPrompt` / `SelectRandomPrompt`. Copy is data, not contract shape (ADR-0039). |
| Discord bot | `apps/bot-discord/cmd/bot/main.go` + `apps/bot-discord/internal/` | Thin HTTP client of the API. Owns `apps/bot-discord/.bot-data/` (evidence, scheduler state). `npm run bot:dev` / `npm run bot:test`. |
| Check CI pipeline | `.github/workflows/ci.yml` | Go + Node |
| Domain terminology | `CONTEXT.md` | Authoritative vocabulary |

## CONVENTIONS

- **Hexagonal architecture**: Domain logic in `internal/game/` (pure functions, injected repo interfaces). Transport in `internal/httpapi/` (routing, JSON, DTO conversion). Domain must never import `net/http` or `database/sql`.
- **Repository-per-flow**: Each `game/*.go` defines its own small repository interface. `PostgresStore` implements the composed interface; tests use per-test fakes or Postgres-backed fixtures.
- **Result pattern**: Domain functions return `XxxResult{Allowed, Created, Err}` structs. `Allowed=false` maps to HTTP 403.
- **Snapshot pattern**: Read-only views use `XxxSnapshot` structs.
- **Stable IDs**: `stableID(kind, value)` generates SHA256-based deterministic IDs, not UUIDs.
- **Clock injection**: `func() time.Time` is injected for testability.
- **Pre-stable migrations**: Fold schema changes into existing migration files. Do not create standalone numbered migrations until DB is declared stable.
- **OpenAPI sync gate**: CI runs `generate:api-client` then `git diff --exit-code`.
- **Hand-rolled mocks**: Go tests use inline `mock*Repo` structs with function fields. No testify/mock or gomock.
- **Co-located tests**: `*_test.go` alongside source. Node tests use `node --test` with `*.test.mjs`.
- **Coverage commands**: `npm run test:coverage` and `npm run api:test:coverage` emit coverage summaries and write files under `coverage/`.

## ANTI-PATTERNS (THIS PROJECT)

- Putting business logic in HTTP handlers (`server.go`). All rules belong in `internal/game/`.
- Hand-writing API client types instead of generating from `openapi.yaml`.
- Creating standalone DB migration files before DB stability.
- Adding ESLint/Prettier/golangci-lint configs without team discussion.

## GIT WORKFLOW

- **Feature branches for all changes** — never push to main directly.
- **Before every push**: `git fetch origin && git rebase origin/main`.
- **If rebase conflicts**: Stop and report them. Never force-push through conflicts.
- **Only push after**: rebase resolves cleanly + build/tests pass.

## MAINTENANCE CONTRACT

These files are maintained by convention, not automation. Follow these rules in every PR:

- **Update in the same PR**: If your PR changes something documented in any AGENTS.md — a convention, pattern, file location, or transitional state — update that AGENTS.md in the same PR.
- **Open a new AGENTS.md on new boundaries**: When you add a new package or app, create a corresponding AGENTS.md. 30-80 lines, reference the parent AGENTS.md, never repeat parent content.

## COMMANDS

```sh
# Development
npm run db:up             # Start Docker Compose Postgres service
npm run db:down           # Stop Docker Compose Postgres without deleting data
npm run db:reset          # Recreate local dev DB and reapply migrations
npm run db:migrate        # Apply migrations to the local Docker Postgres only
npm run api:dev           # Run API against existing DB
npm run bot:dev           # Run Discord bot (needs SUPPERJUMPIN_BOT_TOKEN in env)

# Testing
npm run api:test          # Run Go API tests against Postgres
npm run api:test:coverage  # Run Go API tests with coverage output
npm run bot:test          # Run Go bot tests (in-process, no Postgres)
npm test                  # npm workspace tests (api-client + scripts)
npm run test:coverage     # npm workspace tests with coverage output

# API client regeneration
npm run generate:api-client  # openapi-typescript → packages/api-client/src/generated.d.ts

# sqlc query layer
npm run generate:sqlc        # sqlc generate → apps/api/internal/db/
```

## NOTES

- `SUPPERJUMPIN_DATABASE_URL` is mandatory for the Go binary.
- Infrastructure is local-first: Docker Postgres, local dev bearer auth. Hosted infrastructure will be additive when introduced.
- Auth is local-first: `SUPPERJUMPIN_ADAPTER_TOKEN` defaults to `dev-token` in `npm run api:dev`.
- `npm run api:test` resets a `_test`-suffixed database and applies migrations before running Go tests.
- `npm run api:test:coverage` writes `coverage/api.coverprofile` and appends a summary when `GITHUB_STEP_SUMMARY` is set.
- `npm run test:coverage` runs workspace test coverage and appends a summary when `GITHUB_STEP_SUMMARY` is set.
- No production deployment configs exist in this repo.
- Issues tracked in GitHub Issues (`supperjumpin/supperjumpin`).
- Triage labels: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`.
