-- name: GetExternalIdentity :one
SELECT player_id, community_id
FROM external_identity
WHERE platform = $1 AND platform_server_id = $2 AND platform_user_id = $3;

-- name: InsertExternalIdentity :exec
INSERT INTO external_identity (platform, platform_server_id, platform_user_id, player_id, community_id)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (platform, platform_server_id, platform_user_id) DO NOTHING;

-- name: FindCommunity :one
SELECT id, display_name, created_at
FROM communities
WHERE id = $1;

-- name: CreateCommunity :exec
INSERT INTO communities (id, display_name)
VALUES ($1, $2)
ON CONFLICT (id) DO NOTHING;

-- name: FindPlayer :one
SELECT id, display_name, created_at
FROM players
WHERE id = $1;

-- name: CreatePlayer :exec
INSERT INTO players (id, display_name)
VALUES ($1, $2)
ON CONFLICT (id) DO NOTHING;

-- name: UpdatePlayerDisplayName :exec
UPDATE players
SET display_name = $2
WHERE id = $1;
