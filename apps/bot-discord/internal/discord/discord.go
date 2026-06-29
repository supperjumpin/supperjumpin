package discord

import (
	"context"
	"fmt"

	discordgo "github.com/bwmarrin/discordgo"

	"github.com/supperjumpin/supperjumpin/apps/bot-discord/internal/bot"
)

func EventToIncoming(event *discordgo.InteractionCreate) (bot.IncomingInteraction, error) {
	if event == nil || event.Interaction == nil {
		return bot.IncomingInteraction{}, fmt.Errorf("discord: nil interaction event")
	}
	i := event.Interaction

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		return applicationCommandToIncoming(i), nil
	case discordgo.InteractionMessageComponent:
		return componentToIncoming(i), nil
	default:
		return bot.IncomingInteraction{}, fmt.Errorf("discord: unsupported interaction type %d", i.Type)
	}
}

func applicationCommandToIncoming(i *discordgo.Interaction) bot.IncomingInteraction {
	options := map[string]string{}
	var attachmentURL string
	for _, opt := range i.ApplicationCommandData().Options {
		if opt == nil {
			continue
		}
		if s, ok := opt.Value.(string); ok {
			options[opt.Name] = s
		}
		if opt.Type == discordgo.ApplicationCommandOptionAttachment {
			if attachment, ok := i.ApplicationCommandData().Resolved.Attachments[opt.Value.(string)]; ok {
				attachmentURL = attachment.URL
			}
		}
	}
	return bot.IncomingInteraction{
		Type:          bot.InteractionApplicationCommand,
		GuildID:       i.GuildID,
		ChannelID:     i.ChannelID,
		UserID:        i.User.ID,
		Command: bot.CommandRoute{
			Name:       i.ApplicationCommandData().Name,
			Subcommand: extractSubcommand(i),
		},
		Options:       options,
		AttachmentURL: attachmentURL,
	}
}

func componentToIncoming(i *discordgo.Interaction) bot.IncomingInteraction {
	return bot.IncomingInteraction{
		Type:      bot.InteractionMessageComponent,
		GuildID:   i.GuildID,
		ChannelID: i.ChannelID,
		UserID:    i.User.ID,
		CustomID:  i.MessageComponentData().CustomID,
	}
}

func extractSubcommand(i *discordgo.Interaction) string {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Type == discordgo.ApplicationCommandOptionSubCommand {
			return opt.Name
		}
	}
	return ""
}

type Responder struct {
	session     *discordgo.Session
	interaction *discordgo.Interaction
}

func NewResponder(session *discordgo.Session, interaction *discordgo.Interaction) *Responder {
	return &Responder{session: session, interaction: interaction}
}

func (r *Responder) Respond(ctx context.Context, reply bot.Reply) error {
	data := &discordgo.InteractionResponseData{
		Content: reply.Body,
	}
	if reply.Ephemeral {
		data.Flags = discordgo.MessageFlagsEphemeral
	}
	return r.session.InteractionRespond(r.interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
}

func (r *Responder) PostFollowUp(ctx context.Context, msg bot.FollowUpMessage) error {
	components := make([]discordgo.MessageComponent, 0, len(msg.Buttons))
	row := discordgo.ActionsRow{}
	for _, b := range msg.Buttons {
		style := discordgo.SecondaryButton
		switch b.Style {
		case "Primary":
			style = discordgo.PrimaryButton
		case "Success":
			style = discordgo.SuccessButton
		case "Danger":
			style = discordgo.DangerButton
		}
		row.Components = append(row.Components, &discordgo.Button{
			Label:    b.Label,
			Style:    style,
			CustomID: b.CustomID,
		})
	}
	if len(row.Components) > 0 {
		components = append(components, row)
	}
	_, err := r.session.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{
		Content:    msg.Body,
		Components: components,
	})
	return err
}
