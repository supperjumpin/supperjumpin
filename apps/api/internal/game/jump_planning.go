package game

import (
	"context"
)

// JumpRepository defines persistence operations for the jump planning flow.
type JumpRepository interface {
	// InsertIdea creates a new Idea jump and returns a snapshot of it.
	InsertIdea(ctx context.Context, groupID, playerID, source, destination, food string) (JumpSnapshot, error)
	// Idea returns the jump for the given ID without any status filtering.
	// ok is false when not found.
	Idea(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
	// GroupMembership returns the membership for a player in a group.
	// ok is false when there is no membership.
	GroupMembership(ctx context.Context, playerID, groupID string) (MembershipSnapshot, bool, error)
	// ActiveSeasonForGroup returns the active season for a group, if any.
	// SeasonSnapshot.ID is empty when no active season exists.
	ActiveSeasonForGroup(ctx context.Context, groupID string) (SeasonSnapshot, error)
	// UpdateJumpToPlanned atomically updates an Idea to Planned Jump status
	// and associates it with the given season (nil for off-season).
	UpdateJumpToPlanned(ctx context.Context, jumpID, playerID string, seasonID *string) (JumpSnapshot, error)
}

// CreateIdeaInput bundles the parameters for creating an Idea.
type CreateIdeaInput struct {
	GroupID     string
	PlayerID    string
	Source      string
	Destination string
	Food        string
}

// CreateIdeaResult is the outcome of creating an Idea.
type CreateIdeaResult struct {
	Jump    JumpSnapshot
	Allowed bool
	Err     error
}

// CreatePlannedJumpInput bundles the parameters for planning a Jump.
type CreatePlannedJumpInput struct {
	IdeaID    string
	PlayerID  string
	OffSeason bool
}

// CreatePlannedJumpResult is the outcome of planning a Jump.
type CreatePlannedJumpResult struct {
	Jump    JumpSnapshot
	Allowed bool
	Err     error
}

// CreateIdea evaluates Idea creation rules and persists the result.
//
// Returns Allowed = false when the player is not a group member.
func CreateIdea(ctx context.Context, repo JumpRepository, input CreateIdeaInput) CreateIdeaResult {
	// 1. Player must be a group member
	_, ok, err := repo.GroupMembership(ctx, input.PlayerID, input.GroupID)
	if err != nil {
		return CreateIdeaResult{Err: err}
	}
	if !ok {
		return CreateIdeaResult{Allowed: false}
	}

	// 2. Create the Idea
	jump, err := repo.InsertIdea(ctx, input.GroupID, input.PlayerID, input.Source, input.Destination, input.Food)
	if err != nil {
		return CreateIdeaResult{Err: err}
	}

	return CreateIdeaResult{Jump: jump, Allowed: true}
}

// CreatePlannedJump evaluates jump planning rules and persists the result.
//
// Returns ErrJumpNotFound when the idea does not exist or is not in "Idea" status.
// Returns Allowed = false when the player is not a group member or does not own the idea.
func CreatePlannedJump(ctx context.Context, repo JumpRepository, input CreatePlannedJumpInput) CreatePlannedJumpResult {
	// 1. Look up the idea
	jump, ok, err := repo.Idea(ctx, input.IdeaID)
	if err != nil {
		return CreatePlannedJumpResult{Err: err}
	}
	if !ok || jump.Status != "Idea" {
		return CreatePlannedJumpResult{Err: ErrJumpNotFound}
	}

	// 2. Player must be a group member and the idea owner
	membership, ok, err := repo.GroupMembership(ctx, input.PlayerID, jump.GroupID)
	if err != nil {
		return CreatePlannedJumpResult{Err: err}
	}
	if !ok || jump.PlayerID != input.PlayerID {
		return CreatePlannedJumpResult{Allowed: false}
	}
	_ = membership // role not needed for planning, reserved for future

	// 3. If not explicitly off-season, try to link to an active season
	var seasonID *string
	if !input.OffSeason {
		season, err := repo.ActiveSeasonForGroup(ctx, jump.GroupID)
		if err != nil {
			return CreatePlannedJumpResult{Err: err}
		}
		if season.ID != "" {
			seasonID = &season.ID
		}
	}

	// 4. Update to Planned Jump
	updated, err := repo.UpdateJumpToPlanned(ctx, input.IdeaID, input.PlayerID, seasonID)
	if err != nil {
		return CreatePlannedJumpResult{Err: err}
	}

	return CreatePlannedJumpResult{Jump: updated, Allowed: true}
}
