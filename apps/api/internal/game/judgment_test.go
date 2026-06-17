package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockJudgmentRepo struct {
	jumpFn                      func(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
	seasonFn                    func(ctx context.Context, seasonID string) (SeasonSnapshot, error)
	submitAcceptedJudgmentFn    func(ctx context.Context, input JudgmentInput) (Judgment, error)
	hasJudgedFn                 func(ctx context.Context, jumpID, playerID string) (bool, error)
	hasGuestJudgedFn            func(ctx context.Context, jumpID, guestSessionID string) (bool, error)
	guestSessionJudgmentCountFn func(ctx context.Context, guestSessionID string) (int, error)
	advanceJudgedFn             func(ctx context.Context, jumpID string) error
	incrementGuestSessionFn     func(ctx context.Context, guestSessionID string) error
}

func (m *mockJudgmentRepo) Jump(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
	return m.jumpFn(nil, jumpID)
}

func (m *mockJudgmentRepo) Season(_ context.Context, seasonID string) (SeasonSnapshot, error) {
	if m.seasonFn != nil {
		return m.seasonFn(nil, seasonID)
	}
	return SeasonSnapshot{}, nil
}

func (m *mockJudgmentRepo) SubmitAcceptedJudgment(_ context.Context, input JudgmentInput) (Judgment, error) {
	if m.submitAcceptedJudgmentFn != nil {
		return m.submitAcceptedJudgmentFn(nil, input)
	}

	if input.GuestSessionID != "" {
		if m.guestSessionJudgmentCountFn != nil {
			count, err := m.guestSessionJudgmentCountFn(nil, input.GuestSessionID)
			if err != nil {
				return Judgment{}, err
			}
			if count >= 5 {
				return Judgment{}, ErrGuestCapReached
			}
		}
		if m.incrementGuestSessionFn != nil {
			if err := m.incrementGuestSessionFn(nil, input.GuestSessionID); err != nil {
				return Judgment{}, err
			}
		}
	}

	if m.advanceJudgedFn != nil {
		if err := m.advanceJudgedFn(nil, input.JumpID); err != nil {
			return Judgment{}, err
		}
	}

	return Judgment{
		ID:             "judgment_" + input.JumpID,
		JumpID:         input.JumpID,
		PlayerID:       input.JudgePlayerID,
		GuestSessionID: input.GuestSessionID,
		Provenance:     input.Provenance,
		Commitment:     input.Commitment,
		Transgression:  input.Transgression,
		Creativity:     input.Creativity,
		Presentation:   input.Presentation,
	}, nil
}

func (m *mockJudgmentRepo) HasJudgedJump(_ context.Context, jumpID, playerID string) (bool, error) {
	if m.hasJudgedFn != nil {
		return m.hasJudgedFn(nil, jumpID, playerID)
	}
	return false, nil
}

func (m *mockJudgmentRepo) HasGuestJudgedJump(_ context.Context, jumpID, guestSessionID string) (bool, error) {
	if m.hasGuestJudgedFn != nil {
		return m.hasGuestJudgedFn(nil, jumpID, guestSessionID)
	}
	return false, nil
}

func (m *mockJudgmentRepo) HasJudgedJumps(_ context.Context, playerID string, jumpIDs []string) (map[string]bool, error) {
	return map[string]bool{}, nil
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

func (m *mockJudgmentRepo) CreateGuestSession(_ context.Context, id string) error {
	return nil
}

func performedJump(jumpID string, graceOffset time.Duration) JumpSnapshot {
	return JumpSnapshot{
		ID:                   jumpID,
		PlayerID:             "performer_1",
		Status:               "Performed Jump",
		GracePeriodExpiresAt: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC).Add(graceOffset),
	}
}

func testJudgmentEligibility_SelfJudgingBlocks(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)

	jump := JumpSnapshot{
		ID:                   "jump_1",
		PlayerID:             "performer_1",
		GracePeriodExpiresAt: gracePeriodExpiresAt,
	}

	hint := JudgmentEligibility(jump, "performer_1", false, now)
	if hint.CanJudge {
		t.Fatal("expected self-judging to be blocked")
	}
	if hint.Reason != "self-judging" {
		t.Fatalf("expected reason 'self-judging', got %q", hint.Reason)
	}
}

func testJudgmentEligibility_GracePeriodBlocks(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 12, 10, 0, 0, time.UTC)

	jump := JumpSnapshot{
		ID:                   "jump_1",
		PlayerID:             "performer_1",
		GracePeriodExpiresAt: gracePeriodExpiresAt,
	}

	hint := JudgmentEligibility(jump, "viewer_1", false, now)
	if hint.CanJudge {
		t.Fatal("expected grace-period to block judging")
	}
	if hint.Reason != "grace-period" {
		t.Fatalf("expected reason 'grace-period', got %q", hint.Reason)
	}
	if hint.GracePeriodEndsAt == nil {
		t.Fatal("expected GracePeriodEndsAt to be populated")
	}
	if !hint.GracePeriodEndsAt.Equal(gracePeriodExpiresAt) {
		t.Fatalf("expected GracePeriodEndsAt %v, got %v", gracePeriodExpiresAt, *hint.GracePeriodEndsAt)
	}
}

func testJudgmentEligibility_AlreadyJudgedBlocks(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)

	jump := JumpSnapshot{
		ID:                   "jump_1",
		PlayerID:             "performer_1",
		GracePeriodExpiresAt: gracePeriodExpiresAt,
	}

	hint := JudgmentEligibility(jump, "viewer_1", true, now)
	if hint.CanJudge {
		t.Fatal("expected already-judged to block judging")
	}
	if hint.Reason != "already-judged" {
		t.Fatalf("expected reason 'already-judged', got %q", hint.Reason)
	}
}

func testJudgmentEligibility_EmptyViewerIDIsEligible(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 12, 10, 0, 0, time.UTC)

	jump := JumpSnapshot{
		ID:                   "jump_1",
		PlayerID:             "performer_1",
		GracePeriodExpiresAt: gracePeriodExpiresAt,
	}

	hint := JudgmentEligibility(jump, "", false, now)
	if !hint.CanJudge {
		t.Fatal("expected empty viewerID to be eligible")
	}
	if hint.Reason != "" {
		t.Fatalf("expected no reason, got %q", hint.Reason)
	}
}

func testJudgmentEligibility_OtherPlayerAfterGracePeriodIsEligible(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)

	jump := JumpSnapshot{
		ID:                   "jump_1",
		PlayerID:             "performer_1",
		GracePeriodExpiresAt: gracePeriodExpiresAt,
	}

	hint := JudgmentEligibility(jump, "viewer_1", false, now)
	if !hint.CanJudge {
		t.Fatal("expected other player after grace period to be eligible")
	}
	if hint.Reason != "" {
		t.Fatalf("expected no reason, got %q", hint.Reason)
	}
}

func testValidScore_AcceptsBoundaryValues(t *testing.T) {
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

func testValidScore_AllFourScoresMustBeValid(t *testing.T) {
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

func testNonMemberCanJudgePublicPerformedJumpAfterGracePeriod(t *testing.T) {
	var advancedJumpID string
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -10*time.Minute), true, nil
		},
		hasJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
		advanceJudgedFn: func(_ context.Context, jumpID string) error {
			advancedJumpID = jumpID
			return nil
		},
	}

	judgment, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "nonmember_judge_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    4,
		Presentation:  4,
	}, now)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if judgment.Commitment != 3 || judgment.Transgression != 3 || judgment.Creativity != 4 || judgment.Presentation != 4 {
		t.Fatalf("expected judgment with correct scores, got %#v", judgment)
	}
	if advancedJumpID != "jump_1" {
		t.Fatal("expected jump to be advanced to Judged Jump on first judgment")
	}
}

func testAuthorGracePeriodBlocksJudging(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, 10*time.Minute), true, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    4,
		Presentation:  4,
	}, now)

	if !errors.Is(err, ErrAuthorGracePeriodActive) {
		t.Fatalf("expected ErrAuthorGracePeriodActive, got %v", err)
	}
}

func testJudgingAllowedAfterGracePeriodExpires(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -1*time.Second), true, nil
		},
		hasJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    3,
		Presentation:  3,
	}, now)

	if err != nil {
		t.Fatalf("expected no error after grace period, got %v", err)
	}
}

func testSelfJudgingStillBlocked(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -10*time.Minute), true, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "performer_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    4,
		Presentation:  4,
	}, now)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden for self-judging, got %v", err)
	}
}

func testInvalidJudgmentScoreRejected(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -10*time.Minute), true, nil
		},
		hasJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_1",
		Commitment:    11,
		Transgression: 3,
		Creativity:    3,
		Presentation:  3,
	}, now)

	if !errors.Is(err, ErrInvalidJudgmentScore) {
		t.Fatalf("expected ErrInvalidJudgmentScore for out-of-range score, got %v", err)
	}
}

func testGuestCanJudgePublicPerformedJump(t *testing.T) {
	var advancedJumpID string
	var incrementedSession string
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -10*time.Minute), true, nil
		},
		hasGuestJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
		guestSessionJudgmentCountFn: func(_ context.Context, _ string) (int, error) {
			return 0, nil
		},
		advanceJudgedFn: func(_ context.Context, jumpID string) error {
			advancedJumpID = jumpID
			return nil
		},
		incrementGuestSessionFn: func(_ context.Context, guestSessionID string) error {
			incrementedSession = guestSessionID
			return nil
		},
	}

	judgment, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:         "jump_1",
		GuestSessionID: "guest_session_abc",
		Provenance:     "public",
		Commitment:     2,
		Transgression:  3,
		Creativity:     3,
		Presentation:   4,
	}, now)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if judgment.GuestSessionID != "guest_session_abc" {
		t.Fatalf("expected guest session id on judgment, got %q", judgment.GuestSessionID)
	}
	if judgment.Provenance != "public" {
		t.Fatalf("expected provenance 'public', got %q", judgment.Provenance)
	}
	if advancedJumpID != "jump_1" {
		t.Fatal("expected jump to be advanced to Judged Jump on first guest judgment")
	}
	if incrementedSession != "guest_session_abc" {
		t.Fatalf("expected guest session cap to be incremented, got %q", incrementedSession)
	}
}

func testGuestJudgmentRequiresExactlyOneJudgeIdentity(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -time.Hour), true, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:         "jump_1",
		JudgePlayerID:  "judge_1",
		GuestSessionID: "guest_session_abc",
		Commitment:     2,
		Transgression:  2,
		Creativity:     3,
		Presentation:   3,
	}, now)
	if !errors.Is(err, ErrInvalidJudgeIdentity) {
		t.Fatalf("expected ErrInvalidJudgeIdentity when both identities provided, got %v", err)
	}

	_, err = SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		Commitment:    2,
		Transgression: 2,
		Creativity:    3,
		Presentation:  3,
	}, now)
	if !errors.Is(err, ErrInvalidJudgeIdentity) {
		t.Fatalf("expected ErrInvalidJudgeIdentity when no identity provided, got %v", err)
	}
}

func testGuestCapPreventsJudging(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -10*time.Minute), true, nil
		},
		hasGuestJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
		guestSessionJudgmentCountFn: func(_ context.Context, _ string) (int, error) {
			return 5, nil // cap reached
		},
		submitAcceptedJudgmentFn: func(_ context.Context, _ JudgmentInput) (Judgment, error) {
			t.Fatal("expected guest cap to block before persistence")
			return Judgment{}, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:         "jump_1",
		GuestSessionID: "guest_session_abc",
		Provenance:     "public",
		Commitment:     2,
		Transgression:  2,
		Creativity:     3,
		Presentation:   3,
	}, now)

	if !errors.Is(err, ErrGuestCapReached) {
		t.Fatalf("expected ErrGuestCapReached, got %v", err)
	}
}

func testGuestCannotJudgeTheirOwnJump(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -10*time.Minute), true, nil
		},
		hasGuestJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
		guestSessionJudgmentCountFn: func(_ context.Context, _ string) (int, error) {
			return 0, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:         "jump_1",
		GuestSessionID: "guest_session_abc",
		Commitment:     2,
		Transgression:  2,
		Creativity:     3,
		Presentation:   3,
	}, now)

	if err != nil {
		t.Fatalf("expected no error from guest judging performer's jump, got %v", err)
	}
}

func testPlayerJudgmentStillWorks(t *testing.T) {
	var advancedJumpID string
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -10*time.Minute), true, nil
		},
		hasJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
		advanceJudgedFn: func(_ context.Context, jumpID string) error {
			advancedJumpID = jumpID
			return nil
		},
	}

	judgment, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "nonmember_judge_1",
		Provenance:    "public",
		Commitment:    2,
		Transgression: 3,
		Creativity:    3,
		Presentation:  4,
	}, now)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if judgment.PlayerID != "nonmember_judge_1" {
		t.Fatalf("expected player id on judgment, got %q", judgment.PlayerID)
	}
	if advancedJumpID != "jump_1" {
		t.Fatal("expected jump to be advanced to Judged Jump")
	}
}

func testIsOpenSeasonStatus_ActiveAndJudgingGracePeriodAreOpen(t *testing.T) {
	if !isOpenSeasonStatus("Active") {
		t.Error("expected Active to be open")
	}
	if !isOpenSeasonStatus("Judging Grace Period") {
		t.Error("expected Judging Grace Period to be open")
	}
}

func testIsOpenSeasonStatus_OtherStatusesAreClosed(t *testing.T) {
	if isOpenSeasonStatus("Finalized") {
		t.Error("expected Finalized to be closed")
	}
	if isOpenSeasonStatus("Closed") {
		t.Error("expected Closed to be closed")
	}
	if isOpenSeasonStatus("") {
		t.Error("expected empty status to be closed")
	}
}

func testSubmitJudgment_SeasonLinkedWithClosedSeasonReturnsError(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	seasonID := "season_1"

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			jump := performedJump(jumpID, -10*time.Minute)
			jump.SeasonID = &seasonID
			return jump, true, nil
		},
		hasJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
		seasonFn: func(_ context.Context, _ string) (SeasonSnapshot, error) {
			return SeasonSnapshot{ID: seasonID, Status: "Finalized"}, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    3,
		Presentation:  3,
	}, now)

	if !errors.Is(err, ErrJudgingWindowClosed) {
		t.Fatalf("expected ErrJudgingWindowClosed for finalized season, got %v", err)
	}
}

func testSubmitJudgment_JudgedJumpAllowsAdditionalJudgments(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Judged Jump",
				GracePeriodExpiresAt: time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC),
			}, true, nil
		},
		hasJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_2",
		Commitment:    3,
		Transgression: 3,
		Creativity:    3,
		Presentation:  3,
	}, now)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func testSubmitJudgment_GuardOrder_MissingJumpBeforeInvalidScores(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, _ string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{}, false, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "missing_jump",
		JudgePlayerID: "judge_1",
		Commitment:    11,
		Transgression: 11,
		Creativity:    11,
		Presentation:  11,
	}, now)

	if !errors.Is(err, ErrJumpNotFound) {
		t.Fatalf("expected ErrJumpNotFound to precede invalid scores, got %v", err)
	}
}

func testSubmitJudgment_GuardOrder_SelfJudgingBeforeGracePeriod(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, 10*time.Minute), true, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "performer_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    3,
		Presentation:  3,
	}, now)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden (self-judging) before grace period, got %v", err)
	}
}

func testSubmitJudgment_GuardOrder_AlreadyJudgedBeforeInvalidScores(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -10*time.Minute), true, nil
		},
		hasJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return true, nil
		},
		submitAcceptedJudgmentFn: func(_ context.Context, _ JudgmentInput) (Judgment, error) {
			t.Fatal("expected already-judged to block before persistence")
			return Judgment{}, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_1",
		Commitment:    11,
		Transgression: 11,
		Creativity:    11,
		Presentation:  11,
	}, now)

	if !errors.Is(err, ErrAlreadyJudged) {
		t.Fatalf("expected ErrAlreadyJudged to precede invalid scores, got %v", err)
	}
}

func testSubmitJudgment_GuardOrder_GuestCapBeforeInvalidScores(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return performedJump(jumpID, -10*time.Minute), true, nil
		},
		hasGuestJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
		guestSessionJudgmentCountFn: func(_ context.Context, _ string) (int, error) {
			return 5, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:         "jump_1",
		GuestSessionID: "guest_session_abc",
		Commitment:     11,
		Transgression:  11,
		Creativity:     11,
		Presentation:   11,
	}, now)

	if !errors.Is(err, ErrGuestCapReached) {
		t.Fatalf("expected ErrGuestCapReached to precede invalid scores, got %v", err)
	}
}

func testSubmitJudgment_Immutability_PlayerDuplicate(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Judged Jump",
				GracePeriodExpiresAt: time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC),
			}, true, nil
		},
		hasJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return true, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    3,
		Presentation:  3,
	}, now)

	if !errors.Is(err, ErrAlreadyJudged) {
		t.Fatalf("expected ErrAlreadyJudged for duplicate player judgment, got %v", err)
	}
}

func testSubmitJudgment_Immutability_GuestDuplicate(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Judged Jump",
				GracePeriodExpiresAt: time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC),
			}, true, nil
		},
		hasGuestJudgedFn: func(_ context.Context, _, _ string) (bool, error) {
			return true, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:         "jump_1",
		GuestSessionID: "guest_session_abc",
		Commitment:     3,
		Transgression:  3,
		Creativity:     3,
		Presentation:   3,
	}, now)

	if !errors.Is(err, ErrAlreadyJudged) {
		t.Fatalf("expected ErrAlreadyJudged for duplicate guest judgment, got %v", err)
	}
}

func testSubmitJudgment_RemovedJumpReturnsWindowClosed(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Removed Jump",
				GracePeriodExpiresAt: time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC),
			}, true, nil
		},
	}

	_, err := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_1",
		Commitment:    3,
		Transgression: 3,
		Creativity:    3,
		Presentation:  3,
	}, now)

	if !errors.Is(err, ErrJudgingWindowClosed) {
		t.Fatalf("expected ErrJudgingWindowClosed for removed jump, got %v", err)
	}
}

func TestJudgment(t *testing.T) {
	t.Run("eligibility", func(t *testing.T) {
		t.Run("self judging blocks", testJudgmentEligibility_SelfJudgingBlocks)
		t.Run("grace period blocks", testJudgmentEligibility_GracePeriodBlocks)
		t.Run("already judged blocks", testJudgmentEligibility_AlreadyJudgedBlocks)
		t.Run("empty viewerID is eligible", testJudgmentEligibility_EmptyViewerIDIsEligible)
		t.Run("other player after grace period is eligible", testJudgmentEligibility_OtherPlayerAfterGracePeriodIsEligible)
	})

	t.Run("score validation", func(t *testing.T) {
		t.Run("accepts boundary values", testValidScore_AcceptsBoundaryValues)
		t.Run("all four scores must be valid", testValidScore_AllFourScoresMustBeValid)
		t.Run("rejects invalid score", testInvalidJudgmentScoreRejected)
	})

	t.Run("performed jump judgment", func(t *testing.T) {
		t.Run("non-member can judge public performed jump after grace period", testNonMemberCanJudgePublicPerformedJumpAfterGracePeriod)
		t.Run("guest can judge public performed jump", testGuestCanJudgePublicPerformedJump)
		t.Run("player judgment still works", testPlayerJudgmentStillWorks)
	})

	t.Run("grace period", func(t *testing.T) {
		t.Run("author grace period blocks judging", testAuthorGracePeriodBlocksJudging)
		t.Run("judging allowed after grace period expires", testJudgingAllowedAfterGracePeriodExpires)
		t.Run("self judging still blocked", testSelfJudgingStillBlocked)
	})

	t.Run("guest identity", func(t *testing.T) {
		t.Run("requires exactly one judge identity", testGuestJudgmentRequiresExactlyOneJudgeIdentity)
		t.Run("guest cap prevents judging", testGuestCapPreventsJudging)
		t.Run("guest cannot judge their own jump", testGuestCannotJudgeTheirOwnJump)
	})

	t.Run("season status", func(t *testing.T) {
		t.Run("active and judging grace period are open", testIsOpenSeasonStatus_ActiveAndJudgingGracePeriodAreOpen)
		t.Run("other statuses are closed", testIsOpenSeasonStatus_OtherStatusesAreClosed)
	})

	t.Run("season-linked jump", func(t *testing.T) {
		t.Run("closed season returns error", testSubmitJudgment_SeasonLinkedWithClosedSeasonReturnsError)
		t.Run("judged jump allows additional judgments", testSubmitJudgment_JudgedJumpAllowsAdditionalJudgments)
	})

	t.Run("guard order", func(t *testing.T) {
		t.Run("missing jump before invalid scores", testSubmitJudgment_GuardOrder_MissingJumpBeforeInvalidScores)
		t.Run("self judging before grace period", testSubmitJudgment_GuardOrder_SelfJudgingBeforeGracePeriod)
		t.Run("already judged before invalid scores", testSubmitJudgment_GuardOrder_AlreadyJudgedBeforeInvalidScores)
		t.Run("guest cap before invalid scores", testSubmitJudgment_GuardOrder_GuestCapBeforeInvalidScores)
	})

	t.Run("immutability", func(t *testing.T) {
		t.Run("player duplicate rejected", testSubmitJudgment_Immutability_PlayerDuplicate)
		t.Run("guest duplicate rejected", testSubmitJudgment_Immutability_GuestDuplicate)
	})

	t.Run("removed jump", func(t *testing.T) {
		t.Run("removed jump returns window closed", testSubmitJudgment_RemovedJumpReturnsWindowClosed)
	})
}
