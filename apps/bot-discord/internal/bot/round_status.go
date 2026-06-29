package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

type RoundStatusHandler struct {
	client   *APIClient
	resolver ActorResolver
}

func NewRoundStatusHandler(client *APIClient, resolver ActorResolver) *RoundStatusHandler {
	return &RoundStatusHandler{client: client, resolver: resolver}
}

func (h *RoundStatusHandler) Handle(ctx context.Context, i IncomingInteraction) (Reply, error) {
	actor, err := h.resolver.Resolve(ctx, i.GuildID, i.UserID)
	if err != nil {
		return Reply{}, fmt.Errorf("round status: resolve actor: %w", err)
	}

	roundID := i.Options["roundId"]
	if roundID == "" {
		return Reply{}, fmt.Errorf("round status: missing roundId option")
	}

	resp, err := h.client.ListRoundJumps(ctx, roundID, actor)
	if err != nil {
		return Reply{}, fmt.Errorf("round status: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Reply{}, fmt.Errorf("round status: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Reply{}, fmt.Errorf("round status: read body: %w", err)
	}

	var parsed struct {
		CommitCount    int `json:"commitCount"`
		SubmissionCount int `json:"submissionCount"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return Reply{}, fmt.Errorf("round status: parse: %w", err)
	}

	return Reply{
		Body:      fmt.Sprintf("Round %s: %d committed, %d submitted.", roundID, parsed.CommitCount, parsed.SubmissionCount),
		Ephemeral: true,
	}, nil
}

func (h *RoundStatusHandler) AsHandlerFunc() HandlerFunc {
	return h.Handle
}
