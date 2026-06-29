package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type RoundStartHandler struct {
	client    *APIClient
	resolver  ActorResolver
	scheduler RevealScheduler
	registry  *RoundRegistry
}

func NewRoundStartHandler(client *APIClient, resolver ActorResolver) *RoundStartHandler {
	return &RoundStartHandler{client: client, resolver: resolver}
}

func (h *RoundStartHandler) SetRevealScheduler(s RevealScheduler) {
	h.scheduler = s
}

func (h *RoundStartHandler) SetRegistry(r *RoundRegistry) {
	h.registry = r
}

func (h *RoundStartHandler) Handle(ctx context.Context, i IncomingInteraction) (Reply, error) {
	actor, err := h.resolver.Resolve(ctx, i.GuildID, i.UserID)
	if err != nil {
		return Reply{}, fmt.Errorf("resolve actor: %w", err)
	}

	req := StartRoundRequestBody{
		CommunityID:       i.Options["communityId"],
		RevealTimeframeID: i.Options["revealTimeframeId"],
	}
	if promptID, ok := i.Options["promptId"]; ok && promptID != "" {
		req.PromptID = promptID
	}

	resp, err := h.client.StartRound(ctx, req, actor)
	if err != nil {
		return Reply{}, fmt.Errorf("start round: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Reply{}, fmt.Errorf("start round: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Reply{}, fmt.Errorf("start round: read body: %w", err)
	}

	var parsed struct {
		Round struct {
			ID       string    `json:"id"`
			RevealBy time.Time `json:"revealBy"`
			Prompt   struct {
				Copy string `json:"copy"`
			} `json:"prompt"`
		} `json:"round"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return Reply{}, fmt.Errorf("start round: parse response: %w", err)
	}

	if h.scheduler != nil {
		if err := h.scheduler.Schedule(ctx, parsed.Round.ID, parsed.Round.RevealBy); err != nil {
			return Reply{}, fmt.Errorf("start round: schedule reveal: %w", err)
		}
	}

	if h.registry != nil {
		h.registry.Remember(RoundInfo{
			RoundID:    parsed.Round.ID,
			ChannelID:  i.ChannelID,
			PromptCopy: parsed.Round.Prompt.Copy,
		})
	}

	return Reply{
		Body:      fmt.Sprintf("**Prompt:** %s\n\nRound ID: `%s`\n\nTap **I'm In** below, then `/jump submit` to add your photo.", parsed.Round.Prompt.Copy, parsed.Round.ID),
		FollowUps: []FollowUpMessage{{
			ChannelID: i.ChannelID,
			Body:      fmt.Sprintf("**Round started.**\n**Prompt:** %s\n\nTap **I'm In** to commit, then submit your Jump before reveal.", parsed.Round.Prompt.Copy),
			Buttons: []FollowUpButton{{
				Label:    "I'm In",
				Style:    "Primary",
				CustomID: "commit:" + parsed.Round.ID,
			}},
		}},
	}, nil
}

func (h *RoundStartHandler) AsHandlerFunc() HandlerFunc {
	return h.Handle
}
