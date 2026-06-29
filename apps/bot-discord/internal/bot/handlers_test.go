package bot

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestRoundStatusHandler_FetchesCommitAndSubmissionCounts(t *testing.T) {
	h := newTestHarness(t)

	h.api.responder = func(path string, _ []byte) (int, []byte) {
		if path == "/v1/rounds/round-1/jumps" {
			return http.StatusOK, []byte(`{"jumps":[],"commitCount":3,"submissionCount":2}`)
		}
		return http.StatusNotFound, []byte(`{"error":"not_found","message":"not found"}`)
	}

	handler := NewRoundStatusHandler(h.client, testActorResolver("discord:guild-1:user-1"))

	interaction := IncomingInteraction{
		Type:    InteractionApplicationCommand,
		GuildID: "guild-1",
		UserID:  "user-1",
		Command: CommandRoute{Name: "round", Subcommand: "status"},
		Options: map[string]string{"roundId": "round-1"},
	}

	reply, err := handler.Handle(context.Background(), interaction)
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !reply.Ephemeral {
		t.Error("reply not ephemeral")
	}
	if !strings.Contains(reply.Body, "3 committed") {
		t.Errorf("reply body: got %q, want contains '3 committed'", reply.Body)
	}
	if !strings.Contains(reply.Body, "2 submitted") {
		t.Errorf("reply body: got %q, want contains '2 submitted'", reply.Body)
	}
}

func TestCommentHandler_RoundLevelPostsToRoundEndpoint(t *testing.T) {
	h := newTestHarness(t)

	h.api.responder = func(path string, _ []byte) (int, []byte) {
		if path == "/v1/rounds/round-1/comments" {
			return http.StatusCreated, []byte(`{"comment":{"id":"comment-1","roundId":"round-1","body":"hello","createdAt":"2026-06-01T12:00:00Z"}}`)
		}
		return http.StatusNotFound, []byte(`{"error":"not_found","message":"not found"}`)
	}

	handler := NewCommentHandler(h.client, testActorResolver("discord:guild-1:user-1"))

	interaction := IncomingInteraction{
		Type:    InteractionApplicationCommand,
		GuildID: "guild-1",
		UserID:  "user-1",
		Command: CommandRoute{Name: "comment", Subcommand: "round"},
		Options: map[string]string{"roundId": "round-1", "body": "hello"},
	}

	if _, err := handler.Handle(context.Background(), interaction); err != nil {
		t.Fatalf("handle: %v", err)
	}

	if got, want := len(h.api.received), 1; got != want {
		t.Fatalf("API calls: got %d, want %d", got, want)
	}
	if h.api.received[0].path != "/v1/rounds/round-1/comments" {
		t.Errorf("path: got %q", h.api.received[0].path)
	}
}

func TestCommentHandler_JumpLevelPostsToJumpEndpoint(t *testing.T) {
	h := newTestHarness(t)

	h.api.responder = func(path string, _ []byte) (int, []byte) {
		if path == "/v1/rounds/round-1/jumps/jump-1/comments" {
			return http.StatusCreated, []byte(`{"comment":{"id":"comment-1","roundId":"round-1","jumpId":"jump-1","body":"hello","createdAt":"2026-06-01T12:00:00Z"}}`)
		}
		return http.StatusNotFound, []byte(`{"error":"not_found","message":"not found"}`)
	}

	handler := NewCommentHandler(h.client, testActorResolver("discord:guild-1:user-1"))

	interaction := IncomingInteraction{
		Type:    InteractionApplicationCommand,
		GuildID: "guild-1",
		UserID:  "user-1",
		Command: CommandRoute{Name: "comment", Subcommand: "jump"},
		Options: map[string]string{"roundId": "round-1", "jumpId": "jump-1", "body": "hello"},
	}

	if _, err := handler.Handle(context.Background(), interaction); err != nil {
		t.Fatalf("handle: %v", err)
	}

	if got, want := len(h.api.received), 1; got != want {
		t.Fatalf("API calls: got %d, want %d", got, want)
	}
	if h.api.received[0].path != "/v1/rounds/round-1/jumps/jump-1/comments" {
		t.Errorf("path: got %q", h.api.received[0].path)
	}
}

func TestRecapHandler_FetchesAndFormatsRecap(t *testing.T) {
	h := newTestHarness(t)

	h.api.responder = func(path string, _ []byte) (int, []byte) {
		if path == "/v1/rounds/round-1/recap" {
			return http.StatusOK, []byte(`{
				"jumps":[{"caption":"First","totalStamps":2}],
				"comments":[{"body":"Nice work"}]
			}`)
		}
		return http.StatusNotFound, []byte(`{"error":"not_found","message":"not found"}`)
	}

	handler := NewRecapHandler(h.client, testActorResolver("discord:guild-1:user-1"))

	interaction := IncomingInteraction{
		Type:    InteractionApplicationCommand,
		GuildID: "guild-1",
		UserID:  "user-1",
		Command: CommandRoute{Name: "recap", Subcommand: ""},
		Options: map[string]string{"roundId": "round-1"},
	}

	reply, err := handler.Handle(context.Background(), interaction)
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !strings.Contains(reply.Body, "First") {
		t.Errorf("body: got %q, want contains 'First'", reply.Body)
	}
	if !strings.Contains(reply.Body, "Nice work") {
		t.Errorf("body: got %q, want contains 'Nice work'", reply.Body)
	}
}
