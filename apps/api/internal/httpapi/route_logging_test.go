package httpapi

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

type loggingStore struct {
	bootstrapErr error
	packs        []game.PromptPackSnapshot
	prompts      []game.PromptSnapshot
}

func (s loggingStore) BootstrapIdentity(context.Context, AuthIdentity) (MeResponse, error) {
	if s.bootstrapErr != nil {
		return MeResponse{}, s.bootstrapErr
	}
	return MeResponse{
		Account: Account{ID: "account_alice", Email: "alice@example.com"},
		Player:  Player{ID: "player_alice", DisplayName: "alice"},
	}, nil
}

func (s loggingStore) UpdateDisplayName(context.Context, string, string) (Player, error) {
	return Player{}, errors.New("not implemented")
}

func (s loggingStore) ListPromptPacks(context.Context) ([]game.PromptPackSnapshot, error) {
	return s.packs, nil
}

func (s loggingStore) ListPrompts(context.Context) ([]game.PromptSnapshot, error) {
	return s.prompts, nil
}

func (s loggingStore) GetPrompt(_ context.Context, id string) (game.PromptSnapshot, error) {
	for _, p := range s.prompts {
		if p.ID == id {
			return p, nil
		}
	}
	return game.PromptSnapshot{}, game.ErrPromptNotFound
}

func (s loggingStore) ListRevealTimeframes(context.Context) ([]game.RevealTimeframeSnapshot, error) {
	return nil, errors.New("not implemented")
}

func (s loggingStore) GetRevealTimeframe(context.Context, string) (game.RevealTimeframeSnapshot, error) {
	return game.RevealTimeframeSnapshot{}, errors.New("not implemented")
}

func (s loggingStore) FindPlayer(context.Context, string) (game.PlayerSnapshot, bool, error) {
	return game.PlayerSnapshot{}, false, errors.New("not implemented")
}

func (s loggingStore) FindCommunity(context.Context, string) (game.CommunitySnapshot, bool, error) {
	return game.CommunitySnapshot{}, false, errors.New("not implemented")
}

func (s loggingStore) FindActiveRound(context.Context, string) (*game.RoundSnapshot, error) {
	return nil, errors.New("not implemented")
}

func (s loggingStore) CreateRound(context.Context, game.RoundSnapshot) error {
	return errors.New("not implemented")
}

func TestRouteLoggingAddsSuccessOperationAndActorFields(t *testing.T) {
	server, logs := newRouteLoggingTestServer(ServerConfig{
		Store: loggingStore{},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	req.Header.Set("User-Agent", "supperjumpin-test")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	entry := decodeLogEntry(t, logs)
	assertLogField(t, entry, "route", "GET /v1/me")
	assertLogField(t, entry, "operation", "bootstrap_identity")
	assertLogField(t, entry, "outcome", "success")
	assertLogField(t, entry, "actor_type", "player")
	assertLogField(t, entry, "player_id", "player_alice")
	assertLogField(t, entry, "user_agent", "supperjumpin-test")
}

func TestRouteLoggingAddsClientErrorClassification(t *testing.T) {
	server, logs := newRouteLoggingTestServer(ServerConfig{
		Store: loggingStore{},
	})

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/me", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
	entry := decodeLogEntry(t, logs)
	assertLogField(t, entry, "level", "INFO")
	assertLogField(t, entry, "route", "GET /v1/me")
	assertLogField(t, entry, "outcome", "forbidden")
	assertLogField(t, entry, "error_code", "missing_bearer_token")
}

func TestRouteLoggingAddsServerErrorStackAndInternalError(t *testing.T) {
	server, logs := newRouteLoggingTestServer(ServerConfig{
		Store: loggingStore{bootstrapErr: errors.New("SQL SELECT email FROM accounts WHERE secret")},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d: %s", rec.Code, rec.Body.String())
	}
	entry := decodeLogEntry(t, logs)
	assertLogField(t, entry, "level", "ERROR")
	assertLogField(t, entry, "outcome", "server_error")
	assertLogField(t, entry, "error_code", "bootstrap_identity_failed")
	assertLogField(t, entry, "internal_error", "redacted internal error")
	for _, forbidden := range []string{"SELECT", "email", "secret"} {
		if strings.Contains(logs.String(), forbidden) {
			t.Fatalf("log should not contain %q: %s", forbidden, logs.String())
		}
	}
	stack, ok := entry["stack"].(string)
	if !ok || !strings.Contains(stack, "runtime/debug.Stack") {
		t.Fatalf("expected stack trace, got %#v", entry["stack"])
	}
}

func TestRouteLoggingUpdateDisplayName(t *testing.T) {
	server, logs := newRouteLoggingTestServer(ServerConfig{
		Store: loggingStore{},
	})

	body := strings.NewReader(`{"displayName":"New Name"}`)
	req := httptest.NewRequest(http.MethodPatch, "/v1/me/display-name", body)
	req.Header.Set("Authorization", "Bearer alice-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d: %s", rec.Code, rec.Body.String())
	}
	entry := decodeLogEntry(t, logs)
	assertLogField(t, entry, "route", "PATCH /v1/me/display-name")
	assertLogField(t, entry, "operation", "update_display_name")
	assertLogField(t, entry, "error_code", "update_display_name_failed")
}

func TestRouteLoggingPromptCatalogMarksPublicActorType(t *testing.T) {
	server, logs := newRouteLoggingTestServer(ServerConfig{
		Store: loggingStore{
			packs: []game.PromptPackSnapshot{
				{ID: "pack-1", DisplayName: "Kitchen Classics"},
			},
			prompts: []game.PromptSnapshot{
				{ID: "prompt-1", PackID: "pack-1", Copy: "X", Theme: "T", CostTier: "tier_1"},
			},
		},
	})

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/prompt-catalog", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	entry := decodeLogEntry(t, logs)
	assertLogField(t, entry, "route", "GET /v1/prompt-catalog")
	assertLogField(t, entry, "operation", "list_prompt_catalog")
	assertLogField(t, entry, "outcome", "success")
	assertLogField(t, entry, "actor_type", "public")
}

func newRouteLoggingTestServer(config ServerConfig) (http.Handler, *bytes.Buffer) {
	var logs bytes.Buffer
	config.Auth = StaticAuthVerifier{
		"alice-token": {Provider: "test-provider", Subject: "alice-auth", Email: "alice@example.com"},
	}
	config.Logger = slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	config.Now = fixedClock(time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC), time.Date(2026, 6, 17, 12, 0, 0, 1500000, time.UTC))
	return NewServer(config), &logs
}

func assertLogField(t *testing.T, entry map[string]any, field string, want any) {
	t.Helper()
	if got := entry[field]; got != want {
		t.Fatalf("expected log field %s=%#v, got %#v in %#v", field, want, got, entry)
	}
}
