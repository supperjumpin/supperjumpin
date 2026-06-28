package game_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

var frozenNow = time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)

type fakeStartRoundRepo struct {
	players      map[string]game.PlayerSnapshot
	communities  map[string]game.CommunitySnapshot
	prompts      map[string]game.PromptSnapshot
	promptOrder  []string
	timeframes   map[string]game.RevealTimeframeSnapshot
	activeRound  *game.RoundSnapshot
	createdRound *game.RoundSnapshot
	createErr    error
	findPlayerErr error
	findCommunityErr error
	findActiveErr error
	promptErr    error
	timeframeErr error
	listPromptsErr error
}

func newFakeStartRoundRepo() *fakeStartRoundRepo {
	return &fakeStartRoundRepo{
		players:     make(map[string]game.PlayerSnapshot),
		communities: make(map[string]game.CommunitySnapshot),
		prompts:     make(map[string]game.PromptSnapshot),
		timeframes:  make(map[string]game.RevealTimeframeSnapshot),
	}
}

func (f *fakeStartRoundRepo) FindPlayer(ctx context.Context, id string) (game.PlayerSnapshot, bool, error) {
	if f.findPlayerErr != nil {
		return game.PlayerSnapshot{}, false, f.findPlayerErr
	}
	p, ok := f.players[id]
	return p, ok, nil
}

func (f *fakeStartRoundRepo) FindCommunity(ctx context.Context, id string) (game.CommunitySnapshot, bool, error) {
	if f.findCommunityErr != nil {
		return game.CommunitySnapshot{}, false, f.findCommunityErr
	}
	c, ok := f.communities[id]
	return c, ok, nil
}

func (f *fakeStartRoundRepo) FindActiveRound(ctx context.Context, communityID string) (*game.RoundSnapshot, error) {
	if f.findActiveErr != nil {
		return nil, f.findActiveErr
	}
	return f.activeRound, nil
}

func (f *fakeStartRoundRepo) GetPrompt(ctx context.Context, id string) (game.PromptSnapshot, error) {
	if f.promptErr != nil {
		return game.PromptSnapshot{}, f.promptErr
	}
	p, ok := f.prompts[id]
	if !ok {
		return game.PromptSnapshot{}, game.ErrPromptNotFound
	}
	return p, nil
}

func (f *fakeStartRoundRepo) ListPrompts(ctx context.Context) ([]game.PromptSnapshot, error) {
	if f.listPromptsErr != nil {
		return nil, f.listPromptsErr
	}
	var result []game.PromptSnapshot
	for _, p := range f.promptOrder {
		if prompt, ok := f.prompts[p]; ok {
			result = append(result, prompt)
		}
	}
	if len(result) == 0 {
		// fallback: iterate map (order doesn't matter in this case)
		for _, p := range f.prompts {
			result = append(result, p)
		}
	}
	return result, nil
}

func (f *fakeStartRoundRepo) GetRevealTimeframe(ctx context.Context, id string) (game.RevealTimeframeSnapshot, error) {
	if f.timeframeErr != nil {
		return game.RevealTimeframeSnapshot{}, f.timeframeErr
	}
	tf, ok := f.timeframes[id]
	if !ok {
		return game.RevealTimeframeSnapshot{}, game.ErrRevealTimeframeNotFound
	}
	return tf, nil
}

func (f *fakeStartRoundRepo) CreateRound(ctx context.Context, round game.RoundSnapshot) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.createdRound = &round
	return nil
}

func TestStartRoundCreatesRoundWithChosenPrompt(t *testing.T) {
	ctx := context.Background()
	repo := newFakeStartRoundRepo()
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}
	repo.communities["comm-x"] = game.CommunitySnapshot{ID: "comm-x", DisplayName: "X"}
	repo.prompts["prompt-1"] = game.PromptSnapshot{ID: "prompt-1", Copy: "Do a thing", Theme: "Theme A", CostTier: "tier_1"}
	repo.timeframes["tf-24h"] = game.RevealTimeframeSnapshot{ID: "tf-24h", Label: "24 hours", DurationHours: 24}

	result, err := game.StartRound(ctx, repo, game.StartRoundInput{
		CommunityID:       "comm-x",
		PlayerID:          "player-a",
		PromptID:          "prompt-1",
		RevealTimeframeID: "tf-24h",
	}, frozenNow)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if result.Round.CommunityID != "comm-x" {
		t.Fatalf("expected CommunityID comm-x, got %s", result.Round.CommunityID)
	}
	if result.Round.PromptID != "prompt-1" {
		t.Fatalf("expected PromptID prompt-1, got %s", result.Round.PromptID)
	}
	if result.Round.Status != "active" {
		t.Fatalf("expected Status active, got %s", result.Round.Status)
	}
	if result.Round.CreatedBy != "player-a" {
		t.Fatalf("expected CreatedBy player-a, got %s", result.Round.CreatedBy)
	}
	if result.Round.CreatedAt != frozenNow {
		t.Fatalf("expected CreatedAt %v, got %v", frozenNow, result.Round.CreatedAt)
	}
	expectedRevealBy := frozenNow.Add(24 * time.Hour)
	if !result.Round.RevealBy.Equal(expectedRevealBy) {
		t.Fatalf("expected RevealBy %v, got %v", expectedRevealBy, result.Round.RevealBy)
	}
}

func TestStartRoundFailsWhenActiveRoundExists(t *testing.T) {
	ctx := context.Background()
	repo := newFakeStartRoundRepo()
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}
	repo.communities["comm-x"] = game.CommunitySnapshot{ID: "comm-x", DisplayName: "X"}
	repo.prompts["prompt-1"] = game.PromptSnapshot{ID: "prompt-1", Copy: "Do a thing"}
	repo.timeframes["tf-24h"] = game.RevealTimeframeSnapshot{ID: "tf-24h", Label: "24 hours", DurationHours: 24}
	existing := &game.RoundSnapshot{ID: "round-existing", CommunityID: "comm-x", Status: "active"}
	repo.activeRound = existing

	result, err := game.StartRound(ctx, repo, game.StartRoundInput{
		CommunityID:       "comm-x",
		PlayerID:          "player-a",
		PromptID:          "prompt-1",
		RevealTimeframeID: "tf-24h",
	}, frozenNow)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when active round exists")
	}
	if !errors.Is(result.Err, game.ErrRoundAlreadyActive) {
		t.Fatalf("expected ErrRoundAlreadyActive, got %v", result.Err)
	}
}

func TestStartRoundUsesRandomPromptWhenNoPromptID(t *testing.T) {
	ctx := context.Background()
	repo := newFakeStartRoundRepo()
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}
	repo.communities["comm-x"] = game.CommunitySnapshot{ID: "comm-x", DisplayName: "X"}
	repo.prompts["prompt-1"] = game.PromptSnapshot{ID: "prompt-1", Copy: "First"}
	repo.prompts["prompt-2"] = game.PromptSnapshot{ID: "prompt-2", Copy: "Second"}
	repo.prompts["prompt-3"] = game.PromptSnapshot{ID: "prompt-3", Copy: "Third"}
	repo.promptOrder = []string{"prompt-1", "prompt-2", "prompt-3"}
	repo.timeframes["tf-24h"] = game.RevealTimeframeSnapshot{ID: "tf-24h", Label: "24 hours", DurationHours: 24}

	pickIndex := func(n int) int {
		if n != 3 {
			t.Fatalf("expected picker to see 3 prompts, got %d", n)
		}
		return 1
	}

	result, err := game.StartRound(ctx, repo, game.StartRoundInput{
		CommunityID:       "comm-x",
		PlayerID:          "player-a",
		RevealTimeframeID: "tf-24h",
	}, frozenNow, pickIndex)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if result.Round.PromptID != "prompt-2" {
		t.Fatalf("expected PromptID prompt-2 (picker index 1), got %s", result.Round.PromptID)
	}
}

func TestStartRoundFailsWhenPlayerNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeStartRoundRepo()
	repo.communities["comm-x"] = game.CommunitySnapshot{ID: "comm-x", DisplayName: "X"}

	result, err := game.StartRound(ctx, repo, game.StartRoundInput{
		CommunityID:       "comm-x",
		PlayerID:          "player-missing",
		PromptID:          "prompt-1",
		RevealTimeframeID: "tf-24h",
	}, frozenNow)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when player not found")
	}
	if !errors.Is(result.Err, game.ErrPlayerNotFound) {
		t.Fatalf("expected ErrPlayerNotFound, got %v", result.Err)
	}
}

func TestStartRoundFailsWhenCommunityNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeStartRoundRepo()
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}

	result, err := game.StartRound(ctx, repo, game.StartRoundInput{
		CommunityID:       "comm-missing",
		PlayerID:          "player-a",
		PromptID:          "prompt-1",
		RevealTimeframeID: "tf-24h",
	}, frozenNow)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when community not found")
	}
	if !errors.Is(result.Err, game.ErrCommunityNotFound) {
		t.Fatalf("expected ErrCommunityNotFound, got %v", result.Err)
	}
}

func TestStartRoundFailsWhenPromptNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeStartRoundRepo()
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}
	repo.communities["comm-x"] = game.CommunitySnapshot{ID: "comm-x", DisplayName: "X"}
	repo.timeframes["tf-24h"] = game.RevealTimeframeSnapshot{ID: "tf-24h", Label: "24 hours", DurationHours: 24}

	result, err := game.StartRound(ctx, repo, game.StartRoundInput{
		CommunityID:       "comm-x",
		PlayerID:          "player-a",
		PromptID:          "prompt-missing",
		RevealTimeframeID: "tf-24h",
	}, frozenNow)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when prompt not found")
	}
	if !errors.Is(result.Err, game.ErrPromptNotFound) {
		t.Fatalf("expected ErrPromptNotFound, got %v", result.Err)
	}
}

func TestStartRoundFailsWhenTimeframeNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeStartRoundRepo()
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}
	repo.communities["comm-x"] = game.CommunitySnapshot{ID: "comm-x", DisplayName: "X"}
	repo.prompts["prompt-1"] = game.PromptSnapshot{ID: "prompt-1", Copy: "Do a thing"}

	result, err := game.StartRound(ctx, repo, game.StartRoundInput{
		CommunityID:       "comm-x",
		PlayerID:          "player-a",
		PromptID:          "prompt-1",
		RevealTimeframeID: "tf-missing",
	}, frozenNow)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when timeframe not found")
	}
	if !errors.Is(result.Err, game.ErrRevealTimeframeNotFound) {
		t.Fatalf("expected ErrRevealTimeframeNotFound, got %v", result.Err)
	}
}

func TestStartRoundFailsWhenNoPromptsAvailable(t *testing.T) {
	ctx := context.Background()
	repo := newFakeStartRoundRepo()
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}
	repo.communities["comm-x"] = game.CommunitySnapshot{ID: "comm-x", DisplayName: "X"}
	repo.timeframes["tf-24h"] = game.RevealTimeframeSnapshot{ID: "tf-24h", Label: "24 hours", DurationHours: 24}

	pickIndex := func(n int) int { return 0 }

	result, err := game.StartRound(ctx, repo, game.StartRoundInput{
		CommunityID:       "comm-x",
		PlayerID:          "player-a",
		RevealTimeframeID: "tf-24h",
	}, frozenNow, pickIndex)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when no prompts available")
	}
	if !errors.Is(result.Err, game.ErrNoPromptsAvailable) {
		t.Fatalf("expected ErrNoPromptsAvailable, got %v", result.Err)
	}
}

func TestListRevealTimeframesReturnsSeeded(t *testing.T) {
	ctx := context.Background()
	repo := &fakeListTimeframesRepo{
		timeframes: []game.RevealTimeframeSnapshot{
			{ID: "tf-24h", Label: "24 hours", DurationHours: 24},
			{ID: "tf-72h", Label: "3 days", DurationHours: 72},
		},
	}

	result, err := game.ListRevealTimeframes(ctx, repo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false")
	}
	if len(result.Timeframes) != 2 {
		t.Fatalf("expected 2 timeframes, got %d", len(result.Timeframes))
	}
	if result.Timeframes[0].ID != "tf-24h" {
		t.Fatalf("expected first timeframe tf-24h, got %s", result.Timeframes[0].ID)
	}
}

type fakeListTimeframesRepo struct {
	timeframes []game.RevealTimeframeSnapshot
	err        error
}

func (f *fakeListTimeframesRepo) ListRevealTimeframes(ctx context.Context) ([]game.RevealTimeframeSnapshot, error) {
	return f.timeframes, f.err
}
