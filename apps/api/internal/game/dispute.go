package game

import (
	"context"
	"errors"
)

var (
	ErrInvalidDisputeConcern   = errors.New("invalid dispute concern")
	ErrInvalidDisputeResolution = errors.New("invalid dispute resolution")
	ErrDisputeNotFound         = errors.New("dispute not found")
)

type DisputeSnapshot struct {
	ID                 string
	JumpID            string
	RaisedByPlayerID   string
	Concern            string
	Details            string
	Status             string
	Resolution         *string
	ResolutionReason   *string
	ResolvedByPlayerID *string
	OverrideResolution *string
	OverrideReason     *string
	OverrideByPlayerID *string
}

type DisputeRepository interface {
	JumpByID(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
	Season(ctx context.Context, seasonID string) (SeasonSnapshot, error)
	GroupMembership(ctx context.Context, playerID, groupID string) (MembershipSnapshot, bool, error)
	InsertDispute(ctx context.Context, jumpID, raisedByPlayerID, concern, details string) (DisputeSnapshot, error)
	Dispute(ctx context.Context, disputeID string) (DisputeSnapshot, error)
	UpdateDisputeResolution(ctx context.Context, disputeID, resolution, resolutionReason, resolvedByPlayerID string) error
	UpdateDisputeOverride(ctx context.Context, disputeID, overrideResolution, overrideReason, overrideByPlayerID string) error
	UpdateJumpStatusAfterDispute(ctx context.Context, jumpID, status string) error
}

type CreateDisputeInput struct {
	PlayerID string
	JumpID  string
	Concern  string
	Details  string
}

type CreateDisputeResult struct {
	Dispute DisputeSnapshot
	Allowed bool
	Err     error
}

type ResolveDisputeInput struct {
	PlayerID         string
	DisputeID        string
	Resolution       string
	ResolutionReason string
}

type ResolveDisputeResult struct {
	Dispute DisputeSnapshot
	Jump    JumpSnapshot
	Allowed bool
	Err     error
}

func CreateDispute(ctx context.Context, repo DisputeRepository, input CreateDisputeInput) CreateDisputeResult {
	if !validDisputeConcern(input.Concern) {
		return CreateDisputeResult{Err: ErrInvalidDisputeConcern}
	}

	jump, ok, err := repo.JumpByID(ctx, input.JumpID)
	if err != nil {
		return CreateDisputeResult{Err: err}
	}
	if !ok || !disputableJumpStatus(jump.Status) {
		return CreateDisputeResult{Err: ErrJumpNotFound}
	}

	_, ok, err = repo.GroupMembership(ctx, input.PlayerID, jump.GroupID)
	if err != nil {
		return CreateDisputeResult{Err: err}
	}
	if !ok {
		return CreateDisputeResult{Allowed: false}
	}

	dispute, err := repo.InsertDispute(ctx, input.JumpID, input.PlayerID, input.Concern, input.Details)
	if err != nil {
		return CreateDisputeResult{Err: err}
	}

	return CreateDisputeResult{Dispute: dispute, Allowed: true}
}

func ResolveDispute(ctx context.Context, repo DisputeRepository, input ResolveDisputeInput) ResolveDisputeResult {
	if !validDisputeResolution(input.Resolution) {
		return ResolveDisputeResult{Err: ErrInvalidDisputeResolution}
	}

	dispute, err := repo.Dispute(ctx, input.DisputeID)
	if err != nil {
		return ResolveDisputeResult{Err: err}
	}
	if dispute.ID == "" {
		return ResolveDisputeResult{Err: ErrDisputeNotFound}
	}

	jump, ok, err := repo.JumpByID(ctx, dispute.JumpID)
	if err != nil {
		return ResolveDisputeResult{Err: err}
	}
	if !ok {
		return ResolveDisputeResult{Err: ErrJumpNotFound}
	}

	membership, ok, err := repo.GroupMembership(ctx, input.PlayerID, jump.GroupID)
	if err != nil {
		return ResolveDisputeResult{Err: err}
	}
	if !ok {
		return ResolveDisputeResult{Allowed: false}
	}

	if dispute.Status == "Open" {
		if jump.SeasonID == nil {
			if membership.Role != "Group Admin" || input.Resolution == "Disqualified Jump" {
				return ResolveDisputeResult{Allowed: false}
			}
		} else {
			if input.Resolution == "Removed Jump" {
				return ResolveDisputeResult{Allowed: false}
			}
			season, err := repo.Season(ctx, *jump.SeasonID)
			if err != nil {
				return ResolveDisputeResult{Err: err}
			}
			if season.CommissionerPlayerID != input.PlayerID {
				return ResolveDisputeResult{Allowed: false}
			}
		}
		if err := repo.UpdateDisputeResolution(ctx, input.DisputeID, input.Resolution, input.ResolutionReason, input.PlayerID); err != nil {
			return ResolveDisputeResult{Err: err}
		}
		dispute.Status = "Resolved"
		dispute.Resolution = &input.Resolution
		dispute.ResolutionReason = &input.ResolutionReason
		dispute.ResolvedByPlayerID = &input.PlayerID
	} else {
		if membership.Role != "Group Admin" || input.Resolution == "No Action" {
			return ResolveDisputeResult{Allowed: false}
		}
		if err := repo.UpdateDisputeOverride(ctx, input.DisputeID, input.Resolution, input.ResolutionReason, input.PlayerID); err != nil {
			return ResolveDisputeResult{Err: err}
		}
		dispute.Status = "Overridden"
		dispute.OverrideResolution = &input.Resolution
		dispute.OverrideReason = &input.ResolutionReason
		dispute.OverrideByPlayerID = &input.PlayerID
	}

	effectiveResolution := input.Resolution
	if effectiveResolution == "No Action" {
		effectiveResolution = ""
	}
	if effectiveResolution != "" {
		jump = applyDisputeResolutionToJump(jump, effectiveResolution)
		if err := repo.UpdateJumpStatusAfterDispute(ctx, jump.ID, jump.Status); err != nil {
			return ResolveDisputeResult{Err: err}
		}
	}

	return ResolveDisputeResult{
		Dispute: dispute,
		Jump:    jump,
		Allowed: true,
	}
}

func validDisputeConcern(concern string) bool {
	switch concern {
	case "House Rules", "Credibility", "Source", "Destination", "Food", "duplicate", "other":
		return true
	default:
		return false
	}
}

func validDisputeResolution(resolution string) bool {
	switch resolution {
	case "No Action", "Disqualified Jump", "Removed Jump":
		return true
	default:
		return false
	}
}

func disputableJumpStatus(status string) bool {
	return status == "Performed Jump" || status == "Judged Jump" || status == "Unjudged Jump" || status == "Disqualified Jump"
}

func applyDisputeResolutionToJump(jump JumpSnapshot, resolution string) JumpSnapshot {
	switch resolution {
	case "Disqualified Jump":
		jump.Status = "Disqualified Jump"
		jump.FinalScore = nil
	case "Removed Jump":
		jump.Status = "Removed Jump"
		jump.FinalScore = nil
	}
	return jump
}
