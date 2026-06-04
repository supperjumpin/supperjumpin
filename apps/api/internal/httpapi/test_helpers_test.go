package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

// ---------------------------------------------------------------------------
// Response body types
// ---------------------------------------------------------------------------

type jumpBody struct {
	ID          string `json:"id"`
	PlayerID    string `json:"playerId"`
	Status      string `json:"status"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Food        string `json:"food"`
}

type judgmentBody struct {
	ID            string `json:"id"`
	JumpID        string `json:"jumpId"`
	PlayerID      string `json:"playerId"`
	GuestSessionID string `json:"guestSessionId,omitempty"`
	Provenance     string `json:"provenance,omitempty"`
	Commitment    int    `json:"commitment"`
	Transgression int    `json:"transgression"`
	Creativity    int    `json:"creativity"`
	Presentation  int    `json:"presentation"`
}

// ---------------------------------------------------------------------------
// HTTP test helpers
// ---------------------------------------------------------------------------

func doJSON(server http.Handler, method string, path string, token string, body any) *httptest.ResponseRecorder {
	var requestBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&requestBody).Encode(body); err != nil {
			panic(err)
		}
	}
	req := httptest.NewRequest(method, path, &requestBody)
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	return rec
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Server constructors
// ---------------------------------------------------------------------------

func newTestServer(t *testing.T) http.Handler {
	return newTestServerWithStore(newCleanPostgresTestStore(t))
}

func newTestServerAndStore(t *testing.T) (http.Handler, *httpapi.PostgresStore) {
	store := newCleanPostgresTestStore(t)
	return newTestServerWithStore(store), store
}

func newTestServerWithStore(store *httpapi.PostgresStore) http.Handler {
	return httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {Provider: "supabase", Subject: "alice-auth", Email: "alice@example.com"},
			"bob-token":   {Provider: "supabase", Subject: "bob-auth", Email: "bob@example.com"},
			"carol-token": {Provider: "supabase", Subject: "carol-auth", Email: "carol@example.com"},
		},
		Store:        store,
		Now:          store.Now,
		JumpPlanning: store,
		Judgment:     store,
		PublicRead:   store,
		Open:         store,
	})
}

// ---------------------------------------------------------------------------
// Lifecycle helpers (minimum set for surviving tests)
// ---------------------------------------------------------------------------

func performJump(t *testing.T, server http.Handler, token string) jumpBody {
	t.Helper()
	rec := doJSON(server, http.MethodPost, "/v1/jumps", token, map[string]string{
		"source":         "Taco Bell",
		"destination":    "Olive Garden parking lot",
		"food":           "Crunchwrap",
		"caption":        "Crunchwrap successfully smuggled into the parking lot.",
		"mediaObjectKey": "evidence_object_123",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var body jumpBody
	decodeResponse(t, rec, &body)
	return body
}

func submitJudgment(t *testing.T, server http.Handler, token string, jumpID string, commitment int, transgression int, creativity int, presentation int, expectedStatus int) judgmentBody {
	t.Helper()
	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+jumpID+"/judgment", token, map[string]int{
		"commitment":    commitment,
		"transgression": transgression,
		"creativity":    creativity,
		"presentation":  presentation,
	})
	if rec.Code != expectedStatus {
		t.Fatalf("expected Judgment status %d, got %d: %s", expectedStatus, rec.Code, rec.Body.String())
	}
	var judgment judgmentBody
	decodeResponse(t, rec, &judgment)
	return judgment
}
