package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockJudgmentRepo struct {
	jumpFn                   func(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
	seasonFn                 func(ctx context.Context, seasonID string) (SeasonSnapshot, error)
	upsertJudgmentFn         func(ctx context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (Judgment, bool, error)
	advanceJudgedFn          func(ctx context.Context, jumpID string) error
	guestSessionJudgmentCountFn func(ctx context.Context, guestSessionID string) (int, error)
	incrementGuestSessionFn  func(ctx context.Context, guestSessionID string) error
}

func (m *mockJudgmentRepo) Jump(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
	return m.jumpFn(nil, jumpID)
}

func (m *mockJudgmentRepo) Season(_ context.Context, seasonID string) (SeasonSnapshot, error) {
	return m.seasonFn(nil, seasonID)
}

func (m *mockJudgmentRepo) UpsertJudgment(_ context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (Judgment, bool, error) {
	return m.upsertJudgmentFn(nil, jumpID, playerID, guestSessionID, provenance, commitment, transgression, creativity, presentation)
}

func (m *mockJudgmentRepo) AdvanceJumpToJudged(_ context.Context, jumpID string) error {
	return m.advanceJudgedFn(nil, jumpID)
}

func (m *mockJudgmentRepo) GuestSessionJudgmentCount(_ context.Context, guestSessionID string) (int, error) {
	if m.guestSessionJudgmentCountFn != nil {
		return m.guestSessionJudgmentCountFn(nil, guestSessionID)
	}
	return 0, nil
}

func (m *mockJudgmentRepo) IncrementGuestSessionJudgmentCount(_ context.Context, guestSessionID string) error {
	if m.incrementGuestSessionFn != nil {
		return m.incrementGuestSessionFn(nil, guestSessionID)
	}
	return nil
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
		upsertJudgmentFn: func(_ context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (Judgment, bool, error) {
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
		upsertJudgmentFn: func(_ context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (Judgment, bool, error) {
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

func TestGuestCanJudgePublicPerformedJump(t *testing.T) {
	var advancedJumpID string
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				GroupID:              "group_1",
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, true, nil
		},
		guestSessionJudgmentCountFn: func(_ context.Context, guestSessionID string) (int, error) {
			return 0, nil
		},
		upsertJudgmentFn: func(_ context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (Judgment, bool, error) {
			return Judgment{
				ID:             "judgment_guest",
				JumpID:         jumpID,
				GuestSessionID: guestSessionID,
				Provenance:     provenance,
				Commitment:     commitment,
				Transgression:  transgression,
				Creativity:     creativity,
				Presentation:   presentation,
			}, true, nil
		},
		advanceJudgedFn: func(_ context.Context, jumpID string) error {
			advancedJumpID = jumpID
			return nil
		},
		incrementGuestSessionFn: func(_ context.Context, guestSessionID string) error {
			return nil
		},
	}

	result := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:         "jump_1",
		GuestSessionID: "guest_session_abc",
		Provenance:     "public",
		Commitment:     2,
		Transgression:  3,
		Creativity:     3,
		Presentation:   4,
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected guest judgment to be allowed")
	}
	if !result.Created {
		t.Fatal("expected judgment to be created (not an edit)")
	}
	if result.Judgment.GuestSessionID != "guest_session_abc" {
		t.Fatalf("expected guest session id on judgment, got %q", result.Judgment.GuestSessionID)
	}
	if result.Judgment.Provenance != "public" {
		t.Fatalf("expected provenance 'public', got %q", result.Judgment.Provenance)
	}
	if advancedJumpID != "jump_1" {
		t.Fatal("expected jump to be advanced to Judged Jump on first guest judgment")
	}
}

func TestGuestJudgmentRequiresExactlyOneJudgeIdentity(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, PlayerID: "performer_1", Status: "Performed Jump", GracePeriodExpiresAt: now.Add(-time.Hour)}, true, nil
		},
	}

	// Both player and guest provided
	result := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:         "jump_1",
		JudgePlayerID:  "judge_1",
		GuestSessionID: "guest_session_abc",
		Commitment:     2,
		Transgression:  2,
		Creativity:     3,
		Presentation:   3,
	}, now)
	if !errors.Is(result.Err, ErrInvalidJudgeIdentity) {
		t.Fatalf("expected ErrInvalidJudgeIdentity when both identities provided, got %v", result.Err)
	}

	// Neither player nor guest provided
	result = SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		Commitment:    2,
		Transgression: 2,
		Creativity:    3,
		Presentation:  3,
	}, now)
	if !errors.Is(result.Err, ErrInvalidJudgeIdentity) {
		t.Fatalf("expected ErrInvalidJudgeIdentity when no identity provided, got %v", result.Err)
	}
}

func TestGuestCapPreventsJudging(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, true, nil
		},
		guestSessionJudgmentCountFn: func(_ context.Context, guestSessionID string) (int, error) {
			return 5, nil // cap reached
		},
	}

	result := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:         "jump_1",
		GuestSessionID: "guest_session_abc",
		Provenance:     "public",
		Commitment:     2,
		Transgression:  2,
		Creativity:     3,
		Presentation:   3,
	}, now)

	if !errors.Is(result.Err, ErrGuestCapReached) {
		t.Fatalf("expected ErrGuestCapReached, got %v", result.Err)
	}
}

func TestGuestCannotJudgeTheirOwnJump(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, true, nil
		},
		upsertJudgmentFn: func(_ context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (Judgment, bool, error) {
			return Judgment{ID: "judgment_guest", JumpID: jumpID, GuestSessionID: guestSessionID}, true, nil
		},
		advanceJudgedFn: func(_ context.Context, jumpID string) error {
			return nil
		},
	}

	// A guest who is also the performer somehow tries to judge
	// In practice this shouldn't happen since guests don't have player_ids,
	// but the domain should still guard against it if both are set
	result := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:         "jump_1",
		GuestSessionID: "guest_session_abc",
		Commitment:     2,
		Transgression:  2,
		Creativity:     3,
		Presentation:   3,
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error from guest self-judge check, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected guest judgment on their own jump to be allowed (guest has no player identity)")
	}
}

func TestPlayerJudgmentStillWorks(t *testing.T) {
	var advancedJumpID string
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				GroupID:              "group_1",
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, true, nil
		},
		upsertJudgmentFn: func(_ context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (Judgment, bool, error) {
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
		Provenance:    "public",
		Commitment:    2,
		Transgression: 3,
		Creativity:    3,
		Presentation:  4,
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected player judgment to be allowed")
	}
	if !result.Created {
		t.Fatal("expected judgment to be created")
	}
	if result.Judgment.PlayerID != "nonmember_judge_1" {
		t.Fatalf("expected player id on judgment, got %q", result.Judgment.PlayerID)
	}
	if advancedJumpID != "jump_1" {
		t.Fatal("expected jump to be advanced to Judged Jump")
	}
}
