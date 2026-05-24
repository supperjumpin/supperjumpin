# supperjumpin

A social mobile game about planning, performing, documenting, and judging absurd food-location stunts.

## Runnable scaffold

Install JavaScript dependencies once from the repo root:

```sh
npm install
```

Run the backend behavior tests:

```sh
npm run api:test
```

Regenerate the TypeScript client types from the OpenAPI contract:

```sh
npm run generate:api-client
```

Run the API locally with a development bearer token:

```sh
npm run api:dev
```

Then call `GET http://localhost:8080/v1/me` with `Authorization: Bearer dev-token`.

Start the Expo app:

```sh
npm --workspace @supperjumpin/mobile run dev
```

Copy `apps/mobile/.env.example` to `apps/mobile/.env` and set the Supabase project URL, anon key, and API base URL before using mobile auth.

## First playable slice

The first playable slice proves the Group Stunt loop:

1. A Player signs in.
2. A Player creates or joins a Group.
3. A Group starts a Season.
4. A Player creates a Planned Stunt with a Source, Destination, and Food.
5. A Player submits photo and Caption Evidence.
6. Other Players Judge the Performed Stunt on Difficulty, Transgression, Creativity, and Documentation.
7. The Group sees Season Standings.
