-- name: UpsertJudgment :one
WITH upsert AS (
  INSERT INTO judgments (id, jump_id, player_id, difficulty, transgression, creativity, presentation)
  VALUES ($1, $2, $3, $4, $5, $6, $7)
  ON CONFLICT (jump_id, player_id) DO UPDATE SET
    difficulty = EXCLUDED.difficulty,
    transgression = EXCLUDED.transgression,
    creativity = EXCLUDED.creativity,
    presentation = EXCLUDED.presentation
  RETURNING (xmax = 0) AS created
)
SELECT created FROM upsert;

-- name: GetJudgment :one
SELECT id, jump_id, player_id, difficulty, transgression, creativity, presentation
FROM judgments
WHERE jump_id = $1 AND player_id = $2;

-- name: ListJudgmentsForJump :many
SELECT id, jump_id, player_id, difficulty, transgression, creativity, presentation
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
