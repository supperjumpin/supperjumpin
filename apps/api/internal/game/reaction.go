package game

import (
	"context"
	"errors"
	"time"
)

var (
	ErrStampNotFound    = errors.New("stamp not found")
	ErrRoundNotRevealed = errors.New("round is not revealed")
)

type StampSnapshot struct {
	ID        string
	Stance    string
	Label     string
	Glyph     string
	Copy      string
	CreatedAt time.Time
}

type ReactionSnapshot struct {
	ID        string
	StampID   string
	JumpID    string
	PlayerID  string
	CreatedAt time.Time
}

type ListStampCatalogResult struct {
	Stamps  []StampSnapshot
	Allowed bool
	Err     error
}

type ApplyReactionInput struct {
	JumpID   string
	StampID  string
	PlayerID string
}

type ApplyReactionResult struct {
	Reaction ReactionSnapshot
	Allowed  bool
	Err      error
}

type ListStampCatalogRepo interface {
	ListStamps(ctx context.Context) ([]StampSnapshot, error)
}

type ApplyReactionRepo interface {
	GetJump(ctx context.Context, jumpID string) (JumpSnapshot, error)
	GetRound(ctx context.Context, roundID string) (RoundSnapshot, bool, error)
	GetStamp(ctx context.Context, stampID string) (StampSnapshot, error)
	FindPlayer(ctx context.Context, playerID string) (PlayerSnapshot, bool, error)
	CreateReaction(ctx context.Context, reaction ReactionSnapshot) error
}

func ListStampCatalog(ctx context.Context, repo ListStampCatalogRepo) (ListStampCatalogResult, error) {
	stamps, err := repo.ListStamps(ctx)
	if err != nil {
		return ListStampCatalogResult{Allowed: false, Err: err}, nil
	}
	return ListStampCatalogResult{Stamps: stamps, Allowed: true}, nil
}

func ApplyReaction(ctx context.Context, repo ApplyReactionRepo, input ApplyReactionInput, now time.Time) (ApplyReactionResult, error) {
	jump, err := repo.GetJump(ctx, input.JumpID)
	if err != nil {
		if errors.Is(err, ErrJumpNotFound) {
			return ApplyReactionResult{Allowed: false, Err: ErrJumpNotFound}, nil
		}
		return ApplyReactionResult{Allowed: false, Err: err}, nil
	}

	round, _, err := repo.GetRound(ctx, jump.RoundID)
	if err != nil {
		return ApplyReactionResult{Allowed: false, Err: err}, nil
	}
	if round.Status != "revealed" {
		return ApplyReactionResult{Allowed: false, Err: ErrRoundNotRevealed}, nil
	}

	_, err = repo.GetStamp(ctx, input.StampID)
	if err != nil {
		if errors.Is(err, ErrStampNotFound) {
			return ApplyReactionResult{Allowed: false, Err: ErrStampNotFound}, nil
		}
		return ApplyReactionResult{Allowed: false, Err: err}, nil
	}

	_, playerExists, err := repo.FindPlayer(ctx, input.PlayerID)
	if err != nil {
		return ApplyReactionResult{Allowed: false, Err: err}, nil
	}
	if !playerExists {
		return ApplyReactionResult{Allowed: false, Err: ErrPlayerNotFound}, nil
	}

	reactionID := domainStableID("reaction", input.JumpID+":"+input.PlayerID+":"+input.StampID)
	reaction := ReactionSnapshot{
		ID:        reactionID,
		StampID:   input.StampID,
		JumpID:    input.JumpID,
		PlayerID:  input.PlayerID,
		CreatedAt: now,
	}

	if err := repo.CreateReaction(ctx, reaction); err != nil {
		return ApplyReactionResult{Allowed: false, Err: err}, nil
	}

	return ApplyReactionResult{Reaction: reaction, Allowed: true}, nil
}
