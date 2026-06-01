-- name: InsertEvidenceUploadAuthorization :one
INSERT INTO evidence_upload_authorizations (id, jump_id, player_id, content_type, media_object_key, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, jump_id, player_id, content_type, media_object_key, expires_at;

-- name: GetEvidenceUploadAuthorizationForUpdate :one
SELECT id, player_id, media_object_key, expires_at
FROM evidence_upload_authorizations
WHERE id = $1 AND jump_id = $2
FOR UPDATE;

-- name: InsertEvidence :exec
INSERT INTO evidences (id, jump_id, player_id, upload_authorization_id, caption, media_object_key, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: DeleteEvidenceUploadAuthorization :exec
DELETE FROM evidence_upload_authorizations WHERE id = $1;
