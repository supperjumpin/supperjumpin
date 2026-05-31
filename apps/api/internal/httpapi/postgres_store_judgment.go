package httpapi

import (
	"context"
	"database/sql"
	"errors"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.JudgmentRepository adapter methods for PostgresStore

func (s *PostgresStore) Jump(ctx context.Context, jumpID string) (game.JumpSnapshot, bool, error) {
	var snap game.JumpSnapshot
	var seasonID sql.NullString
	var gracePeriodExpiresAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, player_id, season_id, status, source, destination, food, grace_period_expires_at
FROM jumps
WHERE id = $1`, jumpID).Scan(
		&snap.ID,
		&snap.GroupID,
		&snap.PlayerID,
		&seasonID,
		&snap.Status,
		&snap.Source,
		&snap.Destination,
		&snap.Food,
		&gracePeriodExpiresAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return game.JumpSnapshot{}, false, nil
	}
	if err != nil {
		return game.JumpSnapshot{}, false, err
	}
	if !visiblePerformedStatus(snap.Status) {
		return game.JumpSnapshot{}, false, nil
	}
	if seasonID.Valid {
		snap.SeasonID = &seasonID.String
	}
	if gracePeriodExpiresAt.Valid {
		snap.GracePeriodExpiresAt = gracePeriodExpiresAt.Time
	}
	return snap, true, nil
}

func (s *PostgresStore) Season(ctx context.Context, seasonID string) (game.SeasonSnapshot, error) {
	var snap game.SeasonSnapshot
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline
FROM seasons WHERE id = $1`, seasonID).Scan(
		&snap.ID, &snap.GroupID, &snap.CommissionerPlayerID, &snap.Status, &snap.SubmissionDeadline, &snap.JudgingDeadline)
	if errors.Is(err, sql.ErrNoRows) {
		return game.SeasonSnapshot{}, nil
	}
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	return snap, nil
}

func (s *PostgresStore) GroupMembership(ctx context.Context, playerID, groupID string) (game.MembershipSnapshot, bool, error) {
	var role string
	err := s.db.QueryRowContext(ctx, `
SELECT role FROM group_memberships
WHERE player_id = $1 AND group_id = $2`, playerID, groupID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return game.MembershipSnapshot{}, false, nil
	}
	if err != nil {
		return game.MembershipSnapshot{}, false, err
	}
	return game.MembershipSnapshot{Role: role}, true, nil
}

func (s *PostgresStore) UpsertJudgment(ctx context.Context, jumpID, playerID string, difficulty, transgression, creativity, presentation int) (game.Judgment, bool, error) {
	judgmentID := stableID("judgment", jumpID+":"+playerID)
	var created bool
	err := s.db.QueryRowContext(ctx, `
WITH upsert AS (
  INSERT INTO judgments (id, jump_id, player_id, difficulty, transgression, creativity, presentation)
  VALUES ($1, $2, $3, $4, $5, $6, $7)
  ON CONFLICT (jump_id, player_id) DO UPDATE SET
    difficulty = EXCLUDED.difficulty,
    transgression = EXCLUDED.transgression,
    creativity = EXCLUDED.creativity,
    presentation = EXCLUDED.presentation
  RETURNING (xmax = 0) AS created
)
SELECT created FROM upsert`, judgmentID, jumpID, playerID, difficulty, transgression, creativity, presentation).Scan(&created)
	if err != nil {
		return game.Judgment{}, false, err
	}
	return game.Judgment{
		ID:            judgmentID,
		JumpID:        jumpID,
		PlayerID:      playerID,
		Difficulty:    difficulty,
		Transgression: transgression,
		Creativity:    creativity,
		Presentation:  presentation,
	}, created, nil
}
