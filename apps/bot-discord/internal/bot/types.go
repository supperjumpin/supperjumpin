package bot

import (
	"context"
	"time"
)

type InteractionType int

const (
	InteractionApplicationCommand InteractionType = iota + 1
	InteractionMessageComponent
)

type CommandRoute struct {
	Name       string
	Subcommand string
}

type IncomingInteraction struct {
	Type          InteractionType
	GuildID       string
	ChannelID     string
	UserID        string
	Command       CommandRoute
	CustomID      string
	Options       map[string]string
	AttachmentURL string
}

type Reply struct {
	Body      string
	Ephemeral bool
	Flags     uint64
	FollowUps []FollowUpMessage
}

// FollowUpMessage is a follow-up the bot posts to a channel AFTER responding
// to an interaction. Used for the round announcement with the "I'm In" button.
type FollowUpMessage struct {
	ChannelID string
	Body      string
	Buttons   []FollowUpButton
}

type FollowUpButton struct {
	Label    string
	Style    string
	CustomID string
}

type Responder interface {
	Respond(ctx context.Context, reply Reply) error
	PostFollowUp(ctx context.Context, msg FollowUpMessage) error
}

type RevealScheduler interface {
	Schedule(ctx context.Context, roundID string, revealBy time.Time) error
}

type EvidenceSaver interface {
	Save(ctx context.Context, sourceURL string) (string, error)
}
