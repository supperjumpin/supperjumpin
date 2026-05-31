# packages/api-client KNOWLEDGE BASE

## OVERVIEW

Generated TypeScript client and types for the Supperjumpin REST/OpenAPI contract. The mobile app consumes backend contracts through this package instead of hand-writing duplicated API payload types.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Runtime API calls | `src/index.js` | Hand-written fetch wrappers. Injectable `fetchImpl`. |
| Type signatures | `src/index.d.ts` | Hand-written function signatures re-exporting from `generated.d.ts`. |
| Generated schema types | `src/generated.d.ts` | Auto-generated from `apps/api/openapi.yaml` via `openapi-typescript`. |
| Tests | `src/index.test.mjs` | Node built-in test runner (`node --test`). Mock `fetchImpl` injection. |

## CONVENTIONS

- **ESM only**: `"type": "module"`. No CommonJS output.
- **Zero runtime dependencies**: Pure JS + TypeScript types. Only `openapi-typescript` as dev dependency.
- **Hand-written runtime, generated types**: `index.js` is maintained manually; `generated.d.ts` is machine-generated. CI enforces sync.
- **Injectable fetch**: Every function accepts `fetchImpl` parameter for mocking in tests and custom fetch behavior.
- **Bearer token injection**: Functions accept `token` and set `Authorization: Bearer <token>` header.

## ANTI-PATTERNS

- Hand-writing types that already exist in `generated.d.ts`. Re-export or reference the generated types.
- Modifying `generated.d.ts` directly. Always regenerate via `npm run generate`.
- Adding runtime dependencies. This package should remain dependency-free.

## NOTES

- Regeneration command: `npm run generate` (or root: `npm run generate:api-client`).
- CI enforces `git diff --exit-code packages/api-client/src` after regeneration. Any OpenAPI change without client regeneration breaks the build.
- Tests use Node's built-in `node --test` with `node:assert/strict`. No Jest, Vitest, or other test framework.
