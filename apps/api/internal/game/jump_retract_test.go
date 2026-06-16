package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRetractJumpRepo struct {
	jumpForRetractFn func(ctx context.Context, jumpID string) (JumpSnapshot, error)
	retractJumpFn    func(ctx context.Context, jumpID string, removedAt time.Time) error
}

func (m *mockRetractJumpRepo) JumpForRetract(_ context.Context, jumpID string) (JumpSnapshot, error) {
	return m.jumpForRetractFn(nil, jumpID)
}

func (m *mockRetractJumpRepo) RetractJump(_ context.Context, jumpID string, removedAt time.Time) error {
	return m.retractJumpFn(nil, jumpID, removedAt)
}

func testRetractJump_AllowedDuringGracePeriod(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 12, 10, 0, 0, time.UTC)

	var retractedJumpID string
	var retractedRemovedAt time.Time
	repo := &mockRetractJumpRepo{
		jumpForRetractFn: func(_ context.Context, jumpID string) (JumpSnapshot, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, nil
		},
		retractJumpFn: func(_ context.Context, jumpID string, removedAt time.Time) error {
			retractedJumpID = jumpID
			retractedRemovedAt = removedAt
			return nil
		},
	}

	result := RetractJump(context.Background(), repo, RetractJumpInput{
		JumpID:   "jump_1",
		PlayerID: "performer_1",
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected retraction to be allowed")
	}
	if retractedJumpID != "jump_1" {
		t.Fatalf("expected jump ID %q, got %q", "jump_1", retractedJumpID)
	}
	if !retractedRemovedAt.Equal(now.UTC()) {
		t.Fatalf("expected removedAt %v, got %v", now.UTC(), retractedRemovedAt)
	}
}

func testRetractJump_RejectedAfterGracePeriodExpires(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 11, 50, 0, 0, time.UTC)

	repo := &mockRetractJumpRepo{
		jumpForRetractFn: func(_ context.Context, jumpID string) (JumpSnapshot, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, nil
		},
	}

	result := RetractJump(context.Background(), repo, RetractJumpInput{
		JumpID:   "jump_1",
		PlayerID: "performer_1",
	}, now)

	if !errors.Is(result.Err, ErrAuthorGracePeriodExpired) {
		t.Fatalf("expected ErrAuthorGracePeriodExpired, got %v", result.Err)
	}
}

func testRetractJump_RejectedForNonPerformer(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 12, 10, 0, 0, time.UTC)

	repo := &mockRetractJumpRepo{
		jumpForRetractFn: func(_ context.Context, jumpID string) (JumpSnapshot, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, nil
		},
	}

	result := RetractJump(context.Background(), repo, RetractJumpInput{
		JumpID:   "jump_1",
		PlayerID: "other_player",
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error for non-performer, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected retraction to be rejected for non-performer")
	}
}

func testRetractJump_RejectedForMissingJump(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	repo := &mockRetractJumpRepo{
		jumpForRetractFn: func(_ context.Context, jumpID string) (JumpSnapshot, error) {
			return JumpSnapshot{}, errors.New("jump not found")
		},
	}

	result := RetractJump(context.Background(), repo, RetractJumpInput{
		JumpID:   "jump_1",
		PlayerID: "performer_1",
	}, now)

	if result.Err == nil || result.Err.Error() != "jump not found" {
		t.Fatalf("expected jump not found error, got %v", result.Err)
	}
}

func testRetractJump_PersistenceErrorPropagates(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	gracePeriodExpiresAt := time.Date(2026, 6, 1, 12, 10, 0, 0, time.UTC)

	repo := &mockRetractJumpRepo{
		jumpForRetractFn: func(_ context.Context, jumpID string) (JumpSnapshot, error) {
			return JumpSnapshot{
				ID:                   jumpID,
				PlayerID:             "performer_1",
				Status:               "Performed Jump",
				GracePeriodExpiresAt: gracePeriodExpiresAt,
			}, nil
		},
		retractJumpFn: func(_ context.Context, jumpID string, removedAt time.Time) error {
			return errors.New("db error")
		},
	}

	result := RetractJump(context.Background(), repo, RetractJumpInput{
		JumpID:   "jump_1",
		PlayerID: "performer_1",
	}, now)

	if result.Err == nil || result.Err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", result.Err)
	}
}

func TestRetractJump(t *testing.T) {
	t.Run("allowed during grace period", testRetractJump_AllowedDuringGracePeriod)
	t.Run("rejected after grace period expires", testRetractJump_RejectedAfterGracePeriodExpires)
	t.Run("rejected for non-performer", testRetractJump_RejectedForNonPerformer)
	t.Run("rejected for missing jump", testRetractJump_RejectedForMissingJump)
	t.Run("persistence error propagates", testRetractJump_PersistenceErrorPropagates)
}
