package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

type RevealActor struct {
	Client   *APIClient
	Resolver ActorResolver
	Registry *RoundRegistry
	Poster   RecapPoster
}

func (a *RevealActor) Fire(roundID string) error {
	return a.FireCtx(context.Background(), roundID)
}

func (a *RevealActor) FireCtx(ctx context.Context, roundID string) error {
	actor, err := a.Resolver.Resolve(ctx, "", "")
	if err != nil {
		return fmt.Errorf("reveal: resolve actor: %w", err)
	}
	resp, err := a.Client.EvaluateReveal(ctx, roundID, actor)
	if err != nil {
		return fmt.Errorf("reveal: evaluate: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("reveal: evaluate status %d: %s", resp.StatusCode, string(body))
	}

	info, ok := a.Registry.Get(roundID)
	if !ok {
		return fmt.Errorf("reveal: no registry entry for round %q", roundID)
	}

	recapResp, err := a.Client.GetRecap(ctx, roundID, actor)
	if err != nil {
		return fmt.Errorf("reveal: get recap: %w", err)
	}
	defer recapResp.Body.Close()
	if recapResp.StatusCode < 200 || recapResp.StatusCode >= 300 {
		body, _ := io.ReadAll(recapResp.Body)
		return fmt.Errorf("reveal: get recap status %d: %s", recapResp.StatusCode, string(body))
	}

	body, err := io.ReadAll(recapResp.Body)
	if err != nil {
		return fmt.Errorf("reveal: read recap: %w", err)
	}

	var parsed struct {
		Jumps []struct {
			JumpID      string         `json:"jumpId"`
			Caption     string         `json:"caption"`
			TotalStamps int            `json:"totalStamps"`
			StampCounts map[string]int `json:"stampCounts"`
		} `json:"jumps"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return fmt.Errorf("reveal: parse recap: %w", err)
	}

	jumps := make([]RecapJump, 0, len(parsed.Jumps))
	for _, j := range parsed.Jumps {
		jumps = append(jumps, RecapJump{
			ID:          j.JumpID,
			Caption:     j.Caption,
			TotalStamps: j.TotalStamps,
			StampCounts: j.StampCounts,
		})
	}

	return a.Poster.PostReveal(ctx, info.ChannelID, RecapMessage{
		PromptCopy: info.PromptCopy,
		Jumps:      jumps,
	})
}
