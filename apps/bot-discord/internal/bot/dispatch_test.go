package bot

import (
	"context"
	"testing"
)

func TestBot_DispatchesRoundStartInteractionToHandler(t *testing.T) {
	h := newTestHarness(t)

	responder := &fakeResponder{}
	b := NewBot(BotConfig{
		RoundStart: h.handler.AsHandlerFunc(),
	})

	interaction := IncomingInteraction{
		Type:      InteractionApplicationCommand,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		UserID:    "user-1",
		Command:   CommandRoute{Name: "round", Subcommand: "start"},
		Options: map[string]string{
			"communityId":       "community-1",
			"revealTimeframeId": "tf-1",
		},
	}

	if err := b.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	if got, want := len(h.api.received), 1; got != want {
		t.Fatalf("API calls: got %d, want %d", got, want)
	}
	if got, want := h.api.received[0].path, "/v1/rounds"; got != want {
		t.Errorf("path: got %q, want %q", got, want)
	}
	if got, want := h.api.received[0].authorization, "Bearer dev-token"; got != want {
		t.Errorf("Authorization: got %q, want %q", got, want)
	}
	if got, want := h.api.received[0].actor, "discord:guild-1:user-1"; got != want {
		t.Errorf("X-Adapter-Actor: got %q, want %q", got, want)
	}

	if got, want := len(responder.captured), 1; got != want {
		t.Fatalf("responder calls: got %d, want %d", got, want)
	}
	got := responder.captured[0]
	if got.body == "" {
		t.Errorf("responder body: empty, want non-empty reply")
	}
	if got.ephemeral {
		t.Errorf("responder ephemeral: got true, want false (round start is channel-visible)")
	}
}

func TestBot_DispatchesStampComponentToHandler(t *testing.T) {
	h := newTestHarness(t)
	stampHandler := NewStampApplyHandler(h.client, testActorResolver("discord:guild-1:user-1"))

	responder := &fakeResponder{}
	b := NewBot(BotConfig{
		StampApply: stampHandler.AsHandlerFunc(),
	})

	interaction := IncomingInteraction{
		Type:      InteractionMessageComponent,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		UserID:    "user-1",
		CustomID:  "stamp:round-1:jump-7:stamp-hype",
	}

	if err := b.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	if got, want := len(h.api.received), 1; got != want {
		t.Fatalf("API calls: got %d, want %d", got, want)
	}
	if got, want := h.api.received[0].path, "/v1/rounds/round-1/jumps/jump-7/reactions"; got != want {
		t.Errorf("path: got %q, want %q", got, want)
	}
}

func TestBot_DispatchesJumpCommentCommandToHandler(t *testing.T) {
	h := newTestHarness(t)
	h.api.responder = func(path string, _ []byte) (int, []byte) {
		if path == "/v1/rounds/round-1/jumps/jump-1/comments" {
			return 201, []byte(`{"comment":{"id":"comment-1"}}`)
		}
		return 404, []byte(`{"error":"not_found"}`)
	}
	commentHandler := NewCommentHandler(h.client, testActorResolver("discord:guild-1:user-1"))

	responder := &fakeResponder{}
	b := NewBot(BotConfig{Comment: commentHandler.AsHandlerFunc()})

	interaction := IncomingInteraction{
		Type:    InteractionApplicationCommand,
		GuildID: "guild-1",
		UserID:  "user-1",
		Command: CommandRoute{Name: "comment", Subcommand: "jump"},
		Options: map[string]string{"roundId": "round-1", "jumpId": "jump-1", "body": "hello"},
	}

	if err := b.Dispatch(context.Background(), interaction, responder); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if got, want := len(h.api.received), 1; got != want {
		t.Fatalf("API calls: got %d, want %d", got, want)
	}
	if got, want := h.api.received[0].path, "/v1/rounds/round-1/jumps/jump-1/comments"; got != want {
		t.Fatalf("path: got %q, want %q", got, want)
	}
}
