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

func (s *PostgresStore) BootstrapIdentity(ctx context.Context, identity AuthIdentity) (MeResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return MeResponse{}, err
	}
	defer tx.Rollback()
	qtx := s.queries.WithTx(tx)

	profile, err := qtx.GetProfileByAuthIdentity(ctx, db.GetProfileByAuthIdentityParams{
		Provider: identity.Provider,
		Subject:  identity.Subject,
	})
	if err == nil {
		return MeResponse{
			Account: Account{ID: profile.ID, Email: profile.Email},
			Player:  Player{ID: profile.ID_2, DisplayName: profile.DisplayName},
		}, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return MeResponse{}, err
	}
	key := identity.Provider + ":" + identity.Subject
	account := Account{ID: stableID("account", key), Email: identity.Email}
	player := Player{ID: stableID("player", account.ID), DisplayName: displayName(identity.Email)}
	if err := qtx.UpsertAccount(ctx, db.UpsertAccountParams{
		ID: account.ID, Email: account.Email,
	}); err != nil {
		return MeResponse{}, err
	}
	if err := qtx.InsertAuthIdentity(ctx, db.InsertAuthIdentityParams{
		Provider: identity.Provider, Subject: identity.Subject, AccountID: account.ID,
	}); err != nil {
		return MeResponse{}, err
	}
	if err := qtx.InsertPlayer(ctx, db.InsertPlayerParams{
		ID: player.ID, AccountID: sql.NullString{String: account.ID, Valid: true}, DisplayName: player.DisplayName,
	}); err != nil {
		return MeResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return MeResponse{}, err
	}
	return MeResponse{Account: account, Player: player}, nil
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
