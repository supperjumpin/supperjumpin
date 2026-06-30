# Mobile app removed

The Expo React Native mobile app is deleted. The pivot to a Discord-only front-end (ADR-0032, ADR-0035) made the mobile tree dead weight: a full Expo/React Native codebase with Jest harness, Babel config, and 8+ screen components that no runner in the project touches. This ADR records the removal and supersedes ADR-0001 (the original "build it in Expo" decision).

The `apps/mobile/` directory and its workspace entry in the root `package.json` are removed. All references in CI, AGENTS.md, README, and the `@supperjumpin/api-client` consumer list are cleaned up. The mobile app is not parked in a `_deprecated/` subfolder — it is gone, on the grounds that half-deleting a pivot leaves landmines for the next engineer.

## What gets deleted

- `apps/mobile/` — entire tree: `App.tsx`, `AuthGate.tsx`, `CreateJumpScreen.tsx`, `DisplayNameSetupScreen.tsx`, `FeedScreen.tsx`, `JudgmentScreen.tsx`, `JumpDetailScreen.tsx`, `flow.ts`, `types.ts`, `package.json`, `jest.config.js`, `jest.setup.ts`, `babel.config.js`, `app.json`, `index.js`, `test/mockApi.ts`, `prototype-brand-185/`, and the `__snapshots__/` and `*.test.tsx` companions.
- Mobile entries in `package.json` workspaces (the `apps/mobile` workspace itself is the only thing it brought in, since `packages/api-client` was the other workspace and is removed separately under ADR-0049).
- The mobile job in `.github/workflows/ci.yml` (rewritten under ADR-0047 anyway).
- The "Mobile" row in the `coverage-diff.mjs` labels map (the script is deleted under ADR-0047).
- References to the mobile stack in `docs/research/testing-strategy-refresh.md`, `docs/agents/domain.md`, and `MEMORY.md`.

## Why "delete entirely" instead of "park in deprecated"

- The pivot already happened in the domain. ADR-0032 made the Round the central thing; ADR-0035 deleted the competition apparatus; the Discord adapter is the only front-end (ADR-0041). The AGENTS.md and CONTEXT.md only mention the API + Discord bot. A parked `apps/mobile/_deprecated/` would be the only reference to a product surface that doesn't exist anywhere else in the repo.
- The Expo/React Native toolchain (`expo`, `jest-expo`, `@testing-library/react-native`, the babel config, the `flow.ts` typecheck, the `prototype-brand-185` static prototype, the `__snapshots__` for screen-level tests) carries its own upgrade and compatibility burden even when no runner executes it.
- A future mobile surface — if the team ever needs one — will be a different stack (likely React Native or Expo SDK 50+ at that point) and a different shape. Reanimating this tree would be more work than starting from the current OpenAPI contract.

## Reintroduction

This ADR does not preclude a future mobile or web client. If the team ever builds one, the new client is its own ADR and its own repo subtree. The contract it talks to is `apps/api/openapi.yaml` (no longer auto-generated into a TS package, see ADR-0049).

## Status

accepted — supersedes ADR-0001.
