# PROJECT KNOWLEDGE BASE

**Generated:** 2026-05-31 20:56:02 UTC
**Commit:** 87510f3
**Branch:** ben/agents-md-update

## OVERVIEW

Go backend (`apps/api`) + Expo React Native mobile app (`apps/mobile`) + generated TypeScript client (`packages/api-client`). Single product context monorepo — domain language lives in root `CONTEXT.md` and ADRs in `docs/adr/`.

## STRUCTURE

```
.
├── apps/
│   ├── api/              # Go backend (see apps/api/AGENTS.md)
│   └── mobile/           # Expo React Native (see apps/mobile/AGENTS.md)
├── packages/
│   └── api-client/       # Generated TS client (see packages/api-client/AGENTS.md)
├── docs/
│   ├── adr/              # Architecture Decision Records
│   ├── agents/           # Agent-specific docs (issue tracker, triage labels)
│   └── design/           # Product/UX/technical design docs
├── scripts/              # Local dev and code-generation helpers
├── worktrees/            # In-repo Git worktrees (e.g., issue-<number>)
├── .agents/ / .claude/   # Agent skills (mattpocock/skills, see skills-lock.json)
├── .work-issue/          # Work-issue skill operator config
├── CONTEXT.md            # Authoritative domain language dictionary
└── AGENTS.md             # This file
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Understand game rules | `apps/api/internal/game/` | Pure functions, no HTTP/DB imports |
| Add/modify API endpoint | `apps/api/internal/httpapi/server.go` | Routes + handler closures |
| Change API contract | `apps/api/openapi.yaml` → regenerate client | CI enforces sync |
| Modify DB schema | `apps/api/db/migrations/*.sql` | Pre-stable: fold into existing |
| Mobile UI changes | `apps/mobile/App.tsx` | Single-file prototype shell |
| Check CI pipeline | `.github/workflows/ci.yml` | Go 1.26.3, Node 24.16.0 |
| Domain terminology | `CONTEXT.md` | "Jump" not "Stunt", "Player" not "User" |

## CONVENTIONS

- **Hexagonal architecture**: Domain logic in `internal/game/` (pure functions, injected repo interfaces). Transport in `internal/httpapi/` (routing, JSON, DTO conversion). Domain must never import `net/http` or `database/sql`.
- **Repository-per-flow**: Each `game/*.go` defines its own small repository interface (e.g., `JudgmentRepository`). `PostgresStore` implements the composed `Persistence` interface; tests should use per-test fakes or Postgres-backed fixtures.
- **Result pattern**: Domain functions return `XxxResult{Allowed, Created, Err}` structs. `Allowed=false` maps to HTTP 403.
- **Snapshot pattern**: Read-only views use `XxxSnapshot` structs (e.g., `JumpSnapshot`, `SeasonSnapshot`).
- **Stable IDs**: `stableID(kind, value)` generates SHA256-based deterministic IDs, not UUIDs.
- **Clock injection**: `func() time.Time` is injected for testability (e.g., `PostgresStore` accepts `now`).
- **Pre-stable migrations**: Fold schema changes into existing migration files. Do not create standalone numbered migrations until DB is declared stable.
- **OpenAPI sync gate**: CI runs `generate:api-client` then `git diff --exit-code`. Any OpenAPI change without regeneration breaks CI.
- **Hand-rolled mocks**: Go tests use inline `mock*Repo` structs with function fields. No testify/mock or gomock.
- **Co-located tests**: `*_test.go` alongside source. Node tests use `node --test` with `*.test.mjs`.
- **Coverage commands**: `npm run test:coverage` and `npm run api:test:coverage` emit coverage summaries and write files under `coverage/`.

## ANTI-PATTERNS (THIS PROJECT)

- Putting business logic in HTTP handlers (`server.go`). All rules belong in `internal/game/`.
- Hand-writing API client types instead of generating from `openapi.yaml`.
- Using "Stunt", "User", or "Vote" instead of "Jump", "Player", "Judgment" (see CONTEXT.md avoid lists).
- Creating standalone DB migration files before DB stability.
- Adding ESLint/Prettier/golangci-lint configs without team discussion — none exist today; minimal tooling is the current stance.
- Using domain-forbidden synonyms in code, comments, or error messages.

## GIT WORKFLOW

- **Feature branches for all changes** — never push to main directly.
- **Before every push**: `git fetch origin && git rebase origin/main` so your branch is always on top of the latest shared state from other devs.
- **If rebase conflicts**: Stop and report them. Never force-push through conflicts.
- **Only push after**: rebase resolves cleanly + build/tests pass.

## UNIQUE STYLES

- **Agent-native repo scaffolding**: `.agents/`, `.claude/`, `.work-issue/`, `skills-lock.json`, `opencode.json` are first-class project infrastructure, not afterthoughts.

## TRANSITIONAL STATE

These are known compromises that exist for pre-MVP speed but should converge toward better patterns. When you resolve one in a PR, remove its row from this table.

| Current state | Why it exists | Converge when |
|---|---|---|
## MAINTENANCE CONTRACT

These files are maintained by convention, not automation. Follow these rules in every PR:

- **Update in the same PR**: If your PR changes something documented in any AGENTS.md — a convention, pattern, file location, or transitional state — update that AGENTS.md in the same PR. Do not leave stale docs.
- **Remove resolved transitional states**: If your PR resolves one of the transitional state items above, delete its row from the table.
- **Open a new AGENTS.md on new boundaries**: When you add a new package (`packages/foo`) or app (`apps/foo`), create a corresponding AGENTS.md. 30-80 lines, reference the parent AGENTS.md, never repeat parent content.
- **Minimal effort rule**: Updating AGENTS.md should add at most 2-3 minutes per PR. If it's taking longer, the file is too verbose — trim it.

## COMMANDS

```sh
# Development
npm run db:up             # Start Docker Compose Postgres service
npm run db:down           # Stop Docker Compose Postgres without deleting data
npm run db:reset          # Recreate local dev DB and reapply migrations
npm run db:migrate        # Apply migrations to local Docker Postgres by default
npm run db:migrate:staging # Apply migrations to SUPPERJUMPIN_DATABASE_URL
npm run api:dev           # Run API against existing DB

# Testing
npm run api:test          # Run Go API tests against Postgres (local Docker or SUPPERJUMPIN_TEST_DATABASE_URL); canonical test path
npm run api:test:coverage  # Run Go API tests with coverage output and summary
npm test                  # npm workspace tests (api-client + scripts)
npm run test:coverage     # npm workspace tests with coverage output and summary
npm --workspace @supperjumpin/mobile test  # Jest + React Native Testing Library mobile tests
npm --workspace @supperjumpin/mobile run typecheck  # tsc --noEmit

# API client regeneration
npm run generate:api-client  # openapi-typescript → packages/api-client/src/generated.d.ts

# sqlc query layer
npm run generate:sqlc        # sqlc generate → apps/api/internal/db/
```

## NOTES

- `DATABASE_URL` is mandatory for the Go binary, but npm scripts set it for local Docker Postgres by default. Use `SUPPERJUMPIN_DATABASE_URL` to intentionally target Supabase staging. See `docs/supabase.md`.
- Supabase Auth JWT verification is enabled by `SUPABASE_JWT_SECRET`; dev auth token defaults to `dev-token` via `SUPPERJUMPIN_DEV_AUTH_TOKEN`.
- Mobile needs `EXPO_PUBLIC_SUPABASE_URL`, `EXPO_PUBLIC_SUPABASE_ANON_KEY`, and `EXPO_PUBLIC_API_BASE_URL` in `.env`. Put the Supabase publishable key in the legacy `EXPO_PUBLIC_SUPABASE_ANON_KEY` variable.
- `npm run api:test` resets a `_test`-suffixed database and applies migrations before running Go tests. Uses local Docker Compose Postgres by default, or `SUPPERJUMPIN_TEST_DATABASE_URL` when set. Refuses destructive reset on non-test databases unless `SUPPERJUMPIN_TEST_ALLOW_UNSAFE_RESET=1` is set.
- `npm run api:test:coverage` writes `coverage/api.coverprofile`, prints `go tool cover -func` output, and appends a summary when `GITHUB_STEP_SUMMARY` is set.
- `npm run test:coverage` runs workspace test coverage and appends a summary when `GITHUB_STEP_SUMMARY` is set.
- Mobile now has a lightweight Jest + React Native Testing Library harness for screen-level tests in `apps/mobile/*.test.tsx`; `apps/mobile/test/mockApi.ts` is the default public-read API mocking seam.
- No production deployment configs (Dockerfile, K8s, Terraform) exist in this repo.
- Issues tracked in GitHub Issues (`supperjumpin/supperjumpin`). See `docs/agents/issue-tracker.md`.
- Triage labels: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`.
- Agent skills evaluated from `.agents/skills/` and `.claude/skills/` (lockfile: `skills-lock.json`).
