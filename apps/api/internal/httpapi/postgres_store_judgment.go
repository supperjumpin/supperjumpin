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

func (s *PostgresStore) UpsertJudgment(ctx context.Context, jumpID, playerID string, commitment, transgression, creativity, presentation int) (game.Judgment, bool, error) {
	judgmentID := stableID("judgment", jumpID+":"+playerID)
	created, err := s.queries.UpsertJudgment(ctx, db.UpsertJudgmentParams{
		ID:            judgmentID,
		JumpID:        jumpID,
		PlayerID:      playerID,
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
		PlayerID:      playerID,
		Commitment:    commitment,
		Transgression: transgression,
		Creativity:    creativity,
		Presentation:  presentation,
	}, created, nil
}

func (s *PostgresStore) AdvanceJumpToJudged(ctx context.Context, jumpID string) error {
	return s.queries.AdvanceJumpToJudged(ctx, jumpID)
}
