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
func (s *Service) SubmitJudgment(ctx context.Context, jumpID, judgePlayerID string, difficulty, transgression, creativity, presentation int) JudgmentResult {
	return SubmitJudgment(ctx, s.JudgmentRepo, JudgmentInput{
		JumpID:        jumpID,
		JudgePlayerID: judgePlayerID,
		Difficulty:    difficulty,
		Transgression: transgression,
		Creativity:    creativity,
		Presentation:  presentation,
	}, s.Now())
}
