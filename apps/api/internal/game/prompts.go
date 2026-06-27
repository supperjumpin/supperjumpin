package game

import (
	"context"
	"time"
)

type PromptPackSnapshot struct {
	ID          string
	DisplayName string
	Description string
	CreatedAt   time.Time
}

type PromptSnapshot struct {
	ID        string
	PackID    string
	Copy      string
	Theme     string
	CostTier  string
	CreatedAt time.Time
}

type CatalogSnapshot struct {
	Packs []PackWithPrompts
}

type PackWithPrompts struct {
	Pack    PromptPackSnapshot
	Prompts []PromptSnapshot
}

type ListCatalogRepo interface {
	ListPromptPacks(ctx context.Context) ([]PromptPackSnapshot, error)
	ListPrompts(ctx context.Context) ([]PromptSnapshot, error)
}

func ListCatalog(ctx context.Context, repo ListCatalogRepo) (CatalogSnapshot, error) {
	packs, err := repo.ListPromptPacks(ctx)
	if err != nil {
		return CatalogSnapshot{}, err
	}

	prompts, err := repo.ListPrompts(ctx)
	if err != nil {
		return CatalogSnapshot{}, err
	}

	promptByPack := make(map[string][]PromptSnapshot)
	for _, p := range prompts {
		promptByPack[p.PackID] = append(promptByPack[p.PackID], p)
	}

	result := CatalogSnapshot{
		Packs: make([]PackWithPrompts, 0, len(packs)),
	}
	for _, pack := range packs {
		result.Packs = append(result.Packs, PackWithPrompts{
			Pack:    pack,
			Prompts: promptByPack[pack.ID],
		})
	}

	return result, nil
}
