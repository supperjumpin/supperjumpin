-- name: CreateJudgment :exec
INSERT INTO judgments (id, jump_id, player_id, difficulty, transgression, creativity, presentation, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
ON CONFLICT (jump_id, player_id) 
DO UPDATE SET 
    difficulty = EXCLUDED.difficulty,
    transgression = EXCLUDED.transgression,
    creativity = EXCLUDED.creativity,
    presentation = EXCLUDED.presentation,
    updated_at = now();

-- name: GetJudgment :one
SELECT * FROM judgments
WHERE jump_id = $1 AND player_id = $2;

-- name: ListJumpsForJudging :many
SELECT s.id, s.player_id, s.source, s.destination, s.food
FROM jumps s
WHERE s.group_id = $1 
  AND s.status = 'Performed Jump'
  AND s.player_id != $2
  AND NOT EXISTS (
      SELECT 1 FROM judgments j 
      WHERE j.jump_id = s.id AND j.player_id = $2
  )
ORDER BY s.created_at DESC;
