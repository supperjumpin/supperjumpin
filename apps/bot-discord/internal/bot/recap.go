package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

type RecapHandler struct {
	client   *APIClient
	resolver ActorResolver
}

func NewRecapHandler(client *APIClient, resolver ActorResolver) *RecapHandler {
	return &RecapHandler{client: client, resolver: resolver}
}

func (h *RecapHandler) Handle(ctx context.Context, i IncomingInteraction) (Reply, error) {
	actor, err := h.resolver.Resolve(ctx, i.GuildID, i.UserID)
	if err != nil {
		return Reply{}, fmt.Errorf("recap: resolve actor: %w", err)
	}

	roundID := i.Options["roundId"]
	if roundID == "" {
		return Reply{}, fmt.Errorf("recap: missing roundId")
	}

	resp, err := h.client.GetRecap(ctx, roundID, actor)
	if err != nil {
		return Reply{}, fmt.Errorf("recap: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Reply{}, fmt.Errorf("recap: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Reply{}, fmt.Errorf("recap: read body: %w", err)
	}

	var parsed struct {
		Jumps []struct {
			Caption     string `json:"caption"`
			TotalStamps int    `json:"totalStamps"`
		} `json:"jumps"`
		Comments []struct {
			Body string `json:"body"`
		} `json:"comments"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return Reply{}, fmt.Errorf("recap: parse: %w", err)
	}

	text := fmt.Sprintf("**Recap for %s**\n", roundID)
	if len(parsed.Jumps) > 0 {
		text += fmt.Sprintf("%d jumps:\n", len(parsed.Jumps))
		for _, j := range parsed.Jumps {
			text += fmt.Sprintf("- %s (%d stamps)\n", j.Caption, j.TotalStamps)
		}
	} else {
		text += "No jumps.\n"
	}
	if len(parsed.Comments) > 0 {
		text += fmt.Sprintf("\n%d comments:\n", len(parsed.Comments))
		for _, c := range parsed.Comments {
			text += fmt.Sprintf("> %s\n", c.Body)
		}
	}

	return Reply{Body: text, Ephemeral: true}, nil
}

func (h *RecapHandler) AsHandlerFunc() HandlerFunc {
	return h.Handle
}
