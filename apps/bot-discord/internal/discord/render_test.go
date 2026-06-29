package discord

import (
	"context"
	"testing"

	discordgo "github.com/bwmarrin/discordgo"

	"github.com/supperjumpin/supperjumpin/apps/bot-discord/internal/bot"
)

type fakeChannelPoster struct {
	channelID  string
	content    string
	components []discordgo.MessageComponent
}

func (p *fakeChannelPoster) PostMessage(_ context.Context, channelID string, content string, _ []*discordgo.MessageEmbed, components []discordgo.MessageComponent) error {
	p.channelID = channelID
	p.content = content
	p.components = components
	return nil
}

func collectButtonCustomIDs(components []discordgo.MessageComponent) []string {
	ids := []string{}
	for _, c := range components {
		if row, ok := c.(discordgo.ActionsRow); ok {
			for _, inner := range row.Components {
				if btn, ok := inner.(*discordgo.Button); ok {
					ids = append(ids, btn.CustomID)
				}
			}
		}
	}
	return ids
}

func TestRenderer_BuildsStampButtonsWithRoundJumpStampCustomID(t *testing.T) {
	poster := &fakeChannelPoster{}
	r := NewRenderer(poster)
	r.SetStampTemplate([]StampTemplate{
		{ID: "stamp-approval", Label: "Approve", Glyph: "✅"},
		{ID: "stamp-chaos", Label: "Chaos", Glyph: "🌀"},
	})

	err := r.PostReveal(context.Background(), "channel-1", bot.RecapMessage{
		RoundID:    "round-1",
		PromptCopy: "Test prompt",
		Jumps: []bot.RecapJump{
			{ID: "jump-1", Caption: "first jump", TotalStamps: 0, StampCounts: map[string]int{}},
		},
	})
	if err != nil {
		t.Fatalf("PostReveal: %v", err)
	}

	ids := collectButtonCustomIDs(poster.components)
	wantIDs := []string{"stamp:round-1:jump-1:stamp-approval", "stamp:round-1:jump-1:stamp-chaos"}
	if len(ids) != len(wantIDs) {
		t.Fatalf("button custom_ids: got %v, want %v", ids, wantIDs)
	}
	for i, got := range ids {
		if got != wantIDs[i] {
			t.Errorf("button custom_id[%d]: got %q, want %q", i, got, wantIDs[i])
		}
	}
}

func TestRenderer_BuildsStampButtonsForMultipleJumps(t *testing.T) {
	poster := &fakeChannelPoster{}
	r := NewRenderer(poster)
	r.SetStampTemplate([]StampTemplate{
		{ID: "stamp-approval", Label: "Approve", Glyph: "✅"},
	})

	err := r.PostReveal(context.Background(), "channel-1", bot.RecapMessage{
		RoundID:    "round-1",
		PromptCopy: "Test prompt",
		Jumps: []bot.RecapJump{
			{ID: "jump-1", Caption: "first jump", TotalStamps: 0, StampCounts: map[string]int{}},
			{ID: "jump-2", Caption: "second jump", TotalStamps: 0, StampCounts: map[string]int{}},
		},
	})
	if err != nil {
		t.Fatalf("PostReveal: %v", err)
	}

	ids := collectButtonCustomIDs(poster.components)
	wantIDs := []string{
		"stamp:round-1:jump-1:stamp-approval",
		"stamp:round-1:jump-2:stamp-approval",
	}
	if len(ids) != len(wantIDs) {
		t.Fatalf("button custom_ids: got %v, want %v", ids, wantIDs)
	}
	for i, got := range ids {
		if got != wantIDs[i] {
			t.Errorf("button custom_id[%d]: got %q, want %q", i, got, wantIDs[i])
		}
	}
}
