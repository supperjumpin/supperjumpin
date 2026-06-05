# Confidence-First Testing Strategy

Supperjumpin treats tests as a confidence tool first: the repo optimizes for catching bad behavior in vibe-coded changes, then for speed, and only later for maintainability once launch pressure eases. The default is plain stdlib-style tests with small local fixtures and one behavior per test; framework-like DSLs are only worth it when they clearly improve readability.

The backend follows a split strategy: narrow unit tests for domain logic, Postgres-backed tests for transport and lifecycle coverage, and no shared in-memory test double as the default path. Mobile now uses a lightweight Jest plus React Native Testing Library harness for screen-level behavior, with `global.fetch` interception as the default public-read mocking seam, while `api-client` only needs tests around the hand-written runtime wrapper, not the generated types.
