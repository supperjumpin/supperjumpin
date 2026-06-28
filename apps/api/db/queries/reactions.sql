-- name: FindReaction :one
SELECT id, stamp_id, jump_id, player_id, created_at
FROM reactions
WHERE stamp_id = $1 AND jump_id = $2 AND player_id = $3;

-- name: CreateReaction :exec
INSERT INTO reactions (id, stamp_id, jump_id, player_id, created_at)
VALUES ($1, $2, $3, $4, $5);
