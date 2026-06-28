package game_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

type fakeListCatalogRepo struct {
	packs   []game.PromptPackSnapshot
	prompts []game.PromptSnapshot
}

func newFakeListCatalogRepo() *fakeListCatalogRepo {
	return &fakeListCatalogRepo{}
}

func (f *fakeListCatalogRepo) ListPromptPacks(_ context.Context) ([]game.PromptPackSnapshot, error) {
	return f.packs, nil
}

func (f *fakeListCatalogRepo) ListPrompts(_ context.Context) ([]game.PromptSnapshot, error) {
	return f.prompts, nil
}

func pack(id, name, desc string) game.PromptPackSnapshot {
	return game.PromptPackSnapshot{
		ID:          id,
		DisplayName: name,
		Description: desc,
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func prompt(id, packID, copy, theme, costTier string) game.PromptSnapshot {
	return game.PromptSnapshot{
		ID:        id,
		PackID:    packID,
		Copy:      copy,
		Theme:     theme,
		CostTier:  costTier,
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestListCatalogReturnsPacksWithPrompts(t *testing.T) {
	repo := newFakeListCatalogRepo()
	repo.packs = []game.PromptPackSnapshot{
		pack("pack-1", "Kitchen Classics", "Fridge and pantry bits"),
		pack("pack-2", "Field Ops", "Real-world audacity"),
	}
	repo.prompts = []game.PromptSnapshot{
		prompt("prompt-1", "pack-1", "Turn your fridge into fine dining.", "Fine Dining", "tier_1"),
		prompt("prompt-2", "pack-1", "Wrong room your meal.", "Wrong Room", "tier_1"),
		prompt("prompt-3", "pack-2", "Across enemy lines.", "Enemy Lines", "tier_3"),
	}

	catalog, err := game.ListCatalog(context.Background(), repo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(catalog.Packs) != 2 {
		t.Fatalf("expected 2 packs, got %d", len(catalog.Packs))
	}

	pack1 := catalog.Packs[0]
	if pack1.Pack.ID != "pack-1" {
		t.Fatalf("expected first pack ID 'pack-1', got %q", pack1.Pack.ID)
	}
	if len(pack1.Prompts) != 2 {
		t.Fatalf("expected 2 prompts in pack-1, got %d", len(pack1.Prompts))
	}

	pack2 := catalog.Packs[1]
	if pack2.Pack.ID != "pack-2" {
		t.Fatalf("expected second pack ID 'pack-2', got %q", pack2.Pack.ID)
	}
	if len(pack2.Prompts) != 1 {
		t.Fatalf("expected 1 prompt in pack-2, got %d", len(pack2.Prompts))
	}
}

func TestListCatalogEachPromptInExactlyOnePack(t *testing.T) {
	repo := newFakeListCatalogRepo()
	repo.packs = []game.PromptPackSnapshot{
		pack("pack-1", "Kitchen Classics", "Fridge bits"),
		pack("pack-2", "Field Ops", "Audacity"),
	}
	repo.prompts = []game.PromptSnapshot{
		prompt("a", "pack-1", "Fine dining.", "Fine Dining", "tier_1"),
		prompt("b", "pack-1", "Wrong room.", "Wrong Room", "tier_1"),
		prompt("c", "pack-2", "Enemy lines.", "Enemy Lines", "tier_3"),
	}

	catalog, err := game.ListCatalog(context.Background(), repo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	promptSeen := make(map[string]int)
	for _, pw := range catalog.Packs {
		for _, pr := range pw.Prompts {
			promptSeen[pr.ID]++
		}
	}
	for id, count := range promptSeen {
		if count != 1 {
			t.Fatalf("prompt %s appeared in %d packs, expected exactly 1", id, count)
		}
	}
	if len(promptSeen) != 3 {
		t.Fatalf("expected 3 unique prompts, got %d", len(promptSeen))
	}
}

func TestListCatalogDropsPromptsWhosePackIsNotInCatalog(t *testing.T) {
	repo := newFakeListCatalogRepo()
	repo.packs = []game.PromptPackSnapshot{
		pack("pack-1", "Kitchen Classics", "Fridge bits"),
	}
	repo.prompts = []game.PromptSnapshot{
		prompt("a", "pack-1", "Fine dining.", "Fine Dining", "tier_1"),
		prompt("orphan", "pack-removed", "Stale prompt.", "Stale", "tier_1"),
	}

	catalog, err := game.ListCatalog(context.Background(), repo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(catalog.Packs) != 1 {
		t.Fatalf("expected 1 pack, got %d", len(catalog.Packs))
	}
	if len(catalog.Packs[0].Prompts) != 1 {
		t.Fatalf("expected 1 reachable prompt, got %d", len(catalog.Packs[0].Prompts))
	}
	if catalog.Packs[0].Prompts[0].ID != "a" {
		t.Fatalf("expected reachable prompt 'a', got %q", catalog.Packs[0].Prompts[0].ID)
	}
}

func TestListCatalogEmptyPacksReturnedWithoutPrompts(t *testing.T) {
	repo := newFakeListCatalogRepo()
	repo.packs = []game.PromptPackSnapshot{
		pack("pack-empty", "Empty Pack", "No prompts"),
	}

	catalog, err := game.ListCatalog(context.Background(), repo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(catalog.Packs) != 1 {
		t.Fatalf("expected 1 pack, got %d", len(catalog.Packs))
	}
	if len(catalog.Packs[0].Prompts) != 0 {
		t.Fatalf("expected 0 prompts in empty pack, got %d", len(catalog.Packs[0].Prompts))
	}
}

func TestListCatalogEmptyRepoReturnsEmptyCatalog(t *testing.T) {
	repo := newFakeListCatalogRepo()

	catalog, err := game.ListCatalog(context.Background(), repo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(catalog.Packs) != 0 {
		t.Fatalf("expected 0 packs, got %d", len(catalog.Packs))
	}
}

func TestListCatalogPackErrorPropagated(t *testing.T) {
	broken := &brokenCatalogRepo{packErr: errors.New("db down")}

	_, err := game.ListCatalog(context.Background(), broken)
	if err == nil {
		t.Fatal("expected error from broken pack repo, got nil")
	}
}

func TestListCatalogPromptErrorPropagated(t *testing.T) {
	broken := &brokenCatalogRepo{promptErr: errors.New("db down")}

	catalog, err := game.ListCatalog(context.Background(), broken)
	if err == nil {
		t.Fatalf("expected error from broken prompt repo, got catalog %+v", catalog)
	}
}

type brokenCatalogRepo struct {
	packErr   error
	promptErr error
}

func (b *brokenCatalogRepo) ListPromptPacks(_ context.Context) ([]game.PromptPackSnapshot, error) {
	if b.packErr != nil {
		return nil, b.packErr
	}
	return []game.PromptPackSnapshot{pack("pack-1", "Test", "desc")}, nil
}

func (b *brokenCatalogRepo) ListPrompts(_ context.Context) ([]game.PromptSnapshot, error) {
	if b.promptErr != nil {
		return nil, b.promptErr
	}
	return []game.PromptSnapshot{}, nil
}
