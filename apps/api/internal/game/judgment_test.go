package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockJudgmentRepo struct {
	jumpFn            func(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
	seasonFn          func(ctx context.Context, seasonID string) (SeasonSnapshot, error)
	upsertJudgmentFn  func(ctx context.Context, jumpID, playerID string, commitment, transgression, creativity, presentation int) (Judgment, bool, error)
	advanceJudgedFn   func(ctx context.Context, jumpID string) error
}

func (m *mockJudgmentRepo) Jump(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
	return m.jumpFn(nil, jumpID)
}

func (m *mockJudgmentRepo) Season(_ context.Context, seasonID string) (SeasonSnapshot, error) {
	return m.seasonFn(nil, seasonID)
}

func (m *mockJudgmentRepo) UpsertJudgment(_ context.Context, jumpID, playerID string, commitment, transgression, creativity, presentation int) (Judgment, bool, error) {
	return m.upsertJudgmentFn(nil, jumpID, playerID, commitment, transgression, creativity, presentation)
}

func (m *mockJudgmentRepo) AdvanceJumpToJudged(_ context.Context, jumpID string) error {
	return m.advanceJudgedFn(nil, jumpID)
}

func TestValidScore_AcceptsBoundaryValues(t *testing.T) {
	cases := []struct {
		score int
		want  bool
	}{
		{0, false},
		{1, true},
		{2, true},
		{3, true},
		{4, true},
		{5, false},
		{-1, false},
		{10, false},
	}
	for _, tc := range cases {
		got := validScore(tc.score)
		if got != tc.want {
			t.Errorf("validScore(%d) = %v; want %v", tc.score, got, tc.want)
		}
	}
}

func TestValidScore_AllFourScoresMustBeValid(t *testing.T) {
	if !validScore(1) {
		t.Error("validScore(1) should be true (minimum valid)")
	}
	if !validScore(4) {
		t.Error("validScore(4) should be true (maximum valid)")
	}
	if !validScore(2) {
		t.Error("validScore(2) should be true")
	}
	if !validScore(3) {
		t.Error("validScore(3) should be true")
	}
}

func TestNonMemberCanJudgePublicPerformedJumpAfterGracePeriod(t *testing.T) {
	var advancedJumpID string
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                  jumpID,
				GroupID:             "group_1",
				PlayerID:            "performer_1",
				Status:              "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, true, nil
		},
		upsertJudgmentFn: func(_ context.Context, jumpID, playerID string, commitment, transgression, creativity, presentation int) (Judgment, bool, error) {
			return Judgment{
				ID:            "judgment_abc",
				JumpID:        jumpID,
				PlayerID:      playerID,
				Commitment:    commitment,
				Transgression: transgression,
				Creativity:    creativity,
				Presentation:  presentation,
			}, true, nil
		},
		advanceJudgedFn: func(_ context.Context, jumpID string) error {
			advancedJumpID = jumpID
			return nil
		},
	}

	result := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "nonmember_judge_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    4,
		Presentation:  4,
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected non-member to be allowed to judge a public performed jump")
	}
	if !result.Created {
		t.Fatal("expected judgment to be created (not an edit)")
	}
	if result.Judgment.Commitment != 3 || result.Judgment.Transgression != 3 || result.Judgment.Creativity != 4 || result.Judgment.Presentation != 4 {
		t.Fatalf("expected judgment with correct scores, got %#v", result.Judgment)
	}
	if advancedJumpID != "jump_1" {
		t.Fatal("expected jump to be advanced to Judged Jump on first judgment")
	}
}

func TestAuthorGracePeriodBlocksJudging(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 12, 10, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                  jumpID,
				GroupID:             "group_1",
				PlayerID:            "performer_1",
				Status:              "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, true, nil
		},
	}

	result := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    4,
		Presentation:  4,
	}, now)

	if !errors.Is(result.Err, ErrAuthorGracePeriodActive) {
		t.Fatalf("expected ErrAuthorGracePeriodActive, got %v", result.Err)
	}
}

func TestJudgingAllowedAfterGracePeriodExpires(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 59, 59, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                  jumpID,
				GroupID:             "group_1",
				PlayerID:            "performer_1",
				Status:              "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, true, nil
		},
		upsertJudgmentFn: func(_ context.Context, jumpID, playerID string, commitment, transgression, creativity, presentation int) (Judgment, bool, error) {
			return Judgment{ID: "judgment_abc", JumpID: jumpID, PlayerID: playerID, Commitment: commitment, Transgression: transgression, Creativity: creativity, Presentation: presentation}, true, nil
		},
		advanceJudgedFn: func(_ context.Context, jumpID string) error {
			return nil
		},
	}

	result := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    3,
		Presentation:  3,
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error after grace period, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected judgment to be allowed after grace period")
	}
}

func TestSelfJudgingStillBlocked(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                  jumpID,
				GroupID:             "group_1",
				PlayerID:            "performer_1",
				Status:              "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, true, nil
		},
	}

	result := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "performer_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    4,
		Presentation:  4,
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error from self-judging, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected self-judging to be denied")
	}
}

func TestInvalidJudgmentScoreRejected(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{}

	result := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_1",
		Commitment:    11,
		Transgression: 3,
		Creativity:    3,
		Presentation:  3,
	}, now)

	if !errors.Is(result.Err, ErrInvalidJudgmentScore) {
		t.Fatalf("expected ErrInvalidJudgmentScore for out-of-range score, got %v", result.Err)
	}
}
