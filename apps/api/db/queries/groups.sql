-- name: CreateGroup :exec
INSERT INTO groups (id, name)
VALUES ($1, $2);

-- name: CreateGroupMembership :exec
INSERT INTO group_memberships (group_id, player_id, role)
VALUES ($1, $2, $3);

-- name: ListGroupMembershipsByPlayerID :many
SELECT groups.id, groups.name, group_memberships.player_id, group_memberships.role
FROM group_memberships
JOIN groups ON groups.id = group_memberships.group_id
WHERE group_memberships.player_id = $1
ORDER BY groups.name;

-- name: GetGroupMembership :one
SELECT groups.id, groups.name, group_memberships.player_id, group_memberships.role
FROM group_memberships
JOIN groups ON groups.id = group_memberships.group_id
WHERE group_memberships.group_id = $1 AND group_memberships.player_id = $2;
