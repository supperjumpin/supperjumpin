# REST and OpenAPI Contract

The Go backend will expose a REST API with an OpenAPI contract, and the Expo React Native app will consume generated TypeScript client types from that contract. We rejected GraphQL for the MVP because Supperjumpin's early API is dominated by explicit game commands and media-backed state transitions rather than highly variable client-defined read graphs.
