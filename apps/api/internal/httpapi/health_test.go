package httpapi_test

import (
	"net/http"
	"testing"
)

func TestLivezReturnsNoContentWithoutAuthentication(t *testing.T) {
	server := newTestServer(t)

	rec := doJSON(server, http.MethodGet, "/livez", "", nil)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("GET /livez status: got %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestReadyzReturnsNoContentWhenDatabaseIsReachable(t *testing.T) {
	server := newTestServer(t)

	rec := doJSON(server, http.MethodGet, "/readyz", "", nil)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("GET /readyz status: got %d, want %d", rec.Code, http.StatusNoContent)
	}
}
