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

func (s *PostgresStore) Season(_ context.Context, _ string) (game.SeasonSnapshot, error) {
	// Seasons table removed; no active seasons exist. Return empty snapshot.
	return game.SeasonSnapshot{}, nil
}

func (s *PostgresStore) SubmitAcceptedJudgment(ctx context.Context, input game.JudgmentInput) (game.Judgment, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return game.Judgment{}, false, err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)
	judgment, created, err := s.upsertJudgmentWithTx(ctx, qtx, input)
	if err != nil {
		return game.Judgment{}, false, err
	}

	if created && input.GuestSessionID != "" {
		if err := s.incrementGuestJudgmentCountIfAllowed(ctx, tx, input.GuestSessionID); err != nil {
			return game.Judgment{}, false, err
		}
	}

	if created {
		if err := qtx.AdvanceJumpToJudged(ctx, input.JumpID); err != nil {
			return game.Judgment{}, false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return game.Judgment{}, false, err
	}

	return judgment, created, nil
}

func (s *PostgresStore) upsertJudgmentWithTx(ctx context.Context, qtx *db.Queries, input game.JudgmentInput) (game.Judgment, bool, error) {
	if input.JudgePlayerID != "" {
		judgmentID := stableID("judgment", input.JumpID+":"+input.JudgePlayerID)
		created, err := qtx.UpsertPlayerJudgment(ctx, db.UpsertPlayerJudgmentParams{
			ID:            judgmentID,
			JumpID:        input.JumpID,
			PlayerID:      sql.NullString{String: input.JudgePlayerID, Valid: true},
			Provenance:    input.Provenance,
			Commitment:    int32(input.Commitment),
			Transgression: int32(input.Transgression),
			Creativity:    int32(input.Creativity),
			Presentation:  int32(input.Presentation),
		})
		if err != nil {
			return game.Judgment{}, false, err
		}
		return game.Judgment{
			ID:            judgmentID,
			JumpID:        input.JumpID,
			PlayerID:      input.JudgePlayerID,
			Provenance:    input.Provenance,
			Commitment:    input.Commitment,
			Transgression: input.Transgression,
			Creativity:    input.Creativity,
			Presentation:  input.Presentation,
		}, created, nil
	}

	judgmentID := stableID("judgment", input.JumpID+":guest:"+input.GuestSessionID)
	created, err := qtx.UpsertGuestJudgment(ctx, db.UpsertGuestJudgmentParams{
		ID:             judgmentID,
		JumpID:         input.JumpID,
		GuestSessionID: sql.NullString{String: input.GuestSessionID, Valid: true},
		Provenance:     input.Provenance,
		Commitment:     int32(input.Commitment),
		Transgression:  int32(input.Transgression),
		Creativity:     int32(input.Creativity),
		Presentation:   int32(input.Presentation),
	})
	if err != nil {
		return game.Judgment{}, false, err
	}
	return game.Judgment{
		ID:             judgmentID,
		JumpID:         input.JumpID,
		GuestSessionID: input.GuestSessionID,
		Provenance:     input.Provenance,
		Commitment:     input.Commitment,
		Transgression:  input.Transgression,
		Creativity:     input.Creativity,
		Presentation:   input.Presentation,
	}, created, nil
}

func (s *PostgresStore) incrementGuestJudgmentCountIfAllowed(ctx context.Context, tx *sql.Tx, guestSessionID string) error {
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT judgment_count FROM guest_sessions WHERE id = $1 FOR UPDATE`, guestSessionID).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return game.ErrGuestCapReached
		}
		return err
	}
	if count >= 5 {
		return game.ErrGuestCapReached
	}
	if _, err := tx.ExecContext(ctx, `UPDATE guest_sessions SET judgment_count = judgment_count + 1 WHERE id = $1`, guestSessionID); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) UpsertJudgment(ctx context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (game.Judgment, bool, error) {
	var judgmentID, identityKey string
	if playerID != "" {
		judgmentID = stableID("judgment", jumpID+":"+playerID)
		identityKey = playerID
		created, err := s.queries.UpsertPlayerJudgment(ctx, db.UpsertPlayerJudgmentParams{
			ID:            judgmentID,
			JumpID:        jumpID,
			PlayerID:      sql.NullString{String: playerID, Valid: true},
			Provenance:    provenance,
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
			PlayerID:      identityKey,
			Provenance:    provenance,
			Commitment:    commitment,
			Transgression: transgression,
			Creativity:    creativity,
			Presentation:  presentation,
		}, created, nil
	}

	judgmentID = stableID("judgment", jumpID+":guest:"+guestSessionID)
	identityKey = guestSessionID
	created, err := s.queries.UpsertGuestJudgment(ctx, db.UpsertGuestJudgmentParams{
		ID:             judgmentID,
		JumpID:         jumpID,
		GuestSessionID: sql.NullString{String: guestSessionID, Valid: true},
		Provenance:     provenance,
		Commitment:     int32(commitment),
		Transgression:  int32(transgression),
		Creativity:     int32(creativity),
		Presentation:   int32(presentation),
	})
	if err != nil {
		return game.Judgment{}, false, err
	}
	return game.Judgment{
		ID:             judgmentID,
		JumpID:         jumpID,
		GuestSessionID: identityKey,
		Provenance:     provenance,
		Commitment:     commitment,
		Transgression:  transgression,
		Creativity:     creativity,
		Presentation:   presentation,
	}, created, nil
}

func (s *PostgresStore) AdvanceJumpToJudged(ctx context.Context, jumpID string) error {
	return s.queries.AdvanceJumpToJudged(ctx, jumpID)
}

func (s *PostgresStore) GuestSessionJudgmentCount(ctx context.Context, guestSessionID string) (int, error) {
	count, err := s.queries.GetGuestSessionJudgmentCount(ctx, guestSessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (s *PostgresStore) IncrementGuestSessionJudgmentCount(ctx context.Context, guestSessionID string) error {
	return s.queries.IncrementGuestSessionJudgmentCount(ctx, guestSessionID)
}

func (s *PostgresStore) CreateGuestSession(ctx context.Context, id string) error {
	_, err := s.queries.CreateGuestSession(ctx, id)
	return err
}

// HasJudgedJump returns true if the player has already submitted a Judgment for this Jump.
func (s *PostgresStore) HasJudgedJump(ctx context.Context, jumpID, playerID string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM judgments WHERE jump_id = $1 AND player_id = $2)`, jumpID, playerID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// HasJudgedJumps returns a map of jumpID → hasJudged for the given player.
// Uses a single SQL query with ANY instead of N individual queries.
func (s *PostgresStore) HasJudgedJumps(ctx context.Context, playerID string, jumpIDs []string) (map[string]bool, error) {
	if len(jumpIDs) == 0 {
		return map[string]bool{}, nil
	}

	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT jump_id FROM judgments WHERE player_id = $1 AND jump_id = ANY($2)`, playerID, jumpIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	judged := make(map[string]bool, len(jumpIDs))
	for rows.Next() {
		var jid string
		if err := rows.Scan(&jid); err != nil {
			return nil, err
		}
		judged[jid] = true
	}
	return judged, rows.Err()
}
