package game

import (
	"context"
	"errors"
	"time"
)

var (
	ErrPromptNotFound      = errors.New("prompt not found")
	ErrNoPromptsAvailable  = errors.New("no prompts available")
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

type ListCatalogResult struct {
	Catalog CatalogSnapshot
	Allowed bool
	Err     error
}

type SelectPromptInput struct {
	PromptID string
}

type SelectPromptResult struct {
	Prompt  PromptSnapshot
	Allowed bool
	Err     error
}

type SelectRandomPromptResult struct {
	Prompt  PromptSnapshot
	Allowed bool
	Err     error
}

type ListCatalogRepo interface {
	ListPromptPacks(ctx context.Context) ([]PromptPackSnapshot, error)
	ListPrompts(ctx context.Context) ([]PromptSnapshot, error)
	GetPrompt(ctx context.Context, id string) (PromptSnapshot, error)
}

func ListCatalog(ctx context.Context, repo ListCatalogRepo) (ListCatalogResult, error) {
	packs, err := repo.ListPromptPacks(ctx)
	if err != nil {
		return ListCatalogResult{Allowed: false, Err: err}, nil
	}

	prompts, err := repo.ListPrompts(ctx)
	if err != nil {
		return ListCatalogResult{Allowed: false, Err: err}, nil
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

	return ListCatalogResult{Catalog: result, Allowed: true}, nil
}

func SelectPrompt(ctx context.Context, repo ListCatalogRepo, input SelectPromptInput) (SelectPromptResult, error) {
	if input.PromptID == "" {
		return SelectPromptResult{Allowed: false, Err: ErrPromptNotFound}, nil
	}
	p, err := repo.GetPrompt(ctx, input.PromptID)
	if err != nil {
		if errors.Is(err, ErrPromptNotFound) {
			return SelectPromptResult{Allowed: false, Err: ErrPromptNotFound}, nil
		}
		return SelectPromptResult{Allowed: false, Err: err}, nil
	}
	return SelectPromptResult{Prompt: p, Allowed: true}, nil
}

func SelectRandomPrompt(ctx context.Context, repo ListCatalogRepo, pickIndex func(n int) int) (SelectRandomPromptResult, error) {
	prompts, err := repo.ListPrompts(ctx)
	if err != nil {
		return SelectRandomPromptResult{Allowed: false, Err: err}, nil
	}
	if len(prompts) == 0 {
		return SelectRandomPromptResult{Allowed: false, Err: ErrNoPromptsAvailable}, nil
	}
	return SelectRandomPromptResult{Prompt: prompts[pickIndex(len(prompts))], Allowed: true}, nil
}
