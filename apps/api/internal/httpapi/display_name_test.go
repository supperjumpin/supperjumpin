package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func TestPatchMeDisplayNameRejectsMissingAuth(t *testing.T) {
	store := httpapi.NewMemoryStore()
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth:  httpapi.StaticAuthVerifier{},
		Store: store,
		DB:    store,
	})

	req := httptest.NewRequest(http.MethodPatch, "/v1/me/display-name", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchMeDisplayNameRejectsEmptyName(t *testing.T) {
	store := httpapi.NewMemoryStore()
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"valid-token": {Provider: "supabase", Subject: "auth-user-123", Email: "player@example.com"},
		},
		Store: store,
		DB:    store,
	})

	for name, body := range map[string]string{
		"empty json":   `{"displayName":""}`,
		"whitespace":   `{"displayName":"  "}`,
		"missing field": `{}`,
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/v1/me/display-name", strings.NewReader(body))
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			server.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestPatchMeDisplayNameAcceptsValidName(t *testing.T) {
	store := httpapi.NewMemoryStore()
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"valid-token": {Provider: "supabase", Subject: "auth-user-123", Email: "player@example.com"},
		},
		Store: store,
		DB:    store,
	})

	body := `{"displayName":"Bobby Cloutier"}`
	req := httptest.NewRequest(http.MethodPatch, "/v1/me/display-name", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result struct {
		Player struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"player"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Player.DisplayName != "Bobby Cloutier" {
		t.Fatalf("expected display name 'Bobby Cloutier', got %q", result.Player.DisplayName)
	}
	if result.Player.ID == "" {
		t.Fatalf("expected non-empty player ID")
	}
}

func TestPatchMeDisplayNamePersistsAndReflectsInGetMe(t *testing.T) {
	store := httpapi.NewMemoryStore()
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"valid-token": {Provider: "supabase", Subject: "auth-user-123", Email: "player@example.com"},
		},
		Store: store,
		DB:    store,
	})

	// First, verify initial display name is derived from email
	getReq := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	getReq.Header.Set("Authorization", "Bearer valid-token")
	getRec := httptest.NewRecorder()
	server.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRec.Code)
	}

	var initial struct {
		Player struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"player"`
	}
	if err := json.NewDecoder(getRec.Body).Decode(&initial); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if initial.Player.DisplayName != "player" {
		t.Fatalf("expected initial display name 'player' from email, got %q", initial.Player.DisplayName)
	}

	playerID := initial.Player.ID

	// Update to a new display name
	patchBody := `{"displayName":"Bobby Cloutier"}`
	patchReq := httptest.NewRequest(http.MethodPatch, "/v1/me/display-name", strings.NewReader(patchBody))
	patchReq.Header.Set("Authorization", "Bearer valid-token")
	patchReq.Header.Set("Content-Type", "application/json")
	patchRec := httptest.NewRecorder()
	server.ServeHTTP(patchRec, patchReq)

	if patchRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on patch, got %d: %s", patchRec.Code, patchRec.Body.String())
	}

	// Verify GET /v1/me returns the updated name
	getReq2 := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	getReq2.Header.Set("Authorization", "Bearer valid-token")
	getRec2 := httptest.NewRecorder()
	server.ServeHTTP(getRec2, getReq2)

	if getRec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRec2.Code)
	}

	var updated struct {
		Player struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"player"`
	}
	if err := json.NewDecoder(getRec2.Body).Decode(&updated); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if updated.Player.DisplayName != "Bobby Cloutier" {
		t.Fatalf("expected 'Bobby Cloutier' after update, got %q", updated.Player.DisplayName)
	}
	if updated.Player.ID != playerID {
		t.Fatalf("expected same player ID %q, got %q", playerID, updated.Player.ID)
	}

	// Update again to a different name
	patchBody2 := `{"displayName":"Bigocb"}`
	patchReq2 := httptest.NewRequest(http.MethodPatch, "/v1/me/display-name", strings.NewReader(patchBody2))
	patchReq2.Header.Set("Authorization", "Bearer valid-token")
	patchReq2.Header.Set("Content-Type", "application/json")
	patchRec2 := httptest.NewRecorder()
	server.ServeHTTP(patchRec2, patchReq2)

	if patchRec2.Code != http.StatusOK {
		t.Fatalf("expected 200 on second patch, got %d: %s", patchRec2.Code, patchRec2.Body.String())
	}

	// Verify the latest name is returned
	getReq3 := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	getReq3.Header.Set("Authorization", "Bearer valid-token")
	getRec3 := httptest.NewRecorder()
	server.ServeHTTP(getRec3, getReq3)

	var final struct {
		Player struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"player"`
	}
	if err := json.NewDecoder(getRec3.Body).Decode(&final); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if final.Player.DisplayName != "Bigocb" {
		t.Fatalf("expected 'Bigocb' after second update, got %q", final.Player.DisplayName)
	}
}
