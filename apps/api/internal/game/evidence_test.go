package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockEvidenceRepo struct {
	plannedJumpFn       func(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
	seasonFn            func(ctx context.Context, seasonID string) (SeasonSnapshot, error)
	createAuthorizationFn func(ctx context.Context, jumpID, playerID, contentType string) (AuthorizationSnapshot, error)
	claimAndAdvanceFn   func(ctx context.Context, authorizationID, jumpID, playerID, caption string) (EvidenceCreateResult, error)
}

func (m *mockEvidenceRepo) PlannedJump(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
	return m.plannedJumpFn(nil, jumpID)
}

func (m *mockEvidenceRepo) Season(_ context.Context, seasonID string) (SeasonSnapshot, error) {
	return m.seasonFn(nil, seasonID)
}

func (m *mockEvidenceRepo) CreateAuthorization(_ context.Context, jumpID, playerID, contentType string) (AuthorizationSnapshot, error) {
	return m.createAuthorizationFn(nil, jumpID, playerID, contentType)
}

func (m *mockEvidenceRepo) ClaimAndAdvance(_ context.Context, authorizationID, jumpID, playerID, caption string) (EvidenceCreateResult, error) {
	return m.claimAndAdvanceFn(nil, authorizationID, jumpID, playerID, caption)
}

func testAuthorizeEvidenceUpload_NonPerformerIsRejected(t *testing.T) {
	repo := &mockEvidenceRepo{
		plannedJumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, PlayerID: "performer_1", Status: "Planned Jump"}, true, nil
		},
	}

	result := AuthorizeEvidenceUpload(context.Background(), repo, AuthorizeEvidenceUploadInput{
		JumpID:     "jump_1",
		PlayerID:   "stranger_1",
		ContentType: "image/jpeg",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected non-performer to be rejected")
	}
}

func testAuthorizeEvidenceUpload_PerformerCanAuthorize(t *testing.T) {
	var createdJumpID, createdPlayerID, createdContentType string

	repo := &mockEvidenceRepo{
		plannedJumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, PlayerID: "performer_1", Status: "Planned Jump"}, true, nil
		},
		createAuthorizationFn: func(_ context.Context, jumpID, playerID, contentType string) (AuthorizationSnapshot, error) {
			createdJumpID = jumpID
			createdPlayerID = playerID
			createdContentType = contentType
			return AuthorizationSnapshot{
				ID:             "auth_1",
				JumpID:        jumpID,
				MediaObjectKey: "uploads/auth_1",
				ExpiresAt:      time.Date(2026, 6, 1, 13, 0, 0, 0, time.UTC),
			}, nil
		},
	}

	result := AuthorizeEvidenceUpload(context.Background(), repo, AuthorizeEvidenceUploadInput{
		JumpID:     "jump_1",
		PlayerID:   "performer_1",
		ContentType: "image/jpeg",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected performer to be allowed")
	}
	if result.Authorization.ID != "auth_1" {
		t.Fatalf("expected auth_1, got %q", result.Authorization.ID)
	}
	if createdJumpID != "jump_1" || createdPlayerID != "performer_1" || createdContentType != "image/jpeg" {
		t.Fatalf("unexpected create args: jump=%s player=%s type=%s", createdJumpID, createdPlayerID, createdContentType)
	}
}

func testAuthorizeEvidenceUpload_JumpNotFoundReturnsError(t *testing.T) {
	repo := &mockEvidenceRepo{
		plannedJumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{}, false, nil
		},
	}

	result := AuthorizeEvidenceUpload(context.Background(), repo, AuthorizeEvidenceUploadInput{
		JumpID:     "jump_1",
		PlayerID:   "performer_1",
		ContentType: "image/jpeg",
	})

	if !errors.Is(result.Err, ErrJumpNotFound) {
		t.Fatalf("expected ErrJumpNotFound, got %v", result.Err)
	}
}

func testAuthorizeEvidenceUpload_PersistenceErrorPropagates(t *testing.T) {
	repo := &mockEvidenceRepo{
		plannedJumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, PlayerID: "performer_1", Status: "Planned Jump"}, true, nil
		},
		createAuthorizationFn: func(_ context.Context, jumpID, playerID, contentType string) (AuthorizationSnapshot, error) {
			return AuthorizationSnapshot{}, errors.New("db error")
		},
	}

	result := AuthorizeEvidenceUpload(context.Background(), repo, AuthorizeEvidenceUploadInput{
		JumpID:     "jump_1",
		PlayerID:   "performer_1",
		ContentType: "image/jpeg",
	})

	if result.Err == nil || result.Err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", result.Err)
	}
}

func testSubmitEvidence_NonPerformerIsRejected(t *testing.T) {
	repo := &mockEvidenceRepo{
		plannedJumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, PlayerID: "performer_1", Status: "Planned Jump"}, true, nil
		},
	}

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	result := SubmitEvidence(context.Background(), repo, SubmitEvidenceInput{
		JumpID:              "jump_1",
		PlayerID:             "stranger_1",
		UploadAuthorizationID: "auth_1",
		Caption:              "Test caption",
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected non-performer to be rejected")
	}
	if result.Jump.ID != "jump_1" {
		t.Fatalf("expected jump snapshot returned, got %#v", result.Jump)
	}
}

func testSubmitEvidence_PerformerCanSubmit(t *testing.T) {
	var claimedAuthID, claimedJumpID, claimedPlayerID, claimedCaption string

	repo := &mockEvidenceRepo{
		plannedJumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, PlayerID: "performer_1", Status: "Planned Jump"}, true, nil
		},
		claimAndAdvanceFn: func(_ context.Context, authorizationID, jumpID, playerID, caption string) (EvidenceCreateResult, error) {
			claimedAuthID = authorizationID
			claimedJumpID = jumpID
			claimedPlayerID = playerID
			claimedCaption = caption
			return EvidenceCreateResult{
				EvidenceID:     "evidence_1",
				MediaObjectKey: "uploads/evidence_1",
			}, nil
		},
	}

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	result := SubmitEvidence(context.Background(), repo, SubmitEvidenceInput{
		JumpID:              "jump_1",
		PlayerID:             "performer_1",
		UploadAuthorizationID: "auth_1",
		Caption:              "Taco Bell to Olive Garden",
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected performer to be allowed")
	}
	if result.Evidence.ID != "evidence_1" {
		t.Fatalf("expected evidence_1, got %q", result.Evidence.ID)
	}
	if result.Evidence.Caption != "Taco Bell to Olive Garden" {
		t.Fatalf("expected caption, got %q", result.Evidence.Caption)
	}
	if result.Jump.Status != "Performed Jump" {
		t.Fatalf("expected Performed Jump status, got %q", result.Jump.Status)
	}
	if claimedAuthID != "auth_1" || claimedJumpID != "jump_1" || claimedPlayerID != "performer_1" || claimedCaption != "Taco Bell to Olive Garden" {
		t.Fatalf("unexpected claim args: auth=%s jump=%s player=%s caption=%s", claimedAuthID, claimedJumpID, claimedPlayerID, claimedCaption)
	}
}

func testSubmitEvidence_JumpNotFoundReturnsError(t *testing.T) {
	repo := &mockEvidenceRepo{
		plannedJumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{}, false, nil
		},
	}

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	result := SubmitEvidence(context.Background(), repo, SubmitEvidenceInput{
		JumpID:              "jump_1",
		PlayerID:             "performer_1",
		UploadAuthorizationID: "auth_1",
		Caption:              "Test",
	}, now)

	if !errors.Is(result.Err, ErrJumpNotFound) {
		t.Fatalf("expected ErrJumpNotFound, got %v", result.Err)
	}
}

func testSubmitEvidence_SeasonLinkedWithClosedSubmissionWindowReturnsError(t *testing.T) {
	seasonID := "season_1"
	repo := &mockEvidenceRepo{
		plannedJumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, PlayerID: "performer_1", Status: "Planned Jump", SeasonID: &seasonID}, true, nil
		},
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{
				ID:                 seasonID,
				Status:             "Active",
				SubmissionDeadline: time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC),
			}, nil
		},
	}

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	result := SubmitEvidence(context.Background(), repo, SubmitEvidenceInput{
		JumpID:              "jump_1",
		PlayerID:             "performer_1",
		UploadAuthorizationID: "auth_1",
		Caption:              "Test",
	}, now)

	if !errors.Is(result.Err, ErrSubmissionWindowClosed) {
		t.Fatalf("expected ErrSubmissionWindowClosed, got %v", result.Err)
	}
}

func testSubmitEvidence_SeasonLinkedWithActiveSeasonSucceeds(t *testing.T) {
	seasonID := "season_1"
	repo := &mockEvidenceRepo{
		plannedJumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, PlayerID: "performer_1", Status: "Planned Jump", SeasonID: &seasonID}, true, nil
		},
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{
				ID:                 seasonID,
				Status:             "Active",
				SubmissionDeadline: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
			}, nil
		},
		claimAndAdvanceFn: func(_ context.Context, authorizationID, jumpID, playerID, caption string) (EvidenceCreateResult, error) {
			return EvidenceCreateResult{EvidenceID: "evidence_1", MediaObjectKey: "uploads/evidence_1"}, nil
		},
	}

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	result := SubmitEvidence(context.Background(), repo, SubmitEvidenceInput{
		JumpID:              "jump_1",
		PlayerID:             "performer_1",
		UploadAuthorizationID: "auth_1",
		Caption:              "Test",
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected allowed")
	}
}

func testSubmitEvidence_SeasonNotActiveReturnsError(t *testing.T) {
	seasonID := "season_1"
	repo := &mockEvidenceRepo{
		plannedJumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, PlayerID: "performer_1", Status: "Planned Jump", SeasonID: &seasonID}, true, nil
		},
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{
				ID:     seasonID,
				Status: "Judging Grace Period",
				SubmissionDeadline: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	result := SubmitEvidence(context.Background(), repo, SubmitEvidenceInput{
		JumpID:              "jump_1",
		PlayerID:             "performer_1",
		UploadAuthorizationID: "auth_1",
		Caption:              "Test",
	}, now)

	if !errors.Is(result.Err, ErrSubmissionWindowClosed) {
		t.Fatalf("expected ErrSubmissionWindowClosed, got %v", result.Err)
	}
}

func testSubmitEvidence_AuthorizationErrorPropagates(t *testing.T) {
	repo := &mockEvidenceRepo{
		plannedJumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, PlayerID: "performer_1", Status: "Planned Jump"}, true, nil
		},
		claimAndAdvanceFn: func(_ context.Context, authorizationID, jumpID, playerID, caption string) (EvidenceCreateResult, error) {
			return EvidenceCreateResult{}, ErrEvidenceUploadAuthorizationNotFound
		},
	}

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	result := SubmitEvidence(context.Background(), repo, SubmitEvidenceInput{
		JumpID:              "jump_1",
		PlayerID:             "performer_1",
		UploadAuthorizationID: "auth_1",
		Caption:              "Test",
	}, now)

	if !errors.Is(result.Err, ErrEvidenceUploadAuthorizationNotFound) {
		t.Fatalf("expected ErrEvidenceUploadAuthorizationNotFound, got %v", result.Err)
	}
}

func testSubmitEvidence_OffSeasonSkipsSeasonCheck(t *testing.T) {
	repo := &mockEvidenceRepo{
		plannedJumpFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, PlayerID: "performer_1", Status: "Planned Jump", SeasonID: nil}, true, nil
		},
		claimAndAdvanceFn: func(_ context.Context, authorizationID, jumpID, playerID, caption string) (EvidenceCreateResult, error) {
			return EvidenceCreateResult{EvidenceID: "evidence_1", MediaObjectKey: "uploads/evidence_1"}, nil
		},
	}

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	result := SubmitEvidence(context.Background(), repo, SubmitEvidenceInput{
		JumpID:              "jump_1",
		PlayerID:             "performer_1",
		UploadAuthorizationID: "auth_1",
		Caption:              "Test",
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected allowed")
	}
}

func TestEvidence(t *testing.T) {
	t.Run("authorize evidence upload", func(t *testing.T) {
		t.Run("non-performer is rejected", testAuthorizeEvidenceUpload_NonPerformerIsRejected)
		t.Run("performer can authorize", testAuthorizeEvidenceUpload_PerformerCanAuthorize)
		t.Run("jump not found returns error", testAuthorizeEvidenceUpload_JumpNotFoundReturnsError)
		t.Run("persistence error propagates", testAuthorizeEvidenceUpload_PersistenceErrorPropagates)
	})

	t.Run("submit evidence", func(t *testing.T) {
		t.Run("non-performer is rejected", testSubmitEvidence_NonPerformerIsRejected)
		t.Run("performer can submit", testSubmitEvidence_PerformerCanSubmit)
		t.Run("jump not found returns error", testSubmitEvidence_JumpNotFoundReturnsError)
		t.Run("season linked with closed submission window returns error", testSubmitEvidence_SeasonLinkedWithClosedSubmissionWindowReturnsError)
		t.Run("season linked with active season succeeds", testSubmitEvidence_SeasonLinkedWithActiveSeasonSucceeds)
		t.Run("season not active returns error", testSubmitEvidence_SeasonNotActiveReturnsError)
		t.Run("authorization error propagates", testSubmitEvidence_AuthorizationErrorPropagates)
		t.Run("off-season skips season check", testSubmitEvidence_OffSeasonSkipsSeasonCheck)
	})
}
