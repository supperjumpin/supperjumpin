package httpapi

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStoreLatestSeasonForGroupPrefersCreationOrderOverIDOrder(t *testing.T) {
	store := NewMemoryStoreWithClock(time.Now)
	ctx := context.Background()

	first, err := store.InsertSeason(ctx, "group_1", "player_1", time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("insert first season: %v", err)
	}
	if _, err := store.InsertSeason(ctx, "group_2", "player_2", time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("insert unrelated season: %v", err)
	}
	latest, err := store.InsertSeason(ctx, "group_1", "player_3", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("insert latest season: %v", err)
	}

	if first.ID <= latest.ID {
		t.Fatalf("expected chosen pair to invert ID order, got first=%q latest=%q", first.ID, latest.ID)
	}

	got, err := store.LatestSeasonForGroup(ctx, "group_1")
	if err != nil {
		t.Fatalf("LatestSeasonForGroup: %v", err)
	}
	if got.ID != latest.ID {
		t.Fatalf("expected latest season %q, got %q", latest.ID, got.ID)
	}
}
