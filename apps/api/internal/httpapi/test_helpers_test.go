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

func newTestServer(t *testing.T) http.Handler {
	return newTestServerWithStore(newCleanPostgresTestStore(t))
}

func newTestServerWithStore(store *httpapi.PostgresStore) http.Handler {
	return httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {Provider: "test-provider", Subject: "alice-auth", Email: "alice@example.com"},
			"bob-token":   {Provider: "test-provider", Subject: "bob-auth", Email: "bob@example.com"},
			"carol-token": {Provider: "test-provider", Subject: "carol-auth", Email: "carol@example.com"},
		},
		Store: store,
		Now:   store.Now,
	})
}
