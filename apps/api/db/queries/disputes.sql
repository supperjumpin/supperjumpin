-- name: CountDisputesForJump :one
SELECT count(*) FROM disputes WHERE jump_id = $1;

-- name: InsertDispute :exec
INSERT INTO disputes (id, jump_id, raised_by_player_id, concern, details, status)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetDispute :one
SELECT id, jump_id, raised_by_player_id, concern, details, status,
       resolution, resolution_reason, resolved_by_player_id,
       override_resolution, override_reason, override_by_player_id
FROM disputes
WHERE id = $1;

-- name: UpdateDisputeResolution :exec
UPDATE disputes
SET status = 'Resolved', resolution = $2, resolution_reason = $3, resolved_by_player_id = $4
WHERE id = $1;

-- name: UpdateDisputeOverride :exec
UPDATE disputes
SET status = 'Overridden', override_resolution = $2, override_reason = $3, override_by_player_id = $4
WHERE id = $1;

-- name: ListDisputesForJump :many
SELECT id, jump_id, raised_by_player_id, concern, details, status,
       resolution, resolution_reason, resolved_by_player_id,
       override_resolution, override_reason, override_by_player_id
FROM disputes
WHERE jump_id = $1
ORDER BY created_at ASC, id ASC;
