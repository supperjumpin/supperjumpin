-- name: GetJump :one
SELECT id, player_id, season_id, status, source, destination, food, final_score, grace_period_expires_at
FROM jumps
WHERE id = $1;

-- name: AdvanceJumpToJudged :exec
UPDATE jumps
SET status = 'Judged Jump'
WHERE id = $1 AND status = 'Performed Jump';
