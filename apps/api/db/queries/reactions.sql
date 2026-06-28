-- name: CreateReaction :exec
INSERT INTO reactions (id, stamp_id, jump_id, player_id, created_at)
VALUES ($1, $2, $3, $4, $5);
