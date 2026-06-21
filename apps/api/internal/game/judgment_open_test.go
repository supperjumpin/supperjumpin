package game

import (
	"context"
	"testing"
	"time"
)

func TestSubmitJudgment_OpenAwareness_ActiveOpen(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	expectedOpenID := "open_2026_06"

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -10*time.Minute), true, nil
		},
		hasJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
		activeOpenFn: func(_ context.Context, _ time.Time) (*OpenMonth, error) {
			return &OpenMonth{ID: expectedOpenID}, nil
		},
		submitAcceptedJudgmentFn: func(_ context.Context, input JudgmentInput) (Judgment, error) {
			return Judgment{
				ID:          "j1",
				OpenMonthID: input.OpenMonthID,
			}, nil
		},
	}

	judgment, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    3,
		Presentation:  3,
	}, now)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if judgment.OpenMonthID == nil || *judgment.OpenMonthID != expectedOpenID {
		t.Fatalf("expected OpenMonthID %s, got %v", expectedOpenID, judgment.OpenMonthID)
	}
}

func TestSubmitJudgment_OpenAwareness_NoActiveOpen(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -10*time.Minute), true, nil
		},
		hasJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
		activeOpenFn: func(_ context.Context, _ time.Time) (*OpenMonth, error) {
			return nil, nil // No active open
		},
		submitAcceptedJudgmentFn: func(_ context.Context, input JudgmentInput) (Judgment, error) {
			return Judgment{
				ID:          "j1",
				OpenMonthID: input.OpenMonthID,
			}, nil
		},
	}

	judgment, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    3,
		Presentation:  3,
	}, now)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if judgment.OpenMonthID != nil {
		t.Fatalf("expected OpenMonthID to be nil when no active open exists, got %v", *judgment.OpenMonthID)
	}
}

func TestSubmitJudgment_OpenAwareness_MonthBoundary(t *testing.T) {
	// Case: Submitted at 23:59:59.999 on the last day of the month.
	// The clock should ensure it tags the current month, not the next.
	now := time.Date(2026, 5, 31, 23, 59, 59, 999_000_000, time.UTC)
	expectedOpenID := "open_2026_05"

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -24*time.Hour), true, nil
		},
		hasJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
		activeOpenFn: func(_ context.Context, t time.Time) (*OpenMonth, error) {
			if t.Month() == time.May && t.Year() == 2026 {
				return &OpenMonth{ID: expectedOpenID}, nil
			}
			return nil, nil
		},
		submitAcceptedJudgmentFn: func(_ context.Context, input JudgmentInput) (Judgment, error) {
			return Judgment{
				ID:          "j1",
				OpenMonthID: input.OpenMonthID,
			}, nil
		},
	}

	judgment, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    3,
		Presentation:  3,
	}, now)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if judgment.OpenMonthID == nil || *judgment.OpenMonthID != expectedOpenID {
		t.Fatalf("expected OpenMonthID %s for boundary submission, got %v", expectedOpenID, judgment.OpenMonthID)
	}
}
