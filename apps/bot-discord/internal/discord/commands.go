package discord

import (
	"fmt"

	discordgo "github.com/bwmarrin/discordgo"
)

func RegisterCommands(session *discordgo.Session, appID, guildID string, commands ...*discordgo.ApplicationCommand) error {
	if _, err := session.ApplicationCommandBulkOverwrite(appID, guildID, commands); err != nil {
		return fmt.Errorf("discord: register commands: %w", err)
	}
	return nil
}

func RoundStartCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "round",
		Description: "Start or inspect a Round.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "start",
				Description: "Start a new Round for a Community.",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{Type: discordgo.ApplicationCommandOptionString, Name: "community_id", Description: "Community ID", Required: true},
					{Type: discordgo.ApplicationCommandOptionString, Name: "prompt_id", Description: "Prompt ID (omit for random pick)", Required: false},
					{Type: discordgo.ApplicationCommandOptionString, Name: "reveal_timeframe_id", Description: "Reveal timeframe ID (e.g. tf-1h)", Required: true},
				},
			},
			{
				Name:        "status",
				Description: "Show the current Round's commit and submission counts.",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	}
}

func JumpSubmitCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "jump",
		Description: "Submit or inspect Jumps.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "submit",
				Description: "Submit a sealed Jump for an active Round.",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{Type: discordgo.ApplicationCommandOptionString, Name: "round_id", Description: "Round ID", Required: true},
					{Type: discordgo.ApplicationCommandOptionString, Name: "caption", Description: "Caption", Required: true},
					{Type: discordgo.ApplicationCommandOptionAttachment, Name: "photo", Description: "Photo evidence", Required: true},
				},
			},
		},
	}
}

func CommentCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "comment",
		Description: "Post a comment on a Round or Jump.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "round",
				Description: "Post a comment on a Round.",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{Type: discordgo.ApplicationCommandOptionString, Name: "round_id", Description: "Round ID", Required: true},
					{Type: discordgo.ApplicationCommandOptionString, Name: "body", Description: "Comment body", Required: true},
				},
			},
			{
				Name:        "jump",
				Description: "Post a comment on a Jump.",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{Type: discordgo.ApplicationCommandOptionString, Name: "round_id", Description: "Round ID", Required: true},
					{Type: discordgo.ApplicationCommandOptionString, Name: "jump_id", Description: "Jump ID", Required: true},
					{Type: discordgo.ApplicationCommandOptionString, Name: "body", Description: "Comment body", Required: true},
				},
			},
		},
	}
}

func RecapCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "recap",
		Description: "Show the Recap for a revealed Round.",
		Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "round_id", Description: "Round ID", Required: true},
		},
	}
}
