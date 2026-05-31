package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.EvidenceRepository adapter methods for PostgresStore

func (s *PostgresStore) PlannedJump(ctx context.Context, jumpID string) (game.JumpSnapshot, bool, error) {
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
	if snap.Status != "Planned Jump" {
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

func (s *PostgresStore) CreateAuthorization(ctx context.Context, jumpID, playerID, contentType string) (game.AuthorizationSnapshot, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return game.AuthorizationSnapshot{}, err
	}
	defer tx.Rollback()

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
	if _, err := tx.ExecContext(ctx, `
INSERT INTO evidence_upload_authorizations (id, jump_id, player_id, content_type, media_object_key, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)`,
		id, jumpID, playerID, contentType, mediaObjectKey, expiresAt,
	); err != nil {
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

	var authMediaObjectKey string
	var authExpiresAt time.Time
	var foundPlayerID string
	err = tx.QueryRowContext(ctx, `
SELECT id, player_id, media_object_key, expires_at
FROM evidence_upload_authorizations
WHERE id = $1 AND jump_id = $2 FOR UPDATE`, authorizationID, jumpID).Scan(
		&authorizationID, &foundPlayerID, &authMediaObjectKey, &authExpiresAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return game.EvidenceCreateResult{}, game.ErrEvidenceUploadAuthorizationNotFound
	}
	if err != nil {
		return game.EvidenceCreateResult{}, err
	}
	if foundPlayerID != playerID || time.Now().After(authExpiresAt) {
		return game.EvidenceCreateResult{}, game.ErrEvidenceUploadAuthorizationNotFound
	}

	evidenceID := stableID("evidence", jumpID+":"+authorizationID)
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `
INSERT INTO evidences (id, jump_id, player_id, upload_authorization_id, caption, media_object_key, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		evidenceID, jumpID, playerID, authorizationID, caption, authMediaObjectKey, now,
	); err != nil {
		return game.EvidenceCreateResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE jumps
SET status = 'Performed Jump'
WHERE id = $1 AND status = 'Planned Jump'`, jumpID,
	); err != nil {
		return game.EvidenceCreateResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM evidence_upload_authorizations WHERE id = $1`, authorizationID); err != nil {
		return game.EvidenceCreateResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return game.EvidenceCreateResult{}, err
	}
	return game.EvidenceCreateResult{
		EvidenceID:     evidenceID,
		MediaObjectKey: authMediaObjectKey,
	}, nil
}
