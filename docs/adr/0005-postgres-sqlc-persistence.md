# Postgres and sqlc Persistence

The Go backend will use Postgres as the source of truth with hand-written SQL queries compiled through sqlc, plus ordinary schema migrations. Supperjumpin's core relationships among Players, Jumps, Evidence, Judges, and scores fit a relational model, and sqlc keeps query behavior explicit without introducing an ORM abstraction over the game's lifecycle rules.

Addendum, post-MVP evolution: The `stunts` table was renamed to `jumps` (ADR-0020). New tables `guest_sessions`, `opens`, and `open_standings` were added for the Guest Judge model (ADR-0025) and the Open competition (ADR-0026). The original entities listed in this ADR (Players, Groups, Seasons, Stunts, Evidence, Judges, scores) are partially renamed and extended; see ADR-0020 and the Backend/Data Architecture doc (#107) for the current schema.
