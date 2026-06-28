package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func testGetMeResolvesPlayerAndCommunityFromAdapterActor(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"adapter-token": {},
		},
		Store: store,
		Now:   store.Now,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer adapter-token")
	req.Header.Set("X-Adapter-Actor", "discord:guild-123:user-abc")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Player struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"player"`
		Community struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"community"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Player.ID == "" {
		t.Fatal("expected player id")
	}
	if body.Community.ID == "" {
		t.Fatal("expected community id")
	}
	if body.Player.DisplayName != "player" {
		t.Fatalf("expected default player display name on first touch, got %q", body.Player.DisplayName)
	}
	if body.Community.DisplayName != "community" {
		t.Fatalf("expected default community display name on first touch, got %q", body.Community.DisplayName)
	}

	second := httptest.NewRecorder()
	server.ServeHTTP(second, req)
	var secondBody struct {
		Player struct {
			ID string `json:"id"`
		} `json:"player"`
		Community struct {
			ID string `json:"id"`
		} `json:"community"`
	}
	if err := json.NewDecoder(second.Body).Decode(&secondBody); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	if secondBody.Player.ID != body.Player.ID || secondBody.Community.ID != body.Community.ID {
		t.Fatalf("expected repeated auth identity to return same internal identities")
	}
}

func testGetMeRejectsMissingAdapterActor(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"adapter-token": {},
		},
		Store: store,
		Now:   store.Now,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer adapter-token")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func testGetMeRejectsMissingBearerToken(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth:  httpapi.StaticAuthVerifier{},
		Store: store,
		Now:   store.Now,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func testGetMeRejectsInvalidBearerToken(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"adapter-token": {},
		},
		Store: store,
		Now:   store.Now,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestMe(t *testing.T) {
	t.Run("resolves player and community from adapter actor", testGetMeResolvesPlayerAndCommunityFromAdapterActor)
	t.Run("rejects missing adapter actor", testGetMeRejectsMissingAdapterActor)
	t.Run("rejects missing bearer token", testGetMeRejectsMissingBearerToken)
	t.Run("rejects invalid bearer token", testGetMeRejectsInvalidBearerToken)
}
