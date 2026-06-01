-- name: GetSeason :one
SELECT id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline
FROM seasons WHERE id = $1;

-- name: GetOpenSeasonForGroup :one
SELECT id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline
FROM seasons
WHERE group_id = $1 AND status IN ('Active', 'Judging Grace Period')
LIMIT 1;

-- name: GetActiveSeasonForGroup :one
SELECT id, group_id, commissioner_player_id, status
FROM seasons
WHERE group_id = $1 AND status = 'Active'
ORDER BY created_at DESC
LIMIT 1;

-- name: GetLatestSeasonForGroup :one
SELECT id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline
FROM seasons
WHERE group_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: ListSeasonsForGroup :many
SELECT id, status, submission_deadline, judging_deadline
FROM seasons
WHERE group_id = $1;

-- name: CountSeasons :one
SELECT count(*) FROM seasons;

-- name: InsertSeason :exec
INSERT INTO seasons (id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdateSeasonStatus :exec
UPDATE seasons SET status = $2 WHERE id = $1;

-- name: InsertSeasonHistoryEntry :exec
INSERT INTO season_history (id, season_id, action, actor_player_id, actor_role, override, from_status, to_status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListSeasonHistoryEntries :many
SELECT id, season_id, action, actor_player_id, actor_role, override, from_status, to_status
FROM season_history
WHERE season_id = $1
ORDER BY created_at ASC, id ASC;
