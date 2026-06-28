package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/db"
	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

type PostgresStore struct {
	db      *sql.DB
	queries *db.Queries
	now     func() time.Time
}

// isUniqueViolation reports whether err (or one of its wrapped errors) is a
// Postgres unique-constraint violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	d, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := d.PingContext(ctx); err != nil {
		d.Close()
		return nil, err
	}
	return &PostgresStore{
		db:      d,
		queries: db.New(d),
		now:     time.Now,
	}, nil
}
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) SetClock(now func() time.Time) {
	s.now = now
}

func (s *PostgresStore) UpdateDisplayName(ctx context.Context, playerID string, displayName string) (Player, error) {
	if err := s.queries.UpdatePlayerDisplayName(ctx, db.UpdatePlayerDisplayNameParams{
		ID:          playerID,
		DisplayName: displayName,
	}); err != nil {
		return Player{}, err
	}
	return Player{ID: playerID, DisplayName: displayName}, nil
}
func (s *PostgresStore) Now() time.Time {
	if s.now == nil {
		return time.Now()
	}
	return s.now()
}

func (s *PostgresStore) ListPromptPacks(ctx context.Context) ([]game.PromptPackSnapshot, error) {
	rows, err := s.queries.ListPromptPacks(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]game.PromptPackSnapshot, 0, len(rows))
	for _, r := range rows {
		result = append(result, game.PromptPackSnapshot{
			ID:          r.ID,
			DisplayName: r.DisplayName,
			Description: r.Description,
			CreatedAt:   r.CreatedAt,
		})
	}
	return result, nil
}

func (s *PostgresStore) ListPrompts(ctx context.Context) ([]game.PromptSnapshot, error) {
	rows, err := s.queries.ListPrompts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]game.PromptSnapshot, 0, len(rows))
	for _, r := range rows {
		result = append(result, game.PromptSnapshot{
			ID:        r.ID,
			PackID:    r.PackID,
			Copy:      r.Copy,
			Theme:     r.Theme,
			CostTier:  r.CostTier,
			CreatedAt: r.CreatedAt,
		})
	}
	return result, nil
}

func (s *PostgresStore) GetPrompt(ctx context.Context, id string) (game.PromptSnapshot, error) {
	row, err := s.queries.GetPrompt(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return game.PromptSnapshot{}, game.ErrPromptNotFound
		}
		return game.PromptSnapshot{}, err
	}
	return game.PromptSnapshot{
		ID:        row.ID,
		PackID:    row.PackID,
		Copy:      row.Copy,
		Theme:     row.Theme,
		CostTier:  row.CostTier,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (s *PostgresStore) ListRevealTimeframes(ctx context.Context) ([]game.RevealTimeframeSnapshot, error) {
	rows, err := s.queries.ListRevealTimeframes(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]game.RevealTimeframeSnapshot, 0, len(rows))
	for _, r := range rows {
		result = append(result, game.RevealTimeframeSnapshot{
			ID:            r.ID,
			Label:         r.Label,
			DurationHours: int(r.DurationHours),
		})
	}
	return result, nil
}

func (s *PostgresStore) GetRevealTimeframe(ctx context.Context, id string) (game.RevealTimeframeSnapshot, error) {
	row, err := s.queries.GetRevealTimeframe(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return game.RevealTimeframeSnapshot{}, game.ErrRevealTimeframeNotFound
		}
		return game.RevealTimeframeSnapshot{}, err
	}
	return game.RevealTimeframeSnapshot{
		ID:            row.ID,
		Label:         row.Label,
		DurationHours: int(row.DurationHours),
	}, nil
}

func (s *PostgresStore) FindActiveRound(ctx context.Context, communityID string) (*game.RoundSnapshot, error) {
	row, err := s.queries.FindActiveRound(ctx, communityID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &game.RoundSnapshot{
		ID:          row.ID,
		CommunityID: row.CommunityID,
		PromptID:    row.PromptID,
		Status:      row.Status,
		RevealBy:    row.RevealBy,
		CreatedBy:   row.CreatedBy,
		CreatedAt:   row.CreatedAt,
	}, nil
}

func (s *PostgresStore) CreateRound(ctx context.Context, round game.RoundSnapshot) error {
	return s.queries.CreateRound(ctx, db.CreateRoundParams{
		ID:          round.ID,
		CommunityID: round.CommunityID,
		PromptID:    round.PromptID,
		Status:      round.Status,
		RevealBy:    round.RevealBy,
		CreatedBy:   round.CreatedBy,
		CreatedAt:   round.CreatedAt,
	})
}

// --- CommitRepo ---

func (s *PostgresStore) FindRound(ctx context.Context, roundID string) (game.RoundSnapshot, bool, error) {
	row, err := s.queries.GetRound(ctx, roundID)
	if errors.Is(err, sql.ErrNoRows) {
		return game.RoundSnapshot{}, false, nil
	}
	if err != nil {
		return game.RoundSnapshot{}, false, err
	}
	return game.RoundSnapshot{
		ID:          row.ID,
		CommunityID: row.CommunityID,
		PromptID:    row.PromptID,
		Status:      row.Status,
		RevealBy:    row.RevealBy,
		CreatedBy:   row.CreatedBy,
		CreatedAt:   row.CreatedAt,
	}, true, nil
}

func (s *PostgresStore) FindCommit(ctx context.Context, roundID, playerID string) (*game.CommitSnapshot, error) {
	row, err := s.queries.FindCommit(ctx, db.FindCommitParams{RoundID: roundID, PlayerID: playerID})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &game.CommitSnapshot{
		ID:          row.ID,
		RoundID:     row.RoundID,
		PlayerID:    row.PlayerID,
		CommittedAt: row.CommittedAt,
	}, nil
}

func (s *PostgresStore) CreateCommit(ctx context.Context, commit game.CommitSnapshot) error {
	return s.queries.CreateCommit(ctx, db.CreateCommitParams{
		ID:          commit.ID,
		RoundID:     commit.RoundID,
		PlayerID:    commit.PlayerID,
		CommittedAt: commit.CommittedAt,
	})
}

// --- SubmitRepo ---

func (s *PostgresStore) FindJump(ctx context.Context, roundID, playerID string) (*game.JumpSnapshot, error) {
	row, err := s.queries.FindJumpByPlayer(ctx, db.FindJumpByPlayerParams{RoundID: roundID, PlayerID: playerID})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &game.JumpSnapshot{
		ID:          row.ID,
		RoundID:     row.RoundID,
		PlayerID:    row.PlayerID,
		Caption:     row.Caption,
		SubmittedAt: row.SubmittedAt,
	}, nil
}

func (s *PostgresStore) CreateJump(ctx context.Context, jump game.JumpSnapshot) error {
	return s.queries.CreateJump(ctx, db.CreateJumpParams{
		ID:          jump.ID,
		RoundID:     jump.RoundID,
		PlayerID:    jump.PlayerID,
		Caption:     jump.Caption,
		SubmittedAt: jump.SubmittedAt,
	})
}

func (s *PostgresStore) InsertEvidence(ctx context.Context, jumpID string, urls []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := s.queries.WithTx(tx)

	for i, url := range urls {
		evidenceID := stableID("evidence", jumpID+":"+url+":"+strconv.Itoa(i))
		if err := qtx.InsertJumpEvidence(ctx, db.InsertJumpEvidenceParams{
			ID:        evidenceID,
			JumpID:    jumpID,
			Url:       url,
			SortOrder: int32(i),
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// --- ListJumpsRepo ---

func (s *PostgresStore) ListJumps(ctx context.Context, roundID string) ([]game.JumpSnapshot, error) {
	rows, err := s.queries.ListJumpsByRoundWithContent(ctx, roundID)
	if err != nil {
		return nil, err
	}
	result := make([]game.JumpSnapshot, 0, len(rows))
	for _, row := range rows {
		result = append(result, game.JumpSnapshot{
			ID:          row.ID,
			RoundID:     row.RoundID,
			PlayerID:    row.PlayerID,
			Caption:     row.Caption,
			SubmittedAt: row.SubmittedAt,
		})
	}
	return result, nil
}

func (s *PostgresStore) ListEvidence(ctx context.Context, jumpIDs []string) (map[string][]string, error) {
	if len(jumpIDs) == 0 {
		return map[string][]string{}, nil
	}
	rows, err := s.queries.ListEvidenceForJumps(ctx, jumpIDs)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]string)
	for _, row := range rows {
		m[row.JumpID] = append(m[row.JumpID], row.Url)
	}
	return m, nil
}

func (s *PostgresStore) GetRoundStatus(ctx context.Context, roundID string) (game.RoundStatus, error) {
	row, err := s.queries.GetRoundStatus(ctx, roundID)
	if errors.Is(err, sql.ErrNoRows) {
		return game.RoundStatus{}, game.ErrRoundNotFound
	}
	if err != nil {
		return game.RoundStatus{}, err
	}
	return game.RoundStatus{
		ID:             row.ID,
		Status:         row.Status,
		RevealBy:       row.RevealBy,
		CommitCount:    int(row.CommitCount),
		SubmissionCount: int(row.SubmissionCount),
	}, nil
}

func (s *PostgresStore) FindCommitForPlayer(ctx context.Context, roundID, playerID string) (*game.CommitSnapshot, error) {
	return s.FindCommit(ctx, roundID, playerID)
}

// --- RevealRepo ---

func (s *PostgresStore) GetRound(ctx context.Context, roundID string) (game.RoundSnapshot, bool, error) {
	return s.FindRound(ctx, roundID)
}

func (s *PostgresStore) UpdateRoundStatus(ctx context.Context, roundID, status string) error {
	return s.queries.UpdateRoundStatus(ctx, db.UpdateRoundStatusParams{
		ID:     roundID,
		Status: status,
	})
}

// --- GetJumpRepo ---

func (s *PostgresStore) GetJumpByID(ctx context.Context, jumpID string) (game.JumpSnapshot, error) {
	row, err := s.queries.GetJumpByID(ctx, jumpID)
	if errors.Is(err, sql.ErrNoRows) {
		return game.JumpSnapshot{}, game.ErrJumpNotFound
	}
	if err != nil {
		return game.JumpSnapshot{}, err
	}
	return game.JumpSnapshot{
		ID:          row.ID,
		RoundID:     row.RoundID,
		PlayerID:    row.PlayerID,
		Caption:     row.Caption,
		SubmittedAt: row.SubmittedAt,
	}, nil
}

func (s *PostgresStore) ListEvidenceForJump(ctx context.Context, jumpID string) ([]string, error) {
	rows, err := s.queries.ListEvidenceForJump(ctx, jumpID)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.Url)
	}
	return result, nil
}

// --- ListStampCatalogRepo ---

func (s *PostgresStore) ListStamps(ctx context.Context) ([]game.StampSnapshot, error) {
	rows, err := s.queries.ListStamps(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]game.StampSnapshot, 0, len(rows))
	for _, row := range rows {
		result = append(result, game.StampSnapshot{
			ID:        row.ID,
			Stance:    row.Stance,
			Label:     row.Label,
			Glyph:     row.Glyph,
			Copy:      row.Copy,
			CreatedAt: row.CreatedAt,
		})
	}
	return result, nil
}

// --- ApplyReactionRepo ---

func (s *PostgresStore) GetStamp(ctx context.Context, stampID string) (game.StampSnapshot, error) {
	row, err := s.queries.GetStamp(ctx, stampID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return game.StampSnapshot{}, game.ErrStampNotFound
		}
		return game.StampSnapshot{}, err
	}
	return game.StampSnapshot{
		ID:        row.ID,
		Stance:    row.Stance,
		Label:     row.Label,
		Glyph:     row.Glyph,
		Copy:      row.Copy,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (s *PostgresStore) GetJump(ctx context.Context, jumpID string) (game.JumpSnapshot, error) {
	return s.GetJumpByID(ctx, jumpID)
}

func (s *PostgresStore) CreateReaction(ctx context.Context, reaction game.ReactionSnapshot) error {
	return s.queries.CreateReaction(ctx, db.CreateReactionParams{
		ID:        reaction.ID,
		StampID:   reaction.StampID,
		JumpID:    reaction.JumpID,
		PlayerID:  reaction.PlayerID,
		CreatedAt: reaction.CreatedAt,
	})
}

func (s *PostgresStore) FindReaction(ctx context.Context, jumpID, playerID, stampID string) (*game.ReactionSnapshot, error) {
	row, err := s.queries.FindReaction(ctx, db.FindReactionParams{
		StampID:  stampID,
		JumpID:   jumpID,
		PlayerID: playerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &game.ReactionSnapshot{
		ID:        row.ID,
		StampID:   row.StampID,
		JumpID:    row.JumpID,
		PlayerID:  row.PlayerID,
		CreatedAt: row.CreatedAt,
	}, nil
}

// --- PostCommentRepo ---

func (s *PostgresStore) CreateComment(ctx context.Context, comment game.CommentSnapshot) error {
	var jumpID sql.NullString
	if comment.JumpID != "" {
		jumpID = sql.NullString{String: comment.JumpID, Valid: true}
	}
	return s.queries.CreateComment(ctx, db.CreateCommentParams{
		ID:        comment.ID,
		RoundID:   comment.RoundID,
		JumpID:    jumpID,
		PlayerID:  comment.PlayerID,
		Body:      comment.Body,
		CreatedAt: comment.CreatedAt,
	})
}

// --- ListCommentsRepo ---

// --- RecapRepo ---

func (s *PostgresStore) ListJumpsWithContent(ctx context.Context, roundID string) ([]game.JumpSnapshot, error) {
	rows, err := s.queries.ListJumpsByRoundWithContent(ctx, roundID)
	if err != nil {
		return nil, err
	}
	result := make([]game.JumpSnapshot, 0, len(rows))
	for _, row := range rows {
		result = append(result, game.JumpSnapshot{
			ID:          row.ID,
			RoundID:     row.RoundID,
			PlayerID:    row.PlayerID,
			Caption:     row.Caption,
			SubmittedAt: row.SubmittedAt,
		})
	}
	return result, nil
}

func (s *PostgresStore) ListReactionsForRound(ctx context.Context, roundID string) ([]game.RecapReactionRow, error) {
	rows, err := s.queries.ListReactionsForRound(ctx, roundID)
	if err != nil {
		return nil, err
	}
	result := make([]game.RecapReactionRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, game.RecapReactionRow{
			JumpID:      row.JumpID,
			StampStance: row.StampStance,
		})
	}
	return result, nil
}

func (s *PostgresStore) ListAllCommentsForRound(ctx context.Context, roundID string) ([]game.CommentSnapshot, error) {
	rows, err := s.queries.ListAllCommentsForRound(ctx, roundID)
	if err != nil {
		return nil, err
	}
	result := make([]game.CommentSnapshot, 0, len(rows))
	for _, row := range rows {
		result = append(result, game.CommentSnapshot{
			ID:        row.ID,
			RoundID:   row.RoundID,
			JumpID:    row.JumpID.String,
			PlayerID:  row.PlayerID,
			Body:      row.Body,
			CreatedAt: row.CreatedAt,
		})
	}
	return result, nil
}

func (s *PostgresStore) ListGhostJumpers(ctx context.Context, roundID string) ([]game.RecapGhostJumperRow, error) {
	rows, err := s.queries.ListGhostJumpers(ctx, roundID)
	if err != nil {
		return nil, err
	}
	result := make([]game.RecapGhostJumperRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, game.RecapGhostJumperRow{
			PlayerID:    row.PlayerID,
			CommittedAt: row.CommittedAt,
		})
	}
	return result, nil
}

// --- LoreRepo ---

func (s *PostgresStore) ListRevealedReactionsForCommunity(ctx context.Context, communityID string) ([]game.LoreReactionRow, error) {
	rows, err := s.queries.ListRevealedReactionsForCommunity(ctx, communityID)
	if err != nil {
		return nil, err
	}
	result := make([]game.LoreReactionRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, game.LoreReactionRow{
			JumpID:       row.JumpID,
			StampStance:  row.StampStance,
			RoundID:      row.RoundID,
			JumpCaption:  row.JumpCaption,
			JumpPlayerID: row.JumpPlayerID,
		})
	}
	return result, nil
}

// --- ListCommentsRepo ---

func (s *PostgresStore) ListComments(ctx context.Context, roundID, jumpID string) ([]game.CommentSnapshot, error) {
	if jumpID != "" {
		rows, err := s.queries.ListCommentsForJump(ctx, db.ListCommentsForJumpParams{
			RoundID: roundID,
			JumpID:  sql.NullString{String: jumpID, Valid: true},
		})
		if err != nil {
			return nil, err
		}
		result := make([]game.CommentSnapshot, 0, len(rows))
		for _, row := range rows {
			result = append(result, game.CommentSnapshot{
				ID:        row.ID,
				RoundID:   row.RoundID,
				JumpID:    row.JumpID.String,
				PlayerID:  row.PlayerID,
				Body:      row.Body,
				CreatedAt: row.CreatedAt,
			})
		}
		return result, nil
	}

	rows, err := s.queries.ListCommentsForRound(ctx, roundID)
	if err != nil {
		return nil, err
	}
	result := make([]game.CommentSnapshot, 0, len(rows))
	for _, row := range rows {
		result = append(result, game.CommentSnapshot{
			ID:        row.ID,
			RoundID:   row.RoundID,
			PlayerID:  row.PlayerID,
			Body:      row.Body,
			CreatedAt: row.CreatedAt,
		})
	}
	return result, nil
}
