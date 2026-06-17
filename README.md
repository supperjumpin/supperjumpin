# supperjumpin

A social mobile game about planning, performing, documenting, and judging absurd food-location stunts.

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

3. Create `apps/mobile/.env` from the example file and set your values:

```sh
cp apps/mobile/.env.example apps/mobile/.env
```

You need:

- `EXPO_PUBLIC_API_BASE_URL` (usually `http://localhost:8080`)
- `EXPO_PUBLIC_DEV_AUTH_TOKEN` (defaults to `dev-token`)
- `EXPO_PUBLIC_MEDIA_BASE_URL` (optional; leave blank for local MVP flows without image hosting)

4. Start the API in one terminal:

```sh
npm run api:dev
```

If you already have Postgres running and only want to reapply migrations, use `npm run db:migrate` instead of `npm run db:reset`.

`npm run db:migrate` always targets the local Docker Postgres database and ignores both ambient `DATABASE_URL` and `SUPPERJUMPIN_DATABASE_URL`.
Other local scripts may still use `SUPPERJUMPIN_DATABASE_URL` where documented.

5. In a second terminal, start the Expo app:

```sh
npm --workspace @supperjumpin/mobile run dev
```

6. Use the Expo prompt to open the app on a simulator, emulator, device, or web.

If you want a fresh local starting point, `npm run db:reset` already starts Postgres, recreates the `supperjumpin` database, reapplies migrations, and leaves Postgres running for `npm run api:dev`.

The development bearer token defaults to `dev-token`.

Smoke-test it from another terminal:

```sh
curl -H "Authorization: Bearer dev-token" http://localhost:8080/v1/me
```

That gets you the backend plus the current mobile shell for local experimentation.

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

The MVP delivers the public Jump loop:

1. A Player signs in (local dev bearer token for development; see `.env.example`).
2. A Player posts a Performed Jump with a Source, Destination, and Food.
3. A Player submits photo and Caption Evidence.
4. Other Players Judge the Jump on Commitment, Transgression, Creativity, and Presentation.
5. Players compete in The Open (monthly competition with Standings).

Groups, Seasons, and Invites are v2 concepts — removed per ADR-0019.

See `docs/design/` for the full design package and `docs/adr/` for architecture decisions.
