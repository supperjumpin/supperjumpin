package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/db"
	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.EvidenceRepository adapter methods for PostgresStore

func (s *PostgresStore) PlannedJump(ctx context.Context, jumpID string) (game.JumpSnapshot, bool, error) {
	jump, err := s.queries.GetJump(ctx, jumpID)
	if errors.Is(err, sql.ErrNoRows) {
		return game.JumpSnapshot{}, false, nil
	}
	if err != nil {
		return game.JumpSnapshot{}, false, err
	}
	snap := game.JumpSnapshot{
		ID:          jump.ID,
		GroupID:     jump.GroupID.String,
		PlayerID:    jump.PlayerID,
		Status:      jump.Status,
		Source:      jump.Source,
		Destination: jump.Destination,
		Food:        jump.Food,
	}
	if snap.Status != "Planned Jump" {
		return game.JumpSnapshot{}, false, nil
	}
	if jump.SeasonID.Valid {
		snap.SeasonID = &jump.SeasonID.String
	}
	if jump.GracePeriodExpiresAt.Valid {
		snap.GracePeriodExpiresAt = jump.GracePeriodExpiresAt.Time
	}
	return snap, true, nil
}

func (s *PostgresStore) CreateAuthorization(ctx context.Context, jumpID, playerID, contentType string) (game.AuthorizationSnapshot, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return game.AuthorizationSnapshot{}, err
	}
	defer tx.Rollback()
	qtx := s.queries.WithTx(tx)

	now := time.Now()
	id, err := randomToken("evidence_upload")
	if err != nil {
		return game.AuthorizationSnapshot{}, err
	}
	mediaObjectKey, err := randomToken("evidence_object")
	if err != nil {
		return game.AuthorizationSnapshot{}, err
	}
	expiresAt := now.Add(15 * time.Minute).UTC()
	if _, err := qtx.InsertEvidenceUploadAuthorization(ctx, db.InsertEvidenceUploadAuthorizationParams{
		ID:             id,
		JumpID:         jumpID,
		PlayerID:       playerID,
		ContentType:    contentType,
		MediaObjectKey: mediaObjectKey,
		ExpiresAt:      expiresAt,
	}); err != nil {
		return game.AuthorizationSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return game.AuthorizationSnapshot{}, err
	}
	return game.AuthorizationSnapshot{
		ID:             id,
		JumpID:         jumpID,
		MediaObjectKey: mediaObjectKey,
		ExpiresAt:      expiresAt,
	}, nil
}

func (s *PostgresStore) ClaimAndAdvance(ctx context.Context, authorizationID, jumpID, playerID, caption string) (game.EvidenceCreateResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return game.EvidenceCreateResult{}, err
	}
	defer tx.Rollback()
	qtx := s.queries.WithTx(tx)

	auth, err := qtx.GetEvidenceUploadAuthorizationForUpdate(ctx, db.GetEvidenceUploadAuthorizationForUpdateParams{
		ID:     authorizationID,
		JumpID: jumpID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return game.EvidenceCreateResult{}, game.ErrEvidenceUploadAuthorizationNotFound
	}
	if err != nil {
		return game.EvidenceCreateResult{}, err
	}
	if auth.PlayerID != playerID || time.Now().After(auth.ExpiresAt) {
		return game.EvidenceCreateResult{}, game.ErrEvidenceUploadAuthorizationNotFound
	}

	evidenceID := stableID("evidence", jumpID+":"+authorizationID)
	now := time.Now().UTC()
	if err := qtx.InsertEvidence(ctx, db.InsertEvidenceParams{
		ID:                    evidenceID,
		JumpID:                jumpID,
		PlayerID:              playerID,
		UploadAuthorizationID: sql.NullString{String: authorizationID, Valid: true},
		Caption:               caption,
		MediaObjectKey:        auth.MediaObjectKey,
		CreatedAt:             now,
	}); err != nil {
		return game.EvidenceCreateResult{}, err
	}
	if err := qtx.AdoptJumpToSeason(ctx, jumpID); err != nil {
		return game.EvidenceCreateResult{}, err
	}
	if err := qtx.DeleteEvidenceUploadAuthorization(ctx, authorizationID); err != nil {
		return game.EvidenceCreateResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return game.EvidenceCreateResult{}, err
	}
	return game.EvidenceCreateResult{
		EvidenceID:     evidenceID,
		MediaObjectKey: auth.MediaObjectKey,
	}, nil
}
