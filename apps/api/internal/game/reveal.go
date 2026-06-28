package game

import (
	"context"
	"time"
)

type EvaluateRevealResult struct {
	Round    RoundSnapshot
	Revealed bool
	Allowed  bool
	Err      error
}

type RevealRepo interface {
	GetRound(ctx context.Context, roundID string) (RoundSnapshot, bool, error)
	UpdateRoundStatus(ctx context.Context, roundID, status string) error
}

func EvaluateReveal(ctx context.Context, repo RevealRepo, roundID string, now time.Time) (EvaluateRevealResult, error) {
	round, ok, err := repo.GetRound(ctx, roundID)
	if err != nil {
		return EvaluateRevealResult{Allowed: false, Err: err}, nil
	}
	if !ok {
		return EvaluateRevealResult{Allowed: false, Err: ErrRoundNotFound}, nil
	}

	if round.Status == "revealed" {
		return EvaluateRevealResult{Round: round, Revealed: true, Allowed: true}, nil
	}

	if now.Before(round.RevealBy) {
		return EvaluateRevealResult{Round: round, Revealed: false, Allowed: true}, nil
	}

	if err := repo.UpdateRoundStatus(ctx, roundID, "revealed"); err != nil {
		return EvaluateRevealResult{Allowed: false, Err: err}, nil
	}

	round.Status = "revealed"
	return EvaluateRevealResult{Round: round, Revealed: true, Allowed: true}, nil
}
