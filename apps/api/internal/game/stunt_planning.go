package game

import (
	"context"
)

// StuntPlanningRepository defines persistence operations for the stunt planning flow.
type StuntPlanningRepository interface {
	// InsertIdea creates a new Idea stunt and returns a snapshot of it.
	InsertIdea(ctx context.Context, groupID, playerID, source, destination, food string) (StuntSnapshot, error)
	// Idea returns the stunt for the given ID without any status filtering.
	// ok is false when not found.
	Idea(ctx context.Context, stuntID string) (StuntSnapshot, bool, error)
	// GroupMembership returns the membership for a player in a group.
	// ok is false when there is no membership.
	GroupMembership(ctx context.Context, playerID, groupID string) (MembershipSnapshot, bool, error)
	// ActiveSeasonForGroup returns the active season for a group, if any.
	// SeasonSnapshot.ID is empty when no active season exists.
	ActiveSeasonForGroup(ctx context.Context, groupID string) (SeasonSnapshot, error)
	// UpdateStuntToPlanned atomically updates an Idea to Planned Stunt status
	// and associates it with the given season (nil for off-season).
	UpdateStuntToPlanned(ctx context.Context, stuntID, playerID string, seasonID *string) (StuntSnapshot, error)
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
	Stunt   StuntSnapshot
	Allowed bool
	Err     error
}

// CreatePlannedStuntInput bundles the parameters for planning a Stunt.
type CreatePlannedStuntInput struct {
	IdeaID    string
	PlayerID  string
	OffSeason bool
}

// CreatePlannedStuntResult is the outcome of planning a Stunt.
type CreatePlannedStuntResult struct {
	Stunt   StuntSnapshot
	Allowed bool
	Err     error
}

// CreateIdea evaluates Idea creation rules and persists the result.
//
// Returns Allowed = false when the player is not a group member.
func CreateIdea(ctx context.Context, repo StuntPlanningRepository, input CreateIdeaInput) CreateIdeaResult {
	// 1. Player must be a group member
	_, ok, err := repo.GroupMembership(ctx, input.PlayerID, input.GroupID)
	if err != nil {
		return CreateIdeaResult{Err: err}
	}
	if !ok {
		return CreateIdeaResult{Allowed: false}
	}

	// 2. Create the Idea
	stunt, err := repo.InsertIdea(ctx, input.GroupID, input.PlayerID, input.Source, input.Destination, input.Food)
	if err != nil {
		return CreateIdeaResult{Err: err}
	}

	return CreateIdeaResult{Stunt: stunt, Allowed: true}
}

// CreatePlannedStunt evaluates stunt planning rules and persists the result.
//
// Returns ErrStuntNotFound when the idea does not exist or is not in "Idea" status.
// Returns Allowed = false when the player is not a group member or does not own the idea.
func CreatePlannedStunt(ctx context.Context, repo StuntPlanningRepository, input CreatePlannedStuntInput) CreatePlannedStuntResult {
	// 1. Look up the idea
	stunt, ok, err := repo.Idea(ctx, input.IdeaID)
	if err != nil {
		return CreatePlannedStuntResult{Err: err}
	}
	if !ok || stunt.Status != "Idea" {
		return CreatePlannedStuntResult{Err: ErrStuntNotFound}
	}

	// 2. Player must be a group member and the idea owner
	membership, ok, err := repo.GroupMembership(ctx, input.PlayerID, stunt.GroupID)
	if err != nil {
		return CreatePlannedStuntResult{Err: err}
	}
	if !ok || stunt.PlayerID != input.PlayerID {
		return CreatePlannedStuntResult{Allowed: false}
	}
	_ = membership // role not needed for planning, reserved for future

	// 3. If not explicitly off-season, try to link to an active season
	var seasonID *string
	if !input.OffSeason {
		season, err := repo.ActiveSeasonForGroup(ctx, stunt.GroupID)
		if err != nil {
			return CreatePlannedStuntResult{Err: err}
		}
		if season.ID != "" {
			seasonID = &season.ID
		}
	}

	// 4. Update to Planned Stunt
	updated, err := repo.UpdateStuntToPlanned(ctx, input.IdeaID, input.PlayerID, seasonID)
	if err != nil {
		return CreatePlannedStuntResult{Err: err}
	}

	return CreatePlannedStuntResult{Stunt: updated, Allowed: true}
}
