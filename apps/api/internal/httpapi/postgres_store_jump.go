package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.JumpRepository adapter methods for PostgresStore

func (s *PostgresStore) InsertIdea(ctx context.Context, groupID, playerID, source, destination, food string) (game.JumpSnapshot, error) {
	for attempts := 0; attempts < 3; attempts++ {
		id, err := randomToken("jump")
		if err != nil {
			return game.JumpSnapshot{}, err
		}
		result, err := s.db.ExecContext(ctx, `
INSERT INTO jumps (id, group_id, player_id, status, source, destination, food)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT DO NOTHING`, id, groupID, playerID, "Idea", source, destination, food)
		if err != nil {
			return game.JumpSnapshot{}, err
		}
		rows, err := result.RowsAffected()
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
	if seasonID.Valid {
		snap.SeasonID = &seasonID.String
	}
	if gracePeriodExpiresAt.Valid {
		snap.GracePeriodExpiresAt = gracePeriodExpiresAt.Time
	}
	return snap, true, nil
}

func (s *PostgresStore) ActiveSeasonForGroup(ctx context.Context, groupID string) (game.SeasonSnapshot, error) {
	season, err := s.activeSeasonForGroup(ctx, groupID)
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	if season == nil {
		return game.SeasonSnapshot{}, nil
	}
	return game.SeasonSnapshot{ID: season.ID, Status: "Active"}, nil
}

func (s *PostgresStore) UpdateJumpToPlanned(ctx context.Context, jumpID, playerID string, seasonID *string) (game.JumpSnapshot, error) {
	var snap game.JumpSnapshot
	var resultSeasonID sql.NullString
	var gracePeriodExpiresAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `
UPDATE jumps
SET status = 'Planned Jump', season_id = $2
WHERE id = $1
  AND status = 'Idea'
RETURNING id, group_id, player_id, season_id, status, source, destination, food, grace_period_expires_at`, jumpID, seasonID).Scan(
		&snap.ID,
		&snap.GroupID,
		&snap.PlayerID,
		&resultSeasonID,
		&snap.Status,
		&snap.Source,
		&snap.Destination,
		&snap.Food,
		&gracePeriodExpiresAt,
	)
	if err != nil {
		return game.JumpSnapshot{}, err
	}
	if resultSeasonID.Valid {
		snap.SeasonID = &resultSeasonID.String
	}
	if gracePeriodExpiresAt.Valid {
		snap.GracePeriodExpiresAt = gracePeriodExpiresAt.Time
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
	if _, err := tx.ExecContext(ctx, `
INSERT INTO evidences (id, jump_id, player_id, caption, media_object_key, created_at)
VALUES ($1, $2, $3, $4, $5, $6)`,
		evidenceID, jumpID, params.PlayerID, params.Caption, params.MediaObjectKey, now,
	); err != nil {
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
