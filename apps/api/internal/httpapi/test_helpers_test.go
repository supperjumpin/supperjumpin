package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func doJSON(server http.Handler, method string, path string, token string, body any) *httptest.ResponseRecorder {
	return doJSONAsActor(server, method, path, token, actorForToken(token), body)
}

func doJSONAsActor(server http.Handler, method string, path string, token string, actor string, body any) *httptest.ResponseRecorder {
	var requestBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&requestBody).Encode(body); err != nil {
			panic(err)
		}
	}
	req := httptest.NewRequest(method, path, &requestBody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		if actor != "" {
			req.Header.Set("X-Adapter-Actor", actor)
		}
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	return rec
}

func actorForToken(token string) string {
	switch token {
	case "alice-token":
		return "discord:test-server:alice-user"
	case "bob-token":
		return "discord:test-server:bob-user"
	case "carol-token":
		return "discord:test-server:carol-user"
	default:
		return ""
	}
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func newTestServer(t *testing.T) http.Handler {
	return newTestServerWithStore(newCleanPostgresTestStore(t))
}

func newTestServerWithStore(store *httpapi.PostgresStore) http.Handler {
	return httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {},
			"bob-token":   {},
			"carol-token": {},
		},
		Store: store,
		Now:   store.Now,
	})
}
