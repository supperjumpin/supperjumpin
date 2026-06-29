package bot

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

type fakeEvidenceSaver struct {
	received []string
	url      string
	err      error
}

func (f *fakeEvidenceSaver) Save(_ context.Context, sourceURL string) (string, error) {
	f.received = append(f.received, sourceURL)
	if f.err != nil {
		return "", f.err
	}
	return f.url, nil
}

func TestJumpSubmitHandler_DownloadsEvidenceAndCallsAPI(t *testing.T) {
	h := newTestHarness(t)

	h.api.responder = func(path string, _ []byte) (int, []byte) {
		if path == "/v1/rounds/round-1/jumps" {
			return http.StatusCreated, []byte(`{"jump":{"id":"jump-1","roundId":"round-1","playerId":"user-1","caption":"Test caption","evidenceUrls":["http://localhost:9999/evidence/abc.png"],"submittedAt":"2026-06-01T12:00:00Z","sealedViewer":true,"playerHasCommitted":true,"playerHasSubmitted":true}}`)
		}
		return http.StatusNotFound, []byte(`{"error":"not_found","message":"not found"}`)
	}

	saver := &fakeEvidenceSaver{url: "http://localhost:9999/evidence/abc.png"}

	handler := NewJumpSubmitHandler(h.client, saver, testActorResolver("discord:guild-1:user-1"))

	interaction := IncomingInteraction{
		Type:      InteractionApplicationCommand,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		UserID:    "user-1",
		Command:   CommandRoute{Name: "jump", Subcommand: "submit"},
		Options: map[string]string{
			"roundId": "round-1",
			"caption": "Test caption",
		},
		AttachmentURL: "https://cdn.discordapp.com/attachments/1/2/photo.png",
	}

	reply, err := handler.Handle(context.Background(), interaction)
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !reply.Ephemeral {
		t.Errorf("reply.Ephemeral: got false, want true (per ADR-0043)")
	}
	if reply.Body == "" {
		t.Errorf("reply.Body: empty, want non-empty")
	}

	if got, want := len(saver.received), 1; got != want {
		t.Fatalf("evidence Save calls: got %d, want %d", got, want)
	}
	if got, want := saver.received[0], "https://cdn.discordapp.com/attachments/1/2/photo.png"; got != want {
		t.Errorf("evidence source URL: got %q, want %q", got, want)
	}

	if got, want := len(h.api.received), 1; got != want {
		t.Fatalf("API calls: got %d, want %d", got, want)
	}
	call := h.api.received[0]
	if got, want := call.path, "/v1/rounds/round-1/jumps"; got != want {
		t.Errorf("path: got %q, want %q", got, want)
	}

	var payload map[string]any
	if err := json.Unmarshal(call.body, &payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got, want := payload["caption"], "Test caption"; got != want {
		t.Errorf("caption: got %v, want %v", got, want)
	}
	urls, ok := payload["evidenceUrls"].([]any)
	if !ok {
		t.Fatalf("evidenceUrls: not an array, got %T", payload["evidenceUrls"])
	}
	if len(urls) != 1 {
		t.Fatalf("evidenceUrls: got %d, want 1", len(urls))
	}
	gotURL, ok := urls[0].(string)
	if !ok {
		t.Fatalf("evidenceUrls[0]: not a string, got %T", urls[0])
	}
	if !strings.HasPrefix(gotURL, "http://localhost:9999/evidence/") {
		t.Errorf("evidenceUrls[0]: got %q, want prefix http://localhost:9999/evidence/", gotURL)
	}
}
