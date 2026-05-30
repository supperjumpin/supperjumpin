# Hexagonal Architecture Formalization

## Context

The Go backend uses an `internal/game` module for domain logic (ADR-0018), with HTTP adapters in `internal/httpapi` and persistence adapters in Postgres stores. Technical Design §5.3 and Backend/Data Architecture §8 formalize this into explicit repository port interfaces.

## Decision

The backend adopts a hexagonal, ports-and-adapters, architecture where:

1. Domain logic in `internal/game` defines repository port interfaces, such as `JumpWriteRepo`, `JumpReadRepo`, `JudgmentWriteRepo`, and `GuestSessionRepo`.
2. HTTP handlers in `internal/httpapi` call domain services, not repositories directly.
3. Postgres implementations in `internal/httpapi/postgres_store.go` implement the port interfaces.
4. Domain services accept a `Now func() time.Time` for testable time.
5. Tests target domain services with in-memory or mock adapters.

## Rationale

Explicit port interfaces make the dependency direction unambiguous: domain logic never imports HTTP or Postgres packages. This enables fast unit tests without database infrastructure, allows swapping persistence implementations, and keeps game rules independent of delivery mechanism.

## Consequences

All new domain services must define their repository dependencies as interfaces, not concrete types. The existing `PostgresStore` must be decomposed to implement multiple small interfaces rather than one large store. HTTP handlers must not contain business logic.

## References

- Technical Design §5.3
- Backend/Data Architecture (#107) §8
- ADR-0018
