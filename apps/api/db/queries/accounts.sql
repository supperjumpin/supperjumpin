-- name: GetAccountByAuthIdentity :one
SELECT accounts.id, accounts.email
FROM accounts
JOIN auth_identities ON auth_identities.account_id = accounts.id
WHERE auth_identities.provider = $1 AND auth_identities.subject = $2;

-- name: GetPlayerByAccountID :one
SELECT id, account_id, display_name
FROM players
WHERE account_id = $1;
