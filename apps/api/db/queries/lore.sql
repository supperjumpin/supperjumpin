-- name: ListRevealedReactionsForCommunity :many
SELECT
    rxn.jump_id,
    s.stance AS stamp_stance,
    j.round_id,
    j.caption AS jump_caption,
    j.player_id AS jump_player_id
FROM reactions rxn
JOIN jumps j ON j.id = rxn.jump_id
JOIN rounds r ON r.id = j.round_id AND r.status = 'revealed'
JOIN stamps s ON s.id = rxn.stamp_id
WHERE r.community_id = $1
ORDER BY j.id, rxn.created_at;
