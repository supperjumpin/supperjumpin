-- name: InsertGroup :exec
INSERT INTO groups (id, name)
VALUES ($1, $2);

-- name: InsertMembership :exec
INSERT INTO group_memberships (group_id, player_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (group_id, player_id) DO UPDATE SET role = EXCLUDED.role;

-- name: GetGroup :one
SELECT id, name FROM groups WHERE id = $1;

-- name: GetMembership :one
SELECT group_id, player_id, role FROM group_memberships
WHERE player_id = $1 AND group_id = $2;

-- name: GetMembershipRole :one
SELECT role FROM group_memberships
WHERE player_id = $1 AND group_id = $2;

-- name: ListMembershipsForPlayer :many
SELECT groups.id, groups.name, group_memberships.group_id, group_memberships.player_id, group_memberships.role
FROM group_memberships
JOIN groups ON groups.id = group_memberships.group_id
WHERE group_memberships.player_id = $1
ORDER BY groups.name;

-- name: InsertInvite :exec
INSERT INTO invites (id, group_id, token, created_by_player_id, expires_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT DO NOTHING;

-- name: GetInviteByToken :one
SELECT id, group_id, token, created_by_player_id, used_by_player_id, expires_at
FROM invites
WHERE token = $1;

-- name: MarkInviteUsed :exec
UPDATE invites
SET used_by_player_id = $1
WHERE token = $2 AND used_by_player_id IS NULL;
