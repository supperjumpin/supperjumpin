package game

import (
	"context"
	"time"
)

// RetractJumpRepository defines persistence operations for the retract flow.
type RetractJumpRepository interface {
	// JumpForRetract returns the Jump for the given ID. The error is non-nil
	// when the jump does not exist or the lookup fails.
	JumpForRetract(ctx context.Context, jumpID string) (JumpSnapshot, error)
	// RetractJump marks the Jump as "Removed Jump" with the given removedAt time.
	RetractJump(ctx context.Context, jumpID string, removedAt time.Time) error
}

// RetractJumpInput bundles the parameters for retracting a Jump.
type RetractJumpInput struct {
	JumpID   string
	PlayerID string
}

// RetractJumpResult is the outcome of a retraction attempt.
type RetractJumpResult struct {
	Allowed bool
	Err     error
}

// RetractJump allows the performer to retract a Jump only while
// the Author Grace Period is active.
//
// Rules:
//   - The Jump must exist.
//   - The Author Grace Period must not have expired.
//   - Only the performer may retract the Jump.
//
// It returns a result with:
//   - Allowed = true on successful retraction.
//   - Allowed = false when the requesting Player is not the performer.
//   - Err set for missing jump, expired grace period, or persistence errors.
func RetractJump(ctx context.Context, repo RetractJumpRepository, input RetractJumpInput, now time.Time) RetractJumpResult {
	// 1. Look up the jump
	jump, err := repo.JumpForRetract(ctx, input.JumpID)
	if err != nil {
		return RetractJumpResult{Err: err}
	}

	// 2. Author Grace Period must still be active
	if !now.Before(jump.GracePeriodExpiresAt) {
		return RetractJumpResult{Err: ErrAuthorGracePeriodExpired}
	}

	// 3. Only the performer may retract
	if jump.PlayerID != input.PlayerID {
		return RetractJumpResult{Allowed: false}
	}

	// 4. Persist the retraction
	if err := repo.RetractJump(ctx, input.JumpID, now.UTC()); err != nil {
		return RetractJumpResult{Err: err}
	}

	return RetractJumpResult{Allowed: true}
}
