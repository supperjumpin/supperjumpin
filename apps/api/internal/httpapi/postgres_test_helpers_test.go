package httpapi_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func postgresTestDatabaseURL(t *testing.T) string {
	t.Helper()
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}
	return databaseURL
}

func openPostgresTestDB(t *testing.T, databaseURL string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open Postgres database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close Postgres database: %v", err)
		}
	})
	return db
}

func cleanTestDatabase(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
		TRUNCATE TABLE open_standings, season_history, disputes, judgments, guest_sessions,
		evidence_upload_authorizations, evidences, jumps, invites, seasons,
		group_memberships, groups, auth_identities, players, accounts CASCADE
	`); err != nil {
		t.Fatalf("clean test database: %v", err)
	}
}

func newPostgresTestStore(t *testing.T, databaseURL string) *httpapi.PostgresStore {
	t.Helper()
	store, err := httpapi.NewPostgresStore(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("new Postgres store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close Postgres store: %v", err)
		}
	})
	return store
}

func newCleanPostgresTestStore(t *testing.T) *httpapi.PostgresStore {
	t.Helper()
	databaseURL := postgresTestDatabaseURL(t)
	cleanTestDatabase(t, openPostgresTestDB(t, databaseURL))
	return newPostgresTestStore(t, databaseURL)
}
