-- name: CreateJudgment :exec
INSERT INTO judgments (id, jump_id, player_id, difficulty, transgression, creativity, presentation, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now())
ON CONFLICT (jump_id, player_id) 
DO UPDATE SET 
    difficulty = EXCLUDED.difficulty,
    transgression = EXCLUDED.transgression,
    creativity = EXCLUDED.creativity,
    presentation = EXCLUDED.presentation;

-- name: GetJudgment :one
SELECT * FROM judgments
WHERE jump_id = $1 AND player_id = $2;

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
