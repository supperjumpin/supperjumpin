package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/db"
	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

type ResolveExternalActorResult struct {
	PlayerID    string
	CommunityID string
	Created     bool
}

func (s *PostgresStore) ResolveExternalActor(ctx context.Context, platform, platformServerID, platformUserID, playerDisplayName, communityDisplayName string) (ResolveExternalActorResult, error) {
	existing, err := s.queries.GetExternalIdentity(ctx, db.GetExternalIdentityParams{
		Platform:         platform,
		PlatformServerID: platformServerID,
		PlatformUserID:   platformUserID,
	})
	if err == nil {
		return ResolveExternalActorResult{
			PlayerID:    existing.PlayerID,
			CommunityID: existing.CommunityID,
		}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return ResolveExternalActorResult{}, fmt.Errorf("lookup external identity: %w", err)
	}

	base := platform + ":" + platformServerID
	communityID := stableID("community", base)
	playerID := stableID("player", base+":"+platformUserID)

	now := s.Now()
	result, gErr := game.EnsurePlayer(ctx, s, game.EnsurePlayerInput{
		PlayerID:             playerID,
		PlayerDisplayName:    playerDisplayName,
		CommunityID:          communityID,
		CommunityDisplayName: communityDisplayName,
	}, now)
	if gErr != nil {
		return ResolveExternalActorResult{}, fmt.Errorf("ensure player: %w", gErr)
	}
	if result.Err != nil {
		return ResolveExternalActorResult{}, fmt.Errorf("ensure player: %w", result.Err)
	}

	if err := s.queries.InsertExternalIdentity(ctx, db.InsertExternalIdentityParams{
		Platform:         platform,
		PlatformServerID: platformServerID,
		PlatformUserID:   platformUserID,
		PlayerID:         playerID,
		CommunityID:      communityID,
	}); err != nil {
		return ResolveExternalActorResult{}, fmt.Errorf("insert external identity: %w", err)
	}

	return ResolveExternalActorResult{
		PlayerID:    playerID,
		CommunityID: communityID,
		Created:     result.Created,
	}, nil
}

func (s *PostgresStore) FindPlayer(ctx context.Context, id string) (game.PlayerSnapshot, bool, error) {
	row, err := s.queries.FindPlayer(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return game.PlayerSnapshot{}, false, nil
	}
	if err != nil {
		return game.PlayerSnapshot{}, false, err
	}
	return game.PlayerSnapshot{
		ID:          row.ID,
		DisplayName: row.DisplayName,
		CreatedAt:   row.CreatedAt,
	}, true, nil
}

func (s *PostgresStore) FindCommunity(ctx context.Context, id string) (game.CommunitySnapshot, bool, error) {
	row, err := s.queries.FindCommunity(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return game.CommunitySnapshot{}, false, nil
	}
	if err != nil {
		return game.CommunitySnapshot{}, false, err
	}
	return game.CommunitySnapshot{
		ID:          row.ID,
		DisplayName: row.DisplayName,
		CreatedAt:   row.CreatedAt,
	}, true, nil
}

func (s *PostgresStore) CreateCommunity(ctx context.Context, id string, displayName string, now time.Time) error {
	return s.queries.CreateCommunity(ctx, db.CreateCommunityParams{
		ID:          id,
		DisplayName: displayName,
	})
}

func (s *PostgresStore) CreatePlayer(ctx context.Context, id string, displayName string, now time.Time) error {
	return s.queries.CreatePlayer(ctx, db.CreatePlayerParams{
		ID:          id,
		DisplayName: displayName,
	})
}
