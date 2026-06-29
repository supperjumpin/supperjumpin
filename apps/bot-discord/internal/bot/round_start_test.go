package bot

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"
)

type recordedCall struct {
	method        string
	path          string
	authorization string
	actor         string
	body          []byte
}

func TestRoundStartHandler_SendsAuthAndActorHeaders(t *testing.T) {
	h := newTestHarness(t)

	interaction := IncomingInteraction{
		Type:    InteractionApplicationCommand,
		GuildID: "guild-1",
		UserID:  "user-1",
		Command: CommandRoute{Name: "round", Subcommand: "start"},
		Options: map[string]string{
			"communityId":       "community-1",
			"revealTimeframeId": "tf-1",
		},
	}

	if _, err := h.handleInteraction(t, interaction); err != nil {
		t.Fatalf("handle: %v", err)
	}

	if got, want := len(h.api.received), 1; got != want {
		t.Fatalf("API calls: got %d, want %d", got, want)
	}
	call := h.api.received[0]
	if call.method != http.MethodPost {
		t.Errorf("method: got %q, want %q", call.method, http.MethodPost)
	}
	if call.path != "/v1/rounds" {
		t.Errorf("path: got %q, want %q", call.path, "/v1/rounds")
	}
	if call.authorization != "Bearer dev-token" {
		t.Errorf("Authorization: got %q, want %q", call.authorization, "Bearer dev-token")
	}
	if call.actor != "discord:guild-1:user-1" {
		t.Errorf("X-Adapter-Actor: got %q, want %q", call.actor, "discord:guild-1:user-1")
	}

	var payload map[string]any
	if err := json.Unmarshal(call.body, &payload); err != nil {
		t.Fatalf("unmarshal body: %v (body=%q)", err, call.body)
	}
	if got, want := payload["communityId"], "community-1"; got != want {
		t.Errorf("communityId: got %v, want %v", got, want)
	}
	if got, want := payload["revealTimeframeId"], "tf-1"; got != want {
		t.Errorf("revealTimeframeId: got %v, want %v", got, want)
	}
	if _, hasPrompt := payload["promptId"]; hasPrompt {
		t.Errorf("promptId: present in body, want absent when not provided (got %v)", payload["promptId"])
	}
}

func TestRoundStartHandler_IncludesPromptIDWhenProvided(t *testing.T) {
	h := newTestHarness(t)

	interaction := IncomingInteraction{
		Type:    InteractionApplicationCommand,
		GuildID: "guild-1",
		UserID:  "user-1",
		Command: CommandRoute{Name: "round", Subcommand: "start"},
		Options: map[string]string{
			"communityId":       "community-1",
			"promptId":          "prompt-42",
			"revealTimeframeId": "tf-1",
		},
	}

	if _, err := h.handleInteraction(t, interaction); err != nil {
		t.Fatalf("handle: %v", err)
	}

	if got, want := len(h.api.received), 1; got != want {
		t.Fatalf("API calls: got %d, want %d", got, want)
	}
	var payload map[string]any
	if err := json.Unmarshal(h.api.received[0].body, &payload); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got, want := payload["promptId"], "prompt-42"; got != want {
		t.Errorf("promptId: got %v, want %v", got, want)
	}
}

type fakeAPI struct {
	mu        sync.Mutex
	received  []recordedCall
	responder func(path string, body []byte) (int, []byte)
}

func (f *fakeAPI) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		f.received = append(f.received, recordedCall{
			method:        r.Method,
			path:          r.URL.Path,
			authorization: r.Header.Get("Authorization"),
			actor:         r.Header.Get("X-Adapter-Actor"),
			body:          body,
		})
		responder := f.responder
		f.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		if responder != nil {
			status, responseBody := responder(r.URL.Path, body)
			w.WriteHeader(status)
			_, _ = w.Write(responseBody)
			return
		}
		_, _ = w.Write([]byte(`{"round":{"id":"round-test","communityId":"community-1","promptId":"prompt-1","status":"open","revealBy":"2026-06-01T12:00:00Z","createdBy":"user-1","createdAt":"2026-06-01T11:00:00Z","prompt":{"id":"prompt-1","copy":"Test","theme":"x","costTier":"low"}}}`))
	})
}

func testActorResolver(actor string) ActorResolver {
	return ActorResolverFunc(func(_ context.Context, guildID, userID string) (Actor, error) {
		return Actor{GuildID: guildID, UserID: userID, Tuple: actor}, nil
	})
}
