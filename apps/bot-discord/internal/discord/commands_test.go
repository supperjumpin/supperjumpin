package discord

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	discordgo "github.com/bwmarrin/discordgo"
)

func TestRegisterCommands_PostsAllCommands(t *testing.T) {
	transport := &recordingTransport{
		responder: func(_ *http.Request) (int, string) {
			return http.StatusOK, `[]`
		},
	}
	session, err := discordgo.New("Bot test-token")
	if err != nil {
		t.Fatalf("discordgo.New: %v", err)
	}
	session.Client = &http.Client{Transport: transport}

	if err := RegisterCommands(session, "app-1", "guild-1", RoundStartCommand(), JumpSubmitCommand()); err != nil {
		t.Fatalf("RegisterCommands: %v", err)
	}

	if got, want := len(transport.requests), 1; got != want {
		t.Fatalf("HTTP requests: got %d, want %d", got, want)
	}
	req := transport.requests[0]
	if req.method != http.MethodPut {
		t.Errorf("method: got %q, want PUT", req.method)
	}
	if !strings.Contains(req.path, "applications/app-1/guilds/guild-1/commands") {
		t.Errorf("path: got %q, want contains applications/app-1/guilds/guild-1/commands", req.path)
	}

	var cmds []*discordgo.ApplicationCommand
	if err := json.Unmarshal(req.body, &cmds); err != nil {
		t.Fatalf("unmarshal body: %v (body=%q)", err, req.body)
	}
	if got, want := len(cmds), 2; got != want {
		t.Fatalf("commands: got %d, want %d", got, want)
	}

	names := map[string]bool{}
	for _, c := range cmds {
		names[c.Name] = true
	}
	if !names["round"] {
		t.Errorf("commands: missing 'round'")
	}
	if !names["jump"] {
		t.Errorf("commands: missing 'jump'")
	}
}
