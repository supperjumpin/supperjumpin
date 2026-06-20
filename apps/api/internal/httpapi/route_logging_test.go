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

type loggingPublicRead struct{}

func (loggingPublicRead) FeedJumps(context.Context, *time.Time, string, int) ([]JumpCard, error) {
	return nil, nil
}

func (loggingPublicRead) JumpDetail(context.Context, string) (JumpDetail, bool, error) {
	return JumpDetail{}, false, nil
}

type loggingJudgmentFlow struct {
	createGuestSessionErr error
}

func (f loggingJudgmentFlow) Jump(context.Context, string) (game.JumpSnapshot, bool, error) {
	return game.JumpSnapshot{}, false, errors.New("not implemented")
}

func (f loggingJudgmentFlow) Season(context.Context, string) (game.SeasonSnapshot, error) {
	return game.SeasonSnapshot{}, errors.New("not implemented")
}

func (f loggingJudgmentFlow) SubmitAcceptedJudgment(context.Context, game.JudgmentInput) (game.Judgment, error) {
	return game.Judgment{}, errors.New("not implemented")
}

func (f loggingJudgmentFlow) HasJudgedJump(context.Context, string, string) (bool, error) {
	return false, errors.New("not implemented")
}

func (f loggingJudgmentFlow) HasJudgedJumps(context.Context, string, []string) (map[string]bool, error) {
	return nil, errors.New("not implemented")
}

func (f loggingJudgmentFlow) HasGuestJudgedJump(context.Context, string, string) (bool, error) {
	return false, errors.New("not implemented")
}

func (f loggingJudgmentFlow) GuestSessionJudgmentCount(context.Context, string) (int, error) {
	return 0, errors.New("not implemented")
}

func (f loggingJudgmentFlow) IncrementGuestSessionJudgmentCount(context.Context, string) error {
	return errors.New("not implemented")
}

func (f loggingJudgmentFlow) CreateGuestSession(context.Context, string) error {
	return f.createGuestSessionErr
}

func (f loggingJudgmentFlow) ActiveOpen(context.Context, time.Time) (*game.OpenMonth, error) {
	return nil, nil
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
		Store: loggingStore{bootstrapErr: errors.New("SQL SELECT email, mediaObjectKey FROM jumps WHERE guest_session_id = 'guest_session_secret'")},
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
	for _, forbidden := range []string{"SELECT", "email", "mediaObjectKey", "guest_session_secret"} {
		if strings.Contains(logs.String(), forbidden) {
			t.Fatalf("log should not contain %q: %s", forbidden, logs.String())
		}
	}
	stack, ok := entry["stack"].(string)
	if !ok || !strings.Contains(stack, "runtime/debug.Stack") {
		t.Fatalf("expected stack trace, got %#v", entry["stack"])
	}
}

func TestRouteLoggingAddsGuestActorWithoutGuestSessionID(t *testing.T) {
	server, logs := newRouteLoggingTestServer(ServerConfig{
		Store:    loggingStore{},
		Judgment: loggingJudgmentFlow{},
	})

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/guest-sessions", nil))

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	entry := decodeLogEntry(t, logs)
	assertLogField(t, entry, "route", "POST /v1/guest-sessions")
	assertLogField(t, entry, "operation", "create_guest_session")
	assertLogField(t, entry, "actor_type", "guest")
	if _, ok := entry["guest_session_id"]; ok {
		t.Fatalf("guest_session_id should not be logged: %#v", entry)
	}
	if strings.Contains(logs.String(), "guest_session_") {
		t.Fatalf("guest session token should not be logged: %s", logs.String())
	}
}

func TestRouteLoggingAddsGuestActorBeforeJudgmentValidation(t *testing.T) {
	server, logs := newRouteLoggingTestServer(ServerConfig{
		Store: loggingStore{},
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/jumps/jump_123/judgment", strings.NewReader(`{"guestSessionId":"guest_session_secret"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
	entry := decodeLogEntry(t, logs)
	assertLogField(t, entry, "route", "POST /v1/jumps/{jumpID}/judgment")
	assertLogField(t, entry, "operation", "submit_judgment")
	assertLogField(t, entry, "actor_type", "guest")
	assertLogField(t, entry, "jump_id", "jump_123")
	assertLogField(t, entry, "error_code", "missing_judgment_scores")
	if strings.Contains(logs.String(), "guest_session_secret") {
		t.Fatalf("guest session token should not be logged: %s", logs.String())
	}
}

func TestRouteLoggingAddsPublicActorAndSafeDomainFields(t *testing.T) {
	server, logs := newRouteLoggingTestServer(ServerConfig{
		Store:      loggingStore{},
		PublicRead: loggingPublicRead{},
	})

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/jumps/jump_123", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", rec.Code, rec.Body.String())
	}
	entry := decodeLogEntry(t, logs)
	assertLogField(t, entry, "route", "GET /v1/jumps/{jumpID}")
	assertLogField(t, entry, "operation", "load_jump_detail")
	assertLogField(t, entry, "actor_type", "public")
	assertLogField(t, entry, "jump_id", "jump_123")
	assertLogField(t, entry, "outcome", "not_found")
	assertLogField(t, entry, "error_code", "not_found")
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
