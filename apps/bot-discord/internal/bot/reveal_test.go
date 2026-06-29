package bot

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

type fakeRecapPoster struct {
	posted    int
	lastRecap RecapMessage
	lastCh    string
}

func (f *fakeRecapPoster) PostReveal(_ context.Context, channelID string, recap RecapMessage) error {
	f.posted++
	f.lastCh = channelID
	f.lastRecap = recap
	return nil
}

func TestRevealActor_FireCallsRevealThenGetRecapThenPost(t *testing.T) {
	h := newTestHarness(t)
	registry := NewRoundRegistry()
	registry.Remember(RoundInfo{
		RoundID:    "round-1",
		ChannelID:  "channel-1",
		PromptCopy: "Test prompt",
	})
	poster := &fakeRecapPoster{}
	actor := &RevealActor{
		Client:   h.client,
		Resolver: testActorResolver("discord::"),
		Registry: registry,
		Poster:   poster,
	}

	h.api.responder = func(path string, _ []byte) (int, []byte) {
		switch path {
		case "/v1/rounds/round-1/reveal":
			return http.StatusOK, []byte(`{"round":{"id":"round-1","status":"revealed"},"revealed":true}`)
		case "/v1/rounds/round-1/recap":
			return http.StatusOK, []byte(`{
				"jumps": [
					{"jumpId":"jump-1","caption":"First","totalStamps":3,"stampCounts":{"stamp-hype":3}}
				]
			}`)
		}
		return http.StatusNotFound, []byte(`{"error":"not_found","message":"not found"}`)
	}

	if err := actor.Fire("round-1"); err != nil {
		t.Fatalf("Fire: %v", err)
	}

	if got, want := poster.posted, 1; got != want {
		t.Fatalf("posted: got %d, want %d", got, want)
	}
	if poster.lastCh != "channel-1" {
		t.Errorf("channelID: got %q, want channel-1", poster.lastCh)
	}
	if poster.lastRecap.PromptCopy != "Test prompt" {
		t.Errorf("PromptCopy: got %q, want %q", poster.lastRecap.PromptCopy, "Test prompt")
	}
	if got, want := len(poster.lastRecap.Jumps), 1; got != want {
		t.Fatalf("Jumps: got %d, want %d", got, want)
	}
	if poster.lastRecap.Jumps[0].ID != "jump-1" {
		t.Errorf("Jump[0].ID: got %q, want jump-1", poster.lastRecap.Jumps[0].ID)
	}
}

func TestRevealActor_FireFailsWithoutRegistryEntry(t *testing.T) {
	h := newTestHarness(t)
	registry := NewRoundRegistry()
	poster := &fakeRecapPoster{}
	actor := &RevealActor{
		Client:   h.client,
		Resolver: testActorResolver("discord::"),
		Registry: registry,
		Poster:   poster,
	}

	h.api.responder = func(path string, _ []byte) (int, []byte) {
		if path == "/v1/rounds/round-1/reveal" {
			return http.StatusOK, []byte(`{"round":{"id":"round-1","status":"revealed"},"revealed":true}`)
		}
		return http.StatusNotFound, []byte(`{"error":"not_found","message":"not found"}`)
	}

	err := actor.Fire("round-1")
	if err == nil {
		t.Fatal("Fire: got nil error, want error when registry has no entry")
	}
	if !strings.Contains(err.Error(), "registry") {
		t.Errorf("error: got %q, want contains 'registry'", err.Error())
	}
}
