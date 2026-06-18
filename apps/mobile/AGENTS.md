# apps/mobile KNOWLEDGE BASE

## OVERVIEW

Expo React Native player-facing app. Pre-MVP prototype shell. Renders the public Jump feed/detail/create loop, captures Evidence, and calls the backend API.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Entry point | `index.js` | `registerRootComponent(App)` |
| App shell / flow | `App.tsx` | Top-level screen switching, auth/profile flow, and Draft lifetime. |
| Screens | `*Screen.tsx` | Feed, Jump detail, create, login, and display-name setup screens. |
| Expo config | `app.json` | Slug, scheme `supperjumpin`, portrait, automatic UI style |
| Env vars | `.env.example` | `EXPO_PUBLIC_API_BASE_URL`, `EXPO_PUBLIC_DEV_AUTH_TOKEN`, optional `EXPO_PUBLIC_MEDIA_BASE_URL` |
| TypeScript config | `tsconfig.json` | Extends `expo/tsconfig.base`, strict, react-jsx, bundler resolution |
| Tests | `*.test.tsx` alongside source | Jest with `jest-expo` preset, `@testing-library/react-native` v14 |
| API mocking | `test/mockApi.ts` | Intercept `global.fetch` for `@supperjumpin/api-client` public-read calls |

## CONVENTIONS

- **No routing library**: Conditional rendering in `App.tsx`. No React Navigation.
- **No state management library**: Raw React state only. No Context, Zustand, Redux, or Jotai.
- **Vanilla until the seam is proven**: Prefer plain React/TypeScript until a concrete seam shows a library would remove more app-specific complexity than it adds. Refactors should preserve behavior while creating testable seams for likely follow-on work; do not pre-build speculative routing, auth, or state infrastructure.
- **Local-first auth**: `App.tsx` uses `EXPO_PUBLIC_DEV_AUTH_TOKEN` for signed-in flows. When hosted auth is added, it will be additive — the dev token remains the local development path.
- **API calls via `@supperjumpin/api-client`**: Direct imports, no custom fetch wrapper in the mobile app.
- **Standard RN primitives**: Prefer existing `SafeAreaView`, `ScrollView`, `StyleSheet`, `TextInput`, and button/touchable patterns unless the app adopts a UI library.

## ANTI-PATTERNS

- Owning game rules in the mobile app. Jump lifecycle transitions, judging eligibility, score aggregation, and House Rules boundaries are backend concerns.
- Duplicating API payload types. Always consume `@supperjumpin/api-client` instead of hand-writing types.
- Adding a state management library before the existing screen split has a concrete state-sharing problem.

## NOTES

- React 19.1.0, React Native 0.81.5, Expo 54.0.34.
- `App.tsx` still owns flow policy; extract plain TypeScript flow modules before adding a routing/state library.
- If navigation is introduced, evaluate React Navigation before custom solutions.
- A local profile switcher will be added when acceptance testing requires switching between two authenticated Players inside one local mobile app session. This is a developer tool, not a user-facing feature — it toggles the `EXPO_PUBLIC_DEV_AUTH_TOKEN` value at runtime.
