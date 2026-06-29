package bot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIClient_PublicEndpointHasNoAuthHeaders(t *testing.T) {
	api := &fakeAPI{}
	server := httptest.NewServer(api.handler())
	defer server.Close()

	client, err := NewAPIClient(APIClientConfig{
		BaseURL:      server.URL,
		AdapterToken: "dev-token",
		HTTPClient:   server.Client(),
	})
	if err != nil {
		t.Fatalf("new API client: %v", err)
	}

	resp, err := client.ListStampCatalog(context.Background())
	if err != nil {
		t.Fatalf("ListStampCatalog: %v", err)
	}
	defer resp.Body.Close()

	if got, want := len(api.received), 1; got != want {
		t.Fatalf("API calls: got %d, want %d", got, want)
	}
	call := api.received[0]
	if call.path != "/v1/stamp-catalog" {
		t.Errorf("path: got %q, want %q", call.path, "/v1/stamp-catalog")
	}
	if call.method != http.MethodGet {
		t.Errorf("method: got %q, want %q", call.method, http.MethodGet)
	}
	if call.authorization != "" {
		t.Errorf("Authorization: got %q, want empty (public endpoint)", call.authorization)
	}
	if call.actor != "" {
		t.Errorf("X-Adapter-Actor: got %q, want empty (public endpoint)", call.actor)
	}
}
