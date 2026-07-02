# No TypeScript API client, no OpenAPI sync gate

`packages/api-client` is deleted. With the mobile app gone (ADR-0048), the only consumer of the TypeScript client was the mobile tree; the Discord bot has its own Go API client (`apps/bot-discord/internal/bot/`) that talks to the API over HTTP and never imported the npm package. There is no remaining TypeScript consumer in the repo, and no planned one.

The OpenAPI sync gate — the CI step that regenerated the TS types from `apps/api/openapi.yaml` and asserted `git diff --exit-code` — is also removed. The gate was the only thing keeping `openapi.yaml` honest in CI. With no consumer of the generated types, the gate has nothing to fail on.

The trade-off: contract drift between the OpenAPI spec and the Go server routes is no longer caught by automation. The team accepts this. If drift becomes a problem, the right replacement is a Go-side route-inventory test (load `openapi.yaml`, walk the routes in `internal/httpapi/server.go`, assert every path/method is documented) — but only when there is a concrete second consumer that needs the spec to be accurate. Speculative drift detection is process cost without payoff.

## What gets deleted

- `packages/api-client/` — entire tree: `package.json`, `README.md`, `src/index.js`, `src/index.d.ts`, `src/index.test.mjs`, `src/generated.d.ts`.
- The `packages/api-client` workspace entry in the root `package.json`.
- The CI step `npm run generate:api-client` and the `git diff --exit-code packages/api-client/src` gate in `.github/workflows/ci.yml`.
- The `npm run generate:api-client` script in the root `package.json`.
- References to the api-client in AGENTS.md, README.md, the `test-coverage.mjs` script (deleted under ADR-0047 anyway), `coverage-diff.mjs` (the `node: "api-client"` label is dropped), `docs/adr/0029-confidence-first-testing-strategy.md`, `docs/adr/0030-coverage-as-visible-signal.md`, and `MEMORY.md`.

## What's left of "the api-client"

The handwritten `index.js` had both live and dead methods:
- **Live for a future TS consumer:** `getMe`, `updateDisplayName`, `createJump`, `createGuestSession`, `getPublicFeed`, `getJumpDetail`.
- **Dead code from the competition era:** `submitJudgment` (superseded by Stamps, ADR-0034) and the `Judgment` type re-export.

The right call was to delete the package rather than curate it: the live methods will be regenerated from `openapi.yaml` by any future TS consumer, with the same hand-written runtime pattern. Carrying the package "just in case" preserves the dead `submitJudgment`/`Judgment` code indefinitely.

## Why not a Go-side contract test in this slice

The case for a Go-side route-inventory test is real: it would catch drift between `server.go` and `openapi.yaml` with zero TS footprint. It's also a non-trivial test (it has to parse YAML, walk the route table, and produce a useful diff on failure). Doing it in *this* slice — when there is no second consumer and no concrete drift incident — is exactly the speculative infra the project has been deliberately avoiding. The ADR is on the record: when the second consumer lands, this is the shape of the replacement.

## Status

accepted — supersedes the api-client + sync gate arrangement that was load-bearing only for the deleted mobile app.
