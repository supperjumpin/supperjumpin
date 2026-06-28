package game_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

var reactionFrozenNow = time.Date(2026, 6, 28, 18, 0, 0, 0, time.UTC)

// --- fake ListStampCatalogRepo ---

type fakeListStampCatalogRepo struct {
	stamps []game.StampSnapshot
	err    error
}

func (f *fakeListStampCatalogRepo) ListStamps(_ context.Context) ([]game.StampSnapshot, error) {
	return f.stamps, f.err
}

func stamp(id, stance, label, glyph, stCopy string) game.StampSnapshot {
	return game.StampSnapshot{
		ID:        id,
		Stance:    stance,
		Label:     label,
		Glyph:     glyph,
		Copy:      stCopy,
		CreatedAt: reactionFrozenNow,
	}
}

func TestListStampCatalogReturnsSeeded(t *testing.T) {
	repo := &fakeListStampCatalogRepo{
		stamps: []game.StampSnapshot{
			stamp("stamp-1", "approval", "Approve", "✅", "Yes."),
			stamp("stamp-2", "chaos", "Chaos", "🌀", "Unsupervised."),
		},
	}

	result, err := game.ListStampCatalog(context.Background(), repo)
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got %+v", result)
	}
	if len(result.Stamps) != 2 {
		t.Fatalf("expected 2 stamps, got %d", len(result.Stamps))
	}
	if result.Stamps[0].Stance != "approval" {
		t.Fatalf("expected stance='approval', got %q", result.Stamps[0].Stance)
	}
}

func TestListStampCatalogEmpty(t *testing.T) {
	repo := &fakeListStampCatalogRepo{}

	result, err := game.ListStampCatalog(context.Background(), repo)
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got %+v", result)
	}
	if len(result.Stamps) != 0 {
		t.Fatalf("expected 0 stamps, got %d", len(result.Stamps))
	}
}

func TestListStampCatalogRepoErrorPropagated(t *testing.T) {
	repo := &fakeListStampCatalogRepo{err: errors.New("db down")}

	result, err := game.ListStampCatalog(context.Background(), repo)
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatalf("expected Allowed=false on repo error, got %+v", result)
	}
	if result.Err == nil || result.Err.Error() != "db down" {
		t.Fatalf("expected 'db down' on Err, got %v", result.Err)
	}
}

// --- fake ApplyReactionRepo ---

type fakeApplyReactionRepo struct {
	jumps       map[string]game.JumpSnapshot
	rounds      map[string]game.RoundSnapshot
	stamps      map[string]game.StampSnapshot
	players     map[string]game.PlayerSnapshot
	reactions   []game.ReactionSnapshot
	getJumpErr  error
	getRoundErr error
	getStampErr error
	createErr   error
}

func newFakeApplyReactionRepo() *fakeApplyReactionRepo {
	return &fakeApplyReactionRepo{
		jumps:   make(map[string]game.JumpSnapshot),
		rounds:  make(map[string]game.RoundSnapshot),
		stamps:  make(map[string]game.StampSnapshot),
		players: make(map[string]game.PlayerSnapshot),
	}
}

func (f *fakeApplyReactionRepo) GetJump(_ context.Context, jumpID string) (game.JumpSnapshot, error) {
	if f.getJumpErr != nil {
		return game.JumpSnapshot{}, f.getJumpErr
	}
	j, ok := f.jumps[jumpID]
	if !ok {
		return game.JumpSnapshot{}, game.ErrJumpNotFound
	}
	return j, nil
}

func (f *fakeApplyReactionRepo) GetRound(_ context.Context, roundID string) (game.RoundSnapshot, bool, error) {
	if f.getRoundErr != nil {
		return game.RoundSnapshot{}, false, f.getRoundErr
	}
	r, ok := f.rounds[roundID]
	return r, ok, nil
}

func (f *fakeApplyReactionRepo) GetStamp(_ context.Context, stampID string) (game.StampSnapshot, error) {
	if f.getStampErr != nil {
		return game.StampSnapshot{}, f.getStampErr
	}
	s, ok := f.stamps[stampID]
	if !ok {
		return game.StampSnapshot{}, game.ErrStampNotFound
	}
	return s, nil
}

func (f *fakeApplyReactionRepo) FindPlayer(_ context.Context, playerID string) (game.PlayerSnapshot, bool, error) {
	p, ok := f.players[playerID]
	return p, ok, nil
}

func (f *fakeApplyReactionRepo) CreateReaction(_ context.Context, r game.ReactionSnapshot) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.reactions = append(f.reactions, r)
	return nil
}

func TestApplyReactionAppliesStampToRevealedJump(t *testing.T) {
	ctx := context.Background()
	repo := newFakeApplyReactionRepo()
	repo.jumps["jump-1"] = game.JumpSnapshot{ID: "jump-1", RoundID: "round-1"}
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "revealed"}
	repo.stamps["stamp-1"] = stamp("stamp-1", "approval", "Approve", "✅", "Yes.")
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}

	result, err := game.ApplyReaction(ctx, repo, game.ApplyReactionInput{
		JumpID:   "jump-1",
		StampID:  "stamp-1",
		PlayerID: "player-a",
	}, reactionFrozenNow)

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if result.Reaction.StampID != "stamp-1" {
		t.Fatalf("expected StampID stamp-1, got %s", result.Reaction.StampID)
	}
	if result.Reaction.JumpID != "jump-1" {
		t.Fatalf("expected JumpID jump-1, got %s", result.Reaction.JumpID)
	}
	if result.Reaction.PlayerID != "player-a" {
		t.Fatalf("expected PlayerID player-a, got %s", result.Reaction.PlayerID)
	}
	if result.Reaction.CreatedAt != reactionFrozenNow {
		t.Fatalf("expected CreatedAt %v, got %v", reactionFrozenNow, result.Reaction.CreatedAt)
	}
	if len(repo.reactions) != 1 {
		t.Fatalf("expected 1 reaction created, got %d", len(repo.reactions))
	}
}

func TestApplyReactionFailsOnNonRevealedRound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeApplyReactionRepo()
	repo.jumps["jump-1"] = game.JumpSnapshot{ID: "jump-1", RoundID: "round-1"}
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "active"}
	repo.stamps["stamp-1"] = stamp("stamp-1", "approval", "Approve", "✅", "Yes.")
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}

	result, err := game.ApplyReaction(ctx, repo, game.ApplyReactionInput{
		JumpID:   "jump-1",
		StampID:  "stamp-1",
		PlayerID: "player-a",
	}, reactionFrozenNow)

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when round not revealed")
	}
	if !errors.Is(result.Err, game.ErrRoundNotRevealed) {
		t.Fatalf("expected ErrRoundNotRevealed, got %v", result.Err)
	}
}

func TestApplyReactionFailsOnUnknownStamp(t *testing.T) {
	ctx := context.Background()
	repo := newFakeApplyReactionRepo()
	repo.jumps["jump-1"] = game.JumpSnapshot{ID: "jump-1", RoundID: "round-1"}
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "revealed"}
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}

	result, err := game.ApplyReaction(ctx, repo, game.ApplyReactionInput{
		JumpID:   "jump-1",
		StampID:  "stamp-missing",
		PlayerID: "player-a",
	}, reactionFrozenNow)

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when stamp not found")
	}
	if !errors.Is(result.Err, game.ErrStampNotFound) {
		t.Fatalf("expected ErrStampNotFound, got %v", result.Err)
	}
}

func TestApplyReactionAllowsRepeatReactions(t *testing.T) {
	ctx := context.Background()
	repo := newFakeApplyReactionRepo()
	repo.jumps["jump-1"] = game.JumpSnapshot{ID: "jump-1", RoundID: "round-1"}
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "revealed"}
	repo.stamps["stamp-1"] = stamp("stamp-1", "approval", "Approve", "✅", "Yes.")
	repo.stamps["stamp-2"] = stamp("stamp-2", "chaos", "Chaos", "🌀", "Unsupervised.")
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}

	r1, err := game.ApplyReaction(ctx, repo, game.ApplyReactionInput{
		JumpID:   "jump-1",
		StampID:  "stamp-1",
		PlayerID: "player-a",
	}, reactionFrozenNow)

	if err != nil || !r1.Allowed {
		t.Fatalf("first should succeed, got err=%v", r1.Err)
	}

	r2, err := game.ApplyReaction(ctx, repo, game.ApplyReactionInput{
		JumpID:   "jump-1",
		StampID:  "stamp-2",
		PlayerID: "player-a",
	}, reactionFrozenNow)

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !r2.Allowed {
		t.Fatalf("expected second to succeed, got err=%v", r2.Err)
	}
	if len(repo.reactions) != 2 {
		t.Fatalf("expected 2 reactions, got %d", len(repo.reactions))
	}
}

func TestApplyReactionFailsOnUnknownJump(t *testing.T) {
	ctx := context.Background()
	repo := newFakeApplyReactionRepo()
	repo.stamps["stamp-1"] = stamp("stamp-1", "approval", "Approve", "✅", "Yes.")
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}

	result, err := game.ApplyReaction(ctx, repo, game.ApplyReactionInput{
		JumpID:   "jump-missing",
		StampID:  "stamp-1",
		PlayerID: "player-a",
	}, reactionFrozenNow)

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when jump not found")
	}
	if !errors.Is(result.Err, game.ErrJumpNotFound) {
		t.Fatalf("expected ErrJumpNotFound, got %v", result.Err)
	}
}

func TestApplyReactionFailsOnUnknownPlayer(t *testing.T) {
	ctx := context.Background()
	repo := newFakeApplyReactionRepo()
	repo.jumps["jump-1"] = game.JumpSnapshot{ID: "jump-1", RoundID: "round-1"}
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "revealed"}
	repo.stamps["stamp-1"] = stamp("stamp-1", "approval", "Approve", "✅", "Yes.")

	result, err := game.ApplyReaction(ctx, repo, game.ApplyReactionInput{
		JumpID:   "jump-1",
		StampID:  "stamp-1",
		PlayerID: "player-missing",
	}, reactionFrozenNow)

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when player not found")
	}
	if !errors.Is(result.Err, game.ErrPlayerNotFound) {
		t.Fatalf("expected ErrPlayerNotFound, got %v", result.Err)
	}
}
