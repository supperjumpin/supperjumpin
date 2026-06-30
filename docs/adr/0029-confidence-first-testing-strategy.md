# Confidence-First Testing Strategy

Supperjumpin treats tests as a confidence tool first: the repo optimizes for catching bad behavior in vibe-coded changes, then for speed, and only later for maintainability once launch pressure eases. The default is plain stdlib-style tests with small local fixtures and one behavior per test; framework-like DSLs are only worth it when they clearly improve readability.

The backend follows a split strategy: narrow unit tests for domain logic, Postgres-backed tests for transport and lifecycle coverage, and no shared in-memory test double as the default path. The Discord bot follows the same split with `node --test` for in-process tests that don't need Postgres. After the build-tooling pivot (ADR-0048, ADR-0049), there is no mobile harness and no api-client package — the only test consumers are the Go API and the Go Discord bot, both invoked through `mage test` with the `-coverage` flag folded in.
