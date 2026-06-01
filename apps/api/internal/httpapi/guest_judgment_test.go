package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func doJSONUnauthenticated(server http.Handler, method string, path string, body any) *httptest.ResponseRecorder {
	var requestBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&requestBody).Encode(body); err != nil {
			panic(err)
		}
	}
	req := httptest.NewRequest(method, path, &requestBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	return rec
}

func TestGuestCanCreateSession(t *testing.T) {
	server := newGroupsTestServer()

	rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/guest-sessions", nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode guest session response: %v", err)
	}
	if body["id"] == "" {
		t.Fatal("expected guest session id in response")
	}
}

func TestGuestCanJudgePerformedJump(t *testing.T) {
	server := newGroupsTestServer()

	// Create a guest session
	sessionRec := doJSON(server, http.MethodPost, "/v1/guest-sessions", "", nil)
	if sessionRec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for guest session, got %d: %s", sessionRec.Code, sessionRec.Body.String())
	}
	var sessionBody map[string]string
	if err := json.NewDecoder(sessionRec.Body).Decode(&sessionBody); err != nil {
		t.Fatalf("decode guest session: %v", err)
	}
	guestSessionID := sessionBody["id"]

	// Create a group and a performed jump by Alice
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	performed := performJump(t, server, "alice-token", group.Group.ID)

	// Guest judges the jump
	rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", map[string]any{
		"guestSessionId": guestSessionID,
		"commitment":     7,
		"transgression":  8,
		"creativity":     9,
		"presentation":   10,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var judgment judgmentBody
	decodeResponse(t, rec, &judgment)
	if judgment.JumpID != performed.Jump.ID {
		t.Fatalf("expected judgment for jump %s, got %s", performed.Jump.ID, judgment.JumpID)
	}
	if judgment.PlayerID != "" {
		t.Fatalf("expected empty player id for guest judgment, got %q", judgment.PlayerID)
	}
}

func TestGuestCapBlocksAdditionalJudgments(t *testing.T) {
	server := newGroupsTestServer()

	// Create a guest session
	sessionRec := doJSON(server, http.MethodPost, "/v1/guest-sessions", "", nil)
	var sessionBody map[string]string
	if err := json.NewDecoder(sessionRec.Body).Decode(&sessionBody); err != nil {
		t.Fatalf("decode guest session: %v", err)
	}
	guestSessionID := sessionBody["id"]

	group := createGroup(t, server, "alice-token", "Breakfast Crew")

	// Guest submits 5 judgments (cap)
	for i := 0; i < 5; i++ {
		performed := performJump(t, server, "alice-token", group.Group.ID)
		rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", map[string]any{
			"guestSessionId": guestSessionID,
			"commitment":     5,
			"transgression":  5,
			"creativity":     5,
			"presentation":   5,
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201 for judgment %d, got %d: %s", i+1, rec.Code, rec.Body.String())
		}
	}

	// 6th judgment should be blocked
	performed := performJump(t, server, "alice-token", group.Group.ID)
	rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", map[string]any{
		"guestSessionId": guestSessionID,
		"commitment":     5,
		"transgression":  5,
		"creativity":     5,
		"presentation":   5,
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 for cap-exceeded judgment, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthenticatedPlayerJudgmentStillWorks(t *testing.T) {
	server := newGroupsTestServer()

	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	performed := performJump(t, server, "alice-token", group.Group.ID)

	// Bob (authenticated) judges Alice's jump
	judgment := submitJudgment(t, server, "bob-token", performed.Jump.ID, 7, 8, 9, 10, http.StatusCreated)
	if judgment.JumpID != performed.Jump.ID {
		t.Fatalf("expected judgment for jump %s, got %s", performed.Jump.ID, judgment.JumpID)
	}
	if judgment.PlayerID == "" {
		t.Fatal("expected player id on authenticated judgment")
	}
}

func TestCannotProvideBothAuthAndGuestSession(t *testing.T) {
	server := newGroupsTestServer()

	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	performed := performJump(t, server, "alice-token", group.Group.ID)

	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", "bob-token", map[string]any{
		"guestSessionId": "some-session-id",
		"commitment":     5,
		"transgression":  5,
		"creativity":     5,
		"presentation":   5,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 when both auth and guestSessionId provided, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGuestSessionIdRequiredWhenUnauthenticated(t *testing.T) {
	server := newGroupsTestServer()

	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	performed := performJump(t, server, "alice-token", group.Group.ID)

	rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", map[string]any{
		"commitment":    5,
		"transgression": 5,
		"creativity":    5,
		"presentation":  5,
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 when unauthenticated without guestSessionId, got %d: %s", rec.Code, rec.Body.String())
	}
}
