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

func (s *PostgresStore) ActiveOpen(_ context.Context, _ time.Time) (*game.OpenMonth, error) {
	return nil, nil
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
		ID: player.ID, AccountID: account.ID, DisplayName: player.DisplayName,
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
