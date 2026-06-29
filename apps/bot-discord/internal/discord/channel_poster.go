package discord

import (
	"context"
	"fmt"

	discordgo "github.com/bwmarrin/discordgo"
)

type ChannelPoster interface {
	PostMessage(ctx context.Context, channelID string, content string, embeds []*discordgo.MessageEmbed, components []discordgo.MessageComponent) error
}

type sessionChannelPoster struct {
	session *discordgo.Session
}

func NewSessionChannelPoster(session *discordgo.Session) ChannelPoster {
	return &sessionChannelPoster{session: session}
}

func (p *sessionChannelPoster) PostMessage(_ context.Context, channelID string, content string, embeds []*discordgo.MessageEmbed, components []discordgo.MessageComponent) error {
	data := &discordgo.MessageSend{
		Content:    content,
		Embeds:     embeds,
		Components: components,
	}
	if _, err := p.session.ChannelMessageSendComplex(channelID, data); err != nil {
		return fmt.Errorf("discord: post message: %w", err)
	}
	return nil
}
