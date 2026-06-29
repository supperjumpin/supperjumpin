package discord

import (
	"context"
	"fmt"
	"strings"

	discordgo "github.com/bwmarrin/discordgo"

	"github.com/supperjumpin/supperjumpin/apps/bot-discord/internal/bot"
)

// RecapMessage is the minimal projection of a Recap the renderer needs.
// The bot parses the API response into this shape; the renderer formats it.
type RecapMessage struct {
	PromptCopy  string
	Jumps       []RecapJump
}

type RecapJump struct {
	ID          string
	Caption     string
	TotalStamps int
	StampCounts map[string]int
}

type StampTemplate struct {
	ID    string
	Label string
	Glyph string
}

// Renderer formats a Recap and posts it to a channel via the ChannelPoster.
type Renderer struct {
	poster        ChannelPoster
	stampTemplate []StampTemplate
}

func NewRenderer(poster ChannelPoster) *Renderer {
	return &Renderer{poster: poster}
}

func (r *Renderer) SetStampTemplate(stamps []StampTemplate) {
	r.stampTemplate = stamps
}

func (r *Renderer) PostReveal(ctx context.Context, channelID string, recap bot.RecapMessage) error {
	content := formatReveal(recap)
	components := r.buildStampButtons(recap.RoundID, recap.Jumps)
	if err := r.poster.PostMessage(ctx, channelID, content, nil, components); err != nil {
		return fmt.Errorf("renderer: post reveal: %w", err)
	}
	return nil
}

func (r *Renderer) PostRoundAnnouncement(ctx context.Context, channelID, promptCopy string, roundID string) error {
	lines := []string{
		fmt.Sprintf("**Prompt:** %s", promptCopy),
		"",
		"Tap **I'm In** to commit, then submit your Jump before reveal.",
		"",
		fmt.Sprintf("Round ID: `%s`", roundID),
	}
	commit := &discordgo.Button{
		Label:    "I'm In",
		Style:    discordgo.PrimaryButton,
		CustomID: "commit:" + roundID,
	}
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{Components: []discordgo.MessageComponent{commit}},
	}
	return r.poster.PostMessage(ctx, channelID, strings.Join(lines, "\n"), nil, components)
}

func (r *Renderer) buildStampButtons(roundID string, jumps []bot.RecapJump) []discordgo.MessageComponent {
	if len(r.stampTemplate) == 0 || len(jumps) == 0 {
		return nil
	}
	rows := make([]discordgo.MessageComponent, 0, len(jumps))
	for _, j := range jumps {
		buttons := make([]discordgo.MessageComponent, 0, len(r.stampTemplate))
		for _, s := range r.stampTemplate {
			label := s.Label
			if s.Glyph != "" {
				label = s.Glyph + " " + label
			}
			buttons = append(buttons, &discordgo.Button{
				Label:    label,
				Style:    discordgo.SecondaryButton,
				CustomID: "stamp:" + roundID + ":" + j.ID + ":" + s.ID,
			})
		}
		rows = append(rows, discordgo.ActionsRow{Components: buttons})
	}
	return rows
}

func formatReveal(recap bot.RecapMessage) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("**Prompt:** %s", recap.PromptCopy))
	lines = append(lines, "")
	if len(recap.Jumps) == 0 {
		lines = append(lines, "_No Jumps submitted._")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, fmt.Sprintf("**%d Jumps revealed:**", len(recap.Jumps)))
	for i, j := range recap.Jumps {
		lines = append(lines, fmt.Sprintf("%d. **%s** — %d stamps", i+1, j.Caption, j.TotalStamps))
	}
	return strings.Join(lines, "\n")
}
