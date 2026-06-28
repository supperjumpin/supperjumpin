package httpapi

import (
	"context"
	"database/sql"
	"errors"
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
