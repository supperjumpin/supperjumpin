-- name: ListPromptPacks :many
SELECT id, display_name, description, created_at
FROM prompt_packs
ORDER BY created_at;

-- name: GetPromptPack :one
SELECT id, display_name, description, created_at
FROM prompt_packs
WHERE id = $1;

-- name: ListPrompts :many
SELECT id, pack_id, copy, theme, cost_tier, created_at
FROM prompts
ORDER BY created_at;

-- name: ListPromptsByPack :many
SELECT id, pack_id, copy, theme, cost_tier, created_at
FROM prompts
WHERE pack_id = $1
ORDER BY created_at;

-- name: GetPrompt :one
SELECT id, pack_id, copy, theme, cost_tier, created_at
FROM prompts
WHERE id = $1;
