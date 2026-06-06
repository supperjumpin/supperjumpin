package httpapi_test

import (
	"net/http"
	"testing"
)

func TestCreatePerformedJump_TDD_Stress(t *testing.T) {
	server := newTestServer(t)

	aliceToken := "alice-token"

	t.Run("Successful Creation", func(t *testing.T) {
		body := map[string]string{
			"source":         "Taco Bell",
			"destination":    "Olive Garden",
			"food":           "Crunchwrap",
			"caption":        "Best jump ever",
			"mediaObjectKey": "media/123",
		}
		rec := doJSON(server, http.MethodPost, "/v1/jumps", aliceToken, body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			ID string `json:"id"`
		}
		decodeResponse(t, rec, &resp)

		if resp.ID == "" {
			t.Fatal("expected created jump to have an ID")
		}
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		body := map[string]string{
			"source":      "Taco Bell",
			"destination": "Olive Garden",
		}
		rec := doJSON(server, http.MethodPost, "/v1/jumps", aliceToken, body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for missing fields, got %d", rec.Code)
		}
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		body := map[string]string{
			"source":         "Taco Bell",
			"destination":    "Olive Garden",
			"food":           "Crunchwrap",
			"caption":        "Best jump ever",
			"mediaObjectKey": "media/123",
		}
		rec := doJSON(server, http.MethodPost, "/v1/jumps", "invalid-token", body)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 for invalid token, got %d", rec.Code)
		}
	})

	t.Run("Successful Off-Season Jump", func(t *testing.T) {
		body := map[string]string{
			"source":         "Taco Bell",
			"destination":    "Olive Garden",
			"food":           "Crunchwrap",
			"caption":        "Best jump ever",
			"mediaObjectKey": "media/456",
		}
		rec := doJSON(server, http.MethodPost, "/v1/jumps", aliceToken, body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201 for jump, got %d", rec.Code)
		}

		var resp struct {
			ID string `json:"id"`
		}
		decodeResponse(t, rec, &resp)

		if resp.ID == "" {
			t.Fatal("expected created jump to have an ID")
		}
	})
}