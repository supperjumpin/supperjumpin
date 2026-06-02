package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/db"
	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.JumpRepository adapter methods for PostgresStore

func (s *PostgresStore) InsertIdea(ctx context.Context, groupID, playerID, source, destination, food string) (game.JumpSnapshot, error) {
	for attempts := 0; attempts < 3; attempts++ {
		id, err := randomToken("jump")
		if err != nil {
			return game.JumpSnapshot{}, err
		}
		rows, err := s.queries.InsertIdea(ctx, db.InsertIdeaParams{
			ID:          id,
			GroupID:     sql.NullString{String: groupID, Valid: true},
			PlayerID:    playerID,
			Status:      "Idea",
			Source:      source,
			Destination: destination,
			Food:        food,
		})
		if err != nil {
			return game.JumpSnapshot{}, err
		}
		if rows == 1 {
			return game.JumpSnapshot{
				ID:          id,
				GroupID:     groupID,
				PlayerID:    playerID,
				Status:      "Idea",
				Source:      source,
				Destination: destination,
				Food:        food,
			}, nil
		}
		if attempts == 2 {
			return game.JumpSnapshot{}, fmt.Errorf("create unique Idea after retries")
		}
	}
	return game.JumpSnapshot{}, fmt.Errorf("create unique Idea: unreachable")
}

func (s *PostgresStore) Idea(ctx context.Context, jumpID string) (game.JumpSnapshot, bool, error) {
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
	if row.SeasonID.Valid {
		snap.SeasonID = &row.SeasonID.String
	}
	if row.GracePeriodExpiresAt.Valid {
		snap.GracePeriodExpiresAt = row.GracePeriodExpiresAt.Time
	}
	return snap, true, nil
}

func (s *PostgresStore) ActiveSeasonForGroup(ctx context.Context, groupID string) (game.SeasonSnapshot, error) {
	season, err := s.queries.GetActiveSeasonForGroup(ctx, groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return game.SeasonSnapshot{}, nil
	}
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	return game.SeasonSnapshot{ID: season.ID, Status: season.Status}, nil
}

func (s *PostgresStore) UpdateJumpToPlanned(ctx context.Context, jumpID, playerID string, seasonID *string) (game.JumpSnapshot, error) {
	_ = playerID
	seasonParam := sql.NullString{}
	if seasonID != nil {
		seasonParam = sql.NullString{String: *seasonID, Valid: true}
	}
	row, err := s.queries.UpdateJumpToPlanned(ctx, db.UpdateJumpToPlannedParams{
		ID:       jumpID,
		SeasonID: seasonParam,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return game.JumpSnapshot{}, game.ErrJumpNotFound
	}
	if err != nil {
		return game.JumpSnapshot{}, err
	}
	row, err := s.queries.UpdateJumpToPlanned(ctx, db.UpdateJumpToPlannedParams{
		ID:       jumpID,
		SeasonID: seasonParam,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return game.JumpSnapshot{}, game.ErrJumpNotFound
	}
	if err != nil {
		return game.JumpSnapshot{}, err
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
	if row.SeasonID.Valid {
		snap.SeasonID = &row.SeasonID.String
	}
	if row.GracePeriodExpiresAt.Valid {
		snap.GracePeriodExpiresAt = row.GracePeriodExpiresAt.Time
	}
	return snap, nil
}

func (s *PostgresStore) InsertPerformedJump(ctx context.Context, params game.InsertPerformedJumpParams) (game.JumpSnapshot, game.EvidenceSnapshot, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return game.JumpSnapshot{}, game.EvidenceSnapshot{}, err
	}
	defer tx.Rollback()

	jumpID, err := randomToken("jump")
	if err != nil {
		return game.JumpSnapshot{}, game.EvidenceSnapshot{}, err
	}

	var groupID any
	if params.GroupID != "" {
		groupID = params.GroupID
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO jumps (id, group_id, player_id, season_id, status, source, destination, food, grace_period_expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		jumpID, groupID, params.PlayerID, params.SeasonID, "Performed Jump",
		params.Source, params.Destination, params.Food, params.GracePeriodExpiresAt,
	); err != nil {
		return game.JumpSnapshot{}, game.EvidenceSnapshot{}, err
	}

	evidenceID := stableID("evidence", jumpID+":"+params.MediaObjectKey)
	now := time.Now().UTC()
	qtx := s.queries.WithTx(tx)
	if err := qtx.InsertEvidence(ctx, db.InsertEvidenceParams{
		ID:                    evidenceID,
		JumpID:                jumpID,
		PlayerID:              params.PlayerID,
		UploadAuthorizationID: sql.NullString{},
		Caption:               params.Caption,
		MediaObjectKey:        params.MediaObjectKey,
		CreatedAt:             now,
	}); err != nil {
		return game.JumpSnapshot{}, game.EvidenceSnapshot{}, err
	}

	if err := tx.Commit(); err != nil {
		return game.JumpSnapshot{}, game.EvidenceSnapshot{}, err
	}

	jumpSnap := game.JumpSnapshot{
		ID:                  jumpID,
		GroupID:             params.GroupID,
		PlayerID:            params.PlayerID,
		Status:              "Performed Jump",
		SeasonID:            params.SeasonID,
		Source:              params.Source,
		Destination:         params.Destination,
		Food:                params.Food,
		GracePeriodExpiresAt: params.GracePeriodExpiresAt,
	}
	evidenceSnap := game.EvidenceSnapshot{
		ID:             evidenceID,
		JumpID:        jumpID,
		PlayerID:       params.PlayerID,
		MediaObjectKey: params.MediaObjectKey,
		Caption:        params.Caption,
		CreatedAt:      now,
	}
	return jumpSnap, evidenceSnap, nil
}
