# API

Go backend API for Supperjumpin.

The API owns game rules, durable domain state, and the REST/OpenAPI contract consumed by the mobile app. It stores core domain data in Postgres and authorizes direct Evidence uploads to object storage.

## Runtime configuration

Set `SUPPERJUMPIN_DATABASE_URL` to a Postgres database with the migrations in `db/migrations` applied before starting the API. The API refuses to start without `SUPPERJUMPIN_DATABASE_URL` so Group and Group Membership state is not accidentally stored in process memory.
