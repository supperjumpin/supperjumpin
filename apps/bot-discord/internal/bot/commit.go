package bot

import (
	"context"
	"fmt"
	"strings"
)

const commitCustomIDPrefix = "commit:"

func ParseCommitCustomID(customID string) (roundID string, ok bool) {
	if !strings.HasPrefix(customID, commitCustomIDPrefix) {
		return "", false
	}
	rest := strings.TrimPrefix(customID, commitCustomIDPrefix)
	if rest == "" {
		return "", false
	}
	return rest, true
}

type CommitHandler struct {
	client   *APIClient
	resolver ActorResolver
}

func NewCommitHandler(client *APIClient, resolver ActorResolver) *CommitHandler {
	return &CommitHandler{client: client, resolver: resolver}
}

func (h *CommitHandler) Handle(ctx context.Context, i IncomingInteraction) (Reply, error) {
	roundID, ok := ParseCommitCustomID(i.CustomID)
	if !ok {
		return Reply{}, fmt.Errorf("commit: invalid custom_id %q", i.CustomID)
	}

	actor, err := h.resolver.Resolve(ctx, i.GuildID, i.UserID)
	if err != nil {
		return Reply{}, fmt.Errorf("commit: resolve actor: %w", err)
	}

	resp, err := h.client.CommitToRound(ctx, roundID, actor)
	if err != nil {
		return Reply{}, fmt.Errorf("commit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Reply{}, fmt.Errorf("commit: unexpected status %d", resp.StatusCode)
	}

	return Reply{Body: "You're in.", Ephemeral: true}, nil
}

func (h *CommitHandler) AsHandlerFunc() HandlerFunc {
	return h.Handle
}
