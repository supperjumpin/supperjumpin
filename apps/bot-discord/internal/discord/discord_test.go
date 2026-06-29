package discord

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	discordgo "github.com/bwmarrin/discordgo"

	"github.com/supperjumpin/supperjumpin/apps/bot-discord/internal/bot"
)


func TestEventToIncoming_ConvertsSlashCommand(t *testing.T) {
	event := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type:      discordgo.InteractionApplicationCommand,
			GuildID:   "guild-1",
			ChannelID: "channel-1",
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "round",
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{Name: "communityId", Value: "community-1"},
					{Name: "revealTimeframeId", Value: "tf-1"},
				},
			},
			User: &discordgo.User{ID: "user-1"},
		},
	}

	got, err := EventToIncoming(event)
	if err != nil {
		t.Fatalf("EventToIncoming: %v", err)
	}
	if got.GuildID != "guild-1" {
		t.Errorf("GuildID: got %q, want %q", got.GuildID, "guild-1")
	}
	if got.ChannelID != "channel-1" {
		t.Errorf("ChannelID: got %q, want %q", got.ChannelID, "channel-1")
	}
	if got.UserID != "user-1" {
		t.Errorf("UserID: got %q, want %q", got.UserID, "user-1")
	}
	if got.Command.Name != "round" {
		t.Errorf("Command.Name: got %q, want %q", got.Command.Name, "round")
	}
	if got.Command.Subcommand != "" {
		t.Errorf("Command.Subcommand: got %q, want empty (no subcommand in this payload)", got.Command.Subcommand)
	}
	if got, want := got.Options["communityId"], "community-1"; got != want {
		t.Errorf("Options[communityId]: got %q, want %q", got, want)
	}
	if got, want := got.Options["revealTimeframeId"], "tf-1"; got != want {
		t.Errorf("Options[revealTimeframeId]: got %q, want %q", got, want)
	}
}

func TestEventToIncoming_ConvertsButtonComponent(t *testing.T) {
	event := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type:      discordgo.InteractionMessageComponent,
			GuildID:   "guild-1",
			ChannelID: "channel-1",
			Data: discordgo.MessageComponentInteractionData{
				CustomID: "stamp:round-1:jump-7:stamp-hype",
			},
			User: &discordgo.User{ID: "user-1"},
		},
	}

	got, err := EventToIncoming(event)
	if err != nil {
		t.Fatalf("EventToIncoming: %v", err)
	}
	if got.Type != bot.InteractionMessageComponent {
		t.Errorf("Type: got %d, want %d (InteractionMessageComponent)", got.Type, bot.InteractionMessageComponent)
	}
	if got.CustomID != "stamp:round-1:jump-7:stamp-hype" {
		t.Errorf("CustomID: got %q, want %q", got.CustomID, "stamp:round-1:jump-7:stamp-hype")
	}
}

func TestDiscordResponder_SendsEphemeralReply(t *testing.T) {
	transport := &recordingTransport{}
	session, err := discordgo.New("Bot test-token")
	if err != nil {
		t.Fatalf("discordgo.New: %v", err)
	}
	session.Client = &http.Client{Transport: transport}

	original := &discordgo.Interaction{
		ID:        "interaction-1",
		AppID:     "app-1",
		Token:     "token-1",
		Type:      discordgo.InteractionApplicationCommand,
		ChannelID: "channel-1",
	}
	responder := NewResponder(session, original)

	err = responder.Respond(context.Background(), bot.Reply{Body: "Round started.", Ephemeral: true})
	if err != nil {
		t.Fatalf("Respond: %v", err)
	}

	if got, want := len(transport.requests), 1; got != want {
		t.Fatalf("HTTP requests: got %d, want %d", got, want)
	}
	req := transport.requests[0]
	if req.method != http.MethodPost {
		t.Errorf("method: got %q, want %q", req.method, http.MethodPost)
	}
	if !strings.Contains(req.path, "interaction-1") || !strings.Contains(req.path, "token-1") {
		t.Errorf("path: got %q, want contains interaction-1 and token-1", req.path)
	}

	var body struct {
		Type int `json:"type"`
		Data struct {
			Content string `json:"content"`
			Flags   int    `json:"flags"`
		} `json:"data"`
	}
	if err := json.Unmarshal(req.body, &body); err != nil {
		t.Fatalf("unmarshal body: %v (body=%q)", err, req.body)
	}
	if body.Type != 4 {
		t.Errorf("type: got %d, want 4 (ChannelMessageWithSource)", body.Type)
	}
	if body.Data.Content != "Round started." {
		t.Errorf("content: got %q, want %q", body.Data.Content, "Round started.")
	}
	const wantEphemeral = 64
	if body.Data.Flags != wantEphemeral {
		t.Errorf("flags: got %d, want %d (MessageFlagsEphemeral)", body.Data.Flags, wantEphemeral)
	}
}

type recordingTransport struct {
	requests  []recordedRequest
	responder func(req *http.Request) (int, string)
}

type recordedRequest struct {
	method string
	path   string
	body   []byte
}

func (r *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	r.requests = append(r.requests, recordedRequest{
		method: req.Method,
		path:   req.URL.Path,
		body:   body,
	})
	if r.responder != nil {
		status, respBody := r.responder(req)
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(respBody)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}
	if req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/messages") {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"id":"0","channel_id":"0","content":""}`)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       http.NoBody,
		Header:     make(http.Header),
		Request:    req,
	}, nil
}
