-- name: FindActiveRound :one
SELECT id, community_id, prompt_id, status, reveal_by, created_by, created_at
FROM rounds
WHERE community_id = $1 AND status = 'active'
LIMIT 1;

-- name: CreateRound :exec
INSERT INTO rounds (id, community_id, prompt_id, status, reveal_by, created_by, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetRound :one
SELECT id, community_id, prompt_id, status, reveal_by, created_by, created_at
FROM rounds
WHERE id = $1;
