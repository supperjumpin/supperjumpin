# apps/mobile KNOWLEDGE BASE

## OVERVIEW

Expo React Native player-facing app. Pre-MVP single-file prototype shell. Renders the Group Jump loop, captures Evidence, and calls the backend API.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Entry point | `index.js` | `registerRootComponent(App)` |
| Entire app UI | `App.tsx` | Single file (~550 lines). All screens, state, auth, API calls inline. |
| Expo config | `app.json` | Slug, scheme `supperjumpin`, portrait, automatic UI style |
| Env vars | `.env.example` | `EXPO_PUBLIC_SUPABASE_URL`, `EXPO_PUBLIC_SUPABASE_ANON_KEY`, `EXPO_PUBLIC_API_BASE_URL` |
| TypeScript config | `tsconfig.json` | Extends `expo/tsconfig.base`, strict, react-jsx, bundler resolution |

## CONVENTIONS

- **No routing library**: Conditional rendering blocks inside a single `ScrollView`. No React Navigation.
- **No state management library**: Raw `useState` hooks only (~15 state variables). No Context, Zustand, Redux, or Jotai.
- **Supabase auth directly**: `GoTrueClient` instantiated in `App.tsx`.
- **API calls via `@supperjumpin/api-client`**: Direct imports, no custom fetch wrapper in the mobile app.
- **PanResponder for gestures**: Judging UI uses `PanResponder` for swipe-based score adjustment.
- **Standard RN primitives**: `SafeAreaView`, `ScrollView`, `StyleSheet`, `TextInput`, `Button`.

## ANTI-PATTERNS

- Owning game rules in the mobile app. Jump lifecycle transitions, judging eligibility, score aggregation, and House Rules boundaries are backend concerns.
- Duplicating API payload types. Always consume `@supperjumpin/api-client` instead of hand-writing types.
- Adding a state management library before the app is decomposed into multiple files/screens.

## NOTES

- React 19.1.0, React Native 0.81.5, Expo 54.0.34.
- As the UI grows, `App.tsx` should be the first file to split into screens and components.
- If navigation is introduced, evaluate React Navigation before custom solutions.
