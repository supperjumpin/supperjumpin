package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.DisputeRepository adapter methods for PostgresStore

func (s *PostgresStore) JumpByID(ctx context.Context, jumpID string) (game.JumpSnapshot, bool, error) {
	var snap game.JumpSnapshot
	var seasonID sql.NullString
	var gracePeriodExpiresAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, player_id, season_id, status, source, destination, food, final_score, grace_period_expires_at
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
		&snap.FinalScore,
		&gracePeriodExpiresAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return game.JumpSnapshot{}, false, nil
	}
	if err != nil {
		return game.JumpSnapshot{}, false, err
	}
	if seasonID.Valid {
		snap.SeasonID = &seasonID.String
	}
	if gracePeriodExpiresAt.Valid {
		snap.GracePeriodExpiresAt = gracePeriodExpiresAt.Time
	}
	return snap, true, nil
}

func (s *PostgresStore) InsertDispute(ctx context.Context, jumpID, raisedByPlayerID, concern, details string) (game.DisputeSnapshot, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return game.DisputeSnapshot{}, err
	}
	defer tx.Rollback()

	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM disputes WHERE jump_id = $1`, jumpID).Scan(&count); err != nil {
		return game.DisputeSnapshot{}, err
	}
	dispute := Dispute{
		ID:               stableID("dispute", jumpID+":"+raisedByPlayerID+":"+strconv.Itoa(count+1)),
		JumpID:           jumpID,
		RaisedByPlayerID: raisedByPlayerID,
		Concern:          concern,
		Details:          details,
		Status:           "Open",
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO disputes (id, jump_id, raised_by_player_id, concern, details, status)
VALUES ($1, $2, $3, $4, $5, $6)`, dispute.ID, dispute.JumpID, dispute.RaisedByPlayerID, dispute.Concern, dispute.Details, dispute.Status); err != nil {
		return game.DisputeSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return game.DisputeSnapshot{}, err
	}
	return disputeToSnapshot(dispute), nil
}

func (s *PostgresStore) Dispute(ctx context.Context, disputeID string) (game.DisputeSnapshot, error) {
	dispute, err := disputeInDB(ctx, s.db, disputeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return game.DisputeSnapshot{}, nil
		}
		return game.DisputeSnapshot{}, err
	}
	return disputeToSnapshot(dispute), nil
}

func (s *PostgresStore) UpdateDisputeResolution(ctx context.Context, disputeID, resolution, resolutionReason, resolvedByPlayerID string) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE disputes
SET status = 'Resolved', resolution = $2, resolution_reason = $3, resolved_by_player_id = $4
WHERE id = $1`, disputeID, resolution, resolutionReason, resolvedByPlayerID)
	return err
}

func (s *PostgresStore) UpdateDisputeOverride(ctx context.Context, disputeID, overrideResolution, overrideReason, overrideByPlayerID string) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE disputes
SET status = 'Overridden', override_resolution = $2, override_reason = $3, override_by_player_id = $4
WHERE id = $1`, disputeID, overrideResolution, overrideReason, overrideByPlayerID)
	return err
}

func (s *PostgresStore) UpdateJumpStatusAfterDispute(ctx context.Context, jumpID, status string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE jumps SET status = $2, final_score = NULL WHERE id = $1`, jumpID, status)
	return err
}
