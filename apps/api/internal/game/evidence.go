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

// EvidenceCreateResult is the outcome of claiming an authorization and advancing a stunt.
type EvidenceCreateResult struct {
	EvidenceID     string
	MediaObjectKey string
}

// EvidenceRepository defines persistence operations for the evidence flow.
type EvidenceRepository interface {
	// PlannedStunt returns the stunt. ok is true only when the stunt exists
	// and is "Planned Jump".
	PlannedStunt(ctx context.Context, stuntID string) (StuntSnapshot, bool, error)
	// Season returns the Season for the given ID.
	Season(ctx context.Context, seasonID string) (SeasonSnapshot, error)
	// CreateAuthorization persists a new upload authorization.
	CreateAuthorization(ctx context.Context, stuntID, playerID, contentType string) (AuthorizationSnapshot, error)
	// ClaimAndAdvance atomically validates+consumes the upload authorization,
	// advances the stunt from Planned to Performed, and records evidence.
	// Returns ErrEvidenceUploadAuthorizationNotFound if the auth is missing,
	// expired, or doesn't match the stunt/player.
	ClaimAndAdvance(ctx context.Context, authorizationID, stuntID, playerID, caption string) (EvidenceCreateResult, error)
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
	Stunt    StuntSnapshot
	Allowed  bool
	Err      error
}

// AuthorizeEvidenceUpload evaluates evidence upload authorization rules.
//
// Returns Allowed=false when the player is not the stunt performer.
// Returns an error when the stunt is not found or persistence fails.
func AuthorizeEvidenceUpload(ctx context.Context, repo EvidenceRepository, input AuthorizeEvidenceUploadInput) AuthorizeEvidenceUploadResult {
	stunt, ok, err := repo.PlannedStunt(ctx, input.JumpID)
	if err != nil {
		return AuthorizeEvidenceUploadResult{Err: err}
	}
	if !ok {
		return AuthorizeEvidenceUploadResult{Err: ErrJumpNotFound}
	}

	// Only the stunt performer may authorize an upload.
	if stunt.PlayerID != input.PlayerID {
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
// Returns Allowed=false when the player is not the stunt performer.
// Returns an error when the stunt/authorization is not found, the
// submission window is closed, or persistence fails.
func SubmitEvidence(ctx context.Context, repo EvidenceRepository, input SubmitEvidenceInput, now time.Time) SubmitEvidenceResult {
	stunt, ok, err := repo.PlannedStunt(ctx, input.JumpID)
	if err != nil {
		return SubmitEvidenceResult{Err: err}
	}
	if !ok {
		return SubmitEvidenceResult{Err: ErrJumpNotFound}
	}

	// Only the stunt performer may submit evidence.
	if stunt.PlayerID != input.PlayerID {
		return SubmitEvidenceResult{Allowed: false, Stunt: stunt}
	}

	// Check submission window for season-linked jumps.
	if stunt.SeasonID != nil {
		season, err := repo.Season(ctx, *stunt.SeasonID)
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

	stunt.Status = "Performed Jump"
	return SubmitEvidenceResult{
		Evidence: EvidenceSnapshot{
			ID:             result.EvidenceID,
			JumpID:        input.JumpID,
			PlayerID:       input.PlayerID,
			MediaObjectKey: result.MediaObjectKey,
			Caption:        input.Caption,
		},
		Stunt:   stunt,
		Allowed: true,
	}
}
