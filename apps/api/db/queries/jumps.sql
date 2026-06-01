-- name: InsertIdea :execrows
INSERT INTO jumps (id, group_id, player_id, status, source, destination, food)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT DO NOTHING;

-- name: GetJump :one
SELECT id, group_id, player_id, season_id, status, source, destination, food, final_score, grace_period_expires_at
FROM jumps
WHERE id = $1;

-- name: GetJumpByID :one
SELECT id, group_id, player_id, season_id, status, source, destination, food, final_score, grace_period_expires_at
FROM jumps
WHERE id = $1;

-- name: UpdateJumpToPlanned :one
UPDATE jumps
SET status = 'Planned Jump', season_id = $2
WHERE id = $1
  AND status = 'Idea'
RETURNING id, group_id, player_id, season_id, status, source, destination, food, grace_period_expires_at;

-- name: AdoptJumpToSeason :exec
UPDATE jumps
SET status = 'Performed Jump', grace_period_expires_at = $2
WHERE id = $1 AND status = 'Planned Jump';

-- name: ListJumpsForSeason :many
SELECT id, group_id, player_id, season_id, status, source, destination, food, final_score, grace_period_expires_at
FROM jumps
WHERE season_id = $1;

-- name: UpdateJumpFinalization :exec
UPDATE jumps
SET status = $2, final_score = $3
WHERE id = $1;

-- name: AdvanceJumpToJudged :exec
UPDATE jumps
SET status = 'Judged Jump'
WHERE id = $1 AND status = 'Performed Jump';

-- name: UpdateJumpStatusAfterDispute :exec
UPDATE jumps
SET status = $2, final_score = NULL
WHERE id = $1;

-- name: ListPerformedJumpsForGroup :many
SELECT jumps.id, jumps.group_id, jumps.player_id, jumps.season_id, jumps.status, jumps.source, jumps.destination, jumps.food, jumps.final_score, jumps.grace_period_expires_at,
       evidences.id, evidences.caption, evidences.media_object_key, evidences.created_at,
       players.id, players.display_name
FROM jumps
JOIN evidences ON evidences.jump_id = jumps.id
JOIN players ON players.id = jumps.player_id
WHERE jumps.group_id = $1 AND jumps.status IN ('Performed Jump', 'Judged Jump', 'Unjudged Jump', 'Disqualified Jump')
ORDER BY evidences.created_at DESC, jumps.id DESC;
