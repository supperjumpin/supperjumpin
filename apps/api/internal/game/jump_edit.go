package game

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrAuthorGracePeriodExpired = errors.New("Author Grace Period has expired")
	ErrInvalidCaption           = errors.New("Caption is required")
)

// CaptionEditRepository defines persistence operations for the caption edit flow.
type CaptionEditRepository interface {
	// JumpForEdit returns the Jump for the given ID. The error is non-nil
	// when the jump does not exist or the lookup fails.
	JumpForEdit(ctx context.Context, jumpID string) (JumpSnapshot, error)
	// UpdateCaption persists the new Caption for the given Jump.
	UpdateCaption(ctx context.Context, jumpID string, caption string) error
}

// EditCaptionInput bundles the parameters for editing a Jump's Caption.
type EditCaptionInput struct {
	JumpID   string
	PlayerID string
	Caption  string
}

// EditCaptionResult is the outcome of a caption edit attempt.
type EditCaptionResult struct {
	Allowed bool
	Err     error
}

// EditCaption allows the performer to edit a Jump's Caption only while
// the Author Grace Period is active.
//
// Rules:
//   - Caption must be non-empty after trimming.
//   - The Jump must exist.
//   - The Author Grace Period must not have expired.
//   - Only the performer may edit the Caption.
//
// It returns a result with:
//   - Allowed = true on successful edit.
//   - Allowed = false when the requesting Player is not the performer.
//   - Err set for invalid caption, missing jump, expired grace period, or persistence errors.
func EditCaption(ctx context.Context, repo CaptionEditRepository, input EditCaptionInput, now time.Time) EditCaptionResult {
	// 1. Validate caption
	if strings.TrimSpace(input.Caption) == "" {
		return EditCaptionResult{Err: ErrInvalidCaption}
	}

	// 2. Look up the jump
	jump, err := repo.JumpForEdit(ctx, input.JumpID)
	if err != nil {
		return EditCaptionResult{Err: err}
	}

	// 3. Author Grace Period must still be active
	if !now.Before(jump.GracePeriodExpiresAt) {
		return EditCaptionResult{Err: ErrAuthorGracePeriodExpired}
	}

	// 4. Only the performer may edit
	if jump.PlayerID != input.PlayerID {
		return EditCaptionResult{Allowed: false}
	}

	// 5. Persist the updated caption
	if err := repo.UpdateCaption(ctx, input.JumpID, input.Caption); err != nil {
		return EditCaptionResult{Err: err}
	}

	return EditCaptionResult{Allowed: true}
}
