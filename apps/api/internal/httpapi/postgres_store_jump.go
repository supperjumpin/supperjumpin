package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/db"
	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// game.JumpRepository adapter methods for PostgresStore

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

	if _, err := tx.ExecContext(ctx, `
INSERT INTO jumps (id, player_id, season_id, status, source, destination, food, grace_period_expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		jumpID, params.PlayerID, nil, "Performed Jump",
		params.Source, params.Destination, params.Food, params.GracePeriodExpiresAt,
	); err != nil {
		return game.JumpSnapshot{}, game.EvidenceSnapshot{}, err
	}

	evidenceID := stableID("evidence", jumpID+":"+params.MediaObjectKey)
	now := s.Now().UTC()
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
		ID:                   jumpID,
		PlayerID:             params.PlayerID,
		Status:               "Performed Jump",
		Source:               params.Source,
		Destination:          params.Destination,
		Food:                 params.Food,
		GracePeriodExpiresAt: params.GracePeriodExpiresAt,
	}
	evidenceSnap := game.EvidenceSnapshot{
		ID:             evidenceID,
		JumpID:         jumpID,
		PlayerID:       params.PlayerID,
		MediaObjectKey: params.MediaObjectKey,
		Caption:        params.Caption,
		CreatedAt:      now,
	}
	return jumpSnap, evidenceSnap, nil
}

// FeedJumps returns a page of public Feed Jumps with cursor-based pagination.
//
// TODO: The running average subquery (judgments LEFT JOIN + AVG aggregate) does a
// full scan of the judgments table per query. At scale, consider materializing the
// running average on the jumps table (periodic batch update) or adding a composite
// index on (jump_id, created_at) to the judgments table. Tracked for Wave 2+.
func (s *PostgresStore) FeedJumps(ctx context.Context, cursorTS *time.Time, cursorID string, limit int) ([]JumpCard, error) {
	var rows *sql.Rows
	var err error

	if cursorTS == nil {
		rows, err = s.db.QueryContext(ctx, `
SELECT j.id, j.player_id, j.season_id, j.status, j.source, j.destination, j.food,
       j.final_score, j.grace_period_expires_at, j.created_at,
       COALESCE(e.id, '') AS evidence_id, COALESCE(e.caption, '') AS caption, COALESCE(e.media_object_key, '') AS media_object_key,
       p.id AS performer_id, p.display_name AS performer_name,
       COALESCE(jg.avg_composite, 0) AS running_average,
       COALESCE(jg.judgment_count, 0) AS judgment_count
FROM jumps j
JOIN players p ON p.id = j.player_id
LEFT JOIN evidences e ON e.jump_id = j.id
LEFT JOIN (
    SELECT jump_id,
           AVG((CAST(commitment AS numeric(4,2)) + CAST(transgression AS numeric(4,2)) + CAST(creativity AS numeric(4,2)) + CAST(presentation AS numeric(4,2))) / 4.0) AS avg_composite,
           COUNT(*) AS judgment_count
    FROM judgments
    GROUP BY jump_id
) jg ON jg.jump_id = j.id
WHERE j.status IN ('Performed Jump', 'Judged Jump', 'Unjudged Jump', 'Disqualified Jump')
ORDER BY j.created_at DESC, j.id DESC
LIMIT $1`, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, `
SELECT j.id, j.player_id, j.season_id, j.status, j.source, j.destination, j.food,
       j.final_score, j.grace_period_expires_at, j.created_at,
       COALESCE(e.id, '') AS evidence_id, COALESCE(e.caption, '') AS caption, COALESCE(e.media_object_key, '') AS media_object_key,
       p.id AS performer_id, p.display_name AS performer_name,
       COALESCE(jg.avg_composite, 0) AS running_average,
       COALESCE(jg.judgment_count, 0) AS judgment_count
FROM jumps j
JOIN players p ON p.id = j.player_id
LEFT JOIN evidences e ON e.jump_id = j.id
LEFT JOIN (
    SELECT jump_id,
           AVG((CAST(commitment AS numeric(4,2)) + CAST(transgression AS numeric(4,2)) + CAST(creativity AS numeric(4,2)) + CAST(presentation AS numeric(4,2))) / 4.0) AS avg_composite,
           COUNT(*) AS judgment_count
    FROM judgments
    GROUP BY jump_id
) jg ON jg.jump_id = j.id
WHERE j.status IN ('Performed Jump', 'Judged Jump', 'Unjudged Jump', 'Disqualified Jump')
  AND (j.created_at < $1 OR (j.created_at = $1 AND j.id < $2))
ORDER BY j.created_at DESC, j.id DESC
LIMIT $3`, *cursorTS, cursorID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []JumpCard
	for rows.Next() {
		var card JumpCard
		var seasonID sql.NullString
		var finalScore sql.NullInt32
		var evidenceID sql.NullString

		if err := rows.Scan(
			&card.ID, &card.PerformerID, &seasonID,
			&card.Status, &card.Source, &card.Destination, &card.Food,
			&finalScore, &card.GracePeriodExpiresAt, &card.CreatedAt,
			&evidenceID, &card.Caption, &card.MediaObjectKey,
			&card.PerformerID, &card.PerformerName,
			&card.RunningAverage, &card.JudgmentCount,
		); err != nil {
			return nil, err
		}
		_ = seasonID
		_ = finalScore
		_ = evidenceID
		cards = append(cards, card)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cards, nil
}

// JumpDetail returns the full detail view of a Jump. ok=false if not found.
func (s *PostgresStore) JumpDetail(ctx context.Context, jumpID string) (JumpDetail, bool, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT j.id, j.player_id, j.season_id, j.status, j.source, j.destination, j.food,
       j.final_score, j.removed_at, j.grace_period_expires_at, j.created_at,
       COALESCE(e.id, '') AS evidence_id, COALESCE(e.caption, '') AS caption, COALESCE(e.media_object_key, '') AS media_object_key,
       p.id AS performer_id, p.display_name AS performer_name,
       COALESCE(jg.avg_composite, 0) AS running_average,
       COALESCE(jg.judgment_count, 0) AS judgment_count
FROM jumps j
JOIN players p ON p.id = j.player_id
LEFT JOIN evidences e ON e.jump_id = j.id
LEFT JOIN (
    SELECT jump_id,
           AVG((CAST(commitment AS numeric(4,2)) + CAST(transgression AS numeric(4,2)) + CAST(creativity AS numeric(4,2)) + CAST(presentation AS numeric(4,2))) / 4.0) AS avg_composite,
           COUNT(*) AS judgment_count
    FROM judgments
    GROUP BY jump_id
) jg ON jg.jump_id = j.id
WHERE j.id = $1`, jumpID)

	var detail JumpDetail
	var seasonID sql.NullString
	var finalScore sql.NullInt32
	var removedAt sql.NullTime
	var evidenceID sql.NullString

	err := row.Scan(
		&detail.ID, &detail.PerformerID, &seasonID,
		&detail.Status, &detail.Source, &detail.Destination, &detail.Food,
		&finalScore, &removedAt, &detail.GracePeriodExpiresAt, &detail.CreatedAt,
		&evidenceID, &detail.Caption, &detail.MediaObjectKey,
		&detail.PerformerID, &detail.PerformerName,
		&detail.RunningAverage, &detail.JudgmentCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return JumpDetail{}, false, nil
	}
	if err != nil {
		return JumpDetail{}, false, err
	}
	_ = seasonID
	_ = evidenceID
	if finalScore.Valid {
		v := int(finalScore.Int32)
		detail.FinalScore = &v
	}
	if removedAt.Valid {
		t := removedAt.Time
		detail.RemovedAt = &t
	}
	return detail, true, nil
}
