# Local-First Auth with Internal Accounts

Supperjumpin will defer hosted authentication until the local MVP is playable end-to-end. The Go backend keeps an internal Account and Player model, while local development uses a static bearer token to exercise signed-in flows. Future external auth provider identities may attach to Accounts; they do not define in-game Player identity, Group membership, Season history, or Jump ownership, preserving a migration path to hosted login methods later.
