     1|-- name: CreateJudgment :exec
     2|INSERT INTO judgments (id, jump_id, player_id, commitment, transgression, creativity, presentation, created_at, updated_at)
     3|VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
     4|ON CONFLICT (jump_id, player_id) 
     5|DO UPDATE SET 
     6|    commitment = EXCLUDED.commitment,
     7|    transgression = EXCLUDED.transgression,
     8|    creativity = EXCLUDED.creativity,
     9|    presentation = EXCLUDED.presentation,
    10|    updated_at = now();
    11|
    12|-- name: GetJudgment :one
    13|SELECT * FROM judgments
    14|WHERE jump_id = $1 AND player_id = $2;
    15|
    16|-- name: ListJumpsForJudging :many
    17|SELECT s.id, s.player_id, s.source, s.destination, s.food
    18|FROM jumps s
    19|WHERE s.group_id = $1 
    20|  AND s.status = 'Performed Jump'
    21|  AND s.player_id != $2
    22|  AND NOT EXISTS (
    23|      SELECT 1 FROM judgments j 
    24|      WHERE j.jump_id = s.id AND j.player_id = $2
    25|  )
    26|ORDER BY s.created_at DESC;
    27|