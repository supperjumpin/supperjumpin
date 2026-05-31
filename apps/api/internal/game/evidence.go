package game

import (
	"context"
	"errors"
	"time"
)

var (
	ErrEvidenceUploadAuthorizationNotFound = errors.New("Evidence upload authorization not found")
	ErrSubmissionWindowClosed              = errors.New("Submission Window closed")
)

// AuthorizationSnapshot is a read-only view of an evidence upload authorization.
type AuthorizationSnapshot struct {
	ID             string
	JumpID        string
	MediaObjectKey string
	ExpiresAt      time.Time
}

// EvidenceSnapshot is a read-only view of an Evidence record.
type EvidenceSnapshot struct {
	ID             string
	JumpID        string
	PlayerID       string
	MediaObjectKey string
	Caption        string
	CreatedAt      time.Time
}

// EvidenceCreateResult is the outcome of claiming an authorization and advancing a jump.
type EvidenceCreateResult struct {
	EvidenceID     string
	MediaObjectKey string
}

// EvidenceRepository defines persistence operations for the evidence flow.
type EvidenceRepository interface {
	// PlannedJump returns the jump. ok is true only when the jump exists
	// and is "Planned Jump".
	PlannedJump(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
	// Season returns the Season for the given ID.
	Season(ctx context.Context, seasonID string) (SeasonSnapshot, error)
	// CreateAuthorization persists a new upload authorization.
	CreateAuthorization(ctx context.Context, jumpID, playerID, contentType string) (AuthorizationSnapshot, error)
	// ClaimAndAdvance atomically validates+consumes the upload authorization,
	// advances the jump from Planned to Performed, and records evidence.
	// Returns ErrEvidenceUploadAuthorizationNotFound if the auth is missing,
	// expired, or doesn't match the jump/player.
	ClaimAndAdvance(ctx context.Context, authorizationID, jumpID, playerID, caption string) (EvidenceCreateResult, error)
}

// AuthorizeEvidenceUploadInput bundles input for AuthorizeEvidenceUpload.
type AuthorizeEvidenceUploadInput struct {
	JumpID     string
	PlayerID    string
	ContentType string
}

// AuthorizeEvidenceUploadResult is the outcome of AuthorizeEvidenceUpload.
type AuthorizeEvidenceUploadResult struct {
	Authorization AuthorizationSnapshot
	Allowed       bool
	Err           error
}

// SubmitEvidenceInput bundles input for SubmitEvidence.
type SubmitEvidenceInput struct {
	JumpID              string
	PlayerID             string
	UploadAuthorizationID string
	Caption              string
}

// SubmitEvidenceResult is the outcome of SubmitEvidence.
type SubmitEvidenceResult struct {
	Evidence EvidenceSnapshot
	Jump     JumpSnapshot
	Allowed  bool
	Err      error
}

// AuthorizeEvidenceUpload evaluates evidence upload authorization rules.
//
// Returns Allowed=false when the player is not the jump performer.
// Returns an error when the jump is not found or persistence fails.
func AuthorizeEvidenceUpload(ctx context.Context, repo EvidenceRepository, input AuthorizeEvidenceUploadInput) AuthorizeEvidenceUploadResult {
	jump, ok, err := repo.PlannedJump(ctx, input.JumpID)
	if err != nil {
		return AuthorizeEvidenceUploadResult{Err: err}
	}
	if !ok {
		return AuthorizeEvidenceUploadResult{Err: ErrJumpNotFound}
	}

	// Only the jump performer may authorize an upload.
	if jump.PlayerID != input.PlayerID {
		return AuthorizeEvidenceUploadResult{Allowed: false}
	}

	auth, err := repo.CreateAuthorization(ctx, input.JumpID, input.PlayerID, input.ContentType)
	if err != nil {
		return AuthorizeEvidenceUploadResult{Err: err}
	}

	return AuthorizeEvidenceUploadResult{
		Authorization: auth,
		Allowed:       true,
	}
}

// SubmitEvidence evaluates evidence submission rules.
//
// Returns Allowed=false when the player is not the jump performer.
// Returns an error when the jump/authorization is not found, the
// submission window is closed, or persistence fails.
func SubmitEvidence(ctx context.Context, repo EvidenceRepository, input SubmitEvidenceInput, now time.Time) SubmitEvidenceResult {
	jump, ok, err := repo.PlannedJump(ctx, input.JumpID)
	if err != nil {
		return SubmitEvidenceResult{Err: err}
	}
	if !ok {
		return SubmitEvidenceResult{Err: ErrJumpNotFound}
	}

	// Only the jump performer may submit evidence.
	if jump.PlayerID != input.PlayerID {
		return SubmitEvidenceResult{Allowed: false, Jump: jump}
	}

	// Check submission window for season-linked jumps.
	if jump.SeasonID != nil {
		season, err := repo.Season(ctx, *jump.SeasonID)
		if err != nil {
			return SubmitEvidenceResult{Err: err}
		}
		if season.Status != "Active" || now.After(season.SubmissionDeadline) {
			return SubmitEvidenceResult{Err: ErrSubmissionWindowClosed}
		}
	}

	result, err := repo.ClaimAndAdvance(ctx, input.UploadAuthorizationID, input.JumpID, input.PlayerID, input.Caption)
	if err != nil {
		return SubmitEvidenceResult{Err: err}
	}

	jump.Status = "Performed Jump"
	return SubmitEvidenceResult{
		Evidence: EvidenceSnapshot{
			ID:             result.EvidenceID,
			JumpID:        input.JumpID,
			PlayerID:       input.PlayerID,
			MediaObjectKey: result.MediaObjectKey,
			Caption:        input.Caption,
		},
		Jump:    jump,
		Allowed: true,
	}
}
