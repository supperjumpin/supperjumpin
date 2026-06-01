package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/db"
	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.SeasonRepository adapter methods for PostgresStore

func (s *PostgresStore) OpenSeasonForGroup(ctx context.Context, groupID string) (game.SeasonSnapshot, error) {
	season, err := s.queries.GetOpenSeasonForGroup(ctx, groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return game.SeasonSnapshot{}, nil
	}
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	return game.SeasonSnapshot{
		ID:                   season.ID,
		GroupID:              season.GroupID,
		CommissionerPlayerID: season.CommissionerPlayerID,
		Status:               season.Status,
		SubmissionDeadline:   season.SubmissionDeadline,
		JudgingDeadline:      season.JudgingDeadline,
	}, nil
}

func (s *PostgresStore) InsertSeason(ctx context.Context, groupID, commissionerPlayerID string, submissionDeadline, judgingDeadline time.Time) (game.SeasonSnapshot, error) {
	count, err := s.queries.CountSeasons(ctx)
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	season := Season{
		ID:                   stableID("season", groupID+":"+strconv.FormatInt(count+1, 10)),
		GroupID:              groupID,
		CommissionerPlayerID: commissionerPlayerID,
		Status:               "Active",
		SubmissionDeadline:   submissionDeadline.UTC(),
		JudgingDeadline:      judgingDeadline.UTC(),
	}
	if err := s.queries.InsertSeason(ctx, db.InsertSeasonParams{
		ID:                   season.ID,
		GroupID:              season.GroupID,
		CommissionerPlayerID: season.CommissionerPlayerID,
		Status:               season.Status,
		SubmissionDeadline:   season.SubmissionDeadline,
		JudgingDeadline:      season.JudgingDeadline,
	}); err != nil {
		return game.SeasonSnapshot{}, err
	}
	return game.SeasonSnapshot{
		ID:                   season.ID,
		GroupID:              season.GroupID,
		CommissionerPlayerID: season.CommissionerPlayerID,
		Status:               season.Status,
		SubmissionDeadline:   season.SubmissionDeadline,
		JudgingDeadline:      season.JudgingDeadline,
	}, nil
}

func (s *PostgresStore) UpdateSeasonStatus(ctx context.Context, seasonID, action, actorPlayerID, actorRole string, override bool, fromStatus, toStatus string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)
	historyID, err := randomToken("season_history")
	if err != nil {
		return err
	}

	if err := qtx.UpdateSeasonStatus(ctx, db.UpdateSeasonStatusParams{ID: seasonID, Status: toStatus}); err != nil {
		return err
	}
	if err := qtx.InsertSeasonHistoryEntry(ctx, db.InsertSeasonHistoryEntryParams{
		ID:            historyID,
		SeasonID:      seasonID,
		Action:        action,
		ActorPlayerID: actorPlayerID,
		ActorRole:     actorRole,
		Override:      override,
		FromStatus:    fromStatus,
		ToStatus:      toStatus,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *PostgresStore) JumpsForSeason(ctx context.Context, seasonID string) ([]game.JumpSnapshot, error) {
	rows, err := s.queries.ListJumpsForSeason(ctx, sql.NullString{String: seasonID, Valid: true})
	if err != nil {
		return nil, err
	}

	var result []game.JumpSnapshot
	for _, row := range rows {
		groupID := ""
		if row.GroupID.Valid {
			groupID = row.GroupID.String
		}
		var seasonID *string
		if row.SeasonID.Valid {
			seasonID = &row.SeasonID.String
		}
		var finalScore *int
		if row.FinalScore.Valid {
			score := int(row.FinalScore.Int32)
			finalScore = &score
		}
		var gracePeriodExpiresAt time.Time
		if row.GracePeriodExpiresAt.Valid {
			gracePeriodExpiresAt = row.GracePeriodExpiresAt.Time
		}
		result = append(result, game.JumpSnapshot{
			ID:                   row.ID,
			GroupID:              groupID,
			PlayerID:             row.PlayerID,
			SeasonID:             seasonID,
			Status:               row.Status,
			Source:               row.Source,
			Destination:          row.Destination,
			Food:                 row.Food,
			FinalScore:           finalScore,
			GracePeriodExpiresAt: gracePeriodExpiresAt,
		})
	}
	return result, nil
}

func (s *PostgresStore) JudgmentsForJump(ctx context.Context, jumpID string) ([]game.Judgment, error) {
	rows, err := s.queries.ListJudgmentsForJump(ctx, jumpID)
	if err != nil {
		return nil, err
	}

	var result []game.Judgment
	for _, row := range rows {
		result = append(result, game.Judgment{
			ID:           row.ID,
			JumpID:       row.JumpID,
			PlayerID:     row.PlayerID,
			Commitment:   int(row.Commitment),
			Transgression: int(row.Transgression),
			Creativity:   int(row.Creativity),
			Presentation: int(row.Presentation),
		})
	}
	return result, nil
}

func (s *PostgresStore) UpdateJumpFinalization(ctx context.Context, jumpID string, status string, finalScore *int) error {
	finalScoreParam := sql.NullInt32{}
	if finalScore != nil {
		finalScoreParam = sql.NullInt32{Int32: int32(*finalScore), Valid: true}
	}
	return s.queries.UpdateJumpFinalization(ctx, db.UpdateJumpFinalizationParams{ID: jumpID, Status: status, FinalScore: finalScoreParam})
}

func (s *PostgresStore) LatestSeasonForGroup(ctx context.Context, groupID string) (game.SeasonSnapshot, error) {
	season, err := s.queries.GetLatestSeasonForGroup(ctx, groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return game.SeasonSnapshot{}, nil
	}
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	return game.SeasonSnapshot{
		ID:                   season.ID,
		GroupID:              season.GroupID,
		CommissionerPlayerID: season.CommissionerPlayerID,
		Status:               season.Status,
		SubmissionDeadline:   season.SubmissionDeadline,
		JudgingDeadline:      season.JudgingDeadline,
	}, nil
}

func (s *PostgresStore) GroupPlayers(ctx context.Context, groupID string) ([]game.PlayerSnapshot, error) {
	rows, err := s.queries.GetGroupPlayers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var result []game.PlayerSnapshot
	for _, row := range rows {
		result = append(result, game.PlayerSnapshot{ID: row.ID, DisplayName: row.DisplayName})
	}
	return result, nil
}

func (s *PostgresStore) SeasonHistoryEntries(ctx context.Context, seasonID string) ([]game.SeasonHistoryEntry, error) {
	rows, err := s.queries.ListSeasonHistoryEntries(ctx, seasonID)
	if err != nil {
		return nil, err
	}

	entries := []game.SeasonHistoryEntry{}
	for _, row := range rows {
		entries = append(entries, game.SeasonHistoryEntry{
			ID:            row.ID,
			SeasonID:      row.SeasonID,
			Action:        row.Action,
			ActorPlayerID: row.ActorPlayerID,
			ActorRole:     row.ActorRole,
			Override:      row.Override,
			FromStatus:    row.FromStatus,
			ToStatus:      row.ToStatus,
		})
	}
	return entries, nil
}
