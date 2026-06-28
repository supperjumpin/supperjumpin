-- name: FindCommit :one
SELECT id, round_id, player_id, committed_at
FROM commits
WHERE round_id = $1 AND player_id = $2;

-- name: CreateCommit :exec
INSERT INTO commits (id, round_id, player_id, committed_at)
VALUES ($1, $2, $3, $4);

-- name: FindJumpByPlayer :one
SELECT id, round_id, player_id, caption, submitted_at
FROM jumps
WHERE round_id = $1 AND player_id = $2;

-- name: CreateJump :exec
INSERT INTO jumps (id, round_id, player_id, caption, submitted_at)
VALUES ($1, $2, $3, $4, $5);

-- name: InsertJumpEvidence :exec
INSERT INTO jump_evidence (id, jump_id, url, sort_order)
VALUES ($1, $2, $3, $4);

-- name: ListJumpsByRound :many
SELECT id, round_id, player_id, submitted_at
FROM jumps
WHERE round_id = $1
ORDER BY submitted_at;

-- name: ListJumpsByRoundWithContent :many
SELECT id, round_id, player_id, caption, submitted_at
FROM jumps
WHERE round_id = $1
ORDER BY submitted_at;

-- name: ListEvidenceForJumps :many
SELECT jump_id, url, sort_order
FROM jump_evidence
WHERE jump_id = ANY($1::TEXT[])
ORDER BY sort_order;

-- name: GetJumpByID :one
SELECT id, round_id, player_id, caption, submitted_at
FROM jumps
WHERE id = $1;

-- name: ListEvidenceForJump :many
SELECT id, jump_id, url
FROM jump_evidence
WHERE jump_id = $1
ORDER BY sort_order;

-- name: GetRoundStatus :one
SELECT
    r.id,
    r.status,
    r.reveal_by,
    COUNT(DISTINCT c.id)::INT AS commit_count,
    COUNT(DISTINCT j.id)::INT AS submission_count
FROM rounds r
LEFT JOIN commits c ON c.round_id = r.id
LEFT JOIN jumps j ON j.round_id = r.id
WHERE r.id = $1
GROUP BY r.id;
