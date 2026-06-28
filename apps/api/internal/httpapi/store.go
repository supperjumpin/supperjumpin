package httpapi

import (
	"context"
)

type UpdateDisplayNameResponse struct {
	Player Player `json:"player"`
}

type Store interface {
	BootstrapIdentity(ctx context.Context, identity AuthIdentity) (MeResponse, error)
	UpdateDisplayName(ctx context.Context, playerID string, displayName string) (Player, error)
}
