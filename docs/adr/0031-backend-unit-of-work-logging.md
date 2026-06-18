# Backend Unit-of-Work Logging

The Go API uses standard-library `log/slog` structured logs for backend unit-of-work logging. A unit of work is an externally meaningful API/domain operation, such as creating a Performed Jump, submitting a Judgment, loading the Feed, or computing Open scores, rather than an individual SQL query.

Logging lives at the HTTP/domain-operation boundary: middleware records request metadata and completion status, handlers mark the operation and safe domain identifiers, and `internal/game` remains logger-free. `PostgresStore` should not emit routine repository logs; transaction or persistence diagnostics can be added later only when they explain failures without producing query-level noise.

Each request carries a small aggregated request log context. Handlers and lower-level backend functions that already receive `context.Context` may add approved metadata or raise severity, and middleware emits one final unit-of-work log using the accumulated fields and highest severity. The first implementation should keep this seam in `internal/httpapi` rather than introducing a general logging package; if other backend packages need it later, the small context-helper API can move mechanically.

Logs should be JSON by default, use snake_case field names.

**Automatically included in every log** (set by middleware): `request_id`, `method`, `path`, `status`, `response_bytes`, `duration_ms`, `user_agent`.

**Approved handler fields** (must be explicitly added via `AddRequestLogField` or `setRequestOperation`): `actor_type`, `player_id`, `route`, `operation`, `jump_id`, `judgment_id`, `open_year`, `open_month`, `outcome`, `error_code`, `internal_error`, `stack`.

The field `outcome` takes values `success`, `client_error`, `forbidden`, `not_found`, `conflict`, or `server_error`. The fields `internal_error` (redacted) and `stack` (full goroutine trace) are set only on 5xx errors or panics. The log accumulator silently drops any field name not in the approved list.

Logs must not include bearer tokens, email addresses, player display names, guest session IDs, media object keys, jump captions, Source/Destination/Food text, raw request bodies, response bodies, SQL query text, or other content fields.

Expected client and domain outcomes are `INFO`; server failures are `ERROR`. The API uses a canonical UUID request ID for every request: it accepts an incoming `X-Request-ID` only when it is a valid UUID, otherwise it generates a UUID v4 with `crypto/rand`. The canonical request ID is attached to request context, returned in the `X-Request-ID` response header, and included in every unit-of-work log.
