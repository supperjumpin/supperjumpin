-- name: ListJumpsForOpenMonth :many
SELECT id, group_id, player_id, season_id, status, source, destination, food, final_score, grace_period_expires_at
FROM jumps
WHERE created_at >= $1 AND created_at < $2
  AND status IN ('Performed Jump', 'Judged Jump', 'Unjudged Jump', 'Disqualified Jump');

-- name: UpdateJumpOpenFinalScore :exec
UPDATE jumps
SET open_final_score = $2
WHERE id = $1;

-- name: ListPlayersForOpenMonth :many
SELECT DISTINCT p.id, p.display_name
FROM players p
JOIN jumps j ON j.player_id = p.id
WHERE j.created_at >= $1 AND j.created_at < $2
  AND j.status IN ('Performed Jump', 'Judged Jump', 'Unjudged Jump', 'Disqualified Jump');

-- name: UpsertOpenStanding :one
INSERT INTO open_standings (id, year, month, player_id, score, judged_jumps)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (year, month, player_id) DO UPDATE SET
    score = EXCLUDED.score,
    judged_jumps = EXCLUDED.judged_jumps,
    updated_at = now()
RETURNING id;
