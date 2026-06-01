package httpapi

import (
	"context"
	"database/sql"
	"errors"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/db"
	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.JudgmentRepository adapter methods for PostgresStore

func (s *PostgresStore) Jump(ctx context.Context, jumpID string) (game.JumpSnapshot, bool, error) {
	row, err := s.queries.GetJump(ctx, jumpID)
	if errors.Is(err, sql.ErrNoRows) {
		return game.JumpSnapshot{}, false, nil
	}
	if err != nil {
		return game.JumpSnapshot{}, false, err
	}
	snap := game.JumpSnapshot{
		ID:          row.ID,
		PlayerID:    row.PlayerID,
		Status:      row.Status,
		Source:      row.Source,
		Destination: row.Destination,
		Food:        row.Food,
	}
	if row.GroupID.Valid {
		snap.GroupID = row.GroupID.String
	}
	if !visiblePerformedStatus(snap.Status) {
		return game.JumpSnapshot{}, false, nil
	}
	if row.SeasonID.Valid {
		snap.SeasonID = &row.SeasonID.String
	}
	if row.GracePeriodExpiresAt.Valid {
		snap.GracePeriodExpiresAt = row.GracePeriodExpiresAt.Time
	}
	if row.FinalScore.Valid {
		finalScore := int(row.FinalScore.Int32)
		snap.FinalScore = &finalScore
	}
	return snap, true, nil
}

func (s *PostgresStore) Season(ctx context.Context, seasonID string) (game.SeasonSnapshot, error) {
	row, err := s.queries.GetSeason(ctx, seasonID)
	if errors.Is(err, sql.ErrNoRows) {
		return game.SeasonSnapshot{}, nil
	}
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	return game.SeasonSnapshot{
		ID:                   row.ID,
		GroupID:              row.GroupID,
		CommissionerPlayerID: row.CommissionerPlayerID,
		Status:               row.Status,
		SubmissionDeadline:   row.SubmissionDeadline,
		JudgingDeadline:      row.JudgingDeadline,
	}, nil
}

func (s *PostgresStore) GroupMembership(ctx context.Context, playerID, groupID string) (game.MembershipSnapshot, bool, error) {
	role, err := s.queries.GetMembershipRole(ctx, db.GetMembershipRoleParams{PlayerID: playerID, GroupID: groupID})
	if errors.Is(err, sql.ErrNoRows) {
		return game.MembershipSnapshot{}, false, nil
	}
	if err != nil {
		return game.MembershipSnapshot{}, false, err
	}
	return game.MembershipSnapshot{Role: role}, true, nil
}

func (s *PostgresStore) UpsertJudgment(ctx context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (game.Judgment, bool, error) {
	var judgmentID, identityKey string
	if playerID != "" {
		judgmentID = stableID("judgment", jumpID+":"+playerID)
		identityKey = playerID
		created, err := s.queries.UpsertPlayerJudgment(ctx, db.UpsertPlayerJudgmentParams{
			ID:            judgmentID,
			JumpID:        jumpID,
			PlayerID:      sql.NullString{String: playerID, Valid: true},
			Provenance:    provenance,
			Commitment:    int32(commitment),
			Transgression: int32(transgression),
			Creativity:    int32(creativity),
			Presentation:  int32(presentation),
		})
		if err != nil {
			return game.Judgment{}, false, err
		}
		return game.Judgment{
			ID:            judgmentID,
			JumpID:        jumpID,
			PlayerID:      identityKey,
			Provenance:    provenance,
			Commitment:    commitment,
			Transgression: transgression,
			Creativity:    creativity,
			Presentation:  presentation,
		}, created, nil
	}

	judgmentID = stableID("judgment", jumpID+":guest:"+guestSessionID)
	identityKey = guestSessionID
	created, err := s.queries.UpsertGuestJudgment(ctx, db.UpsertGuestJudgmentParams{
		ID:             judgmentID,
		JumpID:         jumpID,
		GuestSessionID: sql.NullString{String: guestSessionID, Valid: true},
		Provenance:     provenance,
		Commitment:     int32(commitment),
		Transgression:  int32(transgression),
		Creativity:     int32(creativity),
		Presentation:   int32(presentation),
	})
	if err != nil {
		return game.Judgment{}, false, err
	}
	return game.Judgment{
		ID:             judgmentID,
		JumpID:         jumpID,
		GuestSessionID: identityKey,
		Provenance:     provenance,
		Commitment:     commitment,
		Transgression:  transgression,
		Creativity:     creativity,
		Presentation:   presentation,
	}, created, nil
}

func (s *PostgresStore) AdvanceJumpToJudged(ctx context.Context, jumpID string) error {
	return s.queries.AdvanceJumpToJudged(ctx, jumpID)
}

func (s *PostgresStore) GuestSessionJudgmentCount(ctx context.Context, guestSessionID string) (int, error) {
	count, err := s.queries.GetGuestSessionJudgmentCount(ctx, guestSessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (s *PostgresStore) IncrementGuestSessionJudgmentCount(ctx context.Context, guestSessionID string) error {
	return s.queries.IncrementGuestSessionJudgmentCount(ctx, guestSessionID)
}

func (s *PostgresStore) CreateGuestSession(ctx context.Context, id string) error {
	_, err := s.queries.CreateGuestSession(ctx, id)
	return err
}
