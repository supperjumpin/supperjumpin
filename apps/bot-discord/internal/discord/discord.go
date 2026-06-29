package discord

import (
	"context"
	"fmt"
	"strings"

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
	attachments := map[string]*discordgo.MessageAttachment{}
	if i.ApplicationCommandData().Resolved != nil {
		attachments = i.ApplicationCommandData().Resolved.Attachments
	}
	collectOptions(i.ApplicationCommandData().Options, attachments, options, &attachmentURL)
	return bot.IncomingInteraction{
		Type:          bot.InteractionApplicationCommand,
		GuildID:       i.GuildID,
		ChannelID:     i.ChannelID,
		UserID:        interactionUserID(i),
		Command: bot.CommandRoute{
			Name:       i.ApplicationCommandData().Name,
			Subcommand: extractSubcommand(i),
		},
		Options:       options,
		AttachmentURL: attachmentURL,
	}
}

func collectOptions(opts []*discordgo.ApplicationCommandInteractionDataOption, attachments map[string]*discordgo.MessageAttachment, out map[string]string, attachmentURL *string) {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		switch opt.Type {
		case discordgo.ApplicationCommandOptionSubCommand, discordgo.ApplicationCommandOptionSubCommandGroup:
			collectOptions(opt.Options, attachments, out, attachmentURL)
		case discordgo.ApplicationCommandOptionAttachment:
			if attachment, ok := attachments[opt.Value.(string)]; ok {
				*attachmentURL = attachment.URL
			}
		default:
			if s, ok := opt.Value.(string); ok {
				out[snakeToCamel(opt.Name)] = s
			}
		}
	}
}

func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 0 {
		return s
	}
	for i := 1; i < len(parts); i++ {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}

func componentToIncoming(i *discordgo.Interaction) bot.IncomingInteraction {
	return bot.IncomingInteraction{
		Type:      bot.InteractionMessageComponent,
		GuildID:   i.GuildID,
		ChannelID: i.ChannelID,
		UserID:    interactionUserID(i),
		CustomID:  i.MessageComponentData().CustomID,
	}
}

func interactionUserID(i *discordgo.Interaction) string {
	if i == nil {
		return ""
	}
	if i.User != nil {
		return i.User.ID
	}
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}
	return ""
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
