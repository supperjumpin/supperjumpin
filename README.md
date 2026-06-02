# supperjumpin

A social mobile game about planning, performing, documenting, and judging absurd food-location stunts.

## Local development

Prerequisites:

- Node.js/npm
- Go
- sqlc
- Docker Desktop or Docker Engine with Compose

Install JavaScript dependencies once from the repo root:

```sh
npm install
```

Start the local Postgres service:

```sh
npm run db:up
```

Stop the local Postgres service without deleting its data:

```sh
npm run db:down
```

Reset the local development database back to a clean migrated baseline:

```sh
npm run db:reset
```

Apply API migrations:

```sh
npm run db:migrate
```

Run the Go API with an explicit `DATABASE_URL`:

```sh
DATABASE_URL=postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable npm run api:dev
```

If you want a fresh local starting point, `npm run db:reset` already starts Postgres, recreates the `supperjumpin` database, reapplies migrations, and leaves Postgres running for `npm run api:dev`.

The development bearer token defaults to `dev-token`.

Smoke-test it from another terminal:

```sh
curl -H "Authorization: Bearer dev-token" http://localhost:8080/v1/me
```

Start the Expo app separately if you want to inspect the current mobile shell:

```sh
npm --workspace @supperjumpin/mobile run dev
```

Copy `apps/mobile/.env.example` to `apps/mobile/.env` and set the Supabase project URL, anon key, and API base URL before using mobile auth.

## Runnable scaffold

Run the backend behavior tests against Postgres:

```sh
npm run api:test
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

## First playable slice

The first playable slice proves the Group Stunt loop:

1. A Player signs in.
2. A Player creates or joins a Group.
3. A Group starts a Season.
4. A Player creates a Planned Stunt with a Source, Destination, and Food.
5. A Player submits photo and Caption Evidence.
6. Other Players Judge the Performed Stunt on Difficulty, Transgression, Creativity, and Documentation.
7. The Group sees Season Standings.
