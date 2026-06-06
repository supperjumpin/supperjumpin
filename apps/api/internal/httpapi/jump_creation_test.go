package httpapi_test

import (
	"net/http"
	"testing"
)

func TestCreatePerformedJump_TDD_Stress(t *testing.T) {
	server := newGroupsTestServer()

	// Setup: User needs to be signed in and have a group
	aliceToken := "alice-token"
	group := createGroup(t, server, aliceToken, "Test Group")

	t.Run("Successful Creation", func(t *testing.T) {
		body := map[string]string{
			"source":         "Taco Bell",
			"destination":    "Olive Garden",
			"food":           "Crunchwrap",
			"caption":       "Best jump ever",
			"mediaObjectKey": "media/123",
			"groupId":        group.Group.ID,
		}
		rec := doJSON(server, http.MethodPost, "/v1/jumps", aliceToken, body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			Jump struct {
				ID string `json:"id"`
			} `json:"jump"`
			Evidence struct {
				ID string `json:"id"`
			} `json:"evidence"`
		}
		decodeResponse(t, rec, &resp)

		if resp.Jump.ID == "" {
			t.Fatal("expected created jump to have an ID")
		}
		if resp.Evidence.ID == "" {
			t.Fatal("expected evidence to have an ID")
		}
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		// Missing 'food'
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
			"source":      "Taco Bell",
			"destination": "Olive Garden",
			"food":        "Crunchwrap",
		}
		rec := doJSON(server, http.MethodPost, "/v1/jumps", "invalid-token", body)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 for invalid token, got %d", rec.Code)
		}
	})

	t.Run("Off-Season (No GroupID)", func(t *testing.T) {
		body := map[string]string{
			"source":      "Taco Bell",
			"destination": "Olive Garden",
			"food":        "Crunchwrap",
		}
		rec := doJSON(server, http.MethodPost, "/v1/jumps", aliceToken, body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201 for off-season jump, got %d", rec.Code)
		}

		var resp struct {
			Jump struct {
				GroupID string `json:"groupId"`
			} `json:"jump"`
		}
		decodeResponse(t, rec, &resp)

		if resp.Jump.GroupID != "" {
			t.Fatalf("expected off-season jump to have no groupId, got %q", resp.Jump.GroupID)
		}
	})
}
