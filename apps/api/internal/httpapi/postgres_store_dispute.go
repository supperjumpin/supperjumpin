package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/db"
	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.DisputeRepository adapter methods for PostgresStore

func (s *PostgresStore) JumpByID(ctx context.Context, jumpID string) (game.JumpSnapshot, bool, error) {
	jump, err := s.queries.GetJump(ctx, jumpID)
	if errors.Is(err, sql.ErrNoRows) {
		return game.JumpSnapshot{}, false, nil
	}
	if err != nil {
		return game.JumpSnapshot{}, false, err
	}
	snap := game.JumpSnapshot{
		ID:     jump.ID,
		Status: jump.Status,
		Source: jump.Source,
		Destination: jump.Destination,
		Food:   jump.Food,
	}
	if jump.GroupID.Valid {
		snap.GroupID = jump.GroupID.String
	}
	snap.PlayerID = jump.PlayerID
	if jump.SeasonID.Valid {
		snap.SeasonID = &jump.SeasonID.String
	}
	if jump.FinalScore.Valid {
		score := int(jump.FinalScore.Int32)
		snap.FinalScore = &score
	}
	if jump.GracePeriodExpiresAt.Valid {
		snap.GracePeriodExpiresAt = jump.GracePeriodExpiresAt.Time
	}
	return snap, true, nil
}

func (s *PostgresStore) InsertDispute(ctx context.Context, jumpID, raisedByPlayerID, concern, details string) (game.DisputeSnapshot, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return game.DisputeSnapshot{}, err
	}
	defer tx.Rollback()
	qtx := s.queries.WithTx(tx)

	count, err := qtx.CountDisputesForJump(ctx, jumpID)
	if err != nil {
		return game.DisputeSnapshot{}, err
	}
	dispute := Dispute{
		ID:               stableID("dispute", jumpID+":"+raisedByPlayerID+":"+strconv.FormatInt(count+1, 10)),
		JumpID:           jumpID,
		RaisedByPlayerID: raisedByPlayerID,
		Concern:          concern,
		Details:          details,
		Status:           "Open",
	}
	if err := qtx.InsertDispute(ctx, db.InsertDisputeParams{
		ID:               dispute.ID,
		JumpID:           dispute.JumpID,
		RaisedByPlayerID: dispute.RaisedByPlayerID,
		Concern:          dispute.Concern,
		Details:          dispute.Details,
		Status:           dispute.Status,
	}); err != nil {
		return game.DisputeSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return game.DisputeSnapshot{}, err
	}
	return disputeToSnapshot(dispute), nil
}

func (s *PostgresStore) Dispute(ctx context.Context, disputeID string) (game.DisputeSnapshot, error) {
	row, err := s.queries.GetDispute(ctx, disputeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return game.DisputeSnapshot{}, nil
		}
		return game.DisputeSnapshot{}, err
	}
	snap := game.DisputeSnapshot{
		ID:               row.ID,
		JumpID:           row.JumpID,
		RaisedByPlayerID: row.RaisedByPlayerID,
		Concern:          row.Concern,
		Details:          row.Details,
		Status:           row.Status,
	}
	if row.Resolution.Valid {
		resolution := row.Resolution.String
		snap.Resolution = &resolution
	}
	if row.ResolutionReason.Valid {
		resolutionReason := row.ResolutionReason.String
		snap.ResolutionReason = &resolutionReason
	}
	if row.ResolvedByPlayerID.Valid {
		resolvedByPlayerID := row.ResolvedByPlayerID.String
		snap.ResolvedByPlayerID = &resolvedByPlayerID
	}
	if row.OverrideResolution.Valid {
		overrideResolution := row.OverrideResolution.String
		snap.OverrideResolution = &overrideResolution
	}
	if row.OverrideReason.Valid {
		overrideReason := row.OverrideReason.String
		snap.OverrideReason = &overrideReason
	}
	if row.OverrideByPlayerID.Valid {
		overrideByPlayerID := row.OverrideByPlayerID.String
		snap.OverrideByPlayerID = &overrideByPlayerID
	}
	return snap, nil
}

func (s *PostgresStore) UpdateDisputeResolution(ctx context.Context, disputeID, resolution, resolutionReason, resolvedByPlayerID string) error {
	return s.queries.UpdateDisputeResolution(ctx, db.UpdateDisputeResolutionParams{
		ID:                 disputeID,
		Resolution:         sql.NullString{String: resolution, Valid: true},
		ResolutionReason:   sql.NullString{String: resolutionReason, Valid: true},
		ResolvedByPlayerID: sql.NullString{String: resolvedByPlayerID, Valid: true},
	})
}

func (s *PostgresStore) UpdateDisputeOverride(ctx context.Context, disputeID, overrideResolution, overrideReason, overrideByPlayerID string) error {
	return s.queries.UpdateDisputeOverride(ctx, db.UpdateDisputeOverrideParams{
		ID:                 disputeID,
		OverrideResolution: sql.NullString{String: overrideResolution, Valid: true},
		OverrideReason:     sql.NullString{String: overrideReason, Valid: true},
		OverrideByPlayerID: sql.NullString{String: overrideByPlayerID, Valid: true},
	})
}

func (s *PostgresStore) UpdateJumpStatusAfterDispute(ctx context.Context, jumpID, status string) error {
	return s.queries.UpdateJumpStatusAfterDispute(ctx, db.UpdateJumpStatusAfterDisputeParams{
		ID:     jumpID,
		Status: status,
	})
}
