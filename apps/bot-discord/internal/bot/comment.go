package bot

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type CommentHandler struct {
	client   *APIClient
	resolver ActorResolver
}

func NewCommentHandler(client *APIClient, resolver ActorResolver) *CommentHandler {
	return &CommentHandler{client: client, resolver: resolver}
}

func (h *CommentHandler) Handle(ctx context.Context, i IncomingInteraction) (Reply, error) {
	actor, err := h.resolver.Resolve(ctx, i.GuildID, i.UserID)
	if err != nil {
		return Reply{}, fmt.Errorf("comment: resolve actor: %w", err)
	}

	roundID := i.Options["roundId"]
	body := i.Options["body"]
	if roundID == "" || body == "" {
		return Reply{}, fmt.Errorf("comment: missing roundId or body")
	}

	var resp *http.Response
	var err2 error
	if jumpID := i.Options["jumpId"]; jumpID != "" {
		resp, err2 = h.client.PostJumpComment(ctx, roundID, jumpID, body, actor)
	} else {
		resp, err2 = h.client.PostRoundComment(ctx, roundID, body, actor)
	}
	if err2 != nil {
		return Reply{}, fmt.Errorf("comment: %w", err2)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return Reply{}, fmt.Errorf("comment: unexpected status %d: %s", resp.StatusCode, string(raw))
	}

	return Reply{Body: "Comment posted.", Ephemeral: true}, nil
}

func (h *CommentHandler) AsHandlerFunc() HandlerFunc {
	return h.Handle
}
