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

func TestRoundStartCommand_PutsRequiredOptionsBeforeOptional(t *testing.T) {
	cmd := RoundStartCommand()
	if got, want := len(cmd.Options), 2; got != want {
		t.Fatalf("top-level options: got %d, want %d", got, want)
	}
	start := cmd.Options[0]
	if start.Name != "start" {
		t.Fatalf("first subcommand: got %q, want %q", start.Name, "start")
	}
	if got, want := len(start.Options), 3; got != want {
		t.Fatalf("start options: got %d, want %d", got, want)
	}
	if got, want := start.Options[0].Name, "community_id"; got != want {
		t.Errorf("start.Options[0].Name: got %q, want %q", got, want)
	}
	if got, want := start.Options[1].Name, "reveal_timeframe_id"; got != want {
		t.Errorf("start.Options[1].Name: got %q, want %q", got, want)
	}
	if got, want := start.Options[2].Name, "prompt_id"; got != want {
		t.Errorf("start.Options[2].Name: got %q, want %q", got, want)
	}
	if !start.Options[0].Required || !start.Options[1].Required {
		t.Error("required options must remain required")
	}
	if start.Options[2].Required {
		t.Error("prompt_id: got required=true, want false")
	}
}
