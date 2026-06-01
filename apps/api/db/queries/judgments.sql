-- name: UpsertPlayerJudgment :one
WITH upsert AS (
  INSERT INTO judgments (id, jump_id, player_id, guest_session_id, provenance, commitment, transgression, creativity, presentation)
  VALUES ($1, $2, $3, NULL, $4, $5, $6, $7, $8)
  ON CONFLICT (jump_id, player_id) WHERE player_id IS NOT NULL DO UPDATE SET
    commitment = EXCLUDED.commitment,
    transgression = EXCLUDED.transgression,
    creativity = EXCLUDED.creativity,
    presentation = EXCLUDED.presentation
  RETURNING (xmax = 0) AS created
)
SELECT created FROM upsert;

-- name: UpsertGuestJudgment :one
WITH upsert AS (
  INSERT INTO judgments (id, jump_id, player_id, guest_session_id, provenance, commitment, transgression, creativity, presentation)
  VALUES ($1, $2, NULL, $3, $4, $5, $6, $7, $8)
  ON CONFLICT (jump_id, guest_session_id) WHERE guest_session_id IS NOT NULL DO UPDATE SET
    commitment = EXCLUDED.commitment,
    transgression = EXCLUDED.transgression,
    creativity = EXCLUDED.creativity,
    presentation = EXCLUDED.presentation
  RETURNING (xmax = 0) AS created
)
SELECT created FROM upsert;

-- name: GetPlayerJudgment :one
SELECT id, jump_id, player_id, guest_session_id, provenance, commitment, transgression, creativity, presentation
FROM judgments
WHERE jump_id = $1 AND player_id = $2;

-- name: ListJudgmentsForJump :many
SELECT id, jump_id, player_id, guest_session_id, provenance, commitment, transgression, creativity, presentation
FROM judgments
WHERE jump_id = $1;

-- name: ListJumpsForJudging :many
SELECT j.id, j.player_id, j.source, j.destination, j.food
FROM jumps j
WHERE j.group_id = $1
  AND j.status = 'Performed Jump'
  AND j.player_id != $2
  AND NOT EXISTS (
      SELECT 1 FROM judgments sub
      WHERE sub.jump_id = j.id AND sub.player_id = $2
  )
ORDER BY j.created_at DESC;

-- name: CreateGuestSession :one
INSERT INTO guest_sessions (id)
VALUES ($1)
RETURNING id, judgment_count, created_at;

-- name: GetGuestSession :one
SELECT id, judgment_count, created_at
FROM guest_sessions
WHERE id = $1;

-- name: IncrementGuestSessionJudgmentCount :exec
UPDATE guest_sessions
SET judgment_count = judgment_count + 1
WHERE id = $1;

-- name: GetGuestSessionJudgmentCount :one
SELECT judgment_count
FROM guest_sessions
WHERE id = $1;
