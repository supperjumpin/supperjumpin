package bot

import (
	"context"
	"fmt"
)

type JumpSubmitHandler struct {
	client   *APIClient
	evidence EvidenceSaver
	resolver ActorResolver
}

func NewJumpSubmitHandler(client *APIClient, evidence EvidenceSaver, resolver ActorResolver) *JumpSubmitHandler {
	return &JumpSubmitHandler{client: client, evidence: evidence, resolver: resolver}
}

func (h *JumpSubmitHandler) Handle(ctx context.Context, i IncomingInteraction) (Reply, error) {
	actor, err := h.resolver.Resolve(ctx, i.GuildID, i.UserID)
	if err != nil {
		return Reply{}, fmt.Errorf("jump submit: resolve actor: %w", err)
	}

	roundID := i.Options["roundId"]
	if roundID == "" {
		return Reply{}, fmt.Errorf("jump submit: missing roundId option")
	}
	caption := i.Options["caption"]
	if caption == "" {
		return Reply{}, fmt.Errorf("jump submit: missing caption option")
	}
	if i.AttachmentURL == "" {
		return Reply{}, fmt.Errorf("jump submit: missing attachment URL")
	}

	stableURL, err := h.evidence.Save(ctx, i.AttachmentURL)
	if err != nil {
		return Reply{}, fmt.Errorf("jump submit: save evidence: %w", err)
	}

	resp, err := h.client.SubmitJump(ctx, roundID, caption, []string{stableURL}, actor)
	if err != nil {
		return Reply{}, fmt.Errorf("jump submit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Reply{}, fmt.Errorf("jump submit: unexpected status %d", resp.StatusCode)
	}

	return Reply{Body: "Your Jump is sealed until reveal.", Ephemeral: true}, nil
}

func (h *JumpSubmitHandler) AsHandlerFunc() HandlerFunc {
	return h.Handle
}
