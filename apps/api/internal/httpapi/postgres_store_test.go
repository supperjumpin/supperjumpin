package httpapi_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func TestPostgresStoreLatestSeasonForGroupPrefersCreationOrderOverIDOrder(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store, err := httpapi.NewPostgresStore(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("new postgres store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close postgres store: %v", err)
		}
	})
	server := newGroupsTestServerWithStore(store)
	ctx := context.Background()
	targetGroup := createGroup(t, server, "alice-token", "Latest Season Group")

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open postgres database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close postgres database: %v", err)
		}
	})

	var currentCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM seasons`).Scan(&currentCount); err != nil {
		t.Fatalf("count seasons: %v", err)
	}

	base := currentCount

	firstOffset, secondOffset, found := 0, 0, false
	for a := 1; a <= 1000 && !found; a++ {
		firstID := seasonIDForCount(targetGroup.Group.ID, base+a)
		for b := a + 1; b <= 1000; b++ {
			secondID := seasonIDForCount(targetGroup.Group.ID, base+b)
			if firstID > secondID {
				firstOffset = a
				secondOffset = b
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("expected an ID ordering inversion within the search window")
	}

	makeFillerSeason := func(index int) {
		group := createGroup(t, server, "alice-token", fmt.Sprintf("Filler Group %d", index))
		if _, err := store.InsertSeason(ctx, group.Group.ID, group.Membership.PlayerID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)); err != nil {
			t.Fatalf("insert filler season: %v", err)
		}
	}

	for i := 1; i < firstOffset; i++ {
		makeFillerSeason(i)
	}

	first, err := store.InsertSeason(ctx, targetGroup.Group.ID, targetGroup.Membership.PlayerID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("insert first target season: %v", err)
	}

	for i := firstOffset + 1; i < secondOffset; i++ {
		makeFillerSeason(i)
	}

	latest, err := store.InsertSeason(ctx, targetGroup.Group.ID, targetGroup.Membership.PlayerID, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("insert latest target season: %v", err)
	}

	if first.ID <= latest.ID {
		t.Fatalf("expected chosen pair to invert ID order, got first=%q latest=%q", first.ID, latest.ID)
	}

	got, err := store.LatestSeasonForGroup(ctx, targetGroup.Group.ID)
	if err != nil {
		t.Fatalf("LatestSeasonForGroup: %v", err)
	}
	if got.ID != latest.ID {
		t.Fatalf("expected latest season %q, got %q", latest.ID, got.ID)
	}
}

func seasonIDForCount(groupID string, count int) string {
	sum := sha256.Sum256([]byte("season:" + groupID + ":" + strconv.Itoa(count)))
	return "season_" + hex.EncodeToString(sum[:])[:12]
}
