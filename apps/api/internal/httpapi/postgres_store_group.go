package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.GroupRepository adapter methods for PostgresStore

func (s *PostgresStore) InsertGroup(ctx context.Context, groupID, name string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO groups (id, name) VALUES ($1, $2)`, groupID, name)
	return err
}

func (s *PostgresStore) InsertMembership(ctx context.Context, groupID, playerID, role string) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO group_memberships (group_id, player_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (group_id, player_id) DO UPDATE SET role = EXCLUDED.role`, groupID, playerID, role)
	return err
}

func (s *PostgresStore) Group(ctx context.Context, groupID string) (game.GroupSnapshot, bool, error) {
	var snap game.GroupSnapshot
	err := s.db.QueryRowContext(ctx, `SELECT id, name FROM groups WHERE id = $1`, groupID).Scan(&snap.ID, &snap.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return game.GroupSnapshot{}, false, nil
	}
	if err != nil {
		return game.GroupSnapshot{}, false, err
	}
	return snap, true, nil
}

func (s *PostgresStore) Membership(ctx context.Context, playerID, groupID string) (game.GroupMembershipSnapshot, bool, error) {
	var snap game.GroupMembershipSnapshot
	err := s.db.QueryRowContext(ctx, `
SELECT group_id, player_id, role FROM group_memberships
WHERE player_id = $1 AND group_id = $2`, playerID, groupID).Scan(&snap.GroupID, &snap.PlayerID, &snap.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return game.GroupMembershipSnapshot{}, false, nil
	}
	if err != nil {
		return game.GroupMembershipSnapshot{}, false, err
	}
	return snap, true, nil
}

func (s *PostgresStore) MembershipsForPlayer(ctx context.Context, playerID string) ([]game.MembershipWithGroupSnapshot, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT groups.id, groups.name, group_memberships.group_id, group_memberships.player_id, group_memberships.role
FROM group_memberships
JOIN groups ON groups.id = group_memberships.group_id
WHERE group_memberships.player_id = $1`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []game.MembershipWithGroupSnapshot
	for rows.Next() {
		var m game.MembershipWithGroupSnapshot
		if err := rows.Scan(&m.Group.ID, &m.Group.Name, &m.Membership.GroupID, &m.Membership.PlayerID, &m.Membership.Role); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PostgresStore) InsertInvite(ctx context.Context, groupID, createdByPlayerID string, expiresAt int64) (game.InviteSnapshot, error) {
	id, err := randomToken("invite")
	if err != nil {
		return game.InviteSnapshot{}, err
	}
	token, err := randomToken("invite_token")
	if err != nil {
		return game.InviteSnapshot{}, err
	}
	expiresAtTime := time.Unix(expiresAt, 0).UTC()
	_, err = s.db.ExecContext(ctx, `
INSERT INTO invites (id, group_id, token, created_by_player_id, expires_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT DO NOTHING`, id, groupID, token, createdByPlayerID, expiresAtTime)
	if err != nil {
		return game.InviteSnapshot{}, err
	}
	return game.InviteSnapshot{
		ID:        id,
		GroupID:   groupID,
		Token:     token,
		CreatedBy: createdByPlayerID,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *PostgresStore) InviteByToken(ctx context.Context, token string) (game.InviteSnapshot, bool, error) {
	var snap game.InviteSnapshot
	var usedBy sql.NullString
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, token, created_by_player_id, used_by_player_id, EXTRACT(EPOCH FROM expires_at)::bigint
FROM invites
WHERE token = $1`, token).Scan(&snap.ID, &snap.GroupID, &snap.Token, &snap.CreatedBy, &usedBy, &snap.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return game.InviteSnapshot{}, false, nil
	}
	if err != nil {
		return game.InviteSnapshot{}, false, err
	}
	if usedBy.Valid {
		snap.UsedBy = &usedBy.String
	}
	return snap, true, nil
}

func (s *PostgresStore) MarkInviteUsed(ctx context.Context, token, playerID string) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE invites
SET used_by_player_id = $1
WHERE token = $2 AND used_by_player_id IS NULL`, playerID, token)
	return err
}
