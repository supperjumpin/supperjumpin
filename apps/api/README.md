# API

Go backend API for Supperjumpin.

The API owns game rules, durable domain state, and the REST/OpenAPI contract consumed by the mobile app. It stores core domain data in Postgres. Evidence uploads to object storage are specified but deferred until hosted infrastructure is introduced; the local MVP does not require image hosting.

## Runtime configuration

Set `SUPPERJUMPIN_DATABASE_URL` to a Postgres database with the migrations in `db/migrations` applied before starting the API. The API refuses to start without `SUPPERJUMPIN_DATABASE_URL` so all domain state is stored durably in Postgres, not in process memory.
