package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func testGetMeBootstrapsAccountAndPlayerFromSupabaseIdentity(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"valid-token": {Provider: "supabase", Subject: "auth-user-123", Email: "player@example.com"},
		},
		Store: store,
		DB:    store,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Account struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"account"`
		Player struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"player"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Account.ID == "" || body.Account.ID == "auth-user-123" {
		t.Fatalf("expected internal account id distinct from auth subject, got %q", body.Account.ID)
	}
	if body.Player.ID == "" || body.Player.ID == body.Account.ID || body.Player.ID == "auth-user-123" {
		t.Fatalf("expected separate internal player id, got account=%q player=%q", body.Account.ID, body.Player.ID)
	}
	if body.Account.Email != "player@example.com" {
		t.Fatalf("expected account email from auth identity, got %q", body.Account.Email)
	}
	if body.Player.DisplayName != "player" {
		t.Fatalf("expected display name derived from email, got %q", body.Player.DisplayName)
	}

	second := httptest.NewRecorder()
	server.ServeHTTP(second, req)
	var secondBody struct {
		Account struct {
			ID string `json:"id"`
		} `json:"account"`
		Player struct {
			ID string `json:"id"`
		} `json:"player"`
	}
	if err := json.NewDecoder(second.Body).Decode(&secondBody); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	if secondBody.Account.ID != body.Account.ID || secondBody.Player.ID != body.Player.ID {
		t.Fatalf("expected repeated auth identity to return same internal identities")
	}
}

func testGetMeRejectsMissingBearerToken(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth:  httpapi.StaticAuthVerifier{},
		Store: store,
		DB:    store,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestMe(t *testing.T) {
	t.Run("bootstraps account and player from supabase identity", testGetMeBootstrapsAccountAndPlayerFromSupabaseIdentity)
	t.Run("rejects missing bearer token", testGetMeRejectsMissingBearerToken)
}
