package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	discordgo "github.com/bwmarrin/discordgo"

	"github.com/supperjumpin/supperjumpin/apps/bot-discord/internal/bot"
)

func TestDispatcher_EndToEnd_StartsRound(t *testing.T) {
	apiCalls := 0
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"round":{"id":"round-test","communityId":"community-1","promptId":"prompt-1","status":"open","revealBy":"2026-06-01T12:00:00Z","createdBy":"user-1","createdAt":"2026-06-01T11:00:00Z","prompt":{"id":"prompt-1","copy":"Test","theme":"x","costTier":"low"}}}`))
	}))
	defer apiServer.Close()

	client, err := bot.NewAPIClient(bot.APIClientConfig{
		BaseURL:      apiServer.URL,
		AdapterToken: "dev-token",
		HTTPClient:   apiServer.Client(),
	})
	if err != nil {
		t.Fatalf("new API client: %v", err)
	}

	resolver := bot.ActorResolverFunc(func(_ context.Context, guildID, userID string) (bot.Actor, error) {
		return bot.Actor{GuildID: guildID, UserID: userID, Tuple: "discord:" + guildID + ":" + userID}, nil
	})
	handlers := bot.BotConfig{
		RoundStart: bot.NewRoundStartHandler(client, resolver).AsHandlerFunc(),
	}
	b := bot.NewBot(handlers)
	dispatcher := NewDispatcher(b)

	transport := &recordingTransport{}
	session, err := discordgo.New("Bot test-token")
	if err != nil {
		t.Fatalf("discordgo.New: %v", err)
	}
	session.Client = &http.Client{Transport: transport}

	event := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "interaction-1",
			AppID:     "app-1",
			Token:     "token-1",
			Type:      discordgo.InteractionApplicationCommand,
			GuildID:   "guild-1",
			ChannelID: "channel-1",
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "round",
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "start"},
					{Name: "communityId", Value: "community-1"},
					{Name: "revealTimeframeId", Value: "tf-1"},
				},
			},
			User: &discordgo.User{ID: "user-1"},
		},
	}

	dispatcher.HandleInteraction(session, event)

	if apiCalls != 1 {
		t.Errorf("API calls: got %d, want 1", apiCalls)
	}

	if got, want := len(transport.requests), 2; got != want {
		t.Fatalf("Discord HTTP requests: got %d, want %d (response + channel message)", got, want)
	}
	req := transport.requests[0]
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
		t.Fatalf("unmarshal body: %v", err)
	}
	if body.Type != 4 {
		t.Errorf("type: got %d, want 4 (ChannelMessageWithSource)", body.Type)
	}
	if body.Data.Content == "" {
		t.Errorf("content: empty, want non-empty reply")
	}
	if body.Data.Flags != 0 {
		t.Errorf("flags: got %d, want 0 (round start is channel-visible)", body.Data.Flags)
	}
}
