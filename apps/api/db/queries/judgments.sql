-- name: CreateJudgment :exec
INSERT INTO judgments (id, stunt_id, player_id, difficulty, transgression, creativity, documentation, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
ON CONFLICT (stunt_id, player_id) 
DO UPDATE SET 
    difficulty = EXCLUDED.difficulty,
    transgression = EXCLUDED.transgression,
    creativity = EXCLUDED.creativity,
    documentation = EXCLUDED.documentation,
    updated_at = now();

-- name: GetJudgment :one
SELECT * FROM judgments
WHERE stunt_id = $1 AND player_id = $2;

-- name: ListStuntsForJudging :many
SELECT s.id, s.player_id, s.source, s.destination, s.food
FROM stunts s
WHERE s.group_id = $1 
  AND s.status = 'Performed Stunt'
  AND s.player_id != $2
  AND NOT EXISTS (
      SELECT 1 FROM judgments j 
      WHERE j.stunt_id = s.id AND j.player_id = $2
  )
ORDER BY s.created_at DESC;
