package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockCaptionEditRepo struct {
	jumpForEditFn  func(ctx context.Context, jumpID string) (JumpSnapshot, error)
	updateCaptionFn func(ctx context.Context, jumpID string, caption string) error
}

func (m *mockCaptionEditRepo) JumpForEdit(_ context.Context, jumpID string) (JumpSnapshot, error) {
	return m.jumpForEditFn(nil, jumpID)
}

func (m *mockCaptionEditRepo) UpdateCaption(_ context.Context, jumpID string, caption string) error {
	return m.updateCaptionFn(nil, jumpID, caption)
}

func testEditCaption_AllowedDuringGracePeriod(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 12, 10, 0, 0, time.UTC)

	var updatedJumpID, updatedCaption string
	repo := &mockCaptionEditRepo{
		jumpForEditFn: func(_ context.Context, jumpID string) (JumpSnapshot, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, nil
		},
		updateCaptionFn: func(_ context.Context, jumpID string, caption string) error {
			updatedJumpID = jumpID
			updatedCaption = caption
			return nil
		},
	}

	result := EditCaption(context.Background(), repo, EditCaptionInput{
		JumpID:   "jump_1",
		PlayerID: "performer_1",
		Caption:  "Updated caption",
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected edit to be allowed")
	}
	if updatedJumpID != "jump_1" {
		t.Fatalf("expected jump ID %q, got %q", "jump_1", updatedJumpID)
	}
	if updatedCaption != "Updated caption" {
		t.Fatalf("expected caption %q, got %q", "Updated caption", updatedCaption)
	}
}

func testEditCaption_RejectedAfterGracePeriodExpires(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)

	repo := &mockCaptionEditRepo{
		jumpForEditFn: func(_ context.Context, jumpID string) (JumpSnapshot, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, nil
		},
	}

	result := EditCaption(context.Background(), repo, EditCaptionInput{
		JumpID:   "jump_1",
		PlayerID: "performer_1",
		Caption:  "Updated caption",
	}, now)

	if !errors.Is(result.Err, ErrAuthorGracePeriodExpired) {
		t.Fatalf("expected ErrAuthorGracePeriodExpired, got %v", result.Err)
	}
}

func testEditCaption_RejectedForNonPerformer(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 12, 10, 0, 0, time.UTC)

	repo := &mockCaptionEditRepo{
		jumpForEditFn: func(_ context.Context, jumpID string) (JumpSnapshot, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, nil
		},
	}

	result := EditCaption(context.Background(), repo, EditCaptionInput{
		JumpID:   "jump_1",
		PlayerID: "other_player",
		Caption:  "Updated caption",
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error for non-performer, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected edit to be rejected for non-performer")
	}
}

func testEditCaption_RejectedForMissingJump(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockCaptionEditRepo{
		jumpForEditFn: func(_ context.Context, jumpID string) (JumpSnapshot, error) {
			return JumpSnapshot{}, errors.New("jump not found")
		},
	}

	result := EditCaption(context.Background(), repo, EditCaptionInput{
		JumpID:   "jump_1",
		PlayerID: "performer_1",
		Caption:  "Updated caption",
	}, now)

	if result.Err == nil || result.Err.Error() != "jump not found" {
		t.Fatalf("expected jump not found error, got %v", result.Err)
	}
}

func testEditCaption_RejectedForEmptyCaption(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 12, 10, 0, 0, time.UTC)

	repo := &mockCaptionEditRepo{
		jumpForEditFn: func(_ context.Context, jumpID string) (JumpSnapshot, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, nil
		},
	}

	cases := []string{"", "   ", "\t\n"}
	for _, caption := range cases {
		result := EditCaption(context.Background(), repo, EditCaptionInput{
			JumpID:   "jump_1",
			PlayerID: "performer_1",
			Caption:  caption,
		}, now)
		if !errors.Is(result.Err, ErrInvalidCaption) {
			t.Fatalf("expected ErrInvalidCaption for caption %q, got %v", caption, result.Err)
		}
	}
}

func testEditCaption_PersistenceErrorPropagates(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 12, 10, 0, 0, time.UTC)

	repo := &mockCaptionEditRepo{
		jumpForEditFn: func(_ context.Context, jumpID string) (JumpSnapshot, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, nil
		},
		updateCaptionFn: func(_ context.Context, jumpID string, caption string) error {
			return errors.New("db error")
		},
	}

	result := EditCaption(context.Background(), repo, EditCaptionInput{
		JumpID:   "jump_1",
		PlayerID: "performer_1",
		Caption:  "Updated caption",
	}, now)

	if result.Err == nil || result.Err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", result.Err)
	}
}

func TestEditCaption(t *testing.T) {
	t.Run("allowed during grace period", testEditCaption_AllowedDuringGracePeriod)
	t.Run("rejected after grace period expires", testEditCaption_RejectedAfterGracePeriodExpires)
	t.Run("rejected for non-performer", testEditCaption_RejectedForNonPerformer)
	t.Run("rejected for missing jump", testEditCaption_RejectedForMissingJump)
	t.Run("rejected for empty caption", testEditCaption_RejectedForEmptyCaption)
	t.Run("persistence error propagates", testEditCaption_PersistenceErrorPropagates)
}
