# Remove `MemoryStore` as the Shared Test Double

Supperjumpin will stop using `MemoryStore` as the common backend for HTTP/integration tests and will rely on Postgres-backed tests for transport and lifecycle coverage. Pure unit tests should use small per-test mocks or fakes instead of a shared in-memory store, because that keeps the unit-test seam narrow and prevents coverage drift between the in-memory and Postgres paths.
