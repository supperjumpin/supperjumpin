package bot

import (
	"context"
	"fmt"
	"strings"
)

type HandlerFunc func(ctx context.Context, i IncomingInteraction) (Reply, error)

type BotConfig struct {
	RoundStart HandlerFunc
	RoundStatus HandlerFunc
	JumpSubmit  HandlerFunc
	StampApply  HandlerFunc
	Commit      HandlerFunc
	Comment     HandlerFunc
	Recap       HandlerFunc
}

type Bot struct {
	commands    map[CommandRoute]HandlerFunc
	components  map[string]HandlerFunc
}

func NewBot(cfg BotConfig) *Bot {
	b := &Bot{
		commands:   map[CommandRoute]HandlerFunc{},
		components: map[string]HandlerFunc{},
	}
	if cfg.RoundStart != nil {
		b.commands[CommandRoute{Name: "round", Subcommand: "start"}] = cfg.RoundStart
	}
	if cfg.RoundStatus != nil {
		b.commands[CommandRoute{Name: "round", Subcommand: "status"}] = cfg.RoundStatus
	}
	if cfg.JumpSubmit != nil {
		b.commands[CommandRoute{Name: "jump", Subcommand: "submit"}] = cfg.JumpSubmit
	}
	if cfg.Commit != nil {
		b.components[commitCustomIDPrefix] = cfg.Commit
	}
	if cfg.StampApply != nil {
		b.components[stampCustomIDPrefix] = cfg.StampApply
	}
	if cfg.Comment != nil {
		b.commands[CommandRoute{Name: "comment", Subcommand: "round"}] = cfg.Comment
	}
	if cfg.Recap != nil {
		b.commands[CommandRoute{Name: "recap", Subcommand: ""}] = cfg.Recap
	}
	return b
}

func (b *Bot) Dispatch(ctx context.Context, i IncomingInteraction, res Responder) error {
	handler, err := b.route(i)
	if err != nil {
		return err
	}
	reply, err := handler(ctx, i)
	if err != nil {
		return err
	}
	if res == nil {
		return nil
	}
	if err := res.Respond(ctx, reply); err != nil {
		return err
	}
	for _, fu := range reply.FollowUps {
		if err := res.PostFollowUp(ctx, fu); err != nil {
			return fmt.Errorf("post follow-up: %w", err)
		}
	}
	return nil
}

func (b *Bot) route(i IncomingInteraction) (HandlerFunc, error) {
	switch i.Type {
	case InteractionApplicationCommand:
		h, ok := b.commands[i.Command]
		if !ok {
			return nil, fmt.Errorf("bot: no handler for command %q", i.Command)
		}
		return h, nil
	case InteractionMessageComponent:
		for prefix, h := range b.components {
			if strings.HasPrefix(i.CustomID, prefix) {
				return h, nil
			}
		}
		return nil, fmt.Errorf("bot: no handler for component custom_id %q", i.CustomID)
	default:
		return nil, fmt.Errorf("bot: unsupported interaction type %d", i.Type)
	}
}
