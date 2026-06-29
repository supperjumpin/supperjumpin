package bot

import (
	"context"
	"fmt"
)

type StampApplyHandler struct {
	client   *APIClient
	resolver ActorResolver
}

func NewStampApplyHandler(client *APIClient, resolver ActorResolver) *StampApplyHandler {
	return &StampApplyHandler{client: client, resolver: resolver}
}

func (h *StampApplyHandler) Handle(ctx context.Context, i IncomingInteraction) (Reply, error) {
	roundID, jumpID, stampID, ok := ParseStampCustomID(i.CustomID)
	if !ok {
		return Reply{}, fmt.Errorf("stamp apply: invalid custom_id %q", i.CustomID)
	}

	actor, err := h.resolver.Resolve(ctx, i.GuildID, i.UserID)
	if err != nil {
		return Reply{}, fmt.Errorf("stamp apply: resolve actor: %w", err)
	}

	resp, err := h.client.ApplyReaction(ctx, roundID, jumpID, stampID, actor)
	if err != nil {
		return Reply{}, fmt.Errorf("stamp apply: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Reply{}, fmt.Errorf("stamp apply: unexpected status %d", resp.StatusCode)
	}

	return Reply{}, nil
}

func (h *StampApplyHandler) AsHandlerFunc() HandlerFunc {
	return h.Handle
}
