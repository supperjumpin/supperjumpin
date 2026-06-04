# Testing Strategy Refresh

This memo captures the repo-wide test strategy discussed in issue #210. The goal is confidence first: tests should make vibe-coded changes trustworthy, while keeping the system fast enough to work in day-to-day development.

## Shared baseline

- Prefer plain stdlib-style tests by default.
- Reach for framework-like DSLs only when they materially improve readability.
- Keep fixtures local and minimal.
- Prefer one behavior per test when practical.
- Use readable test names over clever abstraction.

## Go API

- Use narrow unit tests for domain logic.
- Use Postgres-backed tests for transport and lifecycle coverage.
- Keep shared in-memory test doubles out of the default path.
- Treat coverage as a quality signal for this stack.

## `api-client`

- Test the hand-written runtime wrapper.
- Do not test the generated type output itself.
- Focus on request shape, auth header injection, and response/error mapping.

## Mobile

- For now, document manual verification plus typecheck as the baseline.
- Revisit a deeper test harness only when the mobile codebase is large enough to justify it.

## Coverage visibility

- Surface coverage in GitHub Actions summaries.
- Post a PR comment when coverage drops from the previous baseline.
- Keep the signal actionable, but do not block merges yet.

## Durable decisions recorded as ADRs

- Confidence-first testing philosophy and per-stack defaults.
- Coverage as a visible signal, not a hard gate.
