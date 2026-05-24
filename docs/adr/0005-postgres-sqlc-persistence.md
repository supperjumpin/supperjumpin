# Postgres and sqlc Persistence

The Go backend will use Postgres as the source of truth with hand-written SQL queries compiled through sqlc, plus ordinary schema migrations. Supperjumpin's core relationships among Players, Groups, Seasons, Stunts, Evidence, Judges, and scores fit a relational model, and sqlc keeps query behavior explicit without introducing an ORM abstraction over the game's lifecycle rules.
