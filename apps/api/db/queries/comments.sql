-- name: CreateComment :exec
INSERT INTO comments (id, round_id, jump_id, player_id, body, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListCommentsForJump :many
SELECT id, round_id, jump_id, player_id, body, created_at
FROM comments
WHERE round_id = $1 AND jump_id = $2
ORDER BY created_at;

-- name: ListCommentsForRound :many
SELECT id, round_id, player_id, body, created_at
FROM comments
WHERE round_id = $1 AND jump_id IS NULL
ORDER BY created_at;
