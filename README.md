# supperjumpin

Platform-pure core API. Identity surface only (`GET /v1/me`, `PATCH /v1/me/display-name`).

## Local development

Prerequisites:

- Go
- Docker Desktop or Docker Engine with Compose

0. Install Mage and make sure `~/go/bin` is on your `PATH`:

```sh
go install github.com/magefile/mage@v1.17.2
export PATH="$HOME/go/bin:$PATH"
```

1. Install local helper binaries from the repo root:

```sh
mage init:tools
```

2. Start Postgres and reset the local database to a clean migrated baseline:

```sh
mage db:reset
```

3. Start the API in one terminal:

```sh
mage dev:api
```

If you already have Postgres running and only want to reapply migrations, use `mage db:migrate` instead of `mage db:reset`.

`mage db:migrate` always targets the local Docker Postgres database and ignores both ambient `DATABASE_URL` and `SUPPERJUMPIN_DATABASE_URL`.

If you want a fresh local starting point, `mage db:reset` already starts Postgres, recreates the `supperjumpin` database, reapplies migrations, and leaves Postgres running for `mage dev:api`.

The development bearer token defaults to `dev-token`.

Smoke-test the API:

```sh
curl -H "Authorization: Bearer dev-token" http://localhost:8080/v1/me
```

## Runnable scaffold

Run the backend behavior tests against Postgres:

```sh
mage test
mage test -coverage
```

Regenerate the API's Go query layer from the SQL query files:

```sh
mage generate:sqlc
```

Run the API locally against an already configured database:

```sh
mage dev:api
```

Build the local Docker images:

```sh
mage build:api
mage build:bot
```

## Home-server deployment

Local development keeps `docker-compose.yml` focused on Postgres only. The home-server stack uses `docker-compose.prod.yml` to run Postgres, one-shot migrations, the API, and the Discord bot:

```sh
cp .env.example .env
docker compose -f docker-compose.prod.yml --env-file .env up -d --build
```

The detailed initial setup and deploy runbooks live in `~/src/home-server-setup/initial-setup.md` and `~/src/home-server-setup/deploy-supperjumpin.md`.

## Current state

Platform-pure core API with identity surface only (`GET /v1/me`, `PATCH /v1/me/display-name`). Downstream slices build on this skeleton.

See `docs/adr/` for architecture decisions.
