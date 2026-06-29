package bot

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestStampApplyHandler_AppliesReactionToJump(t *testing.T) {
	h := newTestHarness(t)

	handler := NewStampApplyHandler(h.client, testActorResolver("discord:guild-1:user-1"))

	interaction := IncomingInteraction{
		Type:      InteractionMessageComponent,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		UserID:    "user-1",
		CustomID:  "stamp:round-1:jump-7:stamp-hype",
	}

	reply, err := handler.Handle(context.Background(), interaction)
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !reply.Ephemeral {
		t.Fatal("expected ephemeral reply")
	}
	if got, want := reply.Body, "Stamp applied."; got != want {
		t.Fatalf("reply.Body: got %q, want %q", got, want)
	}

	if got, want := len(h.api.received), 1; got != want {
		t.Fatalf("API calls: got %d, want %d", got, want)
	}
	call := h.api.received[0]
	if call.method != http.MethodPost {
		t.Errorf("method: got %q, want %q", call.method, http.MethodPost)
	}
	if call.path != "/v1/rounds/round-1/jumps/jump-7/reactions" {
		t.Errorf("path: got %q, want %q", call.path, "/v1/rounds/round-1/jumps/jump-7/reactions")
	}
	if call.authorization != "Bearer dev-token" {
		t.Errorf("Authorization: got %q, want %q", call.authorization, "Bearer dev-token")
	}
	if call.actor != "discord:guild-1:user-1" {
		t.Errorf("X-Adapter-Actor: got %q, want %q", call.actor, "discord:guild-1:user-1")
	}

	var payload map[string]any
	if err := json.Unmarshal(call.body, &payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got, want := payload["stampId"], "stamp-hype"; got != want {
		t.Errorf("stampId: got %v, want %v", got, want)
	}
}

func TestStampApplyHandler_TreatsDuplicateReactionAsEphemeralSuccess(t *testing.T) {
	h := newTestHarness(t)
	h.api.responder = func(path string, _ []byte) (int, []byte) {
		if path == "/v1/rounds/round-1/jumps/jump-7/reactions" {
			return http.StatusForbidden, []byte(`player already reacted with this stamp on this jump`)
		}
		return http.StatusNotFound, []byte(`{"error":"not_found"}`)
	}

	handler := NewStampApplyHandler(h.client, testActorResolver("discord:guild-1:user-1"))

	interaction := IncomingInteraction{
		Type:      InteractionMessageComponent,
		GuildID:   "guild-1",
		ChannelID: "channel-1",
		UserID:    "user-1",
		CustomID:  "stamp:round-1:jump-7:stamp-hype",
	}

	reply, err := handler.Handle(context.Background(), interaction)
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !reply.Ephemeral {
		t.Fatal("expected ephemeral reply")
	}
	if got, want := reply.Body, "Stamp already applied."; got != want {
		t.Fatalf("reply.Body: got %q, want %q", got, want)
	}
}
