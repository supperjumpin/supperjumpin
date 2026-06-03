package game

import (
	"context"
	"time"
)

// JumpRepository defines persistence operations for the jump planning flow.
type JumpRepository interface {
	// InsertPerformedJump creates a new performed jump with initial evidence
	// in a single operation. Returns the jump and evidence snapshots.
	InsertPerformedJump(ctx context.Context, params InsertPerformedJumpParams) (JumpSnapshot, EvidenceSnapshot, error)
}

// InsertPerformedJumpParams bundles parameters for direct performed jump creation.
type InsertPerformedJumpParams struct {
	PlayerID             string
	Source               string
	Destination          string
	Food                 string
	Caption              string
	MediaObjectKey       string
	GracePeriodExpiresAt time.Time
}

// CreatePerformedJumpInput bundles the parameters for creating a performed jump directly.
type CreatePerformedJumpInput struct {
	PlayerID       string
	Source         string
	Destination    string
	Food           string
	Caption        string
	MediaObjectKey string
}

// CreatePerformedJumpOutput is the outcome of creating a performed jump directly.
type CreatePerformedJumpOutput struct {
	Jump     JumpSnapshot
	Evidence EvidenceSnapshot
	Err      error
}

// CreatePerformedJump creates a performed jump directly, bypassing the
// Idea/Planned/Evidence flow. No group membership check is performed.
func CreatePerformedJump(ctx context.Context, repo JumpRepository, input CreatePerformedJumpInput, now time.Time) CreatePerformedJumpOutput {
	gracePeriodExpiresAt := now.Add(10 * time.Minute).UTC()

	jump, evidence, err := repo.InsertPerformedJump(ctx, InsertPerformedJumpParams{
		PlayerID:             input.PlayerID,
		Source:               input.Source,
		Destination:          input.Destination,
		Food:                 input.Food,
		Caption:              input.Caption,
		MediaObjectKey:       input.MediaObjectKey,
		GracePeriodExpiresAt: gracePeriodExpiresAt,
	})
	if err != nil {
		return CreatePerformedJumpOutput{Err: err}
	}

	return CreatePerformedJumpOutput{
		Jump:     jump,
		Evidence: evidence,
	}
}
