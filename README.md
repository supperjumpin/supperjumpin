     1|# supperjumpin
     2|
     3|A social mobile game about planning, performing, documenting, and judging absurd food-location stunts.
     4|
     5|## Demo locally
     6|
     7|Prerequisites:
     8|
     9|- Node.js/npm
    10|- Go
    11|- Docker Desktop or Docker Engine with Compose
    12|
    13|Install JavaScript dependencies once from the repo root:
    14|
    15|```sh
    16|npm install
    17|```
    18|
    19|Start the local demo API:
    20|
    21|```sh
    22|npm run demo:api
    23|```
    24|
    25|That command starts Postgres with Docker Compose, applies API migrations, and runs the Go API with the development bearer token `dev-token`.
    26|
    27|Smoke-test it from another terminal:
    28|
    29|```sh
    30|curl -H "Authorization: Bearer *** http://localhost:8080/v1/me
    31|```
    32|
    33|Optional Make aliases are available on systems with `make`:
    34|
    35|```sh
    36|make demo-api
    37|```
    38|
    39|Start the Expo app separately if you want to inspect the current mobile shell:
    40|
    41|```sh
    42|npm run demo:mobile
    43|```
    44|
    45|Copy `apps/mobile/.env.example` to `apps/mobile/.env` and set the Supabase project URL, anon key, and API base URL before using mobile auth.
    46|
    47|## Runnable scaffold
    48|
    49|Run the backend behavior tests:
    50|
    51|```sh
    52|npm run api:test
    53|```
    54|
    55|Regenerate the TypeScript client types from the OpenAPI contract:
    56|
    57|```sh
    58|npm run generate:api-client
    59|```
    60|
    61|Run the API locally against an already configured database:
    62|
    63|```sh
    64|npm run api:dev
    65|```
    66|
    67|Set `DATABASE_URL` first when using `api:dev` directly. For most local demos, prefer `npm run demo:api`.
    68|
    69|Start the Expo app:
    70|
    71|```sh
    72|npm --workspace @supperjumpin/mobile run dev
    73|```
    74|
    75|## First playable slice
    76|
    77|The first playable slice proves the Group Stunt loop:
    78|
    79|1. A Player signs in.
    80|2. A Player creates or joins a Group.
    81|3. A Group starts a Season.
    82|4. A Player creates a Planned Stunt with a Source, Destination, and Food.
    83|5. A Player submits photo and Caption Evidence.
    84|6. Other Players Judge the Performed Stunt on Commitment, Transgression, Creativity, and Documentation.
    85|7. The Group sees Season Standings.
    86|