package httpapi

import (
	"context"
	"database/sql"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/db"
	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.OpenRepository adapter methods for PostgresStore

func (s *PostgresStore) JumpsForOpenMonth(ctx context.Context, year, month int) ([]game.JumpSnapshot, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	rows, err := s.queries.ListJumpsForOpenMonth(ctx, db.ListJumpsForOpenMonthParams{
		CreatedAt:   start,
		CreatedAt_2: end,
	})
	if err != nil {
		return nil, err
	}

	var result []game.JumpSnapshot
	for _, row := range rows {
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

func (s *PostgresStore) UpdateJumpOpenFinalScore(ctx context.Context, jumpID string, score *int) error {
	param := sql.NullInt32{}
	if score != nil {
		param = sql.NullInt32{Int32: int32(*score), Valid: true}
	}
	return s.queries.UpdateJumpOpenFinalScore(ctx, db.UpdateJumpOpenFinalScoreParams{
		ID:             jumpID,
		OpenFinalScore: param,
	})
}

func (s *PostgresStore) JudgmentsForJump(ctx context.Context, jumpID string) ([]game.Judgment, error) {
	rows, err := s.queries.ListJudgmentsForJump(ctx, jumpID)
	if err != nil {
		return nil, err
	}
	var result []game.Judgment
	for _, row := range rows {
		result = append(result, game.Judgment{
			ID:             row.ID,
			JumpID:         row.JumpID,
			PlayerID:       row.PlayerID.String,
			GuestSessionID: row.GuestSessionID.String,
			Provenance:     row.Provenance,
			Commitment:     int(row.Commitment),
			Transgression:  int(row.Transgression),
			Creativity:     int(row.Creativity),
			Presentation:   int(row.Presentation),
		})
	}
	return result, nil
}

func (s *PostgresStore) PlayersForOpenMonth(ctx context.Context, year, month int) ([]game.PlayerSnapshot, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	rows, err := s.queries.ListPlayersForOpenMonth(ctx, db.ListPlayersForOpenMonthParams{
		CreatedAt:   start,
		CreatedAt_2: end,
	})
	if err != nil {
		return nil, err
	}

	var result []game.PlayerSnapshot
	for _, row := range rows {
		result = append(result, game.PlayerSnapshot{ID: row.ID, DisplayName: row.DisplayName})
	}
	return result, nil
}

func (s *PostgresStore) UpsertOpenStanding(ctx context.Context, year, month int, playerID string, score, judgedJumps int) error {
	id, err := randomToken("open_standing")
	if err != nil {
		return err
	}
	_, err = s.queries.UpsertOpenStanding(ctx, db.UpsertOpenStandingParams{
		ID:          id,
		Year:        int32(year),
		Month:       int32(month),
		PlayerID:    playerID,
		Score:       int32(score),
		JudgedJumps: int32(judgedJumps),
	})
	return err
}
