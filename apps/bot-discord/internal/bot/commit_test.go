package bot

import (
	"context"
	"net/http"
	"testing"
)

func TestCommitHandler_CommitsToRound(t *testing.T) {
	h := newTestHarness(t)

	h.api.responder = func(path string, _ []byte) (int, []byte) {
		if path == "/v1/rounds/round-1/commits" {
			return http.StatusCreated, []byte(`{"commitId":"commit-1"}`)
		}
		return http.StatusNotFound, []byte(`{"error":"not_found","message":"not found"}`)
	}

	handler := NewCommitHandler(h.client, testActorResolver("discord:guild-1:user-1"))

	interaction := IncomingInteraction{
		Type:      InteractionMessageComponent,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		UserID:    "user-1",
		CustomID:  "commit:round-1",
	}

	if _, err := handler.Handle(context.Background(), interaction); err != nil {
		t.Fatalf("handle: %v", err)
	}

	if got, want := len(h.api.received), 1; got != want {
		t.Fatalf("API calls: got %d, want %d", got, want)
	}
	call := h.api.received[0]
	if call.method != http.MethodPost {
		t.Errorf("method: got %q, want POST", call.method)
	}
	if call.path != "/v1/rounds/round-1/commits" {
		t.Errorf("path: got %q, want /v1/rounds/round-1/commits", call.path)
	}
}

func TestParseCommitCustomID(t *testing.T) {
	roundID, ok := ParseCommitCustomID("commit:round-1")
	if !ok {
		t.Fatal("ParseCommitCustomID: got !ok, want ok")
	}
	if roundID != "round-1" {
		t.Errorf("roundID: got %q, want round-1", roundID)
	}

	if _, ok := ParseCommitCustomID("stamp:round-1:jump-1:stamp-1"); ok {
		t.Error("ParseCommitCustomID on stamp custom_id: got ok, want !ok")
	}
	if _, ok := ParseCommitCustomID("commit:"); ok {
		t.Error("ParseCommitCustomID on empty round: got ok, want !ok")
	}
}
