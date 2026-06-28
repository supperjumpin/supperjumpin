-- name: ListStamps :many
SELECT id, stance, label, glyph, copy, created_at
FROM stamps
ORDER BY created_at;

-- name: GetStamp :one
SELECT id, stance, label, glyph, copy, created_at
FROM stamps
WHERE id = $1;

-- name: GetStampByStance :one
SELECT id, stance, label, glyph, copy, created_at
FROM stamps
WHERE stance = $1;
