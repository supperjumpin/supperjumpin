# supperjumpin

A social mobile game about planning, performing, documenting, and judging absurd food-location stunts.

## Demo locally

Prerequisites:

- Node.js/npm
- Go
- Docker Desktop or Docker Engine with Compose

Install JavaScript dependencies once from the repo root:

```sh
npm install
```

Start the local demo API:

```sh
npm run demo:api
```

That command starts Postgres with Docker Compose, applies API migrations, and runs the Go API with the development bearer token `dev-token`.

Smoke-test it from another terminal:

```sh
curl -H "Authorization: Bearer dev-token" http://localhost:8080/v1/me
```

Optional Make aliases are available on systems with `make`:

```sh
make demo-api
```

Start the Expo app separately if you want to inspect the current mobile shell:

```sh
npm run demo:mobile
```

Copy `apps/mobile/.env.example` to `apps/mobile/.env` and set the Supabase project URL, anon key, and API base URL before using mobile auth.

## Runnable scaffold

Run the backend behavior tests:

```sh
npm run api:test
```

Regenerate the TypeScript client types from the OpenAPI contract:

```sh
npm run generate:api-client
```

Run the API locally against an already configured database:

```sh
npm run api:dev
```

Set `DATABASE_URL` first when using `api:dev` directly. For most local demos, prefer `npm run demo:api`.

Start the Expo app:

```sh
npm --workspace @supperjumpin/mobile run dev
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
