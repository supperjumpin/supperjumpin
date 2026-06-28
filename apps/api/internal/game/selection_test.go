package game_test

import (
	"context"
	"errors"
	"testing"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

type fakeSelectionRepo struct {
	packs     []game.PromptPackSnapshot
	prompts   []game.PromptSnapshot
	promptByID map[string]game.PromptSnapshot
	getErr    error
}

func newFakeSelectionRepo() *fakeSelectionRepo {
	return &fakeSelectionRepo{}
}

func (f *fakeSelectionRepo) ListPromptPacks(_ context.Context) ([]game.PromptPackSnapshot, error) {
	return f.packs, nil
}

func (f *fakeSelectionRepo) ListPrompts(_ context.Context) ([]game.PromptSnapshot, error) {
	return f.prompts, nil
}

func (f *fakeSelectionRepo) GetPrompt(_ context.Context, id string) (game.PromptSnapshot, error) {
	if f.getErr != nil {
		return game.PromptSnapshot{}, f.getErr
	}
	if p, ok := f.promptByID[id]; ok {
		return p, nil
	}
	return game.PromptSnapshot{}, game.ErrPromptNotFound
}

func TestSelectPromptReturnsPromptForKnownID(t *testing.T) {
	repo := newFakeSelectionRepo()
	repo.promptByID = map[string]game.PromptSnapshot{
		"prompt-1": prompt("prompt-1", "pack-1", "Fridge to fine dining.", "Fine Dining", "tier_1"),
	}

	result, err := game.SelectPrompt(context.Background(), repo, game.SelectPromptInput{PromptID: "prompt-1"})
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got %+v", result)
	}
	if result.Prompt.ID != "prompt-1" {
		t.Fatalf("expected prompt 'prompt-1', got %q", result.Prompt.ID)
	}
	if result.Prompt.Copy != "Fridge to fine dining." {
		t.Fatalf("expected copy to be preserved, got %q", result.Prompt.Copy)
	}
}

func TestSelectPromptReturnsNotFoundForUnknownID(t *testing.T) {
	repo := newFakeSelectionRepo()
	repo.promptByID = map[string]game.PromptSnapshot{}

	result, err := game.SelectPrompt(context.Background(), repo, game.SelectPromptInput{PromptID: "missing"})
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatalf("expected Allowed=false for unknown id, got %+v", result)
	}
	if !errors.Is(result.Err, game.ErrPromptNotFound) {
		t.Fatalf("expected ErrPromptNotFound, got %v", result.Err)
	}
}

func TestSelectPromptRepoErrorPropagated(t *testing.T) {
	repo := newFakeSelectionRepo()
	repo.getErr = errors.New("db down")

	result, err := game.SelectPrompt(context.Background(), repo, game.SelectPromptInput{PromptID: "prompt-1"})
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatalf("expected Allowed=false on repo error, got %+v", result)
	}
	if result.Err == nil || result.Err.Error() != "db down" {
		t.Fatalf("expected 'db down' error, got %v", result.Err)
	}
}

func TestSelectRandomPromptReturnsPickerIndex(t *testing.T) {
	repo := newFakeSelectionRepo()
	repo.prompts = []game.PromptSnapshot{
		prompt("prompt-1", "pack-1", "First.", "Fine Dining", "tier_1"),
		prompt("prompt-2", "pack-1", "Second.", "Fine Dining", "tier_1"),
		prompt("prompt-3", "pack-2", "Third.", "Field Ops", "tier_3"),
	}

	result, err := game.SelectRandomPrompt(context.Background(), repo, func(n int) int { return 1 })
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got %+v", result)
	}
	if result.Prompt.ID != "prompt-2" {
		t.Fatalf("expected picker index 1 → 'prompt-2', got %q", result.Prompt.ID)
	}
}

func TestSelectRandomPromptReturnsErrorWhenEmpty(t *testing.T) {
	repo := newFakeSelectionRepo()
	repo.prompts = []game.PromptSnapshot{}

	result, err := game.SelectRandomPrompt(context.Background(), repo, func(n int) int { return 0 })
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatalf("expected Allowed=false on empty catalog, got %+v", result)
	}
	if !errors.Is(result.Err, game.ErrNoPromptsAvailable) {
		t.Fatalf("expected ErrNoPromptsAvailable, got %v", result.Err)
	}
}

func TestSelectRandomPromptListErrorPropagated(t *testing.T) {
	broken := &brokenSelectionRepo{listErr: errors.New("db down")}

	result, err := game.SelectRandomPrompt(context.Background(), broken, func(n int) int { return 0 })
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatalf("expected Allowed=false on list error, got %+v", result)
	}
	if result.Err == nil || result.Err.Error() != "db down" {
		t.Fatalf("expected 'db down', got %v", result.Err)
	}
}

type brokenSelectionRepo struct {
	listErr error
}

func (b *brokenSelectionRepo) ListPromptPacks(_ context.Context) ([]game.PromptPackSnapshot, error) {
	return nil, nil
}

func (b *brokenSelectionRepo) ListPrompts(_ context.Context) ([]game.PromptSnapshot, error) {
	if b.listErr != nil {
		return nil, b.listErr
	}
	return nil, nil
}

func (b *brokenSelectionRepo) GetPrompt(_ context.Context, _ string) (game.PromptSnapshot, error) {
	return game.PromptSnapshot{}, game.ErrPromptNotFound
}
