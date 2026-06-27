package game

import (
	"context"
	"time"
)

type PlayerSnapshot struct {
	ID          string
	DisplayName string
	CreatedAt   time.Time
}

type CommunitySnapshot struct {
	ID          string
	DisplayName string
	CreatedAt   time.Time
}

type EnsurePlayerInput struct {
	PlayerID             string
	PlayerDisplayName    string
	CommunityID          string
	CommunityDisplayName string
}

type EnsurePlayerResult struct {
	Player    PlayerSnapshot
	Community CommunitySnapshot
	Created   bool
	Allowed   bool
	Err       error
}

type EnsurePlayerRepo interface {
	FindPlayer(ctx context.Context, id string) (PlayerSnapshot, bool, error)
	FindCommunity(ctx context.Context, id string) (CommunitySnapshot, bool, error)
	CreateCommunity(ctx context.Context, id string, displayName string, now time.Time) error
	CreatePlayer(ctx context.Context, id string, displayName string, now time.Time) error
}

func EnsurePlayer(ctx context.Context, repo EnsurePlayerRepo, input EnsurePlayerInput, now time.Time) (EnsurePlayerResult, error) {
	player, playerExists, err := repo.FindPlayer(ctx, input.PlayerID)
	if err != nil {
		return EnsurePlayerResult{Allowed: false, Err: err}, nil
	}

	community, communityExists, err := repo.FindCommunity(ctx, input.CommunityID)
	if err != nil {
		return EnsurePlayerResult{Allowed: false, Err: err}, nil
	}

	if playerExists && communityExists {
		return EnsurePlayerResult{
			Player:    player,
			Community: community,
			Allowed:   true,
		}, nil
	}

	if !communityExists {
		if err := repo.CreateCommunity(ctx, input.CommunityID, input.CommunityDisplayName, now); err != nil {
			return EnsurePlayerResult{Allowed: false, Err: err}, nil
		}
		community = CommunitySnapshot{
			ID:          input.CommunityID,
			DisplayName: input.CommunityDisplayName,
			CreatedAt:   now,
		}
	}

	if !playerExists {
		if err := repo.CreatePlayer(ctx, input.PlayerID, input.PlayerDisplayName, now); err != nil {
			return EnsurePlayerResult{Allowed: false, Err: err}, nil
		}
		player = PlayerSnapshot{
			ID:          input.PlayerID,
			DisplayName: input.PlayerDisplayName,
			CreatedAt:   now,
		}
	}

	return EnsurePlayerResult{
		Player:    player,
		Community: community,
		Created:   true,
		Allowed:   true,
	}, nil
}
