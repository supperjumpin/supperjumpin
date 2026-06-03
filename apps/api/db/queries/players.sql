-- name: GetGroupPlayers :many
SELECT players.id, players.display_name
FROM group_memberships
JOIN players ON players.id = group_memberships.player_id
WHERE group_memberships.group_id = $1;

-- name: UpdatePlayerDisplayName :exec
UPDATE players SET display_name = $2 WHERE id = $1;
