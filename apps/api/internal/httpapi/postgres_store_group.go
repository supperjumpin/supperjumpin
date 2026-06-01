package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/db"
	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.GroupRepository adapter methods for PostgresStore

func (s *PostgresStore) InsertGroup(ctx context.Context, groupID, name string) error {
	return s.queries.InsertGroup(ctx, db.InsertGroupParams{ID: groupID, Name: name})
}

func (s *PostgresStore) InsertMembership(ctx context.Context, groupID, playerID, role string) error {
	return s.queries.InsertMembership(ctx, db.InsertMembershipParams{
		GroupID:  groupID,
		PlayerID: playerID,
		Role:     role,
	})
}

func (s *PostgresStore) Group(ctx context.Context, groupID string) (game.GroupSnapshot, bool, error) {
	row, err := s.queries.GetGroup(ctx, groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return game.GroupSnapshot{}, false, nil
	}
	if err != nil {
		return game.GroupSnapshot{}, false, err
	}
	return game.GroupSnapshot{ID: row.ID, Name: row.Name}, true, nil
}

func (s *PostgresStore) Membership(ctx context.Context, playerID, groupID string) (game.GroupMembershipSnapshot, bool, error) {
	row, err := s.queries.GetMembership(ctx, db.GetMembershipParams{PlayerID: playerID, GroupID: groupID})
	if errors.Is(err, sql.ErrNoRows) {
		return game.GroupMembershipSnapshot{}, false, nil
	}
	if err != nil {
		return game.GroupMembershipSnapshot{}, false, err
	}
	return game.GroupMembershipSnapshot{GroupID: row.GroupID, PlayerID: row.PlayerID, Role: row.Role}, true, nil
}

func (s *PostgresStore) MembershipsForPlayer(ctx context.Context, playerID string) ([]game.MembershipWithGroupSnapshot, error) {
	rows, err := s.queries.ListMembershipsForPlayer(ctx, playerID)
	if err != nil {
		return nil, err
	}

	var result []game.MembershipWithGroupSnapshot
	for _, row := range rows {
		result = append(result, game.MembershipWithGroupSnapshot{
			Group: game.GroupSnapshot{
				ID:   row.ID,
				Name: row.Name,
			},
			Membership: game.GroupMembershipSnapshot{
				GroupID:  row.GroupID,
				PlayerID: row.PlayerID,
				Role:     row.Role,
			},
		})
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
	err = s.queries.InsertInvite(ctx, db.InsertInviteParams{
		ID:                id,
		GroupID:           groupID,
		Token:             token,
		CreatedByPlayerID: createdByPlayerID,
		ExpiresAt:         expiresAtTime,
	})
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
	row, err := s.queries.GetInviteByToken(ctx, token)
	if errors.Is(err, sql.ErrNoRows) {
		return game.InviteSnapshot{}, false, nil
	}
	if err != nil {
		return game.InviteSnapshot{}, false, err
	}
	snap := game.InviteSnapshot{
		ID:        row.ID,
		GroupID:   row.GroupID,
		Token:     row.Token,
		CreatedBy: row.CreatedByPlayerID,
		ExpiresAt: row.ExpiresAt.Unix(),
	}
	if row.UsedByPlayerID.Valid {
		usedBy := row.UsedByPlayerID.String
		snap.UsedBy = &usedBy
	}
	return snap, true, nil
}

func (s *PostgresStore) MarkInviteUsed(ctx context.Context, token, playerID string) error {
	return s.queries.MarkInviteUsed(ctx, db.MarkInviteUsedParams{
		UsedByPlayerID: sql.NullString{String: playerID, Valid: true},
		Token:          token,
	})
}
