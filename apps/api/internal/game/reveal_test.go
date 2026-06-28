package game_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

type fakeRevealRepo struct {
	rounds     map[string]game.RoundSnapshot
	statuses   map[string]string
	updateErr  error
}

func newFakeRevealRepo() *fakeRevealRepo {
	return &fakeRevealRepo{
		rounds:   make(map[string]game.RoundSnapshot),
		statuses: make(map[string]string),
	}
}

func (f *fakeRevealRepo) GetRound(ctx context.Context, roundID string) (game.RoundSnapshot, bool, error) {
	r, ok := f.rounds[roundID]
	if !ok {
		return game.RoundSnapshot{}, false, nil
	}
	if s, exists := f.statuses[roundID]; exists {
		r.Status = s
	}
	return r, true, nil
}

func (f *fakeRevealRepo) UpdateRoundStatus(ctx context.Context, roundID, status string) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.statuses[roundID] = status
	return nil
}

func TestEvaluateRevealFiresAtRevealTime(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRevealRepo()
	revealBy := frozenNow.Add(-time.Minute) // 1 minute ago
	repo.rounds["round-1"] = game.RoundSnapshot{
		ID:       "round-1",
		Status:   "active",
		RevealBy: revealBy,
	}

	result, err := game.EvaluateReveal(ctx, repo, "round-1", frozenNow)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if !result.Revealed {
		t.Fatal("expected Revealed=true when now >= revealBy")
	}
	if result.Round.Status != "revealed" {
		t.Fatalf("expected Round.Status=revealed, got %s", result.Round.Status)
	}
	if result.Round.ID != "round-1" {
		t.Fatalf("expected Round.ID=round-1, got %s", result.Round.ID)
	}
}

func TestEvaluateRevealDoesNotFireBeforeRevealTime(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRevealRepo()
	revealBy := frozenNow.Add(time.Hour) // 1 hour in the future
	repo.rounds["round-1"] = game.RoundSnapshot{
		ID:       "round-1",
		Status:   "active",
		RevealBy: revealBy,
	}

	result, err := game.EvaluateReveal(ctx, repo, "round-1", frozenNow)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if result.Revealed {
		t.Fatal("expected Revealed=false when now < revealBy")
	}
	if result.Round.Status != "active" {
		t.Fatalf("expected Round.Status=active, got %s", result.Round.Status)
	}
}

func TestEvaluateRevealIdempotentAlreadyRevealed(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRevealRepo()
	revealBy := frozenNow.Add(-time.Hour)
	repo.rounds["round-1"] = game.RoundSnapshot{
		ID:       "round-1",
		Status:   "revealed",
		RevealBy: revealBy,
	}

	result, err := game.EvaluateReveal(ctx, repo, "round-1", frozenNow)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true for already-revealed round, got false (err=%v)", result.Err)
	}
	if !result.Revealed {
		t.Fatal("expected Revealed=true for already-revealed round")
	}
	if result.Round.Status != "revealed" {
		t.Fatalf("expected Round.Status=revealed, got %s", result.Round.Status)
	}
}

func TestEvaluateRevealRoundNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRevealRepo()

	result, err := game.EvaluateReveal(ctx, repo, "nonexistent", frozenNow)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for nonexistent round")
	}
	if !errors.Is(result.Err, game.ErrRoundNotFound) {
		t.Fatalf("expected ErrRoundNotFound, got %v", result.Err)
	}
}
