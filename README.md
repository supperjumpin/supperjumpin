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

- `EXPO_PUBLIC_SUPABASE_URL`
- `EXPO_PUBLIC_SUPABASE_ANON_KEY` (use the Supabase publishable key)
- `EXPO_PUBLIC_API_BASE_URL` (usually `http://localhost:8080`)

4. Start the API in one terminal:

```sh
npm run api:dev
```

If you already have Postgres running and only want to reapply migrations, use `npm run db:migrate` instead of `npm run db:reset`.

For Supabase staging setup, IPv4 pooler connection strings, and auth env vars, see `docs/supabase.md`.

Local scripts intentionally ignore an ambient shell `DATABASE_URL`. To target a non-local database, set `SUPPERJUMPIN_DATABASE_URL` for that command.

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

## First playable slice

The first playable slice proves the Group Stunt loop:

1. A Player signs in.
2. A Player creates or joins a Group.
3. A Group starts a Season.
4. A Player creates a Planned Stunt with a Source, Destination, and Food.
5. A Player submits photo and Caption Evidence.
6. Other Players Judge the Performed Stunt on Difficulty, Transgression, Creativity, and Documentation.
7. The Group sees Season Standings.
