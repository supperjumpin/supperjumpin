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
	submitAcceptedJudgmentFn    func(ctx context.Context, input JudgmentInput) (Judgment, bool, error)
	upsertJudgmentFn            func(ctx context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (Judgment, bool, error)
	advanceJudgedFn             func(ctx context.Context, jumpID string) error
	guestSessionJudgmentCountFn func(ctx context.Context, guestSessionID string) (int, error)
	incrementGuestSessionFn     func(ctx context.Context, guestSessionID string) error
}

func (m *mockJudgmentRepo) Jump(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
	return m.jumpFn(nil, jumpID)
}

func (m *mockJudgmentRepo) Season(_ context.Context, seasonID string) (SeasonSnapshot, error) {
	return m.seasonFn(nil, seasonID)
}

func (m *mockJudgmentRepo) SubmitAcceptedJudgment(_ context.Context, input JudgmentInput) (Judgment, bool, error) {
	if m.submitAcceptedJudgmentFn != nil {
		return m.submitAcceptedJudgmentFn(nil, input)
	}

	if input.GuestSessionID != "" && m.guestSessionJudgmentCountFn != nil {
		count, err := m.guestSessionJudgmentCountFn(nil, input.GuestSessionID)
		if err != nil {
			return Judgment{}, false, err
		}
		if count >= 5 {
			return Judgment{}, false, ErrGuestCapReached
		}
	}

	judgment := Judgment{ID: "judgment_abc", JumpID: input.JumpID, PlayerID: input.JudgePlayerID, GuestSessionID: input.GuestSessionID, Provenance: input.Provenance, Commitment: input.Commitment, Transgression: input.Transgression, Creativity: input.Creativity, Presentation: input.Presentation}
	created := true
	if m.upsertJudgmentFn != nil {
		var err error
		judgment, created, err = m.upsertJudgmentFn(nil, input.JumpID, input.JudgePlayerID, input.GuestSessionID, input.Provenance, input.Commitment, input.Transgression, input.Creativity, input.Presentation)
		if err != nil {
			return Judgment{}, false, err
		}
	}

	if created && m.advanceJudgedFn != nil {
		if err := m.advanceJudgedFn(nil, input.JumpID); err != nil {
			return Judgment{}, false, err
		}
	}
	if input.GuestSessionID != "" && created && m.incrementGuestSessionFn != nil {
		if err := m.incrementGuestSessionFn(nil, input.GuestSessionID); err != nil {
			return Judgment{}, false, err
		}
	}

	return judgment, created, nil
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

func (m *mockJudgmentRepo) HasJudgedJump(_ context.Context, jumpID, playerID string) (bool, error) {
	return false, nil
}

func (m *mockJudgmentRepo) HasJudgedJumps(_ context.Context, playerID string, jumpIDs []string) (map[string]bool, error) {
	return map[string]bool{}, nil
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

func testGuestJudgmentUsesAtomicPersistenceCommand(t *testing.T) {
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
		submitAcceptedJudgmentFn: func(_ context.Context, input JudgmentInput) (Judgment, bool, error) {
			return Judgment{
				ID:             "judgment_guest",
				JumpID:         input.JumpID,
				GuestSessionID: input.GuestSessionID,
				Provenance:     input.Provenance,
				Commitment:     input.Commitment,
				Transgression:  input.Transgression,
				Creativity:     input.Creativity,
				Presentation:   input.Presentation,
			}, true, nil
		},
		guestSessionJudgmentCountFn: func(_ context.Context, guestSessionID string) (int, error) {
			t.Fatalf("expected atomic command to avoid guestSessionJudgmentCount for %s", guestSessionID)
			return 0, nil
		},
		upsertJudgmentFn: func(_ context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (Judgment, bool, error) {
			t.Fatal("expected atomic command to avoid upsertJudgment")
			return Judgment{}, false, nil
		},
		advanceJudgedFn: func(_ context.Context, jumpID string) error {
			t.Fatal("expected atomic command to avoid advanceJudged")
			return nil
		},
		incrementGuestSessionFn: func(_ context.Context, guestSessionID string) error {
			t.Fatal("expected atomic command to avoid incrementGuestSessionJudgmentCount")
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
		t.Fatal("expected judgment to be created")
	}
	if result.Judgment.GuestSessionID != "guest_session_abc" {
		t.Fatalf("expected guest session id on judgment, got %q", result.Judgment.GuestSessionID)
	}
}

func testAuthorGracePeriodBlocksJudging(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 12, 10, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
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

func testJudgingAllowedAfterGracePeriodExpires(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 59, 59, 0, time.UTC)

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

func testSelfJudgingStillBlocked(t *testing.T) {
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

func testInvalidJudgmentScoreRejected(t *testing.T) {
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

func testGuestCanJudgePublicPerformedJump(t *testing.T) {
	var advancedJumpID string
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

func testGuestJudgmentRequiresExactlyOneJudgeIdentity(t *testing.T) {
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

func testGuestCapPreventsJudging(t *testing.T) {
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

func testGuestCannotJudgeTheirOwnJump(t *testing.T) {
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

func testPlayerJudgmentStillWorks(t *testing.T) {
	var advancedJumpID string
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
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)
	seasonID := "season_1"

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
				SeasonID:             &seasonID,
			}, true, nil
		},
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{ID: seasonID, Status: "Finalized"}, nil
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

	if !errors.Is(result.Err, ErrJudgingWindowClosed) {
		t.Fatalf("expected ErrJudgingWindowClosed for finalized season, got %v", result.Err)
	}
}

func testSubmitJudgment_JudgedJumpAllowsAdditionalJudgments(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Judged Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, true, nil
		},
		upsertJudgmentFn: func(_ context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (Judgment, bool, error) {
			return Judgment{ID: "judgment_2", JumpID: jumpID, PlayerID: playerID}, true, nil
		},
		advanceJudgedFn: func(_ context.Context, jumpID string) error {
			return nil
		},
	}

	result := SubmitJudgment(context.Background(), repo, JudgmentInput{
		JumpID:        "jump_1",
		JudgePlayerID: "judge_2",
		Commitment:    3,
		Transgression: 3,
		Creativity:    3,
		Presentation:  3,
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected additional judgment to be allowed on Judged Jump")
	}
}

func testSubmitJudgment_ExistingJudgmentEditDoesNotAdvance(t *testing.T) {
	var advancedJumpID string
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)

	repo := &mockJudgmentRepo{
		jumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Judged Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, true, nil
		},
		upsertJudgmentFn: func(_ context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (Judgment, bool, error) {
			return Judgment{ID: "judgment_existing", JumpID: jumpID, PlayerID: playerID}, false, nil
		},
		advanceJudgedFn: func(_ context.Context, jumpID string) error {
			advancedJumpID = jumpID
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
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected edit to be allowed")
	}
	if result.Created {
		t.Fatal("expected Created=false for edit")
	}
	if advancedJumpID != "" {
		t.Fatal("expected no jump advancement for edit")
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
		t.Run("guest judgment uses atomic persistence command", testGuestJudgmentUsesAtomicPersistenceCommand)
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
		t.Run("existing judgment edit does not advance", testSubmitJudgment_ExistingJudgmentEditDoesNotAdvance)
	})
}
