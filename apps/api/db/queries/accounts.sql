-- name: GetAccountByAuthIdentity :one
SELECT accounts.id, accounts.email
FROM accounts
JOIN auth_identities ON auth_identities.account_id = accounts.id
WHERE auth_identities.provider = $1 AND auth_identities.subject = $2;

-- name: UpsertAccount :exec
INSERT INTO accounts (id, email)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email;

-- name: InsertAuthIdentity :exec
INSERT INTO auth_identities (provider, subject, account_id)
VALUES ($1, $2, $3)
ON CONFLICT (provider, subject) DO NOTHING;

-- name: InsertPlayer :exec
INSERT INTO players (id, account_id, display_name)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO NOTHING;

-- name: GetPlayerByAccountID :one
SELECT id, account_id, display_name
FROM players
WHERE account_id = $1;

-- name: GetProfileByAuthIdentity :one
SELECT accounts.id, accounts.email, players.id, players.display_name
FROM accounts
JOIN auth_identities ON auth_identities.account_id = accounts.id
JOIN players ON players.account_id = accounts.id
WHERE auth_identities.provider = $1 AND auth_identities.subject = $2;

-- name: UpdatePlayerDisplayName :exec
UPDATE players
SET display_name = $2
WHERE id = $1;
