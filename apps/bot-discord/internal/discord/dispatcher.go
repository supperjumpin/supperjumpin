package discord

import (
	"context"
	"log/slog"

	discordgo "github.com/bwmarrin/discordgo"

	"github.com/supperjumpin/supperjumpin/apps/bot-discord/internal/bot"
)

type Dispatcher struct {
	bot    *bot.Bot
	logger *slog.Logger
}

func NewDispatcher(b *bot.Bot, opts ...DispatcherOption) *Dispatcher {
	d := &Dispatcher{bot: b, logger: slog.Default()}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

type DispatcherOption func(*Dispatcher)

func WithLogger(l *slog.Logger) DispatcherOption {
	return func(d *Dispatcher) { d.logger = l }
}

func (d *Dispatcher) HandleInteraction(s *discordgo.Session, event *discordgo.InteractionCreate) {
	incoming, err := EventToIncoming(event)
	if err != nil {
		d.logger.Warn("discord: failed to convert interaction", "err", err)
		return
	}
	responder := NewResponder(s, event.Interaction)
	if err := d.bot.Dispatch(context.Background(), incoming, responder); err != nil {
		d.logger.Warn("discord: dispatch failed", "err", err, "command", incoming.Command)
	}
}
