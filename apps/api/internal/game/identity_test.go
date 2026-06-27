package game_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

type fakeEnsurePlayerRepo struct {
	players    map[string]game.PlayerSnapshot
	communities map[string]game.CommunitySnapshot
}

func newFakeEnsurePlayerRepo() *fakeEnsurePlayerRepo {
	return &fakeEnsurePlayerRepo{
		players:    make(map[string]game.PlayerSnapshot),
		communities: make(map[string]game.CommunitySnapshot),
	}
}

func (f *fakeEnsurePlayerRepo) FindPlayer(_ context.Context, id string) (game.PlayerSnapshot, bool, error) {
	p, ok := f.players[id]
	return p, ok, nil
}

func (f *fakeEnsurePlayerRepo) FindCommunity(_ context.Context, id string) (game.CommunitySnapshot, bool, error) {
	c, ok := f.communities[id]
	return c, ok, nil
}

func (f *fakeEnsurePlayerRepo) CreateCommunity(_ context.Context, id string, displayName string, now time.Time) error {
	if _, exists := f.communities[id]; exists {
		return nil
	}
	f.communities[id] = game.CommunitySnapshot{ID: id, DisplayName: displayName, CreatedAt: now}
	return nil
}

func (f *fakeEnsurePlayerRepo) CreatePlayer(_ context.Context, id string, displayName string, now time.Time) error {
	if _, exists := f.players[id]; exists {
		return nil
	}
	f.players[id] = game.PlayerSnapshot{ID: id, DisplayName: displayName, CreatedAt: now}
	return nil
}

func TestEnsurePlayerCreatesOnFirstCall(t *testing.T) {
	repo := newFakeEnsurePlayerRepo()
	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	result, err := game.EnsurePlayer(context.Background(), repo, game.EnsurePlayerInput{
		PlayerID:             "player-abc",
		PlayerDisplayName:    "coolkoala",
		CommunityID:          "community-xyz",
		CommunityDisplayName: "Supper Club",
	}, now)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Created {
		t.Fatal("expected Created=true on first call")
	}
	if result.Player.ID != "player-abc" {
		t.Fatalf("expected player ID 'player-abc', got %q", result.Player.ID)
	}
	if result.Player.DisplayName != "coolkoala" {
		t.Fatalf("expected display name 'coolkoala', got %q", result.Player.DisplayName)
	}
	if result.Community.ID != "community-xyz" {
		t.Fatalf("expected community ID 'community-xyz', got %q", result.Community.ID)
	}
	if result.Community.DisplayName != "Supper Club" {
		t.Fatalf("expected community display name 'Supper Club', got %q", result.Community.DisplayName)
	}
}

func TestEnsurePlayerIsIdempotent(t *testing.T) {
	repo := newFakeEnsurePlayerRepo()
	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	input := game.EnsurePlayerInput{
		PlayerID:             "player-abc",
		PlayerDisplayName:    "coolkoala",
		CommunityID:          "community-xyz",
		CommunityDisplayName: "Supper Club",
	}

	result1, err := game.EnsurePlayer(context.Background(), repo, input, now)
	if err != nil {
		t.Fatalf("first call: unexpected error %v", err)
	}
	if !result1.Created {
		t.Fatal("first call: expected Created=true")
	}
	if result1.Err != nil {
		t.Fatalf("first call: expected Err=nil, got %v", result1.Err)
	}

	result2, err := game.EnsurePlayer(context.Background(), repo, input, now)
	if err != nil {
		t.Fatalf("second call: unexpected error %v", err)
	}
	if result2.Created {
		t.Fatal("second call: expected Created=false")
	}
	if result2.Player.ID != result1.Player.ID {
		t.Fatal("second call: expected same player ID")
	}
	if result2.Community.ID != result1.Community.ID {
		t.Fatal("second call: expected same community ID")
	}
}

func TestEnsurePlayerReturnsErrorOnRepoFailure(t *testing.T) {
	broken := &brokenEnsurePlayerRepo{}
	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	result, err := game.EnsurePlayer(context.Background(), broken, game.EnsurePlayerInput{
		PlayerID:             "player-abc",
		PlayerDisplayName:    "coolkoala",
		CommunityID:          "community-xyz",
		CommunityDisplayName: "Supper Club",
	}, now)

	if err != nil {
		t.Fatalf("expected nil error (Err field carries it), got %v", err)
	}
	if result.Err == nil {
		t.Fatal("expected Err to be set when repo fails")
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when repo fails")
	}
}

type brokenEnsurePlayerRepo struct{}

func (b *brokenEnsurePlayerRepo) FindPlayer(_ context.Context, _ string) (game.PlayerSnapshot, bool, error) {
	return game.PlayerSnapshot{}, false, errors.New("db down")
}

func (b *brokenEnsurePlayerRepo) FindCommunity(_ context.Context, _ string) (game.CommunitySnapshot, bool, error) {
	return game.CommunitySnapshot{}, false, nil
}

func (b *brokenEnsurePlayerRepo) CreateCommunity(_ context.Context, _ string, _ string, _ time.Time) error {
	return nil
}

func (b *brokenEnsurePlayerRepo) CreatePlayer(_ context.Context, _ string, _ string, _ time.Time) error {
	return nil
}
