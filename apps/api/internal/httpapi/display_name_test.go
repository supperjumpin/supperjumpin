package httpapi_test

import (
	"net/http"
	"testing"
)

func TestPatchMeDisplayNameRejectsMissingAuth(t *testing.T) {
	server := newTestServer(t)

	rec := doJSON(server, http.MethodPatch, "/v1/me/display-name", "", nil)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchMeDisplayNameRejectsEmptyName(t *testing.T) {
	server := newTestServer(t)

	for name, body := range map[string]map[string]string{
		"empty json":    {"displayName": ""},
		"whitespace":    {"displayName": "  "},
		"missing field": {},
	} {
		t.Run(name, func(t *testing.T) {
			rec := doJSON(server, http.MethodPatch, "/v1/me/display-name", "alice-token", body)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestPatchMeDisplayNameAcceptsValidName(t *testing.T) {
	server := newTestServer(t)

	rec := doJSON(server, http.MethodPatch, "/v1/me/display-name", "alice-token", map[string]string{"displayName": "Bobby Cloutier"})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result struct {
		Player struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"player"`
	}
	decodeResponse(t, rec, &result)
	if result.Player.DisplayName != "Bobby Cloutier" {
		t.Fatalf("expected display name 'Bobby Cloutier', got %q", result.Player.DisplayName)
	}
	if result.Player.ID == "" {
		t.Fatalf("expected non-empty player ID")
	}
}

func TestPatchMeDisplayNamePersistsAndReflectsInGetMe(t *testing.T) {
	server := newTestServer(t)

	getRec := doJSON(server, http.MethodGet, "/v1/me", "alice-token", nil)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRec.Code)
	}

	var initial struct {
		Player struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"player"`
	}
	decodeResponse(t, getRec, &initial)
	if initial.Player.DisplayName != "player" {
		t.Fatalf("expected initial display name 'player' from adapter actor bootstrap, got %q", initial.Player.DisplayName)
	}

	playerID := initial.Player.ID

	patchRec := doJSON(server, http.MethodPatch, "/v1/me/display-name", "alice-token", map[string]string{"displayName": "Bobby Cloutier"})
	if patchRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on patch, got %d: %s", patchRec.Code, patchRec.Body.String())
	}

	getRec2 := doJSON(server, http.MethodGet, "/v1/me", "alice-token", nil)
	if getRec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRec2.Code)
	}

	var updated struct {
		Player struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"player"`
	}
	decodeResponse(t, getRec2, &updated)
	if updated.Player.DisplayName != "Bobby Cloutier" {
		t.Fatalf("expected 'Bobby Cloutier' after update, got %q", updated.Player.DisplayName)
	}
	if updated.Player.ID != playerID {
		t.Fatalf("expected same player ID %q, got %q", playerID, updated.Player.ID)
	}

	patchRec2 := doJSON(server, http.MethodPatch, "/v1/me/display-name", "alice-token", map[string]string{"displayName": "Bigocb"})
	if patchRec2.Code != http.StatusOK {
		t.Fatalf("expected 200 on second patch, got %d: %s", patchRec2.Code, patchRec2.Body.String())
	}

	getRec3 := doJSON(server, http.MethodGet, "/v1/me", "alice-token", nil)
	var final struct {
		Player struct {
			DisplayName string `json:"displayName"`
		} `json:"player"`
	}
	decodeResponse(t, getRec3, &final)
	if final.Player.DisplayName != "Bigocb" {
		t.Fatalf("expected 'Bigocb' after second update, got %q", final.Player.DisplayName)
	}
}
