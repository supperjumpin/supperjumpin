package game

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"
)

var (
	ErrRoundAlreadyActive       = errors.New("community already has an active round")
	ErrRevealTimeframeNotFound  = errors.New("reveal timeframe not found")
	ErrPlayerNotFound           = errors.New("player not found")
	ErrCommunityNotFound        = errors.New("community not found")
)

type RevealTimeframeSnapshot struct {
	ID            string
	Label         string
	DurationHours int
}

type RoundSnapshot struct {
	ID          string
	CommunityID string
	PromptID    string
	Status      string
	RevealBy    time.Time
	CreatedBy   string
	CreatedAt   time.Time
}

type ListRevealTimeframesResult struct {
	Timeframes []RevealTimeframeSnapshot
	Allowed    bool
	Err        error
}

type StartRoundInput struct {
	CommunityID       string
	PlayerID          string
	PromptID          string
	RevealTimeframeID string
}

type StartRoundResult struct {
	Round   RoundSnapshot
	Allowed bool
	Err     error
}

type ListRevealTimeframesRepo interface {
	ListRevealTimeframes(ctx context.Context) ([]RevealTimeframeSnapshot, error)
}

type StartRoundRepo interface {
	FindPlayer(ctx context.Context, id string) (PlayerSnapshot, bool, error)
	FindCommunity(ctx context.Context, id string) (CommunitySnapshot, bool, error)
	FindActiveRound(ctx context.Context, communityID string) (*RoundSnapshot, error)
	GetPrompt(ctx context.Context, id string) (PromptSnapshot, error)
	ListPrompts(ctx context.Context) ([]PromptSnapshot, error)
	GetRevealTimeframe(ctx context.Context, id string) (RevealTimeframeSnapshot, error)
	CreateRound(ctx context.Context, round RoundSnapshot) error
}

func ListRevealTimeframes(ctx context.Context, repo ListRevealTimeframesRepo) (ListRevealTimeframesResult, error) {
	tfs, err := repo.ListRevealTimeframes(ctx)
	if err != nil {
		return ListRevealTimeframesResult{Allowed: false, Err: err}, nil
	}
	return ListRevealTimeframesResult{Timeframes: tfs, Allowed: true}, nil
}

func StartRound(ctx context.Context, repo StartRoundRepo, input StartRoundInput, now time.Time, randomPicker ...func(int) int) (StartRoundResult, error) {
	_, playerExists, err := repo.FindPlayer(ctx, input.PlayerID)
	if err != nil {
		return StartRoundResult{Allowed: false, Err: err}, nil
	}
	if !playerExists {
		return StartRoundResult{Allowed: false, Err: ErrPlayerNotFound}, nil
	}

	_, communityExists, err := repo.FindCommunity(ctx, input.CommunityID)
	if err != nil {
		return StartRoundResult{Allowed: false, Err: err}, nil
	}
	if !communityExists {
		return StartRoundResult{Allowed: false, Err: ErrCommunityNotFound}, nil
	}

	active, err := repo.FindActiveRound(ctx, input.CommunityID)
	if err != nil {
		return StartRoundResult{Allowed: false, Err: err}, nil
	}
	if active != nil {
		return StartRoundResult{Allowed: false, Err: ErrRoundAlreadyActive}, nil
	}

	prompt, err := resolvePrompt(ctx, repo, input.PromptID, randomPicker...)
	if err != nil {
		return StartRoundResult{Allowed: false, Err: err}, nil
	}

	tf, err := repo.GetRevealTimeframe(ctx, input.RevealTimeframeID)
	if err != nil {
		if errors.Is(err, ErrRevealTimeframeNotFound) {
			return StartRoundResult{Allowed: false, Err: ErrRevealTimeframeNotFound}, nil
		}
		return StartRoundResult{Allowed: false, Err: err}, nil
	}

	revealBy := now.Add(time.Duration(tf.DurationHours) * time.Hour)
	roundID := domainStableID("round", input.CommunityID+":"+prompt.ID+":"+strconv.FormatInt(now.Unix(), 10))
	round := RoundSnapshot{
		ID:          roundID,
		CommunityID: input.CommunityID,
		PromptID:    prompt.ID,
		Status:      "active",
		RevealBy:    revealBy,
		CreatedBy:   input.PlayerID,
		CreatedAt:   now,
	}

	if err := repo.CreateRound(ctx, round); err != nil {
		return StartRoundResult{Allowed: false, Err: err}, nil
	}

	return StartRoundResult{Round: round, Allowed: true}, nil
}

func domainStableID(kind string, value string) string {
	sum := sha256.Sum256([]byte(kind + ":" + value))
	return kind + "_" + hex.EncodeToString(sum[:])[:12]
}

func resolvePrompt(ctx context.Context, repo StartRoundRepo, promptID string, randomPicker ...func(int) int) (PromptSnapshot, error) {
	if promptID != "" {
		p, err := repo.GetPrompt(ctx, promptID)
		if err != nil {
			if errors.Is(err, ErrPromptNotFound) {
				return PromptSnapshot{}, ErrPromptNotFound
			}
			return PromptSnapshot{}, err
		}
		return p, nil
	}

	prompts, err := repo.ListPrompts(ctx)
	if err != nil {
		return PromptSnapshot{}, err
	}
	if len(prompts) == 0 {
		return PromptSnapshot{}, ErrNoPromptsAvailable
	}

	picker := randomPicker[0]
	return prompts[picker(len(prompts))], nil
}
