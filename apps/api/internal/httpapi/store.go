package httpapi

import (
	"context"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

type UpdateDisplayNameResponse struct {
	Player Player `json:"player"`
}

type Store interface {
	ResolveExternalActor(ctx context.Context, platform string, platformServerID string, platformUserID string, playerDisplayName string, communityDisplayName string) (ResolveExternalActorResult, error)
	FindPlayer(ctx context.Context, id string) (game.PlayerSnapshot, bool, error)
	FindCommunity(ctx context.Context, id string) (game.CommunitySnapshot, bool, error)
	UpdateDisplayName(ctx context.Context, playerID string, displayName string) (Player, error)
	game.ListCatalogRepo
	game.ListRevealTimeframesRepo
	game.StartRoundRepo
	game.CommitRepo
	game.SubmitRepo
	game.ListJumpsRepo
	game.GetJumpRepo
	game.RevealRepo
	game.ListStampCatalogRepo
	game.ApplyReactionRepo
	game.PostCommentRepo
	game.ListCommentsRepo
	game.LoreRepo
	game.RecapRepo
}
