-- name: InsertEvidence :exec
INSERT INTO evidences (id, jump_id, player_id, upload_authorization_id, caption, media_object_key, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: UpdateEvidenceCaption :exec
UPDATE evidences
SET caption = $2
WHERE jump_id = $1;
