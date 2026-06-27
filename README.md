# supperjumpin

Platform-pure core API. Identity surface only (`GET /v1/me`, `PATCH /v1/me/display-name`).

## Local development

Prerequisites:

- Node.js/npm
- Go
- sqlc
- Docker Desktop or Docker Engine with Compose

1. Install JavaScript dependencies from the repo root:

```sh
npm install
```

2. Start Postgres and reset the local database to a clean migrated baseline:

```sh
npm run db:reset
```

3. Start the API in one terminal:

```sh
npm run api:dev
```

If you already have Postgres running and only want to reapply migrations, use `npm run db:migrate` instead of `npm run db:reset`.

`npm run db:migrate` always targets the local Docker Postgres database and ignores both ambient `DATABASE_URL` and `SUPPERJUMPIN_DATABASE_URL`.
Other local scripts may still use `SUPPERJUMPIN_DATABASE_URL` where documented.

If you want a fresh local starting point, `npm run db:reset` already starts Postgres, recreates the `supperjumpin` database, reapplies migrations, and leaves Postgres running for `npm run api:dev`.

The development bearer token defaults to `dev-token`.

Smoke-test the API:

```sh
curl -H "Authorization: Bearer dev-token" http://localhost:8080/v1/me
```

## Runnable scaffold

Run the backend behavior tests against Postgres:

```sh
npm run api:test
npm run api:test:coverage
npm run test:coverage
```

Regenerate the TypeScript client types from the OpenAPI contract:

```sh
npm run generate:api-client
```

Regenerate the API's Go query layer from the SQL query files:

```sh
npm run generate:sqlc
```

Run the API locally against an already configured database:

```sh
npm run api:dev
```

## Current state

Platform-pure core API with identity surface only (`GET /v1/me`, `PATCH /v1/me/display-name`). Downstream slices build on this skeleton.

See `docs/adr/` for architecture decisions.
