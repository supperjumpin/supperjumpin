package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.SeasonRepository adapter methods for PostgresStore

func (s *PostgresStore) OpenSeasonForGroup(ctx context.Context, groupID string) (game.SeasonSnapshot, error) {
	var snap game.SeasonSnapshot
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline
FROM seasons
WHERE group_id = $1 AND status IN ('Active', 'Judging Grace Period')
LIMIT 1`, groupID).Scan(
		&snap.ID, &snap.GroupID, &snap.CommissionerPlayerID, &snap.Status, &snap.SubmissionDeadline, &snap.JudgingDeadline)
	if errors.Is(err, sql.ErrNoRows) {
		return game.SeasonSnapshot{}, nil
	}
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	return snap, nil
}

func (s *PostgresStore) InsertSeason(ctx context.Context, groupID, commissionerPlayerID string, submissionDeadline, judgingDeadline time.Time) (game.SeasonSnapshot, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM seasons`).Scan(&count); err != nil {
		return game.SeasonSnapshot{}, err
	}
	season := Season{
		ID:                   stableID("season", groupID+":"+strconv.Itoa(count+1)),
		GroupID:              groupID,
		CommissionerPlayerID: commissionerPlayerID,
		Status:               "Active",
		SubmissionDeadline:   submissionDeadline.UTC(),
		JudgingDeadline:      judgingDeadline.UTC(),
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO seasons (id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline)
VALUES ($1, $2, $3, $4, $5, $6)`, season.ID, season.GroupID, season.CommissionerPlayerID, season.Status, season.SubmissionDeadline, season.JudgingDeadline)
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

func (s *PostgresStore) UpdateSeasonStatus(ctx context.Context, seasonID, action, actorPlayerID, actorRole string, override bool, fromStatus, toStatus string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE seasons SET status = $2 WHERE id = $1`, seasonID, toStatus); err != nil {
		return err
	}
	if err := insertSeasonHistoryEntry(ctx, tx, seasonID, action, actorPlayerID, actorRole, override, fromStatus, toStatus); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *PostgresStore) JumpsForSeason(ctx context.Context, seasonID string) ([]game.JumpSnapshot, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, group_id, player_id, season_id, status, source, destination, food, final_score, grace_period_expires_at
FROM jumps
WHERE season_id = $1`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []game.JumpSnapshot
	for rows.Next() {
		var snap game.JumpSnapshot
		var sid sql.NullString
		var gracePeriodExpiresAt sql.NullTime
		if err := rows.Scan(&snap.ID, &snap.GroupID, &snap.PlayerID, &sid, &snap.Status, &snap.Source, &snap.Destination, &snap.Food, &snap.FinalScore, &gracePeriodExpiresAt); err != nil {
			return nil, err
		}
		if sid.Valid {
			snap.SeasonID = &sid.String
		}
		if gracePeriodExpiresAt.Valid {
			snap.GracePeriodExpiresAt = gracePeriodExpiresAt.Time
		}
		result = append(result, snap)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PostgresStore) JudgmentsForJump(ctx context.Context, jumpID string) ([]game.Judgment, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, jump_id, player_id, difficulty, transgression, creativity, presentation
FROM judgments
WHERE jump_id = $1`, jumpID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []game.Judgment
	for rows.Next() {
		var j game.Judgment
		if err := rows.Scan(&j.ID, &j.JumpID, &j.PlayerID, &j.Difficulty, &j.Transgression, &j.Creativity, &j.Presentation); err != nil {
			return nil, err
		}
		result = append(result, j)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PostgresStore) UpdateJumpFinalization(ctx context.Context, jumpID string, status string, finalScore *int) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE jumps
SET status = $2, final_score = $3
WHERE id = $1`, jumpID, status, finalScore)
	return err
}

func (s *PostgresStore) LatestSeasonForGroup(ctx context.Context, groupID string) (game.SeasonSnapshot, error) {
	var snap game.SeasonSnapshot
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline
FROM seasons
WHERE group_id = $1
ORDER BY created_at DESC
LIMIT 1`, groupID).Scan(
		&snap.ID, &snap.GroupID, &snap.CommissionerPlayerID, &snap.Status, &snap.SubmissionDeadline, &snap.JudgingDeadline)
	if errors.Is(err, sql.ErrNoRows) {
		return game.SeasonSnapshot{}, nil
	}
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	return snap, nil
}

func (s *PostgresStore) GroupPlayers(ctx context.Context, groupID string) ([]game.PlayerSnapshot, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT players.id, players.display_name
FROM group_memberships
JOIN players ON players.id = group_memberships.player_id
WHERE group_memberships.group_id = $1`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []game.PlayerSnapshot
	for rows.Next() {
		var p game.PlayerSnapshot
		if err := rows.Scan(&p.ID, &p.DisplayName); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PostgresStore) SeasonHistoryEntries(ctx context.Context, seasonID string) ([]game.SeasonHistoryEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, season_id, action, actor_player_id, actor_role, override, from_status, to_status
FROM season_history
WHERE season_id = $1
ORDER BY created_at ASC, id ASC`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []game.SeasonHistoryEntry{}
	for rows.Next() {
		var entry game.SeasonHistoryEntry
		if err := rows.Scan(&entry.ID, &entry.SeasonID, &entry.Action, &entry.ActorPlayerID, &entry.ActorRole, &entry.Override, &entry.FromStatus, &entry.ToStatus); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
