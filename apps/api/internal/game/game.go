package game

import (
	"context"
	"time"
)

// Service provides game-rule orchestration, delegating persistence
// to injected adapters.
type Service struct {
	JudgmentRepo JudgmentRepository
	Now          func() time.Time
}

// NewService creates a Game service with the given persistence adapters.
func NewService(judgmentRepo JudgmentRepository) *Service {
	return &Service{
		JudgmentRepo: judgmentRepo,
		Now:          time.Now,
	}
}

// SubmitJudgment evaluates judgment rules and persists the result.
func (s *Service) SubmitJudgment(ctx context.Context, stuntID, judgePlayerID string, difficulty, transgression, creativity, documentation int) JudgmentResult {
	return SubmitJudgment(ctx, s.JudgmentRepo, JudgmentInput{
		StuntID:       stuntID,
		JudgePlayerID: judgePlayerID,
		Difficulty:    difficulty,
		Transgression: transgression,
		Creativity:    creativity,
		Documentation: documentation,
	}, s.Now())
}
