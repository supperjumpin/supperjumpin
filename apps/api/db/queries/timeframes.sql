-- name: ListRevealTimeframes :many
SELECT id, label, duration_hours, sort_order, created_at
FROM reveal_timeframes
ORDER BY sort_order;

-- name: GetRevealTimeframe :one
SELECT id, label, duration_hours, sort_order, created_at
FROM reveal_timeframes
WHERE id = $1;
