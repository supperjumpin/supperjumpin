-- name: ListGhostJumpers :many
SELECT c.player_id, c.committed_at
FROM commits c
LEFT JOIN jumps j ON j.round_id = c.round_id AND j.player_id = c.player_id
WHERE c.round_id = $1 AND j.id IS NULL
ORDER BY c.committed_at;

-- name: ListReactionsForRound :many
SELECT rxn.jump_id, s.stance AS stamp_stance
FROM reactions rxn
JOIN jumps j ON j.id = rxn.jump_id
JOIN stamps s ON s.id = rxn.stamp_id
WHERE j.round_id = $1
ORDER BY rxn.jump_id, rxn.created_at;

-- name: ListAllCommentsForRound :many
SELECT id, round_id, jump_id, player_id, body, created_at
FROM comments
WHERE round_id = $1
ORDER BY created_at;
