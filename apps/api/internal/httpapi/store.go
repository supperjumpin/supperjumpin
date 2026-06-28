package httpapi

import (
	"context"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

type UpdateDisplayNameResponse struct {
	Player Player `json:"player"`
}

type Store interface {
	BootstrapIdentity(ctx context.Context, identity AuthIdentity) (MeResponse, error)
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
