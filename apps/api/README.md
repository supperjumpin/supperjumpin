# API

Go backend API for Supperjumpin.

The API owns domain logic, durable state, and the REST/OpenAPI contract. It stores core domain data in Postgres.

## Runtime configuration

Set `SUPPERJUMPIN_DATABASE_URL` to a Postgres database with the migrations in `db/migrations` applied before starting the API. The API refuses to start without `SUPPERJUMPIN_DATABASE_URL` so all domain state is stored durably in Postgres, not in process memory.
